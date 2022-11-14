// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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

package main

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
)

const (
	commonSpecPath         = "/api3/common/openapi.yaml"
	coreSpecPath           = "/api3/core/openapi.yaml"
	notificationStubPath   = "/api3/notifications/openapi.stub.yaml"
	notificationSpecPath   = "/api3/notifications/openapi.gen.yaml"
	globalSpecPath         = "/api3/global/openapi.gen.yaml"
	modulesPath            = "/backend/pkg/modules/internal"
	inModulePath           = "restapi/openapi.yaml"
	rootPath               = "../.."
	parentNotificationName = "APIClarityNotification"
	baseNotificationName   = "BaseNotification"
	featureEnumName        = "APIClarityFeatureEnum"
)

func loadSpec(path string, fail bool) *openapi3.T {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	spec, err := loader.LoadFromFile(path)
	if err != nil {
		if fail {
			log.Fatalf("Unable to load spec %s: %v", path, err)
		}
		log.Warnf("Unable to load spec %s: %v", path, err)
		return nil
	}
	log.Debugf("Spec %s successfully loaded\n", path)
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
	/* core spec is the starting point to build global spec */
	/* loading corespec and then enriching it with the module spec elements */
	coreSpec := loadSpec(rootPath+"/"+coreSpecPath, true)

	featureEnum := coreSpec.Components.Schemas[featureEnumName]
	coreComponents := reflect.ValueOf(coreSpec.Components)

	moduleSpecs := loadModuleSpecs()
	for module, moduleSpec := range moduleSpecs {
		log.Infof("Aggregating components for module %s", module)
		moduleComponents := reflect.ValueOf(moduleSpec.Components)
		featureEnum.Value.Enum = append(featureEnum.Value.Enum, module)
		for i := 0; i < coreComponents.Type().NumField(); i++ {
			fieldName := coreComponents.Type().Field(i).Name
			if fieldName == "ExtensionProps" {
				continue
			}

			log.Debugf("Merging Components/%s for module %s\n", fieldName, module)
			coreMap := coreComponents.FieldByName(fieldName)
			moduleMap := moduleComponents.FieldByName(fieldName)
			_ = coreMap
			for _, key := range moduleMap.MapKeys() {
				log.Debugf("Handling Components/%s[%s] for module %s", fieldName, key, module)
				moduleV := moduleMap.MapIndex(key)
				coreV := coreMap.MapIndex(key)
				if coreV.IsValid() {
					log.Fatalf("Key collision in %s: Components/%s[%s]\n. Components keys must be unique across modules\n", module, fieldName, key.String())

				}
				coreMap.SetMapIndex(key, moduleV)
			}
		}

		log.Infof("Generating OperationIds for module %s", module)
		for key, value := range moduleSpec.Paths {
			if value.Get != nil && value.Get.OperationID != "" {
				value.Get.OperationID = module + value.Get.OperationID
			}
			if value.Post != nil && value.Post.OperationID != "" {
				value.Post.OperationID = module + value.Post.OperationID
			}
			if value.Put != nil && value.Put.OperationID != "" {
				value.Put.OperationID = module + value.Put.OperationID
			}
			if value.Delete != nil && value.Delete.OperationID != "" {
				value.Delete.OperationID = module + value.Delete.OperationID
			}
			if value.Patch != nil && value.Patch.OperationID != "" {
				value.Patch.OperationID = module + value.Patch.OperationID
			}
			if value.Head != nil && value.Head.OperationID != "" {
				value.Head.OperationID = module + value.Head.OperationID
			}

			coreSpec.Paths["/modules/"+module+key] = value
		}
	}

	/* Sort the elements of the feature enum to force a deterministic outcome */
	sort.Slice(featureEnum.Value.Enum, func(i, j int) bool {
		return strings.Compare(string(featureEnum.Value.Enum[i].(string)), featureEnum.Value.Enum[j].(string)) < 0
	})

	/* now coreSpec holds the new global spec. We can write to file */
	specout, err := coreSpec.MarshalJSON()
	if err != nil {
		log.Fatalf("%v", err)
	}
	yamlout, err := yaml.JSONToYAML(specout)
	if err != nil {
		log.Fatalf("%v", err)
	}

	/* Reference to common folder from module folder */
	commonOldPath, _ := filepath.Rel(filepath.Dir(modulesPath+"/MODULE/"+inModulePath), commonSpecPath)
	/* Reference to common folder from global folder */
	commonNewPath, _ := filepath.Rel(filepath.Dir(globalSpecPath), commonSpecPath)
	/* Perform the replacement */
	yamlout = bytes.ReplaceAll(yamlout, []byte(commonOldPath), []byte(commonNewPath))
	err = ioutil.WriteFile(rootPath+"/"+globalSpecPath, yamlout, 0644)
	if err != nil {
		log.Fatalf("%v", err)
	}

	/* Try to load it back to verify that it is valid */
	loadSpec(rootPath+"/"+globalSpecPath, true)
}

