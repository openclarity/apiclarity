package main

import (
	"testing"

	"github.com/apiclarity/apiclarity/plugins/common/trace_sampling_client"
)

func TestAgent_shouldTrace(t *testing.T) {
	type fields struct {
		traceSamplingManager *trace_sampling_client.Client
	}
	type args struct {
		host string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "traceSamplingManager == nil",
			fields: fields{
				traceSamplingManager: nil,
			},
			args: args{
				host: "host1",
			},
			want: true,
		},
		{
			name: "found",
			fields: fields{
				traceSamplingManager: &trace_sampling_client.Client{
					Hosts: map[string]bool{
						"host1.ns1": true,
						"host2.ns2": true,
					},
				},
			},
			args: args{
				host: "host1.ns1",
			},
			want: true,
		},
		{
			name: "not found",
			fields: fields{
				traceSamplingManager: &trace_sampling_client.Client{
					Hosts: map[string]bool{
						"host1.ns1": true,
						"host2.ns2": true,
					},
				},
			},
			args: args{
				host: "host3",
			},
			want: false,
		},
		{
			name: "all",
			fields: fields{
				traceSamplingManager: &trace_sampling_client.Client{
					Hosts: map[string]bool{
						"host1.ns1": true,
						"host2.ns2": true,
						"*":         true,
					},
				},
			},
			args: args{
				host: "host3",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Agent{
				traceSamplingClient: tt.fields.traceSamplingManager,
			}
			if got := a.shouldTrace(tt.args.host); got != tt.want {
				t.Errorf("shouldTrace() = %v, want %v", got, tt.want)
			}
		})
	}
}
