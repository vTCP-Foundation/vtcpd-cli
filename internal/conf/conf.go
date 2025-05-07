package conf

import (
	"fmt"
	"strconv"

	"github.com/spf13/viper"
)

type HandlerSettings struct {
	NodeDirPath              string `mapstructure:"node_path"`
	ClientExecutableFullPath string `mapstructure:"executable_path"`
	HTTPInterfaceHost        string `mapstructure:"http_host"`
	HTTPInterfacePort        uint16 `mapstructure:"http_port"`
}

type TestHandlerSettings struct {
	HTTPInterfaceHost string `mapstructure:"http_host"`
	HTTPInterfacePort uint16 `mapstructure:"http_port"`
}

type SecuritySettings struct {
	ApiKey       string   `mapstructure:"api_key"`
	AllowableIPs []string `mapstructure:"allowable_ips"`
}

type Settings struct {
	Handler     HandlerSettings     `mapstructure:"handler"`
	Security    SecuritySettings    `mapstructure:"security"`
	TestHandler TestHandlerSettings `mapstructure:"test_handler"`
}

func (s HandlerSettings) HTTPInterface() string {
	return s.HTTPInterfaceHost + ":" + strconv.Itoa(int(s.HTTPInterfacePort))
}

func (s TestHandlerSettings) HTTPInterface() string {
	return s.HTTPInterfaceHost + ":" + strconv.Itoa(int(s.HTTPInterfacePort))
}

var (
	Params = Settings{}
)

func LoadSettings() error {
	v := viper.New()
	v.SetConfigName("conf") // configuration file without extension
	v.SetConfigType("yaml") // configuration type
	v.AddConfigPath(".")    // path to the configuration file

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := v.Unmarshal(&Params); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}
