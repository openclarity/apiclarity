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
		resOp := global.AuthorizationModelOperation{Method: o.Method, Path: o.Path}
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
