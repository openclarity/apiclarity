package bfladetector

import (
	"testing"

	"github.com/go-openapi/spec"
)

func TestDetectedUser_IsMismatchedScopes(t *testing.T) {
	type fields struct {
		scope string
	}
	type args struct {
		security []map[string][]string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{{
		name: "success",
		fields: fields{
			scope: "pets:read pets:write",
		},
		args: args{
			security: []map[string][]string{{
				"oauth": {"pets:read", "pets:write"},
			}},
		},
		want: false,
	}, {
		name: "missing scope",
		fields: fields{
			scope: "pets:read",
		},
		args: args{
			security: []map[string][]string{{
				"oauth": {"pets:read", "pets:write"},
			}},
		},
		want: true,
	}, {
		name: "bad space",
		fields: fields{
			scope: "pets:read   pets:write",
		},
		args: args{
			security: []map[string][]string{{
				"oauth": {"pets:read", "pets:write"},
			}},
		},
		want: false,
	}, {
		name:   "no scope",
		fields: fields{},
		args: args{
			security: []map[string][]string{{
				"oauth": {"pets:read", "pets:write"},
			}},
		},
		want: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &DetectedUser{
				Source:    DetectedUserSourceJWT,
				ID:        "123",
				IPAddress: "0.0.0.0",
				JWTClaims: &JWTClaimsWithScopes{},
			}
			if tt.fields.scope != "" {
				u.JWTClaims.Scope = &tt.fields.scope
			}
			if got := u.IsMismatchedScopes(&spec.Operation{
				OperationProps: spec.OperationProps{Security: tt.args.security},
			}); got != tt.want {
				t.Errorf("IsMismatchedScopes() = %v, want %v", got, tt.want)
			}
		})
	}
}
