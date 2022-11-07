// Copyright © 2021 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package backend

import (
	"context"
	"fmt"
	"mime"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-openapi/strfmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/backend/pkg/backend/speculatoraccessor"
	_config "github.com/openclarity/apiclarity/backend/pkg/config"
	_database "github.com/openclarity/apiclarity/backend/pkg/database"
	"github.com/openclarity/apiclarity/backend/pkg/healthz"
	"github.com/openclarity/apiclarity/backend/pkg/k8smonitor"
	"github.com/openclarity/apiclarity/backend/pkg/modules"
	_notifier "github.com/openclarity/apiclarity/backend/pkg/notifier"
	"github.com/openclarity/apiclarity/backend/pkg/rest"
	"github.com/openclarity/apiclarity/backend/pkg/traces"
	speculatorutils "github.com/openclarity/apiclarity/backend/pkg/utils/speculator"
	tls "github.com/openclarity/apiclarity/backend/pkg/utils/tls"
	pluginsmodels "github.com/openclarity/apiclarity/plugins/api/server/models"
	_spec "github.com/openclarity/speculator/pkg/spec"
	_speculator "github.com/openclarity/speculator/pkg/speculator"
	_mimeutils "github.com/openclarity/speculator/pkg/utils"
	"github.com/openclarity/trace-sampling-manager/manager/pkg/manager"
	interfacemanager "github.com/openclarity/trace-sampling-manager/manager/pkg/manager/interface"
	restmanager "github.com/openclarity/trace-sampling-manager/manager/pkg/rest"
)

type Backend struct {
	speculator          *_speculator.Speculator
	stateBackupInterval time.Duration
	stateBackupFileName string
	monitor             *k8smonitor.Monitor
	apiInventoryLock    sync.RWMutex
	dbHandler           _database.Database
	modulesManager      modules.ModulesManager
	notifier            *_notifier.Notifier
}

func CreateBackend(config *_config.Config, monitor *k8smonitor.Monitor, speculator *_speculator.Speculator, dbHandler *_database.Handler, modulesManager modules.ModulesManager, notifier *_notifier.Notifier) *Backend {
	return &Backend{
		speculator:          speculator,
		stateBackupInterval: time.Second * time.Duration(config.StateBackupIntervalSec),
		stateBackupFileName: config.StateBackupFileName,
		monitor:             monitor,
		dbHandler:           dbHandler,
		modulesManager:      modulesManager,
		notifier:            notifier,
	}
}

func createDatabaseConfig(config *_config.Config) *_database.DBConfig {
	return &_database.DBConfig{
		DriverType:     config.DatabaseDriver,
		EnableInfoLogs: config.EnableDBInfoLogs,
		DBPassword:     config.DBPassword,
		DBUser:         config.DBUser,
		DBHost:         config.DBHost,
		DBPort:         config.DBPort,
		DBName:         config.DBName,
	}
}

const defaultChanSize = 100

func getCoreFeatures() []modules.ModuleInfo {
	features := []modules.ModuleInfo{
		{
			Name:        string(models.APIClarityFeatureEnumSpecreconstructor),
			Description: "Reconstructs OAPI specifications from traces",
		},
	}
	return features
}

