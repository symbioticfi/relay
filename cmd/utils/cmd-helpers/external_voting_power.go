package cmdhelpers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/symbioticfi/relay/symbiotic/client/votingpower"
)

func ExternalVotingPowerProviderConfigs(
	entries []string,
) ([]votingpower.ProviderConfig, error) {
	providerConfigs := make([]votingpower.ProviderConfig, 0, len(entries))
	for i, entry := range entries {
		cfg, err := parseProviderConfigEntry(entry)
		if err != nil {
			return nil, fmt.Errorf("invalid external voting power provider entry %d: %w", i+1, err)
		}
		providerConfigs = append(providerConfigs, cfg)
	}

	return providerConfigs, nil
}

func parseProviderConfigEntry(entry string) (votingpower.ProviderConfig, error) {
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return votingpower.ProviderConfig{}, fmt.Errorf("entry is empty")
	}

	cfg := votingpower.ProviderConfig{}
	seen := make(map[string]struct{})

	parts := strings.Split(entry, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return votingpower.ProviderConfig{}, fmt.Errorf("invalid key=value pair %q", part)
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		if key == "" {
			return votingpower.ProviderConfig{}, fmt.Errorf("field name is empty")
		}
		if _, ok := seen[key]; ok {
			return votingpower.ProviderConfig{}, fmt.Errorf("duplicate field %q", key)
		}
		seen[key] = struct{}{}

		switch key {
		case "id":
			cfg.ID = value
		case "url":
			cfg.URL = value
		case "secure":
			secure, err := strconv.ParseBool(value)
			if err != nil {
				return votingpower.ProviderConfig{}, fmt.Errorf("invalid secure value %q: %w", value, err)
			}
			cfg.Secure = secure
		case "ca-cert-file":
			cfg.CACertFile = value
		case "server-name":
			cfg.ServerName = value
		case "timeout":
			timeout, err := time.ParseDuration(value)
			if err != nil {
				return votingpower.ProviderConfig{}, fmt.Errorf("invalid timeout value %q: %w", value, err)
			}
			cfg.Timeout = timeout
		case "headers":
			headers, err := parseHeaders(value)
			if err != nil {
				return votingpower.ProviderConfig{}, fmt.Errorf("invalid headers value %q: %w", value, err)
			}
			cfg.Headers = headers
		default:
			return votingpower.ProviderConfig{}, fmt.Errorf("unknown field %q", key)
		}
	}

	if strings.TrimSpace(cfg.ID) == "" {
		return votingpower.ProviderConfig{}, fmt.Errorf("id is required")
	}
	if strings.TrimSpace(cfg.URL) == "" {
		return votingpower.ProviderConfig{}, fmt.Errorf("url is required")
	}

	return cfg, nil
}

func parseHeaders(raw string) (map[string]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]string{}, nil
	}

	headers := make(map[string]string)
	items := strings.Split(raw, "|")
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		kv := strings.SplitN(item, ":", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("expected key:value, got %q", item)
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		if key == "" {
			return nil, fmt.Errorf("header key is empty")
		}
		headers[key] = value
	}

	return headers, nil
}
