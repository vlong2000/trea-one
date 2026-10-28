package factory

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigPath = "./config/nrfcfg.yaml"
	DefaultScheme     = "http"
	DefaultBindIP     = "127.0.0.1"
	DefaultPort       = 8000
	NfmPrefix         = "/nnrf-nfm/v1"
	DiscPrefix        = "/nnrf-disc/v1"
	BootPrefix        = "/bootstrapping"
)

type Config struct {
	Info          *Info          `yaml:"info"`
	Configuration *Configuration `yaml:"configuration"`
	Logger        *Logger        `yaml:"logger"`
}

type Info struct {
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}

type Configuration struct {
	Sbi             *Sbi     `yaml:"sbi"`
	DefaultPlmnId   PlmnId   `yaml:"DefaultPlmnId"`
	ServiceNameList []string `yaml:"serviceNameList"`
}

type Sbi struct {
	Scheme       string `yaml:"scheme"`
	RegisterIPv4 string `yaml:"registerIPv4"`
	BindingIPv4  string `yaml:"bindingIPv4"`
	Port         int    `yaml:"port"`
	OAuth        bool   `yaml:"oauth"`
}

type PlmnId struct {
	Mcc string `yaml:"mcc"`
	Mnc string `yaml:"mnc"`
}

type Logger struct {
	Enable       bool   `yaml:"enable"`
	Level        string `yaml:"level"`
	ReportCaller bool   `yaml:"reportCaller"`
}

func ReadConfig(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigPath
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("nrf: read config %s: %w", path, err)
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(content, cfg); err != nil {
		return nil, fmt.Errorf("nrf: parse config %s: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("nrf: invalid config %s: %w", path, err)
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Info == nil {
		return fmt.Errorf("info is required")
	}
	if c.Configuration == nil {
		return fmt.Errorf("configuration is required")
	}
	if c.Logger == nil {
		return fmt.Errorf("logger is required")
	}
	sbi := c.Configuration.Sbi
	if sbi == nil {
		return fmt.Errorf("configuration.sbi is required")
	}
	if sbi.Scheme != "http" && sbi.Scheme != "https" {
		return fmt.Errorf("sbi.scheme must be http or https")
	}
	if sbi.BindingIPv4 == "" {
		return fmt.Errorf("sbi.bindingIPv4 is required")
	}
	if sbi.Port <= 0 || sbi.Port > 65535 {
		return fmt.Errorf("sbi.port must be in 1-65535")
	}
	plmn := c.Configuration.DefaultPlmnId
	if len(plmn.Mcc) != 3 || !isDigits(plmn.Mcc) {
		return fmt.Errorf("DefaultPlmnId.mcc must be 3 digits")
	}
	if n := len(plmn.Mnc); n < 2 || n > 3 || !isDigits(plmn.Mnc) {
		return fmt.Errorf("DefaultPlmnId.mnc must be 2 or 3 digits")
	}
	if len(c.Configuration.ServiceNameList) == 0 {
		return fmt.Errorf("serviceNameList is required")
	}
	for _, name := range c.Configuration.ServiceNameList {
		if name != "nnrf-nfm" && name != "nnrf-disc" {
			return fmt.Errorf("unsupported service name %s", name)
		}
	}
	level := strings.ToLower(c.Logger.Level)
	switch level {
	case "trace", "debug", "info", "warn", "error", "fatal", "panic":
	default:
		return fmt.Errorf("logger.level is invalid")
	}
	return nil
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}

func (c *Config) BindingAddr() string {
	ip := DefaultBindIP
	port := DefaultPort
	if c.Configuration != nil && c.Configuration.Sbi != nil {
		if c.Configuration.Sbi.BindingIPv4 != "" {
			if envIP := os.Getenv(c.Configuration.Sbi.BindingIPv4); envIP != "" {
				ip = envIP
			} else {
				ip = c.Configuration.Sbi.BindingIPv4
			}
		}
		if c.Configuration.Sbi.Port != 0 {
			port = c.Configuration.Sbi.Port
		}
	}
	return ip + ":" + strconv.Itoa(port)
}

func (c *Config) Scheme() string {
	if c.Configuration != nil && c.Configuration.Sbi != nil && c.Configuration.Sbi.Scheme != "" {
		return c.Configuration.Sbi.Scheme
	}
	return DefaultScheme
}

func (c *Config) RegisterIP() string {
	if c.Configuration != nil && c.Configuration.Sbi != nil && c.Configuration.Sbi.RegisterIPv4 != "" {
		return c.Configuration.Sbi.RegisterIPv4
	}
	return DefaultBindIP
}

func (c *Config) Port() int {
	if c.Configuration != nil && c.Configuration.Sbi != nil && c.Configuration.Sbi.Port != 0 {
		return c.Configuration.Sbi.Port
	}
	return DefaultPort
}

func (c *Config) SbiURI() string {
	return c.Scheme() + "://" + c.RegisterIP() + ":" + strconv.Itoa(c.Port())
}
