package cluster

import (
	"context"
	"time"

	"github.com/router-for-me/CLIProxyAPIHome/internal/node"
	"gorm.io/gorm/clause"
)

// PluginPanelRecord is ephemeral registration state, excluded from exports.
type PluginPanelRecord struct {
	NodeID    string `gorm:"primaryKey;size:255"`
	IP        string
	Port      int
	Token     string
	UpdatedAt time.Time
}

func (PluginPanelRecord) TableName() string { return "node_plugin_panel" }

func (r *Repository) RegisterPluginPanel(ctx context.Context, endpoint node.PluginPanelEndpoint) error {
	db, errDB := r.database()
	if errDB != nil {
		return errDB
	}
	record := PluginPanelRecord{NodeID: endpoint.NodeID, IP: endpoint.IP, Port: endpoint.Port, Token: endpoint.Token, UpdatedAt: time.Now().UTC()}
	return db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "node_id"}}, DoUpdates: clause.AssignmentColumns([]string{"ip", "port", "token", "updated_at"})}).Create(&record).Error
}

func (r *Repository) GetPluginPanel(ctx context.Context, nodeID string) (PluginPanelRecord, error) {
	db, errDB := r.database()
	if errDB != nil {
		return PluginPanelRecord{}, errDB
	}
	var record PluginPanelRecord
	errFind := db.WithContext(ctx).Where("node_id = ?", nodeID).First(&record).Error
	return record, errFind
}