func Run() {
	config, err := _config.LoadConfig()
	if err != nil {
		log.Errorf("Failed to load config: %v", err)
		return
	}
	errChan := make(chan struct{}, defaultChanSize)

	healthServer := healthz.NewHealthServer(config.HealthCheckAddress)
	healthServer.Start(errChan)

	healthServer.SetIsReady(false)

	globalCtx, globalCancel := context.WithCancel(context.Background())
	defer globalCancel()

	log.Info("APIClarity backend is running")

	dbConfig := createDatabaseConfig(config)
	dbHandler := _database.Init(dbConfig)
	dbHandler.StartReviewTableCleaner(globalCtx, time.Duration(config.DatabaseCleanerIntervalSec)*time.Second)
	var clientset kubernetes.Interface
	var monitor *k8smonitor.Monitor
	var samplingManager *manager.Manager
	if config.EnableK8s {
		if config.K8sLocal {
			clientset, err = k8smonitor.CreateLocalK8sClientset()
			if err != nil {
				log.Fatalf("failed to create K8s clientset: %v", err)
			}
		} else {
			clientset, err = k8smonitor.CreateK8sClientset()
			if err != nil {
				log.Fatalf("failed to create K8s clientset: %v", err)
			}
		}

		samplingManager, err = manager.Create(clientset, &restmanager.Config{
			RestServerPort:             config.HTTPTraceSamplingManagerPort,
			GRPCServerPort:             config.GRPCTraceSamplingManagerPort,
			HostToTraceSecretName:      config.HostToTraceSecretName,
			HostToTraceSecretNamespace: config.HostToTraceSecretNamespace,
			HostToTraceSecretOwnerName: config.HostToTraceSecretOwnerName,
			EnableTLS:                  config.EnableTLS,
			TLSServerCertFilePath:      config.TLSServerCertFilePath,
			TLSServerKeyFilePath:       config.TLSServerKeyFilePath,
			RootCertFilePath:           config.RootCertFilePath,
			RestServerTLSPort:          config.HTTPSTraceSamplingManagerPort,
		})
		if err != nil {
			log.Errorf("Failed to create a trace sampling manager: %v", err)
			return
		}
		if err := samplingManager.Start(errChan); err != nil {
			log.Errorf("Failed to start trace sampling manager: %v", err)
			return
		}

		if !config.K8sLocal && !viper.GetBool(_config.NoMonitorEnvVar) {
			monitor, err = k8smonitor.CreateMonitor(clientset)
			if err != nil {
				log.Errorf("Failed to create a monitor: %v", err)
				return
			}
			monitor.Start()
			defer monitor.Stop()
		}
	}

	if viper.GetBool(_database.FakeTracesEnvVar) || viper.GetBool(_database.FakeDataEnvVar) {
		go dbHandler.CreateFakeData()
	}

	speculator, err := _speculator.DecodeState(config.StateBackupFileName, config.SpeculatorConfig)
	if err != nil {
		log.Infof("No speculator state to decode, creating new: %v", err)
		speculator = _speculator.CreateSpeculator(config.SpeculatorConfig)
	} else {
		log.Infof("Using encoded speculator state")
	}

	var notifier *_notifier.Notifier
	if config.NotificationPrefix != "" {
		tlsOptions, err := tls.CreateClientTLSOptions(config)
		if err != nil {
			log.Errorf("failed to create client tls options: %v", err)
			return
		}

		notifier = _notifier.NewNotifier(config.NotificationPrefix, _notifier.NotificationMaxQueueSize, _notifier.NotificationWorkers, tlsOptions)
		notifier.Start(context.Background())
	}

	modulesWrapper, modInfos, err := modules.New(globalCtx, dbHandler, clientset, samplingManager, speculatoraccessor.NewSpeculatorAccessor(speculator), notifier, config)
	if err != nil {
		log.Errorf("Failed to create module wrapper and info: %v", err)
		return
	}

	features := append(modInfos, getCoreFeatures()...)
	if config.EnableK8s && !config.TraceSamplingEnabled {
		for _, f := range features {
			samplingManager.AddHostsToTrace(&interfacemanager.HostsByComponentID{
				Hosts:       []string{"*"},
				ComponentID: f.Name,
			})
		}
	}

	backend := CreateBackend(config, monitor, speculator, dbHandler, modulesWrapper, notifier)

	serverConfig := &rest.ServerConfig{
		EnableTLS:             config.EnableTLS,
		Port:                  config.BackendRestPort,
		TLSPort:               config.BackendRestTLSPort,
		TLSServerCertFilePath: config.TLSServerCertFilePath,
		TLSServerKeyFilePath:  config.TLSServerKeyFilePath,
		Speculator:            speculator,
		DBHandler:             dbHandler,
		ModulesManager:        modulesWrapper,
		SamplingManager:       samplingManager,
		Features:              features,
		Notifier:              notifier,
	}
	restServer, err := rest.CreateRESTServer(serverConfig)
	if err != nil {
		log.Fatalf("Failed to create REST server: %v", err)
	}
	restServer.Start(errChan)
	defer restServer.Stop()

	httpTracesServerConfig := &traces.HTTPTracesServerConfig{
		EnableTLS:             config.EnableTLS,
		Port:                  config.HTTPTracesPort,
		TLSPort:               config.HTTPTracesTLSPort,
		TLSServerCertFilePath: config.TLSServerCertFilePath,
		TLSServerKeyFilePath:  config.TLSServerKeyFilePath,
		TraceHandleFunc:       backend.handleHTTPTrace,
	}
	tracesServer, err := traces.CreateHTTPTracesServer(httpTracesServerConfig)
	if err != nil {
		log.Fatalf("Failed to create trace server: %v", err)
	}
	tracesServer.Start(errChan)
	defer tracesServer.Stop()

	backend.startStateBackup(globalCtx)

	healthServer.SetIsReady(true)
	log.Info("APIClarity backend is ready")

	if viper.GetBool(_database.FakeTracesEnvVar) {
		go backend.startSendingFakeTraces()
	}

	// Wait for deactivation
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-errChan:
		log.Errorf("Received an error - shutting down")
	case s := <-sig:
		log.Warningf("Received a termination signal: %v", s)
	}
}

