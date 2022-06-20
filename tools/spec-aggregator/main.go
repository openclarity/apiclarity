package main

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
)

var commonSpecPath = "/api3/common/openapi.yaml"
var coreSpecPath = "/api3/core/openapi.yaml"
var notificationStubPath = "/api3/notifications/openapi.stub.yaml"
var notificationSpecPath = "/api3/notifications/openapi.gen.yaml"
var globalSpecPath = "/api3/global/openapi.gen.yaml"
var modulesPath = "/backend/pkg/modules/internal"
var inModulePath = "restapi/openapi.yaml"
var rootPath = "../.."

func loadSpec(path string, fail bool) *openapi3.T {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	spec, err := loader.LoadFromFile(path)
	if err != nil {
		log.Errorf("Unable to load spec %s: %v", path, err)
		if fail {
			log.Fatal("Fatal Error!")
		}
		return nil
	}
	log.Debugf("Spec %s succesfully loaded\n", path)
	return spec
}

func loadModuleSpecs() map[string]*openapi3.T {
	modules, err := ioutil.ReadDir(rootPath + modulesPath)
	if err != nil {
		log.Fatal(err)
	}

	moduleSpecPaths := map[string]*openapi3.T{}

	for _, m := range modules {
		moduleSpecPath := rootPath + modulesPath + "/" + m.Name() + "/" + inModulePath
		spec := loadSpec(moduleSpecPath, false)
		if spec != nil {
			moduleSpecPaths[m.Name()] = spec
		}
	}
	return moduleSpecPaths
}

func main() {
	aggregateGlobalSpecs()
	aggregateNotificationSpecs()
}

func aggregateGlobalSpecs() {
	coreSpec := loadSpec(rootPath+"/"+coreSpecPath, true)

	moduleSpecs := loadModuleSpecs()
	for module, moduleSpec := range moduleSpecs {
		coreComponents := reflect.ValueOf(coreSpec.Components)
		moduleComponents := reflect.ValueOf(moduleSpec.Components)
		for i := 0; i < coreComponents.Type().NumField(); i++ {
			fieldName := coreComponents.Type().Field(i).Name
			if fieldName == "ExtensionProps" {
				continue
			}

			log.Infof("Merging Components/%s\n", fieldName)
			coreMap := coreComponents.FieldByName(fieldName)
			moduleMap := moduleComponents.FieldByName(fieldName)
			_ = coreMap
			for _, k := range moduleMap.MapKeys() {
				log.Infof("---- %s[%s] ", fieldName, k)
				moduleV := moduleMap.MapIndex(k)
				coreV := coreMap.MapIndex(k)
				if coreV.IsValid() {
					log.Fatalf("Key collision in %s: %s[%s]\n", module, fieldName, k.String())

				}
				coreMap.SetMapIndex(k, moduleV)

			}
		}

		for k, v := range moduleSpec.Paths {
			if v.Get != nil && v.Get.OperationID != "" {
				v.Get.OperationID = module + v.Get.OperationID
			}
			if v.Post != nil && v.Post.OperationID != "" {
				v.Post.OperationID = module + v.Post.OperationID
			}
			if v.Put != nil && v.Put.OperationID != "" {
				v.Put.OperationID = module + v.Put.OperationID
			}
			if v.Delete != nil && v.Delete.OperationID != "" {
				v.Delete.OperationID = module + v.Delete.OperationID
			}
			if v.Patch != nil && v.Patch.OperationID != "" {
				v.Patch.OperationID = module + v.Patch.OperationID
			}
			if v.Head != nil && v.Head.OperationID != "" {
				v.Head.OperationID = module + v.Head.OperationID
			}

			coreSpec.Paths["/modules/"+module+k] = v
		}
	}
	specout, err := coreSpec.MarshalJSON()
	if err != nil {
		log.Errorf("ERROR: %v", err)
	}
	yamlout, err := yaml.JSONToYAML(specout)
	if err != nil {
		log.Errorf("ERROR: %v", err)
	}
	yamlout = bytes.ReplaceAll(yamlout, []byte("../../../../../../api3/common/"), []byte("../common/"))
	err = ioutil.WriteFile(rootPath+"/"+globalSpecPath, yamlout, 0644)
	if err != nil {
		log.Errorf("ERROR: %v", err)
	}

	loadSpec(rootPath+"/"+globalSpecPath, true)
	log.Info("SUCCESS!")
}

