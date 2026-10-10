package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/router-for-me/CLIProxyAPI/v8/sdk/pluginapi"
)

// ReadPluginAuth backs node host callbacks with the authoritative DB inventory.
func (r *Repository) ReadPluginAuth(ctx context.Context, method, index string) ([]byte, error) {
	if method != "list" && method != "get" {
		return nil, fmt.Errorf("unsupported plugin auth method")
	}
	auths, errList := r.ListAuths(ctx)
	if errList != nil {
		return nil, errList
	}
	files := make([]pluginapi.HostAuthFileEntry, 0, len(auths))
	for _, auth := range auths {
		if method == "get" {
			if auth.ID == index {
				return json.Marshal(map[string]any{"auth_index": auth.ID, "name": auth.FileName, "json": auth.Metadata})
			}
			continue
		}
		name := auth.FileName
		if name == "" {
			name = auth.ID
		}
		email, _ := auth.Metadata["email"].(string)
		files = append(files, pluginapi.HostAuthFileEntry{ID: auth.ID, AuthIndex: auth.ID, Name: name, Type: auth.Provider, Provider: auth.Provider, Label: auth.Label, Status: string(auth.Status), StatusMessage: auth.StatusMessage, Disabled: auth.Disabled, Unavailable: auth.Unavailable, RuntimeOnly: strings.EqualFold(auth.Attributes["runtime_only"], "true"), Source: "db", Email: email, NextRetryAfter: auth.NextRetryAfter, CreatedAt: auth.CreatedAt, UpdatedAt: auth.UpdatedAt})
	}
	if method == "get" {
		return nil, fmt.Errorf("credential not found")
	}
	return json.Marshal(map[string]any{"files": files})
}
