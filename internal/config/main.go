package config

import (
	"encoding/json"
	"fmt"
	"instagram-stats/internal/stats"
	"os"
)

type InstagramStatsConfig struct {
	CollectConfig *stats.ConversationRetrievalConfig `json:"collect_config"`
	OutputConfig  *stats.OutputConfig                `json:"output_config"`
}

/*
 * GetConfig reads the configuration file from the specified path and unmarshals it into the InstagramStatsConfig struct.
 */
func GetConfig(configPath string) (*InstagramStatsConfig, error) {
	config := &InstagramStatsConfig{}
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}
	if config.CollectConfig == nil {
		return nil, fmt.Errorf("missing collect_config in config file")
	} else if config.OutputConfig == nil {
		return nil, fmt.Errorf("missing output_config in config file")
	}

	err = config.CollectConfig.CheckConfig()

	return config, err
}
