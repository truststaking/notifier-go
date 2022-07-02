package config

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

// GeneralConfig defines the config setup based on main config file
type GeneralConfig struct {
	ConnectorApi ConnectorApiConfig
	Redis        RedisConfig
	RabbitMQ     RabbitMQConfig
	Flags        *FlagsConfig
}

// ConnectorApiConfig maps the connector configuration
type ConnectorApiConfig struct {
	Port            string
	Username        string
	Password        string
	CheckDuplicates bool
}

// RedisConfig maps the redis configuration
type RedisConfig struct {
	Url            string
	Channel        string
	MasterName     string
	SentinelUrl    string
	ConnectionType string
	TTL            uint32
}

// RabbitMQConfig maps the rabbitMQ configuration
type RabbitMQConfig struct {
	Url                     string
	EventsExchange          string
	RevertEventsExchange    string
	FinalizedEventsExchange string
	BlockTxsExchange        string
	BlockScrsExchange       string
}

// FlagsConfig holds the values for CLI flags
type FlagsConfig struct {
	LogLevel          string
	SaveLogFile       bool
	GeneralConfigPath string
	WorkingDir        string
	APIType           string
}

var log = logger.GetOrCreate("eventNotifier")

// LoadConfig return a GeneralConfig instance by reading the provided toml file
func LoadConfig(filePath string) (*GeneralConfig, error) {

	cfg := &GeneralConfig{}
	err := core.LoadTomlFile(cfg, filePath)
	if err != nil {
		return nil, err
	}
	keyVaultUrl := fmt.Sprintf("https://%s.vault.azure.net/", "TrustMarketVault")

	log.Info(keyVaultUrl)
	// cred, err := azidentity.NewDefaultAzureCredential(nil)
	// if err != nil {
	// 	return nil, err
	// }
	log.Info("After cred")
	// client, err := azsecrets.NewClient(keyVaultUrl, cred, nil)
	// if err != nil {
	// 	return nil, err
	// }

	log.Info("Client created")
	// rabbitURL, err := client.GetSecret(context.TODO(), "RabbitMqConnectionString", nil)
	// if err != nil {
	// 	log.Error("failed to get the secret: %v", err)
	// 	return nil, err
	// }

	// log.Info("After url")
	// nodesUsername, err := client.GetSecret(context.TODO(), "SquadNotifierUsername", nil)
	// if err != nil {
	// 	return nil, err
	// }

	// nodesPassword, err := client.GetSecret(context.TODO(), "SquadNotifierPassword", nil)
	// if err != nil {
	// 	return nil, err
	// }

	// redisURL, err := client.GetSecret(context.TODO(), "NotifierRedisURL", nil)
	// if err != nil {
	// 	return nil, err
	// }

	cfg.ConnectorApi.Username = "xoxnouser"
	cfg.ConnectorApi.Password = "bG5atWZqkiEAxWS"
	// cfg.Redis.Url = "rediss://:sD2VfqU37NdUl7AoA4H9juXt9sFw6YhDNAzCaJr3vbA=@trustmarket.redis.cache.windows.net:6380/0"
	cfg.RabbitMQ.Url = "amqp://user:IExbuMmsviBP2@trustrabbitmqkdemcstsrkh42-vm0.westeurope.cloudapp.azure.com:5672"
	return cfg, err
}
