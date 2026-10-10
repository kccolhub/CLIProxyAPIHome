package management

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPIHome/internal/cluster"
	"github.com/router-for-me/CLIProxyAPIHome/internal/node"
)

func TestPluginPanelProxyUsesLiveNodeAndDoesNotForwardHomeSecret(t *testing.T) {
	db, cleanup := openManagementLogTestDB(t)
	defer cleanup()
	repo := cluster.NewRepository(db)
	now := time.Now().UTC()
	home := cluster.ClusterNodeRecord{IP: "home", Port: 8327, StartedAt: now.Add(-time.Hour), LastSeenAt: now}
	if err := db.Create(&home).Error; err != nil {
		t.Fatal(err)
	}
	seedActiveManagementCPA(t, db, home, "fp", "node", "127.0.0.1", now.Add(-time.Minute), now)
	calls := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/v8/internal/home-plugin/plugins" || r.Header.Get("Authorization") != "Bearer node-secret" || r.Header.Get("X-Management-Key") != "" || r.Header.Get("Cookie") != "" {
			t.Errorf("unexpected proxy request")
		}
		fmt.Fprint(w, `{"plugins":[]}`)
	}))
	defer upstream.Close()
	_, port, _ := net.SplitHostPort(upstream.Listener.Addr().String())
	n, _ := strconv.Atoi(port)
	if err := repo.RegisterPluginPanel(context.Background(), node.PluginPanelEndpoint{NodeID: "node", IP: "127.0.0.1", Port: n, Token: "node-secret"}); err != nil {
		t.Fatal(err)
	}
	handler := NewHandler(repo, nil, "home", 8327)
	engine := gin.New()
	engine.Any("/nodes/:node_id/plugin-panel/*path", handler.ProxyNodePluginPanel)
	for _, test := range []struct {
		path   string
		status int
	}{{"plugins", 200}, {"v1/models", 404}} {
		req := httptest.NewRequest("GET", "/nodes/node/plugin-panel/"+test.path, nil)
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		req = req.WithContext(ctx)
		req.Header.Set("Authorization", "Bearer home-secret")
		req.Header.Set("X-Management-Key", "home-secret")
		req.Header.Set("Cookie", "private=value")
		response := httptest.NewRecorder()
		engine.ServeHTTP(response, req)
		if response.Code != test.status {
			t.Fatalf("%s: %d %s", test.path, response.Code, response.Body.String())
		}
	}
	if err := repo.RegisterPluginPanel(context.Background(), node.PluginPanelEndpoint{NodeID: "node", IP: "127.0.0.2", Port: n, Token: "node-secret"}); err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, httptest.NewRequest("GET", "/nodes/node/plugin-panel/plugins", nil))
	if response.Code != 409 || calls != 1 {
		t.Fatalf("stale registration forwarded: status=%d calls=%d", response.Code, calls)
	}
}
