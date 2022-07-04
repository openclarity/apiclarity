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

package fuzzer

import (
	"fmt"
)

/*
*  Error used when an object does not exists on the system.
 */
type NotFoundError struct {
	msg string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("Not found error: %v", e.msg)
}

/*
*  Error used when action is impossible because nor enough parameters of invalid parameter.
 */
type InvalidParameterError struct {
	msg string
}

func (e *InvalidParameterError) Error() string {
	return fmt.Sprintf("Invalid parameter error: %v", e.msg)
}

/*
*  General error.
 */
type PluginError struct {
	msg string
}

func (e *PluginError) Error() string {
	return fmt.Sprintf("General Fuzzer error: %v", e.msg)
}

/*
*  Error for an usupported action/parameter.
 */
type NotSupportedError struct {
	msg string
}

func (e *NotSupportedError) Error() string {
	return fmt.Sprintf("Not supported action or parameter: %v", e.msg)
}
