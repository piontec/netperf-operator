package operator

import (
	"testing"

	"github.com/piontec/netperf-operator/pkg/apis/app/fakekube"
	"github.com/piontec/netperf-operator/pkg/apis/app/kube"
)

func TestNetperf_parseNetperfResult(t *testing.T) {
	type fields struct {
		provider kube.Provider
	}
	type args struct {
		result string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    float64
		wantErr bool
	}{
		{
			name:    "Parse fail",
			fields:  fields{provider: fakekube.NewFakeProvider()},
			args:    args{result: "netperf output"},
			want:    0.0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Netperf{
				provider: tt.fields.provider,
			}
			got, err := n.parseNetperfResult(tt.args.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("Netperf.parseNetperfResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Netperf.parseNetperfResult() = %v, want %v", got, tt.want)
			}
		})
	}
}
