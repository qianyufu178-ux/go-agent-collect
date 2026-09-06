package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// EdgeConfig represents the configuration for Nexus-Edge.
type EdgeConfig struct {
	General GeneralConfig `yaml:"general"`
	Proxy   ProxyConfig   `yaml:"proxy"`
	Limiter LimiterConfig `yaml:"limiter"`
	Collect CollectorConfig `yaml:"collector"`
	Cache   CacheConfig   `yaml:"cache"`
	GRPC    GRPCConfig    `yaml:"grpc"`
	API     APIConfig     `yaml:"api"`
}

// GeneralConfig holds general settings.
type GeneralConfig struct {
	NodeID      string `yaml:"node_id"`
	NodeName    string `yaml:"node_name"`
	LogLevel    string `yaml:"log_level"`
	ConfigFile  string `yaml:"-"`
}

// ProxyConfig holds HTTP reverse proxy settings.
type ProxyConfig struct {
	ListenAddr      string `yaml:"listen_addr"`
	UpstreamAddr    string `yaml:"upstream_addr"`
	UpstreamTimeout int    `yaml:"upstream_timeout"`
	MaxConcurrency  int    `yaml:"max_concurrency"`
}

// LimiterConfig holds rate limiting settings.
type LimiterConfig struct {
	Algorithm          string             `yaml:"algorithm"`
	GlobalMaxQPS       int                `yaml:"global_max_qps"`
	TokenBucketRate    float64            `yaml:"token_bucket_rate"`
	TokenBucketBurst   float64            `yaml:"token_bucket_burst"`
	WindowSizeMs       int64              `yaml:"window_size_ms"`
	WindowMaxRequests  int64              `yaml:"window_max_requests"`
	MaxConcurrency     int64              `yaml:"max_concurrency"`
	IPWhitelist        []string           `yaml:"ip_whitelist"`
	HeaderWhitelist    []HeaderRule       `yaml:"header_whitelist"`
	BlockedStatusCode  int                `yaml:"blocked_status_code"`
	BlockedResponse    string             `yaml:"blocked_response"`
	OverloadCPUThresh  float64            `yaml:"overload_cpu_threshold"`
	OverloadMemThresh  float64            `yaml:"overload_mem_threshold"`
	URLRules           map[string]int     `yaml:"url_rules"`
}

// HeaderRule defines a header-based whitelist rule.
type HeaderRule struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

// CollectorConfig holds metrics/log collection settings.
type CollectorConfig struct {
	MetricsInterval int      `yaml:"metrics_interval"`
	LogPaths        []string `yaml:"log_paths"`
	LogFilterMode   string   `yaml:"log_filter_mode"`
	LogKeywords     []string `yaml:"log_keywords"`
	LogMaxLineSize  int      `yaml:"log_max_line_size"`
	LogPollInterval int      `yaml:"log_poll_interval"`
	DiskMountPoint  string   `yaml:"disk_mount_point"`
}

// CacheConfig holds offline cache settings.
type CacheConfig struct {
	StorageType   string `yaml:"storage_type"`
	MaxSizeMB     int    `yaml:"max_size_mb"`
	ResendRate    int    `yaml:"resend_rate"`
	ReconnInterval int   `yaml:"reconn_interval"`
}

// GRPCConfig holds gRPC connection settings.
type GRPCConfig struct {
	ControllerAddr  string `yaml:"controller_addr"`
	HeartbeatSec    int    `yaml:"heartbeat_sec"`
	TimeoutSec      int    `yaml:"timeout_sec"`
	MaxRetries      int    `yaml:"max_retries"`
	AuthToken       string `yaml:"auth_token"`
}

// APIConfig holds local HTTP API settings.
type APIConfig struct {
	ListenAddr string `yaml:"listen_addr"`
}

// ControllerConfig represents the configuration for Nexus-Controller.
type ControllerConfig struct {
	General    ControllerGeneralConfig `yaml:"general"`
	GRPC       ControllerGRPCConfig    `yaml:"grpc"`
	Store      StoreConfig             `yaml:"store"`
	API        ControllerAPIConfig     `yaml:"api"`
}

// ControllerGeneralConfig holds general controller settings.
type ControllerGeneralConfig struct {
	LogLevel string `yaml:"log_level"`
}

