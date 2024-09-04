package utils

import (
	"fmt"
	"strings"

	"github.com/gookit/slog"
	"github.com/spf13/viper"
)

type MQTT struct {
	ClientId string `mapstructure:"client_id"`
	Endpoint string `mapstructure:"endpoint"`
	Protocol string `mapstructure:"protocol"`
	Port     int    `mapstructure:"port"`
	CaFile   string `mapstructure:"ca_file"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
	TopicIn  string `mapstructure:"topic_in"`
	TopicOut string `mapstructure:"topic_out"`
}

type UDP struct {
	IpIn    string `mapstructure:"ip_in"`
	IpOut   string `mapstructure:"ip_out"`
	PortIn  int    `mapstructure:"port_in"`
	PortOut int    `mapstructure:"port_out"`
}

// Config represents the application configuration.
type Config struct {
	MQTT     MQTT   `mapstructure:"mqtt"`
	UDP      UDP    `mapstructure:"udp"`
	LogLevel string `mapstructure:"log_level"`
}

func LoadConfig() (*Config, error) {
	config := &Config{}

	v := viper.New()

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	v.SetEnvPrefix("bridge")

	v.SetConfigName("config") // name of config file (without extension)
	v.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name

	v.AddConfigPath("/etc/udp-mqtt-bridge")             // path to look for the config file in
	v.AddConfigPath("$HOME/.udp-mqtt-bridge")           // call multiple times to add many search paths
	v.AddConfigPath("$XDG_CONFIG_HOME/udp-mqtt-bridge") // call multiple times to add many search paths
	v.AddConfigPath("./configs")                        // optionally look for config in the working directory
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return config, err
		}
	}
	slog.Debugf("Using config file: %s", v.ConfigFileUsed())

	bcast, err := GetBroadcastAddress()
	if err != nil {
		return config, fmt.Errorf("error getting broadcast address: %v", err)
	}

	v.SetDefault("mqtt.client_id", "udp-mqtt-bridge")
	v.SetDefault("log_level", "INFO")
	v.SetDefault("mqtt.protocol", "SSL")
	v.SetDefault("mqtt.port", 8883)
	v.SetDefault("mqtt.topic_in", "topic/in")
	v.SetDefault("mqtt.topic_out", "topic/out")
	v.SetDefault("udp.ip_in", "0.0.0.0")
	v.SetDefault("udp.ip_out", bcast)
	v.SetDefault("udp.port_in", 6000)
	v.SetDefault("udp.port_out", 6001)

	err = v.Unmarshal(config)
	if err != nil {
		return config, fmt.Errorf("unable to decode into config struct, %v", err)
	}

	return config, nil
}
