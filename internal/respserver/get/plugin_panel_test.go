package get

import (
	"context"
	"testing"

	"github.com/router-for-me/CLIProxyAPIHome/internal/node"
	"github.com/router-for-me/CLIProxyAPIHome/internal/respserver/dispatch"
)

func TestPluginPanelRegistrationUsesCertificateIdentity(t *testing.T) {
	runtime := pluginSyncTestRuntime(t)
	runtime.SetPluginSyncNodeActive(func(context.Context, string, string) (bool, error) { return true, nil })
	var stored node.PluginPanelEndpoint
	runtime.SetPluginPanelRegister(func(_ context.Context, endpoint node.PluginPanelEndpoint) error { stored = endpoint; return nil })
	args := []string{"GET", "plugin-panel", `{"port":8317,"token":"12345678901234567890123456789012","node_id":"victim","ip":"192.0.2.7"}`}
	env := dispatch.Env{Runtime: runtime, NodeID: "node", ClientCertificateFingerprint: "fp", ClientIP: "127.0.0.1"}
	if result := handlePluginPanel(context.Background(), env, args); result.Kind != dispatch.ReplyKindSimpleString {
		t.Fatalf("unexpected reply: %+v", result)
	}
	if stored.NodeID != "node" || stored.IP != "127.0.0.1" {
		t.Fatal("client overrode trusted identity")
	}
	runtime.SetPluginSyncNodeActive(func(context.Context, string, string) (bool, error) { return false, nil })
	if result := handlePluginPanel(context.Background(), env, args); result.Kind != dispatch.ReplyKindRedisError {
		t.Fatal("revoked certificate accepted")
	}
}
