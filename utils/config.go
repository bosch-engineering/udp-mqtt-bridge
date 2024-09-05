package utils

import (
	"fmt"
	"os"
	"path/filepath"
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
	MQTT         MQTT   `mapstructure:"mqtt"`
	UDP          UDP    `mapstructure:"udp"`
	LogLevel     string `mapstructure:"log_level"`
	ConfigFolder string
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

	v.Set("title", "udp-mqtt-bridge")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return config, err
		}
	}
	config.ConfigFolder = filepath.Dir(v.ConfigFileUsed())
	slog.Debugf("Using config folder: %s", config.ConfigFolder)

	bcast, err := GetBroadcastAddress()
	if err != nil {
		return config, fmt.Errorf("error getting broadcast address: %v", err)
	}

	hostname := os.Getenv("HOSTNAME")
	if hostname == "" {
		hostname, _ = os.Hostname()
	}
	if hostname == "" {
		hostname = "udp-mqtt-bridge"
	}

	v.SetDefault("log_level", "INFO")
	v.SetDefault("mqtt.ca_file", "certs/AmazonRootCA1.pem")
	v.SetDefault("mqtt.cert_file", "certs/certificate.pem.crt")
	v.SetDefault("mqtt.client_id", hostname)
	v.SetDefault("mqtt.key_file", "certs/private.pem.key")
	v.SetDefault("mqtt.port", 8883)
	v.SetDefault("mqtt.protocol", "SSL")
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
