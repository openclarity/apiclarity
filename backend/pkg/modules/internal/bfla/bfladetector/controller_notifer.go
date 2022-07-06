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

package bfladetector

import (
	"context"
	"github.com/openclarity/apiclarity/api3/global"
	"github.com/openclarity/apiclarity/api3/notifications"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
)

type ControllerNotifier interface {
	Notify(ctx context.Context, apiID uint, notification AuthzModelNotification) error
}

type AuthzModelNotification struct {
	Learning   bool
	AuthzModel AuthorizationModel
	SpecType   SpecType
}

func NewControllerNotifier(accessor core.BackendAccessor) *notifier {
	return &notifier{
		accessor: accessor,
	}
}

type notifier struct {
	accessor core.BackendAccessor
}

func (n *notifier) Notify(ctx context.Context, apiID uint, notification AuthzModelNotification) error {
	ntf := notifications.APIClarityNotification{}
	if err := ntf.FromAuthorizationModelNotification(notifications.AuthorizationModelNotification{
		Learning:   notification.Learning,
		Operations: ToGlobalOperations(notification.AuthzModel.Operations),
		SpecType:   global.SpecType(ToRestapiSpecType(notification.SpecType)),
	}); err != nil {
		return err
	}
	return n.accessor.Notify(ctx, ModuleName, apiID, ntf)
}

func ToGlobalOperations(authzModelOps Operations) (ops []global.AuthorizationModelOperation) {
	for _, o := range authzModelOps {
		resOp := global.AuthorizationModelOperation{
			Method: o.Method,
			Path:   o.Path,
			Tags:   o.Tags,
		}
		for _, aud := range o.Audience {
			resAud := global.AuthorizationModelAudience{
				Authorized:    aud.Authorized,
				External:      aud.External,
				K8sObject:     (*global.K8sObjectRef)(aud.K8sObject),
				StatusCode:    int(aud.StatusCode),
				LastTime:      &aud.LastTime,
				WarningStatus: global.BFLAStatus(aud.WarningStatus),
			}
			for _, user := range aud.EndUsers {
				resAud.EndUsers = append(resAud.EndUsers, global.DetectedUser{
					Id:        user.ID,
					Source:    global.DetectedUserSource(user.Source.String()),
					IpAddress: user.IPAddress,
				})
			}
			resOp.Audience = append(resOp.Audience, resAud)
		}
		ops = append(ops, resOp)
	}

	return
}
