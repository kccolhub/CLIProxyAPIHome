package get

import (
	"context"
	"encoding/json"

	"github.com/router-for-me/CLIProxyAPIHome/internal/respserver/dispatch"
)

func handlePluginAuth(ctx context.Context, env dispatch.Env, args []string) dispatch.Reply {
	if env.Runtime == nil || len(args) != 3 || !env.ConnectionLifetime.Controlled || env.NodeID == "" {
		return dispatch.Err("plugin auth requires controlled node connection")
	}
	active, errActive := env.Runtime.PluginSyncNodeActive(ctx, env.NodeID, env.ClientCertificateFingerprint)
	if errActive != nil || !active {
		return dispatch.Err("plugin auth certificate is revoked or unknown")
	}
	var request struct {
		PluginID string `json:"plugin_id"`
		Method   string `json:"method"`
		Index    string `json:"auth_index"`
	}
	if json.Unmarshal([]byte(args[2]), &request) != nil {
		return dispatch.Err("invalid plugin auth request")
	}
	raw, errRead := env.Runtime.ReadPluginAuth(ctx, request.PluginID, request.Method, request.Index)
	if errRead != nil {
		return dispatch.Err(errRead.Error())
	}
	return dispatch.SensitiveBulkString(raw)
}
