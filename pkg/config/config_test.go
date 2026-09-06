package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultEdgeConfig(t *testing.T) {
	cfg := DefaultEdgeConfig()
	if cfg.General.NodeID != "edge-node-1" {
		t.Errorf("expected edge-node-1, got %s", cfg.General.NodeID)
	}
	if cfg.Limiter.Algorithm != "token_bucket" {
		t.Errorf("expected token_bucket, got %s", cfg.Limiter.Algorithm)
	}
	if cfg.Limiter.GlobalMaxQPS != 1000 {
		t.Errorf("expected 1000, got %d", cfg.Limiter.GlobalMaxQPS)
	}
	if cfg.Collect.MetricsInterval != 10 {
		t.Errorf("expected 10, got %d", cfg.Collect.MetricsInterval)
	}
	if cfg.GRPC.ControllerAddr != "127.0.0.1:9999" {
		t.Errorf("expected 127.0.0.1:9999, got %s", cfg.GRPC.ControllerAddr)
	}
}

func TestDefaultControllerConfig(t *testing.T) {
	cfg := DefaultControllerConfig()
	if cfg.GRPC.ListenAddr != ":9999" {
		t.Errorf("expected :9999, got %s", cfg.GRPC.ListenAddr)
	}
	if cfg.Store.DBPath != "nexus_center.db" {
		t.Errorf("expected nexus_center.db, got %s", cfg.Store.DBPath)
	}
	if cfg.Store.MetricsRetainDays != 7 {
		t.Errorf("expected 7, got %d", cfg.Store.MetricsRetainDays)
	}
}

func TestLoadEdgeConfig(t *testing.T) {
	// Create a temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "edge.yaml")

	content := `
general:
  node_id: "test-node"
  log_level: "debug"
limiter:
  algorithm: "sliding_window"
  global_max_qps: 500
collector:
  metrics_interval: 30
grpc:
  controller_addr: "10.0.0.1:9999"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadEdgeConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.General.NodeID != "test-node" {
		t.Errorf("expected test-node, got %s", cfg.General.NodeID)
	}
	if cfg.General.LogLevel != "debug" {
		t.Errorf("expected debug, got %s", cfg.General.LogLevel)
	}
	if cfg.Limiter.Algorithm != "sliding_window" {
		t.Errorf("expected sliding_window, got %s", cfg.Limiter.Algorithm)
	}
	if cfg.Limiter.GlobalMaxQPS != 500 {
		t.Errorf("expected 500, got %d", cfg.Limiter.GlobalMaxQPS)
	}
	if cfg.Collect.MetricsInterval != 30 {
		t.Errorf("expected 30, got %d", cfg.Collect.MetricsInterval)
	}
	if cfg.GRPC.ControllerAddr != "10.0.0.1:9999" {
		t.Errorf("expected 10.0.0.1:9999, got %s", cfg.GRPC.ControllerAddr)
	}

	// Verify defaults for unset fields
	if cfg.Proxy.ListenAddr != ":80" {
		t.Errorf("expected :80, got %s", cfg.Proxy.ListenAddr)
	}
	if cfg.Limiter.TokenBucketRate != 1000 {
		t.Errorf("expected 1000, got %f", cfg.Limiter.TokenBucketRate)
	}
}

func TestLoadEdgeConfigFileNotFound(t *testing.T) {
	cfg, err := LoadEdgeConfig("/nonexistent/path.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.General.NodeID != "edge-node-1" {
		t.Errorf("expected default edge-node-1, got %s", cfg.General.NodeID)
	}
}

func TestLoadControllerConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "controller.yaml")

	content := `
general:
  log_level: "debug"
grpc:
  listen_addr: ":19999"
  heartbeat_timeout: 30
store:
  db_path: "/data/nexus.db"
  metrics_retain_days: 14
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadControllerConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.General.LogLevel != "debug" {
		t.Errorf("expected debug, got %s", cfg.General.LogLevel)
	}
	if cfg.GRPC.ListenAddr != ":19999" {
		t.Errorf("expected :19999, got %s", cfg.GRPC.ListenAddr)
	}
	if cfg.GRPC.HeartbeatTimeout != 30 {
		t.Errorf("expected 30, got %d", cfg.GRPC.HeartbeatTimeout)
	}
	if cfg.Store.DBPath != "/data/nexus.db" {
		t.Errorf("expected /data/nexus.db, got %s", cfg.Store.DBPath)
	}
	if cfg.Store.MetricsRetainDays != 14 {
		t.Errorf("expected 14, got %d", cfg.Store.MetricsRetainDays)
	}

	// Verify defaults
	if cfg.Store.LogsRetainDays != 30 {
		t.Errorf("expected 30, got %d", cfg.Store.LogsRetainDays)
	}
}
