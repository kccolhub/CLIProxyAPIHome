package node

// PluginPanelEndpoint is registered through the authenticated node channel.
// Token must never be returned by the public management API.
type PluginPanelEndpoint struct {
	NodeID string `json:"-"`
	IP     string `json:"-"`
	Port   int    `json:"port"`
	Token  string `json:"token"`
}
