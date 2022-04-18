package bfladetector

import (
	"reflect"
	"testing"
)

func TestGetUserID(t *testing.T) {
	type args struct {
		headers map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    *DetectedUser
		wantErr bool
	}{{
		name: "success jwt",
		args: args{
			headers: map[string]string{
				"authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0MCJ9.Go08qgDIwwiCvcWQ9wA2O2-G4urRxGIbvRKGMRu5uyw",
			},
		},
		want: &DetectedUser{
			Source: DetectedUserSourceJWT,
			ID:     "test0",
		},
		wantErr: false,
	}, {
		name: "success kong x-customer-id",
		args: args{
			headers: map[string]string{
				"x-customer-id": "test1",
			},
		},
		want: &DetectedUser{
			Source: DetectedUserSourceXConsumerIDHeader,
			ID:     "test1",
		},
		wantErr: false,
	}, {
		name: "success basic",
		args: args{
			headers: map[string]string{
				"authorization": "Basic dGVzdDI6cGFzczEK",
			},
		},
		want: &DetectedUser{
			Source: DetectedUserSourceBasic,
			ID:     "test2",
		},
		wantErr: false,
	}, {
		name: "no user detected",
		args: args{
			headers: map[string]string{},
		},
		want:    nil,
		wantErr: false,
	}, {
		name: "want error",
		args: args{
			headers: map[string]string{
				"authorization": "Bearer 123123123",
			},
		},
		want:    nil,
		wantErr: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUserID(tt.args.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserID() got = %v, want %v", got, tt.want)
			}
		})
	}
}
