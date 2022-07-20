package notifier

import "testing"

func Test_setSchemeIfNeeded(t *testing.T) {
	type args struct {
		url    string
		scheme string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "url already have a scheme",
			args: args{
				url:    "https://example.com",
				scheme: "http",
			},
			want: "https://example.com",
		},
		{
			name: "url is missing a scheme",
			args: args{
				url:    "example.com",
				scheme: "http",
			},
			want: "http://example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := setSchemeIfNeeded(tt.args.url, tt.args.scheme); got != tt.want {
				t.Errorf("setSchemeIfNeeded() = %v, want %v", got, tt.want)
			}
		})
	}
}
