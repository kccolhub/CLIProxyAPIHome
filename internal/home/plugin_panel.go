package home

import (
	"context"
	"fmt"
	"net/http"

	coreauth "github.com/router-for-me/CLIProxyAPIHome/internal/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPIHome/internal/node"
)

func (r *Runtime) SetPluginPanelRegister(register func(context.Context, node.PluginPanelEndpoint) error) {
	r.pluginPanelRegister = register
}

func (r *Runtime) SetPluginAuthRead(read func(context.Context, string, string) ([]byte, error)) {
	r.pluginAuthRead = read
}

func (r *Runtime) ReadPluginAuth(ctx context.Context, pluginID, method, index string) ([]byte, error) {
	if !r.PluginEnabled(ctx, pluginID) || r.pluginAuthRead == nil {
		return nil, fmt.Errorf("plugin credential access unavailable")
	}
	return r.pluginAuthRead(ctx, method, index)
}

func (r *Runtime) RegisterPluginPanel(ctx context.Context, endpoint node.PluginPanelEndpoint) error {
	if r.pluginPanelRegister == nil {
		return fmt.Errorf("plugin panel storage unavailable")
	}
	return r.pluginPanelRegister(ctx, endpoint)
}

func (r *Runtime) PluginEnabled(ctx context.Context, pluginID string) bool {
	cfg, errConfig := r.pluginSyncConfig(ctx)
	if errConfig != nil || cfg == nil || !cfg.Plugins.Enabled || pluginID == "" {
		return false
	}
	item, ok := cfg.Plugins.Configs[pluginID]
	return ok && item.Enabled != nil && *item.Enabled
}

// DispatchForPluginWithConcurrency retains normal admission and release accounting.
// Only the authenticated RESP node handler may enter this path.
func (r *Runtime) DispatchForPluginWithConcurrency(ctx context.Context, model string, headers http.Header, pluginID, policy string, excluded []string, pinned string, concurrency DispatchConcurrencyContext) (*DispatchResult, error) {
	if !concurrency.Controlled || pinned == "" || !r.PluginEnabled(ctx, pluginID) {
		return nil, fmt.Errorf("plugin dispatch requires an enabled plugin and a pinned credential")
	}
	metadata := map[string]any{coreauth.AllowedAuthIDsMetadataKey: []string{pinned}, coreauth.CredentialPolicyMetadataKey: policy}
	if len(excluded) > 0 {
		metadata[coreauth.ExcludedAuthIDsMetadataKey] = excluded
	}
	if concurrency.RetryRoundPresent || concurrency.RetryRound > 0 {
		metadata[coreauth.RequestRetryRoundMetadataKey] = concurrency.RetryRound
	}
	return r.dispatchWithOptions(ctx, model, coreauth.Options{Headers: headers.Clone(), Metadata: metadata}, concurrency)
}
