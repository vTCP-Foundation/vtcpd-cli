package conf

import (
	"fmt"
	"strconv"

	"github.com/spf13/viper"
)

type HTTPSettings struct {
	Host string `mapstructure:"host"`
	Port uint16 `mapstructure:"port"`
}

type SecuritySettings struct {
	ApiKey       string   `mapstructure:"api_key"`
	AllowableIPs []string `mapstructure:"allowable_ips"`
}

type Settings struct {
	WorkDir     string           `mapstructure:"workdir"`
	VTCPDPath   string           `mapstructure:"vtcpd_path"`
	HTTP        HTTPSettings     `mapstructure:"http"`
	HTTPTesting HTTPSettings     `mapstructure:"http_testing"`
	Security    SecuritySettings `mapstructure:"security"`
}

func (s HTTPSettings) HTTPInterface() string {
	return s.Host + ":" + strconv.Itoa(int(s.Port))
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
