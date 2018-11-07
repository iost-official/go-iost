package common

import (
	"os"

	"github.com/iost-official/go-iost/ilog"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// ACCConfig account of the system
type ACCConfig struct {
	ID        string
	SecKey    string
	Algorithm string
}

// Witness config of the genesis block
type Witness struct {
	ID      string
	Owner   string
	Active  string
	Balance int64
}

// GenesisConfig config of the genesis block
type GenesisConfig struct {
	CreateGenesis    bool
	InitialTimestamp string
	WitnessInfo      []*Witness
	ContractPath     string
	AdminInfo        *Witness
}

// DBConfig config of the database
type DBConfig struct {
	LdbPath string
}

// VMConfig config of the v8vm
type VMConfig struct {
	JsPath   string
	LogLevel string
}

// P2PConfig is the config for p2p network.
type P2PConfig struct {
	ListenAddr   string
	SeedNodes    []string
	ChainID      uint32
	Version      uint16
	DataPath     string
	InboundConn  int
	OutboundConn int
	BlackPID     []string
	BlackIP      []string
	AdminPort    string
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
	Username string
	Password string
	Enable   bool
	ID       string
}

// DebugConfig is the config of debug.
type DebugConfig struct {
	ListenAddr string
}

// VersionConfig contrains nettype(mainnet / testnet etc) and protocol info
type VersionConfig struct {
	NetType         string
	ProtocolVersion string
}

// Config provide all configuration for the application
type Config struct {
	ACC     *ACCConfig
	Genesis string
	VM      *VMConfig
	DB      *DBConfig
	P2P     *P2PConfig
	RPC     *RPCConfig
	Log     *LogConfig
	Metrics *MetricsConfig
	Debug   *DebugConfig
	Version *VersionConfig
}

// LoadYamlAsViper load yaml file as viper object
func LoadYamlAsViper(configfile string) *viper.Viper {
	v := viper.GetViper()
	v.SetConfigType("yaml")

	f, err := os.Open(configfile)
	if err != nil {
		ilog.Fatalf("Failed to open config file '%v', %v", configfile, err)
	}

	if err := v.ReadConfig(f); err != nil {
		ilog.Fatalf("Failed to read config file: %v", err)
	}

	return v
}

// NewConfig returns a new instance of Config
func NewConfig(configfile string) *Config {
	v := LoadYamlAsViper(configfile)
	c := &Config{}
	if err := v.Unmarshal(c); err != nil {
		ilog.Fatalf("Unable to decode into struct, %v", err)
	}

	return c
}

// YamlString config to string
func (c *Config) YamlString() string {
	bs, err := yaml.Marshal(c)
	if err != nil {
		ilog.Fatalf("Unable to marshal config to YAML: %v", err)
	}
	return string(bs)
}
