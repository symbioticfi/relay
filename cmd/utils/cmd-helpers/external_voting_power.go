package cmdhelpers

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/symbioticfi/relay/symbiotic/client/votingpower"
)

func ExternalVotingPowerProviderConfigs(
	configPath string,
	providers map[string]string,
) ([]votingpower.ProviderConfig, error) {
	if configPath != "" {
		v := viper.New()
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file %q: %w", configPath, err)
		}

		var cfg struct {
			ExternalVotingPowerProviders []votingpower.ProviderConfig `mapstructure:"external-voting-power-providers"`
		}
		if err := v.Unmarshal(&cfg); err != nil {
			return nil, fmt.Errorf("failed to parse external voting power providers from config: %w", err)
		}
		return cfg.ExternalVotingPowerProviders, nil
	}

	providerConfigs := make([]votingpower.ProviderConfig, 0, len(providers))
	for providerID, url := range providers {
		providerConfigs = append(providerConfigs, votingpower.ProviderConfig{
			ID:  providerID,
			URL: url,
		})
	}

	return providerConfigs, nil
}
