package management

import (
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProxyNodePluginPanel exposes only plugin-owned routes on a live CPA node.
func (h *Handler) ProxyNodePluginPanel(c *gin.Context) {
	nodeID := strings.TrimSpace(c.Param("node_id"))
	path := c.Param("path")
	if path != "/plugins" && !strings.HasPrefix(path, "/v0/management/") && !strings.HasPrefix(path, "/v0/resource/plugins/") {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	if h == nil || h.repo == nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	nodes, errNodes := h.activeCPANodes(c.Request.Context())
	if errNodes != nil {
		c.JSON(503, gin.H{"error": "node registry unavailable"})
		return
	}
	endpoint, errEndpoint := h.repo.GetPluginPanel(c.Request.Context(), nodeID)
	if errEndpoint != nil {
		if errors.Is(errEndpoint, gorm.ErrRecordNotFound) {
			c.JSON(409, gin.H{"error": "节点尚未注册插件面板，请升级 CPA 节点"})
		} else {
			c.JSON(503, gin.H{"error": "节点插件面板存储不可用，请检查 Home 数据库迁移"})
		}
		return
	}
	live := false
	for _, node := range nodes {
		if node.NodeID == nodeID && node.IP == endpoint.IP && node.Health == topologyHealthHealthy {
			live = true
			break
		}
	}
	if !live || net.ParseIP(endpoint.IP) == nil || endpoint.Port < 1 || endpoint.Port > 65535 {
		c.JSON(409, gin.H{"error": "节点当前不在线"})
		return
	}
	address := endpoint.IP
	if endpoint.Address != "" {
		address = endpoint.Address
	}
	if net.ParseIP(address) == nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}
	target := &url.URL{Scheme: "http", Host: net.JoinHostPort(address, strconv.Itoa(endpoint.Port))}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(request *http.Request) {
		request.URL.Scheme, request.URL.Host = target.Scheme, target.Host
		request.URL.Path, request.URL.RawPath = "/v8/internal/home-plugin"+path, ""
		request.Host = target.Host
		contentType := request.Header.Get("Content-Type")
		request.Header = make(http.Header)
		if contentType != "" {
			request.Header.Set("Content-Type", contentType)
		}
		request.Header.Set("Authorization", "Bearer "+endpoint.Token)
	}
	proxy.ModifyResponse = func(response *http.Response) error {
		response.Header.Del("Set-Cookie")
		response.Header.Set("Cache-Control", "no-store")
		return nil
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, _ error) {
		http.Error(w, "CPA plugin panel unavailable", http.StatusBadGateway)
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}
