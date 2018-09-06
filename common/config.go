package common

import (
	"os"

	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

type ACCConfig struct {
	ID        string
	SecKey    string
	Algorithm string
}

type GenesisConfig struct {
	CreateGenesis bool
	GenesisHash   string
	WitnessInfo   []string
}

type DBConfig struct {
	LdbPath string
}

type VMConfig struct {
	JsPath   string
	LogLevel string
}

// P2PConfig is the config for p2p network.
type P2PConfig struct {
	ListenAddr string
	SeedNodes  []string
	ChainID    uint32
	Version    uint16
	DataPath   string
}

//RPCConfig is the config for RPC Server.
type RPCConfig struct {
	JSONPort int
	GRPCPort int
}

// FileLogConfig is the config for filewriter of ilog.
type FileLogConfig struct {
	Path   string
	Level  string
	Enable bool
}

// ConsoleLogConfig is the config for consolewriter of ilog.
type ConsoleLogConfig struct {
	Level  string
	Enable bool
}

// LogConfig is the config of ilog.
type LogConfig struct {
	FileLog    *FileLogConfig
	ConsoleLog *ConsoleLogConfig
	AsyncWrite bool
}

// MetricsConfig is the config of metrics.
type MetricsConfig struct {
	PushAddr string
	Enable   bool
	ID       string
}

// DebugConfig is the config of debug.
type DebugConfig struct {
	ListenAddr string
}

// Config provide all configuration for the application
type Config struct {
	ACC     *ACCConfig
	Genesis *GenesisConfig
	VM      *VMConfig
	DB      *DBConfig
	P2P     *P2PConfig
	RPC     *RPCConfig
	Log     *LogConfig
	Metrics *MetricsConfig
	Debug   *DebugConfig
}

// NewConfig returns a new instance of Config
func NewConfig(configfile string) *Config {
	v := viper.GetViper()
	v.SetConfigType("yaml")

	f, err := os.Open(configfile)
	if err != nil {
		ilog.Fatalf("Failed to open config file '%v', %v", configfile, err)
	}

	if err := v.ReadConfig(f); err != nil {
		ilog.Fatalf("Failed to read config file: %v", err)
	}

	c := &Config{}
	if err := v.Unmarshal(c); err != nil {
		ilog.Fatalf("Unable to decode into struct, %v", err)
	}

	return c
}

func (c *Config) YamlString() string {
	bs, err := yaml.Marshal(c)
	if err != nil {
		ilog.Fatalf("Unable to marshal config to YAML: %v", err)
	}
	return string(bs)
}
