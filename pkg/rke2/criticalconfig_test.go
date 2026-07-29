package rke2

import (
	"flag"
	"testing"

	rke2cli "github.com/rancher/rke2/pkg/cli"
	"github.com/urfave/cli/v2"
)

func newTestContext(t *testing.T, cni []string, prime bool) *cli.Context {
	t.Helper()
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	set.Var(cli.NewStringSlice(cni...), "cni", "")
	set.Bool("prime", prime, "")
	return cli.NewContext(nil, set, nil)
}

func newConfig(ingress ...string) rke2cli.Config {
	cfg := rke2cli.Config{}
	for _, i := range ingress {
		cfg.IngressController.Set(i)
	}
	return cfg
}

func TestSerializeCriticalExtraConfig(t *testing.T) {
	tests := []struct {
		name    string
		cni     []string
		ingress []string
		prime   bool
		want    string
	}{
		{
			name:    "defaults canal traefik prime",
			cni:     []string{"canal"},
			ingress: []string{"traefik"},
			prime:   true,
			want:    `{"cni":["canal"],"ingressController":["traefik"],"prime":true}`,
		},
		{
			name:    "multus ordering preserved",
			cni:     []string{"multus", "canal"},
			ingress: []string{"ingress-nginx"},
			prime:   false,
			want:    `{"cni":["multus","canal"],"ingressController":["ingress-nginx"],"prime":false}`,
		},
		{
			name:    "empty ingress serializes as empty list",
			cni:     []string{"cilium"},
			ingress: nil,
			prime:   false,
			want:    `{"cni":["cilium"],"ingressController":null,"prime":false}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clx := newTestContext(t, tt.cni, tt.prime)
			cfg := newConfig(tt.ingress...)

			got, err := serializeCriticalExtraConfig(clx, cfg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("serializeCriticalExtraConfig() =\n  %s\nwant\n  %s", got, tt.want)
			}

			// Serialization must be byte-stable across repeated calls so that two
			// control-plane nodes with identical settings never see a false mismatch.
			again, err := serializeCriticalExtraConfig(clx, cfg)
			if err != nil {
				t.Fatalf("unexpected error on repeat: %v", err)
			}
			if again != got {
				t.Errorf("serialization not stable:\n  %s\n  %s", got, again)
			}
		})
	}
}
