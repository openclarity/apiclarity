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

package clients

const (
	fuzzerContainerName      = "scn-dast"
	jobNamePrefix            = "scn-fuzzer-"
	requestScopeDefaultValue = "global/internalservices/portshift_request"
	tmpFolderPath            = "/tmp"
	tmpEmptyDirVolumeName    = "tmp-volume"
	user                     = 1001

	uriEnvVar               = "URI"
	fuzzersEnvVar           = "FUZZER"
	apiIDEnvVar             = "API_ID"
	platformHostEnvVar      = "PLATFORM_HOST"
	platformTypeEnvVar      = "PLATFORM_TYPE"
	requestScopeEnvVar      = "REQUEST_SCOPE"
	debugEnvVar             = "DEBUG"
	authEnvVar              = "SERVICE_AUTH"
	restlerRootPathEnvVar   = "RESTLER_ROOT_PATH"
	restlerTimeBudgetEnvVar = "RESTLER_TIME_BUDGET"
	authInjectorPathEnvVar  = "RESTLER_TOKEN_INJECTOR_PATH"
)