func (b *Backend) handleHTTPTrace(ctx context.Context, trace *pluginsmodels.Telemetry) error {
	var err error

	log.Debugf("Handling telemetry: %+v", trace)

	// TODO: Selective tracing for spec diffs and spec reconstruction

	// get host name first from headers if not exist
	if trace.Request.Host == "" {
		headers := convertHeadersToMap(trace.Request.Common.Headers)
		if host, ok := headers["host"]; ok {
			trace.Request.Host = host
		}
	}

	trace.Request.Host, err = getHostname(trace.Request.Host)
	if err != nil {
		return fmt.Errorf("failed to get hostname from host: %v", err)
	}

	// we need to convert the trace to speculator trace format in order to call speculator methods on that trace.
	// from here on, we work only with speculator telemetry
	telemetry := speculatorutils.ConvertModelsToSpeculatorTelemetry(trace)

	destInfo, err := _speculator.GetAddressInfoFromAddress(telemetry.DestinationAddress)
	if err != nil {
		return fmt.Errorf("failed to get destination info: %v", err)
	}
	destPort, err := strconv.Atoi(destInfo.Port)
	if err != nil {
		return fmt.Errorf("failed to convert destination port: %v", err)
	}
	srcInfo, err := _speculator.GetAddressInfoFromAddress(telemetry.SourceAddress)
	if err != nil {
		return fmt.Errorf("failed to get source info: %v", err)
	}
	specKey := _speculator.GetSpecKey(telemetry.Request.Host, destInfo.Port)

	// Initialize API info
	apiInfo := _database.APIInfo{
		Name: telemetry.Request.Host,
		Port: int64(destPort),

		DestinationNamespace: trace.DestinationNamespace,
	}

	// Set API Info type
	if b.monitor.IsInternalCIDR(destInfo.IP) {
		apiInfo.Type = models.APITypeINTERNAL
	} else {
		apiInfo.Type = models.APITypeEXTERNAL
	}

	isNonAPI := isNonAPI(telemetry)

	// Don't link non APIs to an API in the inventory
	if !isNonAPI {
		// lock the API inventory to avoid creating API entries twice on trace handling races
		b.apiInventoryLock.Lock()
		created, err := b.dbHandler.APIInventoryTable().FirstOrCreate(&apiInfo)
		if err != nil {
			b.apiInventoryLock.Unlock()
			return fmt.Errorf("failed to get or create API info: %v", err)
		}
		b.apiInventoryLock.Unlock()
		log.Infof("API Info in DB: %+v", apiInfo)
		if created {
			log.Infof("Sending notification for new created API %+v", apiInfo)
		}

		// Handle trace telemetry by Speculator
		if !b.speculator.HasApprovedSpec(specKey) {
			err := b.speculator.LearnTelemetry(telemetry)
			if err != nil {
				return fmt.Errorf("failed to learn telemetry: %v", err)
			}
		}
	}

	path, query := _spec.GetPathAndQuery(telemetry.Request.Path)

	var providedPathID string
	var reconstructedPathID string
	if b.speculator.HasProvidedSpec(specKey) {
		providedPathID, err = b.speculator.GetPathID(specKey, path, _spec.SpecSourceProvided)
		if err != nil {
			return fmt.Errorf("failed to get path id of provided spec: %v", err)
		}
	}
	if b.speculator.HasApprovedSpec(specKey) {
		reconstructedPathID, err = b.speculator.GetPathID(specKey, path, _spec.SpecSourceReconstructed)
		if err != nil {
			return fmt.Errorf("failed to get path id of reconstructed spec: %v", err)
		}
	}

	// Update API event in DB
	statusCode, err := strconv.Atoi(telemetry.Response.StatusCode)
	if err != nil {
		return fmt.Errorf("failed to convert status code: %v", err)
	}

	event := &_database.APIEvent{
		APIInfoID:           apiInfo.ID,
		Time:                strfmt.DateTime(time.Now().UTC()),
		Method:              models.HTTPMethod(telemetry.Request.Method),
		RequestTime:         strfmt.DateTime(time.UnixMilli(trace.Request.Common.Time).UTC()),
		Path:                path,
		Query:               query,
		StatusCode:          int64(statusCode),
		SourceIP:            srcInfo.IP,
		DestinationIP:       destInfo.IP,
		DestinationPort:     int64(destPort),
		HostSpecName:        telemetry.Request.Host,
		IsNonAPI:            isNonAPI,
		EventType:           apiInfo.Type,
		ProvidedPathID:      providedPathID,
		ReconstructedPathID: reconstructedPathID,
	}

	b.dbHandler.APIEventsTable().CreateAPIEvent(event)

	b.modulesManager.EventNotify(ctx, &modules.Event{APIEvent: event, Telemetry: trace})

	return nil
}

