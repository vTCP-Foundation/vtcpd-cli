package conf

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/viper"
)

const (
	ConfigurationFileName = "cli-conf"
)

var (
	ErrConfigNotFound = errors.New("configuration file not found")
	ErrConfigRead     = errors.New("can't read configuration")
)

func LoadConfiguration() error {
	viper.SetConfigName(ConfigurationFileName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/vtcpd/")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("can't read configuration file `%s`: %w", ConfigurationFileName, ErrConfigNotFound)
	}

	return nil
}

func HTTPInterfaceEnabled() bool {
	return viper.GetBool("interface.http.enabled")
}

func HTTPInterface() string {
	return viper.GetString("interface.http.host") + ":" + strconv.Itoa(int(viper.GetUint16("interface.http.port")))
}

func CommandsFifoPath() string {
	return viper.GetString("vtcpd.commands_fifo_path")
}

func ResultsFifoPath() string {
	return viper.GetString("vtcpd.results_fifo_path")
}

func LogLevel() string {
	return viper.GetString("log.level")
}
