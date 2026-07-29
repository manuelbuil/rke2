package rke2

import (
	"encoding/json"

	rke2cli "github.com/rancher/rke2/pkg/cli"
	"github.com/urfave/cli/v2"
)

// criticalExtraConfig is the set of RKE2-specific settings that must be
// identical across all control-plane nodes in an HA cluster. It is serialized
// into K3s's opaque CriticalExtraConfig field and compared verbatim at join
// time by K3s's bootstrap validation.
//
// Field order is fixed by the struct definition and slice order is preserved
// (never sorted): both are semantically significant here — multus must be the
// first CNI value, and the first ingress-controller value becomes the default
// ingress class — so the JSON is deterministic without any additional
// canonicalization.
type criticalExtraConfig struct {
	CNI               []string `json:"cni"`
	IngressController []string `json:"ingressController"`
	Prime             bool     `json:"prime"`
}

// serializeCriticalExtraConfig produces a stable JSON serialization of the
// RKE2-specific critical settings for the given server invocation. The output
// is fed into K3s's CriticalExtraConfig so that a joining server with divergent
// settings fails fast during bootstrap.
func serializeCriticalExtraConfig(clx *cli.Context, cfg rke2cli.Config) (string, error) {
	c := criticalExtraConfig{
		CNI:               clx.StringSlice("cni"),
		IngressController: cfg.IngressController.Value(),
		Prime:             clx.Bool("prime"),
	}
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
