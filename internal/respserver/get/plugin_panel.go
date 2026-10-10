package get

import (
	"context"
	"encoding/json"
	"net"

	"github.com/router-for-me/CLIProxyAPIHome/internal/node"
	"github.com/router-for-me/CLIProxyAPIHome/internal/respserver/dispatch"
)

func handlePluginPanel(ctx context.Context, env dispatch.Env, args []string) dispatch.Reply {
	if env.Runtime == nil || len(args) != 3 {
		return dispatch.Err("invalid plugin panel registration")
	}
	if env.NodeID == "" || env.ClientCertificateFingerprint == "" {
		return dispatch.Err("plugin panel requires mTLS node identity")
	}
	active, errActive := env.Runtime.PluginSyncNodeActive(ctx, env.NodeID, env.ClientCertificateFingerprint)
	if errActive != nil || !active {
		return dispatch.Err("plugin panel certificate is revoked or unknown")
	}
	var endpoint node.PluginPanelEndpoint
	if json.Unmarshal([]byte(args[2]), &endpoint) != nil || endpoint.Port < 1 || endpoint.Port > 65535 || len(endpoint.Token) < 32 || len(endpoint.Token) > 128 || net.ParseIP(env.ClientIP) == nil {
		return dispatch.Err("invalid plugin panel endpoint")
	}
	endpoint.NodeID, endpoint.IP = env.NodeID, env.ClientIP
	if errRegister := env.Runtime.RegisterPluginPanel(ctx, endpoint); errRegister != nil {
		return dispatch.Err("plugin panel registration failed")
	}
	return dispatch.SimpleString("OK")
}
