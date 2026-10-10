package dynamic

import (
	"context"
	"testing"

	"github.com/router-for-me/CLIProxyAPIHome/internal/config"
	"github.com/tidwall/gjson"
)

func TestPluginDispatchPinsCredentialAndPreservesAdmission(t *testing.T) {
	runtime, admitter := newConcurrencyDispatchRuntime(t, []string{"plugin-cred-a", "plugin-cred-b"})
	cfg := runtime.Config()
	enabled := true
	cfg.Plugins.Enabled = true
	cfg.Plugins.Configs = map[string]config.PluginInstanceConfig{"evaluation": {Enabled: &enabled}}
	runtime.SetPluginSyncConfigLoader(func(context.Context) (*config.Config, error) { return cfg, nil })
	runtime.SetPluginSyncNodeActive(func(_ context.Context, nodeID, fp string) (bool, error) { return nodeID == "node" && fp == "fp", nil })
	env := protocolOneDispatchEnv(runtime, "fp")
	env.NodeID = "node"
	env.ClientCertificateFingerprint = "fp"
	args := []string{"RPOP", `{"type":"auth","model":"gpt","plugin_id":"evaluation","pinned_auth_id":"plugin-cred-b","concurrency_protocol":1}`}
	reply := handleAuth(context.Background(), env, args)
	if gjson.GetBytes(reply.BulkString, "auth.id").String() != "plugin-cred-b" || !reply.AccountedAdmission || admitter.Calls() != 1 {
		t.Fatalf("plugin dispatch did not preserve selection/admission: %s", reply.BulkString)
	}
	for _, mutate := range []func(){func() { env.ConnectionLifetime.Controlled = false }, func() { env.ConnectionLifetime.Controlled = true; env.ClientCertificateFingerprint = "revoked" }, func() { env.ClientCertificateFingerprint = "fp"; enabled = false }} {
		mutate()
		reply = handleAuth(context.Background(), env, args)
		if !gjson.GetBytes(reply.BulkString, "error").Exists() || admitter.Calls() != 1 {
			t.Fatal("unauthorized plugin dispatch reached admission")
		}
	}
}
