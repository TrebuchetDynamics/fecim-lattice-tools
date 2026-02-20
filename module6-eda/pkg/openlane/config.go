package openlane

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Config holds OpenLane configuration
type Config struct {
	PDKRoot          string        `json:"pdk_root"`
	PDKVariant       string        `json:"pdk_variant"`
	SCLibrary        string        `json:"sc_library"`
	PreferredMode    Mode          `json:"preferred_mode"`
	TimeoutPlacement time.Duration `json:"-"`
	TimeoutSynthesis time.Duration `json:"-"`
	TimeoutRouting   time.Duration `json:"-"`
	DockerImage      string        `json:"docker_image"`
}

// configJSON is for JSON serialization with string durations
type configJSON struct {
	PDKRoot          string `json:"pdk_root"`
	PDKVariant       string `json:"pdk_variant"`
	SCLibrary        string `json:"sc_library"`
	PreferredMode    string `json:"preferred_mode"`
	TimeoutPlacement string `json:"timeout_placement"`
	TimeoutSynthesis string `json:"timeout_synthesis"`
	TimeoutRouting   string `json:"timeout_routing"`
	DockerImage      string `json:"docker_image"`
}

// DefaultConfig returns configuration with SKY130 defaults
func DefaultConfig() *Config {
	pdkRoot := os.Getenv("PDK_ROOT")
	if pdkRoot == "" {
		pdkRoot = filepath.Join(os.Getenv("HOME"), ".volare")
	}
	return &Config{
		PDKRoot:          pdkRoot,
		PDKVariant:       "sky130A",
		SCLibrary:        "sky130_fd_sc_hd",
		PreferredMode:    ModeDocker,
		TimeoutPlacement: 5 * time.Minute,
		TimeoutSynthesis: 10 * time.Minute,
		TimeoutRouting:   15 * time.Minute,
		DockerImage:      "ghcr.io/the-openroad-project/openlane:latest",
	}
}

// GetConfigPath returns the default config file path
func GetConfigPath() string {
	home := os.Getenv("HOME")
	return filepath.Join(home, ".fecim", "openlane-config.json")
}

// LoadConfig loads configuration from file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig(), err
	}

	var cj configJSON
	if err := json.Unmarshal(data, &cj); err != nil {
		return DefaultConfig(), err
	}

	cfg := DefaultConfig()
	cfg.PDKRoot = cj.PDKRoot
	cfg.PDKVariant = cj.PDKVariant
	cfg.SCLibrary = cj.SCLibrary
	cfg.DockerImage = cj.DockerImage

	if cj.PreferredMode == "native" {
		cfg.PreferredMode = ModeNative
	}

	if d, err := time.ParseDuration(cj.TimeoutPlacement); err == nil {
		cfg.TimeoutPlacement = d
	}
	if d, err := time.ParseDuration(cj.TimeoutSynthesis); err == nil {
		cfg.TimeoutSynthesis = d
	}
	if d, err := time.ParseDuration(cj.TimeoutRouting); err == nil {
		cfg.TimeoutRouting = d
	}

	return cfg, nil
}

// SaveConfig saves configuration to file
func SaveConfig(cfg *Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	modeStr := "docker"
	if cfg.PreferredMode == ModeNative {
		modeStr = "native"
	}

	cj := configJSON{
		PDKRoot:          cfg.PDKRoot,
		PDKVariant:       cfg.PDKVariant,
		SCLibrary:        cfg.SCLibrary,
		PreferredMode:    modeStr,
		TimeoutPlacement: cfg.TimeoutPlacement.String(),
		TimeoutSynthesis: cfg.TimeoutSynthesis.String(),
		TimeoutRouting:   cfg.TimeoutRouting.String(),
		DockerImage:      cfg.DockerImage,
	}

	data, err := json.MarshalIndent(cj, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetTechLEFPath returns the path to the tech LEF file
func (c *Config) GetTechLEFPath() string {
	return filepath.Join(c.PDKRoot, c.PDKVariant, "libs.tech", "openlane", c.SCLibrary, c.SCLibrary+".tlef")
}

// GetCellLEFPath returns the path to the cell LEF file
func (c *Config) GetCellLEFPath() string {
	return filepath.Join(c.PDKRoot, c.PDKVariant, "libs.ref", c.SCLibrary, "lef", c.SCLibrary+".lef")
}

// GetLibertyPath returns the path to the Liberty timing file
func (c *Config) GetLibertyPath() string {
	return filepath.Join(c.PDKRoot, c.PDKVariant, "libs.ref", c.SCLibrary, "lib", c.SCLibrary+"__tt_025C_1v80.lib")
}

// DefaultIHPConfig returns configuration with IHP SG13G2 defaults.
// IHP SG13G2 uses OpenLane 2 (not volare) and has its own PDK directory structure.
// PDK root typically: $IHP_PDK_ROOT or ~/ihp-sg13g2-pdk
func DefaultIHPConfig() *Config {
	pdkRoot := os.Getenv("IHP_PDK_ROOT")
	if pdkRoot == "" {
		pdkRoot = filepath.Join(os.Getenv("HOME"), "ihp-sg13g2-pdk")
	}
	return &Config{
		PDKRoot:          pdkRoot,
		PDKVariant:       "sg13g2",
		SCLibrary:        "sg13g2_stdcell",
		PreferredMode:    ModeDocker,
		TimeoutPlacement: 5 * time.Minute,
		TimeoutSynthesis: 10 * time.Minute,
		TimeoutRouting:   15 * time.Minute,
		DockerImage:      "ghcr.io/the-openroad-project/openlane:latest",
	}
}

// GetIHPTechLEFPath returns path to IHP SG13G2 technology LEF.
func (c *Config) GetIHPTechLEFPath() string {
	return filepath.Join(c.PDKRoot, c.PDKVariant, "libs.ref", c.SCLibrary, "lef", c.SCLibrary+"_tech.lef")
}

// GetIHPCellLEFPath returns path to IHP SG13G2 standard cell LEF.
func (c *Config) GetIHPCellLEFPath() string {
	return filepath.Join(c.PDKRoot, c.PDKVariant, "libs.ref", c.SCLibrary, "lef", c.SCLibrary+".lef")
}

// GetVolareSetupInstructions returns volare setup instructions
func GetVolareSetupInstructions() string {
	return `To set up SKY130 PDK using volare:

1. Install volare:
   pip install volare

2. Enable SKY130 PDK:
   volare enable --pdk sky130 sky130A

3. Set PDK_ROOT environment variable:
   export PDK_ROOT=~/.volare

4. (Optional) Add to shell profile:
   echo 'export PDK_ROOT=~/.volare' >> ~/.bashrc
`
}

// GetIHPSetupInstructions returns IHP SG13G2 PDK setup instructions.
func GetIHPSetupInstructions() string {
	return `To set up IHP SG13G2 PDK:

1. Clone the IHP-Open-PDK:
   git clone https://github.com/IHP-GmbH/IHP-Open-PDK.git ~/ihp-sg13g2-pdk

2. Set IHP_PDK_ROOT environment variable:
   export IHP_PDK_ROOT=~/ihp-sg13g2-pdk

3. (Optional) Add to shell profile:
   echo 'export IHP_PDK_ROOT=~/ihp-sg13g2-pdk' >> ~/.bashrc

4. Run LibreLane with IHP config:
   python3 -m librelane --config-file config.json

IHP Open MPW shuttles (active as of 2026):
   See: https://www.ihp-microelectronics.com/services/research-and-prototyping-service/mpw-prototyping-service
`
}
