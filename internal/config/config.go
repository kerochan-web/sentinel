package config

import (
	"os"
	"time"
	"gopkg.in/yaml.v3"
)

// Config is the root configuration object for Sentinel
type Config struct {
	ServiceNow  ServiceNowConfig `yaml:"servicenow"`
	Services    []Service        `yaml:"services"`
	Remediation Remediation      `yaml:"remediation_defaults"`
	Notifications NotificationsConfig `yaml:"notifications"` // Added for alerts
}

// ServiceNowConfig handles connectivity to the mock ITSM
type ServiceNowConfig struct {
	InstanceURL string `yaml:"instance_url"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"` // In prod, use env vars/secrets
}

// NotificationsConfig holds our alerting channels
type NotificationsConfig struct {
	NtfyTopic string `yaml:"ntfy_topic"`
}

// Service defines what we are monitoring
type Service struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`   // http, tcp, systemd
	Target      string `yaml:"target"` // URL, Port, or Service Name
	CheckInterval time.Duration `yaml:"check_interval"`
	Maintenance      bool          `yaml:"maintenance"`
	MaintenanceUntil time.Time     `yaml:"maintenance_until"`
}

// Remediation defines the safety guardrails
type Remediation struct {
	MaxRetries      int           `yaml:"max_retries"`
	CooldownPeriod  time.Duration `yaml:"cooldown_period"`
	CircuitBreaker  int           `yaml:"circuit_breaker_threshold"`
}

func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	return &cfg, err
}
