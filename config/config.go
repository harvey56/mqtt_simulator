package config

import (
	"encoding/json"
	"fmt"
	"os"

	viper_config "github.com/spf13/viper"
)

var (
	Main *Config
)

const main_config_filename string = "mqtt_sender_config"

const main_config_data string = "mqtt_sender_data"

func Config_Initialization() {
	Main = newConfig()
}

type Config struct {
	General     General_Config
	MQTT        MQTT_Broker_Config
	Server_DB   DB_Config
	Main_config *viper_config.Viper
	MqttData    []map[string]interface{}
}

type General_Config struct {
	Broker_Address string `json:"broker_address" mapstructure:"broker_address"`
	Broker_Port    uint   `json:"broker_port" mapstructure:"broker_port"`
	Username       string `json:"username" mapstructure:"username"`
	Password       string `json:"password" mapstructure:"password"`
	DebugLevel     string `json:"debug_level" mapstructure:"debug_level"`
	LogFile        string `json:"log_file" mapstructure:"log_file"`
	Http_Port      uint   `json:"http_port" mapstructure:"http_port"`
}

type MQTT_Broker_Config struct {
	MQTT_Broker_Address  string         `json:"address" mapstructure:"address"`
	MQTT_Broker_Port     PortConfig     `json:"port"`
	MQTT_Broker_Protocol ProtocolConfig `json:"protocol"`
}

type PortConfig struct {
	HTTP  string `json:"http"`
	HTTPS string `json:"https"`
}

type ProtocolConfig struct {
	HTTP  string `json:"http"`
	HTTPS string `json:"https"`
}

type DB_Config struct {
	Server_Address string `json:"server_address" mapstructure:"server_address"`
	Server_Port    uint   `json:"server_port" mapstructure:"server_port"`
	Database_Name  string `json:"database_name" mapstructure:"database_name"`
	Username       string `json:"username" mapstructure:"username"`
	Password       string `json:"password" mapstructure:"password"`
}

func newConfig() *Config {
	config := &Config{}
	config.Main_config = viper_config.New()

	// Read all configuration information in main config
	config.Main_config.SetConfigName(main_config_filename)
	config.Main_config.AddConfigPath("./")
	config.Main_config.SetConfigType("json")

	// General
	config.Main_config.SetDefault("general", config.General)

	config.General.Broker_Address = "localhost"
	config.General.Broker_Port = 1883
	config.General.Username = ""
	config.General.Password = ""

	// MQTT
	config.Main_config.SetDefault("mqtt_broker", config.MQTT)

	config.MQTT.MQTT_Broker_Address = "localhost"
	config.MQTT.MQTT_Broker_Port.HTTP = "1885"
	config.MQTT.MQTT_Broker_Protocol.HTTP = "ws"

	// IoT Server DB

	config.Main_config.SetDefault("server_db", config.Server_DB)

	return config
}

func (config *Config) Close() {
	// Place here config close routine
}

func (config *Config) LoadConfig() {
	err := config.Main_config.SafeWriteConfig()
	if _, ok := err.(viper_config.ConfigFileAlreadyExistsError); err != nil && !ok {
		fmt.Printf("Error while checking %s.json file\n", main_config_filename)
	}

	err = config.Main_config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	dataBytes, err := os.ReadFile(main_config_data + ".json")
	if err != nil {
		panic(fmt.Errorf("fatal error reading data file %s.json: %w", main_config_data, err))
	}

	err = json.Unmarshal(dataBytes, &config.MqttData)
	if err != nil {
		panic(fmt.Errorf("fatal error unmarshalling data file %s.json: %w", main_config_data, err))
	}

	err = config.Main_config.UnmarshalKey("general", &config.General)
	if err != nil {
		fmt.Printf("Error while parsing %s.json file. Section: general\n", main_config_filename)
	}

	err = config.Main_config.UnmarshalKey("mqtt_broker", &config.MQTT)
	if err != nil {
		fmt.Printf("Error while parsing %s.json file. Section: mqtt\n", main_config_filename)
	}

	err = config.Main_config.UnmarshalKey("server_db", &config.Server_DB)
	if err != nil {
		fmt.Printf("Error while parsing %s.json file. Section: server_db\n", main_config_filename)
	}

}