func convertHeadersToMap(headers []*pluginsmodels.Header) map[string]string {
	ret := make(map[string]string)
	for _, header := range headers {
		ret[header.Key] = header.Value
	}
	return ret
}

// getHostname will return only hostname without scheme and port
// ex. https://example.org:8000 --> example.org.
func getHostname(host string) (string, error) {
	if !strings.Contains(host, "://") {
		// need to add scheme to host in order for url.Parse to parse properly
		host = "http://" + host
	}

	parsedHost, err := url.Parse(host)
	if err != nil {
		return "", fmt.Errorf("failed to parse host. host=%v: %v", host, err)
	}

	if parsedHost.Hostname() == "" {
		return "", fmt.Errorf("hostname is empty. host=%v", host)
	}

	return parsedHost.Hostname(), nil
}

const (
	contentTypeHeaderName      = "content-type"
	contentTypeApplicationJSON = "application/json"
)

func isNonAPI(telemetry *_spec.Telemetry) bool {
	respHeaders := _spec.ConvertHeadersToMap(telemetry.Response.Common.Headers)

	// If response content-type header is missing, we will classify it as API
	respContentType, ok := respHeaders[contentTypeHeaderName]
	if !ok {
		return false
	}

	mediaType, _, err := mime.ParseMediaType(respContentType)
	if err != nil {
		log.Errorf("Failed to parse response media type - classifying as non-API. Content-Type=%v: %v", respContentType, err)
		return true
	}

	// If response content-type is not application/json, need to classify the telemetry as non-api
	return !_mimeutils.IsApplicationJSONMediaType(mediaType)
}

func (b *Backend) startStateBackup(ctx context.Context) {
	go func() {
		stateBackupInterval := b.stateBackupInterval
		for {
			select {
			case <-ctx.Done():
				log.Debugf("Stopping state backup")
				return
			case <-time.After(stateBackupInterval):
				if err := b.speculator.EncodeState(b.stateBackupFileName); err != nil {
					log.Errorf("Failed to encode state: %v", err)
				}
			}
		}
	}()
}
