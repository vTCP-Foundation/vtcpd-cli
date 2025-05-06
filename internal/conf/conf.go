package conf

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
)

type HandlerSettings struct {
	NodeDirPath              string `json:"node_path"`
	ClientExecutableFullPath string `json:"executable_path"`
	HTTPInterfaceHost        string `json:"http_host"`
	HTTPInterfacePort        uint16 `json:"http_port"`
}

type SecuritySettings struct {
	ApiKey       string   `json:"api_key"`
	AllowableIPs []string `json:"allowable_ips"`
}

type Settings struct {
	Handler  HandlerSettings  `json:"handler"`
	Security SecuritySettings `json:"security"`
}

func (s HandlerSettings) HTTPInterface() string {
	return s.HTTPInterfaceHost + ":" + strconv.Itoa(int(s.HTTPInterfacePort))
}

var (
	Params = Settings{}
)

func LoadSettings() error {
	// Reading configuration file
	bytes, err := os.ReadFile("conf.json")
	if err != nil {
		return errors.New("Can't read configuration. Details: " + err.Error())
	}

	err = json.Unmarshal(bytes, &Params)
	if err != nil {
		return errors.New("Can't read configuration. Details: " + err.Error())
	}

	return nil
}
