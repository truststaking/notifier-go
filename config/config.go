package config

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/multiversx/mx-chain-core-go/core"
)

// GeneralConfig defines the config setup based on main config file
type GeneralConfig struct {
	ConnectorApi ConnectorApiConfig
	Redis        RedisConfig
	AzureVault        string
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
	AzureCredentials        string
	EventsExchange          RabbitMQExchangeConfig
	RevertEventsExchange    RabbitMQExchangeConfig
	FinalizedEventsExchange RabbitMQExchangeConfig
	BlockTxsExchange        RabbitMQExchangeConfig
	BlockScrsExchange       RabbitMQExchangeConfig
}

// RabbitMQExchangeConfig holds the configuration for a rabbitMQ exchange
type RabbitMQExchangeConfig struct {
	Name string
	Type string
}

// FlagsConfig holds the values for CLI flags
type FlagsConfig struct {
	LogLevel          string
	SaveLogFile       bool
	GeneralConfigPath string
	WorkingDir        string
	APIType           string
}

// LoadConfig return a GeneralConfig instance by reading the provided toml file
func LoadConfig(filePath string) (*GeneralConfig, error) {

	cfg := &GeneralConfig{}
	err := core.LoadTomlFile(cfg, filePath)
	if err != nil {
		return nil, err
	}
	fmt.Println(cfg)
	keyVaultUrl := fmt.Sprintf("https://%s.vault.azure.net/", cfg.AzureVault)
	fmt.Println(keyVaultUrl)
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	client, err := azsecrets.NewClient(keyVaultUrl, cred, nil)
	if err != nil {
		return nil, err
	}

	rabbitURL, err := client.GetSecret(context.TODO(), "RabbitMqConnectionString", nil)
	if err != nil {
		return nil, err
	}

	nodesUsername, err := client.GetSecret(context.TODO(), "SquadNotifierUsername", nil)
	if err != nil {
		return nil, err
	}

	nodesPassword, err := client.GetSecret(context.TODO(), "SquadNotifierPassword", nil)
	if err != nil {
		return nil, err
	}

	redisURL, err := client.GetSecret(context.TODO(), "NotifierRedisURL", nil)
	if err != nil {
		return nil, err
	}

	serviceBus, err := client.GetSecret(context.TODO(), "ServiceBusConnectionString", nil)
	if err != nil {
		return nil, err
	}

	cfg.ConnectorApi.Username = *nodesUsername.Value
	cfg.ConnectorApi.Password = *nodesPassword.Value
	cfg.RabbitMQ.Url = *rabbitURL.Value
	cfg.Redis.Url = *redisURL.Value
	cfg.RabbitMQ.AzureCredentials = *serviceBus.Value
	return cfg, err
}
