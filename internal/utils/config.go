package utils

import (
	"intermark/internal/files"
	"log"
)

const ConfigPath = "config.json"

var Config ImConfig

type ImConfig struct {
	Title         string `json:"title"`
	EditPassword  string `json:"edit_password"`
	UpdateToken   string `json:"update_token"`
	UpdateTimeout int    `json:"update_timeout"`
	LogLevel      string `json:"log_level"`
	ContentRepo   struct {
		URL       string `json:"url"` // ssh clone url
		Branch    string `json:"branch"`
		AssetsDir string `json:"assets_dir"`
		SshHost   string `json:"ssh_host"`
	} `json:"content_repo"`
	Server struct {
		Port        int    `json:"port"` // empty defaults to http or https if tls key/cert are set
		TrustProxy  bool   `json:"trust_proxy"`
		CacheMaxAge int    `json:"cache_max_age"`
		TLSKeyPath  string `json:"tls_key_path"`
		TLSCertPath string `json:"tls_cert_path"`
	} `json:"server"`
}

func genDefaultConfig() ImConfig {
	var newConfig = ImConfig{}

	newConfig.Title = "Intermark"
	newConfig.UpdateTimeout = 60 // 1 minute
	newConfig.LogLevel = "warn"
	newConfig.ContentRepo.Branch = "main"
	newConfig.ContentRepo.AssetsDir = "assets"
	newConfig.ContentRepo.SshHost = "github-intermark"
	newConfig.Server.Port = 9292
	newConfig.Server.TrustProxy = true
	newConfig.Server.CacheMaxAge = 300 // 5 minutes

	return newConfig
}

// Load loads the configuration file. If the file does not exist, it creates a new one and returns false.
func (c *ImConfig) Load() bool {
	if ok, err := files.LoadJSON(ConfigPath, c); err != nil {
		log.Fatal(err)
	} else if !ok {
		*c = genDefaultConfig()
		c.Save()
		return false
	}
	return true
}

// Save saves the configuration to the file.
func (c *ImConfig) Save() {
	if err := files.SaveJSON(ConfigPath, c, 0777); err != nil {
		log.Fatalf("Error saving config: %s\n", err)
	}
}