func aggregateNotificationSpecs() {
	commonSpec := loadSpec(rootPath+"/"+commonSpecPath, true)
	notificationSpec := loadSpec(rootPath+"/"+notificationStubPath, true)

	// Add Base notification to the notification spec
	parentNotificationSchema := notificationSpec.Components.Schemas["APIClarityNotification"].Value
	notificationSpec.Components.Schemas["BaseNotification"] = commonSpec.Components.Schemas["BaseNotification"]

	for k, notificationSchema := range commonSpec.Components.Schemas {
		// Look for schemas that have allOf and BaseNotification as first element
		if notificationSchema.Value == nil || notificationSchema.Value.AllOf == nil || notificationSchema.Value.AllOf[0].Ref != "#/components/schemas/BaseNotification" {
			continue
		}

		notificationSpec.Components.Schemas[k] = notificationSchema
		parentNotificationSchema.OneOf = append(parentNotificationSchema.OneOf, openapi3.NewSchemaRef("#/components/schemas/"+k, nil))
		parentNotificationSchema.Discriminator.Mapping[k] = "#/components/schemas/" + k

		for _, refSchema := range notificationSchema.Value.AllOf[1:] {
			refPath := strings.Split(refSchema.Ref, "/")
			newRef := refSchema.Value.NewRef()
			for _, innerSchema := range newRef.Value.Properties {
				if innerSchema.Ref != "" {
					innerSchema.Ref = "../common/openapi.yaml" + innerSchema.Ref
					continue
				}
				if innerSchema.Value.Type == "array" && innerSchema.Value.Items.Ref != "" {
					innerSchema.Value.Items.Ref = "../common/openapi.yaml" + innerSchema.Value.Items.Ref
					continue
				}
			}
			notificationSpec.Components.Schemas[refPath[len(refPath)-1]] = newRef
		}
	}
	moduleSpecs := loadModuleSpecs()
	for _, moduleSpec := range moduleSpecs {

		for k, notificationSchema := range moduleSpec.Components.Schemas {
			// Look for schemas that have allOf and BaseNotification as first element
			if notificationSchema.Value == nil || notificationSchema.Value.AllOf == nil || notificationSchema.Value.AllOf[0].Ref != "../../../../../../api3/common/openapi.yaml#/components/schemas/BaseNotification" {
				continue
			}

			notificationSchema.Value.AllOf[0].Ref = "#/components/schemas/BaseNotification"

			notificationSpec.Components.Schemas[k] = notificationSchema
			parentNotificationSchema.OneOf = append(parentNotificationSchema.OneOf, openapi3.NewSchemaRef("#/components/schemas/"+k, nil))
			parentNotificationSchema.Discriminator.Mapping[k] = "#/components/schemas/" + k

			for _, refSchema := range notificationSchema.Value.AllOf[1:] {
				refPath := strings.Split(refSchema.Ref, "/")
				newRef := refSchema.Value.NewRef()
				for _, innerSchema := range newRef.Value.Properties {
					if strings.HasPrefix(innerSchema.Ref, "#/components") {
						innerSchema.Ref = "../global/openapi.gen.yaml" + innerSchema.Ref
						continue
					}
					if strings.HasPrefix(innerSchema.Ref, "../../../../../../api3/common/openapi.yaml#") {
						innerSchema.Ref = strings.Replace(innerSchema.Ref, "../../../../../../api3/common/openapi.yaml#", "../common/openapi.yaml#", 1)
						continue
					}
					if innerSchema.Value.Type == "array" && strings.HasPrefix(innerSchema.Value.Items.Ref, "#/components") {
						innerSchema.Value.Items.Ref = "../global/openapi.gen.yaml" + innerSchema.Value.Items.Ref
						continue
					}
					if innerSchema.Value.Type == "array" && strings.HasPrefix(innerSchema.Value.Items.Ref, "../../../../../../api3/common/openapi.yaml") {
						innerSchema.Value.Items.Ref = strings.Replace(innerSchema.Value.Items.Ref, "../../../../../../api3/common/openapi.yaml#", "../common/openapi.yaml#", 1)

						continue
					}
				}
				notificationSpec.Components.Schemas[refPath[len(refPath)-1]] = newRef
			}
		}
	}
	a := parentNotificationSchema.OneOf[0].Ref
	_ = a
	sort.Slice(parentNotificationSchema.OneOf, func(i, j int) bool {
		return strings.Compare(parentNotificationSchema.OneOf[i].Ref, parentNotificationSchema.OneOf[j].Ref) < 0
	})

	specout, err := notificationSpec.MarshalJSON()
	if err != nil {
		log.Errorf("ERROR: %v", err)
	}
	yamlout, err := yaml.JSONToYAML(specout)
	if err != nil {
		log.Errorf("ERROR: %v", err)
	}
	err = ioutil.WriteFile(rootPath+"/"+notificationSpecPath, yamlout, 0644)
	if err != nil {
		log.Errorf("ERROR: %v", err)
	}

	loadSpec(rootPath+"/"+notificationSpecPath, true)
	log.Info("SUCCESS!")
}
