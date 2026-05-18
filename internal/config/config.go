package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Port           string   `mapstructure:"PORT"`
	DBURL          string   `mapstructure:"DATABASE_URL"`
	KafkaBrokers   []string `mapstructure:"KAFKA_BROKERS"`
	KafkaTopic     string   `mapstructure:"KAFKA_TOPIC"`
	KafkaGroupID   string   `mapstructure:"KAFKA_GROUP_ID"`
	AzureAPIKey    string   `mapstructure:"AZURE_API_KEY"`
	AzureEndpoint  string   `mapstructure:"AZURE_ENDPOINT"`
	GCPProjectID   string   `mapstructure:"GCP_PROJECT_ID"`
	GCPLocation    string   `mapstructure:"GCP_LOCATION"`
	GCPProcessorID string   `mapstructure:"GCP_PROCESSOR_ID"`
	GeminiAPIKey   string   `mapstructure:"GEMINI_API_KEY"`
	WorkerCount    int      `mapstructure:"WORKER_COUNT"`
}

func LoadConfig() (*Config, error) {
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("WORKER_COUNT", 10)
	viper.SetDefault("KAFKA_TOPIC", "extraction_jobs")
	viper.SetDefault("KAFKA_GROUP_ID", "erad-ai-service")
	
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if .env is missing in production
	}
	
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}
	
	// Handle string slice for KafkaBrokers if provided as comma-separated env var
	// Viper usually handles this but sometimes manual check is better
	
	return &config, nil
}
