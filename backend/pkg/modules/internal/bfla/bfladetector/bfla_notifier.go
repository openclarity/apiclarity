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

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/api3/global"
	"github.com/openclarity/apiclarity/api3/notifications"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
)

type BFLANotifier interface {
	NotifyAuthzModel(ctx context.Context, apiID uint, notification AuthzModelNotification) error
	NotifyFindings(ctx context.Context, apiID uint, notification notifications.ApiFindingsNotification) error
}

type AuthzModelNotification struct {
	Learning   bool
	AuthzModel AuthorizationModel
	SpecType   SpecType
}

func NewBFLANotifier(moduleName string, accessor core.BackendAccessor) *Notifier {
	return &Notifier{
		accessor:   accessor,
		moduleName: moduleName,
	}
}

type Notifier struct {
	accessor   core.BackendAccessor
	moduleName string
}

func (n *Notifier) NotifyAuthzModel(ctx context.Context, apiID uint, notification AuthzModelNotification) error {
	ntf := notifications.APIClarityNotification{}
	if err := ntf.FromAuthorizationModelNotification(notifications.AuthorizationModelNotification{
		Learning:   notification.Learning,
		Operations: ToGlobalOperations(notification.AuthzModel.Operations),
		SpecType:   global.SpecType(ToRestapiSpecType(notification.SpecType)),
	}); err != nil {
		return err //nolint:wrapcheck
	}
	if log.IsLevelEnabled(log.DebugLevel) {
		jntf, err := ntf.MarshalJSON()
		if err == nil {
			log.Debugf("Auth model notification: %s", jntf)
		}
	}
	return n.accessor.Notify(ctx, n.moduleName, apiID, ntf) //nolint:wrapcheck
}

func (n *Notifier) NotifyFindings(ctx context.Context, apiID uint, notification notifications.ApiFindingsNotification) error {
	ntf := notifications.APIClarityNotification{}
	if err := ntf.FromApiFindingsNotification(notification); err != nil {
		return err //nolint:wrapcheck
	}
	if log.IsLevelEnabled(log.DebugLevel) {
		jntf, err := ntf.MarshalJSON()
		if err == nil {
			log.Debugf("Finding notification: %s", jntf)
		}
	}
	return n.accessor.Notify(ctx, n.moduleName, apiID, ntf) //nolint:wrapcheck
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
