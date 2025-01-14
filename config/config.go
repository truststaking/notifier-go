package config

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/multiversx/mx-chain-core-go/core"
)

// Configs holds all configs
type Configs struct {
	GeneralConfig   GeneralConfig
	ApiRoutesConfig APIRoutesConfig
	Flags           FlagsConfig
}

// GeneralConfig defines the config setup based on main config file
type GeneralConfig struct {
	ConnectorApi ConnectorApiConfig
	Redis        RedisConfig
	Azure        AzureConfig
	RabbitMQ     RabbitMQConfig
}

// ConnectorApiConfig maps the connector configuration
type ConnectorApiConfig struct {
	Host            string
	Username        string
	Password        string
	CheckDuplicates bool
}

type AzureConfig struct {
	KeyVault string
	Topic    string
}

// APIRoutesConfig holds the configuration related to Rest API routes
type APIRoutesConfig struct {
	APIPackages map[string]APIPackageConfig
}

// APIPackageConfig holds the configuration for the routes of each package
type APIPackageConfig struct {
	Routes []RouteConfig
}

// RouteConfig holds the configuration for a single route
type RouteConfig struct {
	Name string
	Open bool
	Auth bool
}

// RedisConfig maps the redis configuration
type RedisConfig struct {
	Url            string
	MasterName     string
	SentinelUrl    string
	ConnectionType string
	TTL            uint32
}

// RabbitMQConfig maps the rabbitMQ configuration
type RabbitMQConfig struct {
	Url                     string
	AzureCredentials        string
	Topic                   string
	EventsExchange          RabbitMQExchangeConfig
	RevertEventsExchange    RabbitMQExchangeConfig
	FinalizedEventsExchange RabbitMQExchangeConfig
	BlockTxsExchange        RabbitMQExchangeConfig
	BlockScrsExchange       RabbitMQExchangeConfig
	BlockEventsExchange     RabbitMQExchangeConfig
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
	APIConfigPath     string
	WorkingDir        string
	APIType           string
	RestApiInterface  string
}

// LoadGeneralConfig returns a GeneralConfig instance by reading the provided toml file
func LoadGeneralConfig(filePath string) (*GeneralConfig, error) {
	cfg := &GeneralConfig{}
	err := core.LoadTomlFile(cfg, filePath)
	if err != nil {
		return nil, err
	}
	keyVaultUrl := fmt.Sprintf("https://%s.vault.azure.net/", cfg.Azure.KeyVault)
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	client, err := azsecrets.NewClient(keyVaultUrl, cred, nil)
	if err != nil {
		return nil, err
	}

	rabbitURL, err := client.GetSecret(context.TODO(), "RabbitMqConnectionString", "", nil)
	if err != nil {
		return nil, err
	}

	nodesUsername, err := client.GetSecret(context.TODO(), "SquadNotifierUsername", "", nil)
	if err != nil {
		return nil, err
	}

	nodesPassword, err := client.GetSecret(context.TODO(), "SquadNotifierPassword", "", nil)
	if err != nil {
		return nil, err
	}

	redisURL, err := client.GetSecret(context.TODO(), "NotifierRedisURL", "", nil)
	if err != nil {
		return nil, err
	}

	serviceBus, err := client.GetSecret(context.TODO(), "ServiceBusConnectionString", "", nil)
	if err != nil {
		return nil, err
	}

	cfg.ConnectorApi.Username = *nodesUsername.Value
	cfg.ConnectorApi.Password = *nodesPassword.Value
	cfg.RabbitMQ.Url = *rabbitURL.Value
	cfg.Redis.Url = *redisURL.Value
	cfg.RabbitMQ.AzureCredentials = *serviceBus.Value
	cfg.RabbitMQ.Topic = cfg.Azure.Topic
	return cfg, err
}

// LoadAPIConfig returns a APIRoutesConfig instance by reading the provided toml file
func LoadAPIConfig(filePath string) (*APIRoutesConfig, error) {
	cfg := &APIRoutesConfig{}
	err := core.LoadTomlFile(cfg, filePath)
	if err != nil {
		return nil, err
	}
	return cfg, err
}