// ControllerGRPCConfig holds controller gRPC settings.
type ControllerGRPCConfig struct {
	ListenAddr       string `yaml:"listen_addr"`
	HeartbeatTimeout int    `yaml:"heartbeat_timeout"`
	CheckInterval    int    `yaml:"check_interval"`
	AuthToken        string `yaml:"auth_token"`
}

// StoreConfig holds SQLite storage settings.
type StoreConfig struct {
	DBPath          string `yaml:"db_path"`
	BatchIntervalMs int    `yaml:"batch_interval_ms"`
	BatchSize       int    `yaml:"batch_size"`
	MetricsRetainDays int  `yaml:"metrics_retain_days"`
	LogsRetainDays  int    `yaml:"logs_retain_days"`
	CleanupHour     int    `yaml:"cleanup_hour"`
}

// ControllerAPIConfig holds controller HTTP API settings.
type ControllerAPIConfig struct {
	ListenAddr string `yaml:"listen_addr"`
}

// DefaultEdgeConfig returns the default configuration for Nexus-Edge.
func DefaultEdgeConfig() *EdgeConfig {
	return &EdgeConfig{
		General: GeneralConfig{
			NodeID:   "edge-node-1",
			NodeName: "Nexus Edge Node",
			LogLevel: "info",
		},
		Proxy: ProxyConfig{
			ListenAddr:      ":80",
			UpstreamAddr:    ":9090",
			UpstreamTimeout: 10,
			MaxConcurrency:  1024,
		},
		Limiter: LimiterConfig{
			Algorithm:          "token_bucket",
			GlobalMaxQPS:       1000,
			TokenBucketRate:    1000,
			TokenBucketBurst:   2000,
			WindowSizeMs:       1000,
			WindowMaxRequests:  1000,
			MaxConcurrency:     500,
			IPWhitelist:        []string{},
			HeaderWhitelist:    []HeaderRule{},
			BlockedStatusCode:  429,
			BlockedResponse:    `{"code":429,"msg":"系统繁忙，请稍后再试"}`,
			OverloadCPUThresh:  80.0,
			OverloadMemThresh:  85.0,
			URLRules:           map[string]int{},
		},
		Collect: CollectorConfig{
			MetricsInterval: 10,
			LogPaths:        []string{},
			LogFilterMode:   "include",
			LogKeywords:     []string{},
			LogMaxLineSize:  10240,
			LogPollInterval: 1000,
			DiskMountPoint:  "/",
		},
		Cache: CacheConfig{
			StorageType:    "sqlite",
			MaxSizeMB:      100,
			ResendRate:     50,
			ReconnInterval: 3,
		},
		GRPC: GRPCConfig{
			ControllerAddr: "127.0.0.1:9999",
			HeartbeatSec:   5,
			TimeoutSec:     10,
			MaxRetries:     3,
			AuthToken:      "",
		},
		API: APIConfig{
			ListenAddr: ":8081",
		},
	}
}

// DefaultControllerConfig returns the default configuration for Nexus-Controller.
func DefaultControllerConfig() *ControllerConfig {
	return &ControllerConfig{
		General: ControllerGeneralConfig{
			LogLevel: "info",
		},
		GRPC: ControllerGRPCConfig{
			ListenAddr:       ":9999",
			HeartbeatTimeout: 15,
			CheckInterval:    5,
			AuthToken:        "",
		},
		Store: StoreConfig{
			DBPath:            "nexus_center.db",
			BatchIntervalMs:   1000,
			BatchSize:         100,
			MetricsRetainDays: 7,
			LogsRetainDays:    30,
			CleanupHour:       2,
		},
		API: ControllerAPIConfig{
			ListenAddr: ":8080",
		},
	}
}

// LoadEdgeConfig loads Edge configuration from a YAML file.
func LoadEdgeConfig(path string) (*EdgeConfig, error) {
	cfg := DefaultEdgeConfig()
	cfg.General.ConfigFile = path

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return cfg, nil
}

// LoadControllerConfig loads Controller configuration from a YAML file.
func LoadControllerConfig(path string) (*ControllerConfig, error) {
	cfg := DefaultControllerConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return cfg, nil
}