func aggregateNotificationSpecs() {
	/**
	* This functions is meant to generate the global notification specs.
	* Check /api3/README.md for details.
	*
	* Code generation does not support oneOf and allOf when they refer to external objects (i.e. objects defined in other oapi specs).
	* To work around this issue this function will copy in the generated specs all objects that are used in the oneOf and allOf clauses
	*
	 */
	commonSpec := loadSpec(rootPath+"/"+commonSpecPath, true)
	notificationSpec := loadSpec(rootPath+"/"+notificationStubPath, true)

	// Generic notification schema which will be the parent of all notifications by adding them in the oneOf clause
	parentNotificationSchema := notificationSpec.Components.Schemas[parentNotificationName].Value

	// Add Base notification to the notification spec as this is used in allOf clause of all notification instances
	notificationSpec.Components.Schemas[baseNotificationName] = commonSpec.Components.Schemas[baseNotificationName]

	// Path of common spec from ntification spec
	commonSpecRelPathFromNotification, _ := filepath.Rel(filepath.Dir(notificationSpecPath), commonSpecPath)
	commonSpecRelPathFromModule, _ := filepath.Rel(filepath.Dir(modulesPath+"/MODULE/"+inModulePath), commonSpecPath)
	globalSpecRelPathFromNotification, _ := filepath.Rel(filepath.Dir(notificationSpecPath), globalSpecPath)

	log.Info("Building notification specs from common specs")
	// Look for notification defined in the common Specs
	for key, notificationSchema := range commonSpec.Components.Schemas {
		// Look for schemas that have allOf and BaseNotification as first element
		if notificationSchema.Value == nil || notificationSchema.Value.AllOf == nil || notificationSchema.Value.AllOf[0].Ref != "#/components/schemas/"+baseNotificationName {
			continue
		}

		// Add the notification schema locally
		notificationSpec.Components.Schemas[key] = notificationSchema

		// Add the notification schema to the parent Notification
		parentNotificationSchema.OneOf = append(parentNotificationSchema.OneOf, openapi3.NewSchemaRef("#/components/schemas/"+key, nil))
		parentNotificationSchema.Discriminator.Mapping[key] = "#/components/schemas/" + key

		// Since we are copying the scema in the notification spec
		// we have to make sure that the copied objects refers back to the common spec fo references
		for _, refSchema := range notificationSchema.Value.AllOf[1:] {
			refPath := strings.Split(refSchema.Ref, "/")
			newRef := refSchema.Value.NewRef()
			for _, innerSchema := range newRef.Value.Properties {
				if innerSchema.Ref != "" {
					innerSchema.Ref = commonSpecRelPathFromNotification + innerSchema.Ref
					continue
				}
				if innerSchema.Value.Type == "array" && innerSchema.Value.Items.Ref != "" {
					innerSchema.Value.Items.Ref = commonSpecRelPathFromNotification + innerSchema.Value.Items.Ref
					continue
				}
			}
			notificationSpec.Components.Schemas[refPath[len(refPath)-1]] = newRef
		}
	}
	moduleSpecs := loadModuleSpecs()
	for module, moduleSpec := range moduleSpecs {
		log.Infof("Building notification specs from module %s specs", module)
		for key, notificationSchema := range moduleSpec.Components.Schemas {
			// Look for schemas that have allOf and BaseNotification as first element
			if notificationSchema.Value == nil || notificationSchema.Value.AllOf == nil || notificationSchema.Value.AllOf[0].Ref != commonSpecRelPathFromModule+"#/components/schemas/BaseNotification" {
				continue
			}

			notificationSchema.Value.AllOf[0].Ref = "#/components/schemas/" + baseNotificationName

			// Add the notification schema locally
			notificationSpec.Components.Schemas[key] = notificationSchema

			// Add the notification schema to the parent Notification
			parentNotificationSchema.OneOf = append(parentNotificationSchema.OneOf, openapi3.NewSchemaRef("#/components/schemas/"+key, nil))
			parentNotificationSchema.Discriminator.Mapping[key] = "#/components/schemas/" + key

			// Since we are copying the scema in the notification spec
			// we have to make sure that the copied objects refers back to the common spec fo references
			for _, refSchema := range notificationSchema.Value.AllOf[1:] {
				refPath := strings.Split(refSchema.Ref, "/")
				newRef := refSchema.Value.NewRef()
				for _, innerSchema := range newRef.Value.Properties {
					if strings.HasPrefix(innerSchema.Ref, "#/components") {
						innerSchema.Ref = globalSpecRelPathFromNotification + innerSchema.Ref
						continue
					}
					if strings.HasPrefix(innerSchema.Ref, commonSpecRelPathFromModule+"#") {
						innerSchema.Ref = strings.Replace(innerSchema.Ref, commonSpecRelPathFromModule+"#", commonSpecRelPathFromNotification+"#", 1)
						continue
					}
					if innerSchema.Value.Type == "array" && strings.HasPrefix(innerSchema.Value.Items.Ref, "#/components") {
						innerSchema.Value.Items.Ref = globalSpecRelPathFromNotification + innerSchema.Value.Items.Ref
						continue
					}
					if innerSchema.Value.Type == "array" && strings.HasPrefix(innerSchema.Value.Items.Ref, commonSpecRelPathFromModule+"#") {
						innerSchema.Value.Items.Ref = strings.Replace(innerSchema.Value.Items.Ref, commonSpecRelPathFromModule+"#", commonSpecRelPathFromNotification+"#", 1)

						continue
					}
				}
				notificationSpec.Components.Schemas[refPath[len(refPath)-1]] = newRef
			}
		}
	}

	/* Sort the elements of oneOf to force a deterministic outcome */
	sort.Slice(parentNotificationSchema.OneOf, func(i, j int) bool {
		return strings.Compare(parentNotificationSchema.OneOf[i].Ref, parentNotificationSchema.OneOf[j].Ref) < 0
	})

	specout, err := notificationSpec.MarshalJSON()
	if err != nil {
		log.Fatalf("%v", err)
	}
	yamlout, err := yaml.JSONToYAML(specout)
	if err != nil {
		log.Fatalf("%v", err)
	}
	err = ioutil.WriteFile(rootPath+"/"+notificationSpecPath, yamlout, 0644)
	if err != nil {
		log.Fatalf("%v", err)
	}

	loadSpec(rootPath+"/"+notificationSpecPath, true)
}
