package lib

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv          string
	BrokerAddresses []string
	LogFilePath     string
	LogLevel        string
	MessageLimit    int
	SleepTimeout    time.Duration
	TopicName       string
}

// Helper
// Converts string to a map
func convertStrToMap(secretStr string) map[string]string {
	lines := strings.Split(secretStr, "\n")
	newMap := make(map[string]string)

	// Parse each line and extract key-value pairs
	for _, line := range lines {
		// skip empty lines
		if line == "" {
			continue
		}

		// split line into key and value by '='
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			panic(fmt.Sprintln("Malformed line:", line))
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		newMap[key] = value
	}

	return newMap
}

// Load env vars via docker swarm - https://docs.docker.com/engine/swarm/
func loadDockerEnvConfig(envConfigPath string) Config {
	fmt.Print("Loading docker config via docker swarm.\n")

	secretData, err := os.ReadFile(envConfigPath)
	if err != nil {
		panic(fmt.Sprintf("\nStack enabled? Error: %v", err))
	}

	// convert to map
	mapSecrets := convertStrToMap(string(secretData))

	return Config{
		AppEnv:          mapSecrets["APP_ENV"],
		BrokerAddresses: strings.Split(mapSecrets["DOCKER_BROKER_ADDRESSES"], ","),
		LogLevel:        mapSecrets["LOG_LEVEL"],
		LogFilePath:     mapSecrets["LOG_FILE_PATH"],
		MessageLimit:    ConvertStrToInt(mapSecrets["MESSAGE_LIMIT"]),
		SleepTimeout:    time.Duration(ConvertStrToInt(mapSecrets["SLEEP_TIMEOUT"])) * time.Millisecond,
		TopicName:       mapSecrets["TOPIC_NAME"],
	}
}

// Load env vars from local .env file
func loadLocalEnvConfig() Config {
	fmt.Print("Loading local .env file.\n")

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	return Config{
		AppEnv:          os.Getenv("APP_ENV"),
		BrokerAddresses: strings.Split(os.Getenv("LOCAL_BROKER_ADDRESSES"), ","),
		LogLevel:        os.Getenv("LOG_LEVEL"),
		LogFilePath:     os.Getenv("LOG_FILE_PATH"),
		MessageLimit:    ConvertStrToInt(os.Getenv("MESSAGE_LIMIT")),
		SleepTimeout:    time.Duration(ConvertStrToInt(os.Getenv("SLEEP_TIMEOUT"))) * time.Millisecond,
		TopicName:       os.Getenv("TOPIC_NAME"),
	}
}

// Validates config values are not empty
// Helps maintain consistency for local and docker config
func validateConfig(config Config) {
	value := reflect.ValueOf(config)
	for i := 0; i < value.NumField(); i++ {
		if value.Field(i).IsZero() {
			typ := reflect.TypeOf(config)
			panic(fmt.Sprintln("Invalid value:", typ.Field(i).Name))
		}
	}
}

// Allows program to determine whether it is being ran locally or within a docker container
// It then loads environment variables
func LoadConfig() Config {
	var envConfig Config

	envConfigPath := ".env"
	_, envFileErr := os.ReadFile(envConfigPath)
	if envFileErr != nil {
		const dockerSwarmSecretsPath = "/run/secrets/kafka-producer-secrets"
		envConfig = loadDockerEnvConfig(dockerSwarmSecretsPath)
	} else {
		envConfig = loadLocalEnvConfig()
	}

	validateConfig(envConfig)

	return envConfig
}
