package votingpower

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"log/slog"
	"math/big"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-errors/errors"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"

	votingpowerv1 "github.com/symbioticfi/relay/internal/gen/votingpower/v1"
	"github.com/symbioticfi/relay/pkg/tracing"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

const (
	defaultTimeout          = 5 * time.Second
	methodGetVotingPowersAt = "GetVotingPowersAt"
)

type ProviderID [10]byte

// ProviderConfig describes one external voting power provider.
type ProviderConfig struct {
	ID         string            `mapstructure:"id"`
	URL        string            `mapstructure:"url"`
	Secure     bool              `mapstructure:"secure"`
	CACertFile string            `mapstructure:"ca-cert-file"`
	ServerName string            `mapstructure:"server-name"`
	Headers    map[string]string `mapstructure:"headers"`
	Timeout    time.Duration     `mapstructure:"timeout"`
}

// Client routes GetVotingPowers calls to configured external providers.
type Client struct {
	providers map[ProviderID]providerEntry
}

type providerEntry struct {
	conn   *grpc.ClientConn
	client votingpowerv1.VotingPowerProviderServiceClient
	cfg    ProviderConfig
}

// NewClient creates a new external voting power client and validates provider connectivity.
func NewClient(ctx context.Context, providerConfigs []ProviderConfig) (*Client, error) {
	c := &Client{
		providers: make(map[ProviderID]providerEntry, len(providerConfigs)),
	}
	orderedProviderIDs := make([]ProviderID, 0, len(providerConfigs))

	for _, cfg := range providerConfigs {
		providerID, err := ParseProviderID(cfg.ID)
		if err != nil {
			return nil, errors.Errorf("invalid provider id %q: %w", cfg.ID, err)
		}
		if cfg.URL == "" {
			return nil, errors.Errorf("provider %s url is required", providerIDString(providerID))
		}
		if _, exists := c.providers[providerID]; exists {
			return nil, errors.Errorf("duplicate provider id: %s", providerIDString(providerID))
		}
		c.providers[providerID] = providerEntry{cfg: cfg}
		orderedProviderIDs = append(orderedProviderIDs, providerID)
	}

	for _, providerID := range orderedProviderIDs {
		entry := c.providers[providerID]
		providerCfg := entry.cfg
		conn, err := dialProvider(providerCfg)
		if err != nil {
			c.closeAll()
			return nil, errors.Errorf("dial provider %s: %w", providerIDString(providerID), err)
		}

		healthClient := grpc_health_v1.NewHealthClient(conn)
		if err := checkProviderHealth(ctx, providerID, providerCfg, healthClient); err != nil {
			c.closeAll()
			return nil, errors.Errorf("health check provider %s: %w", providerIDString(providerID), err)
		}

		entry.conn = conn
		entry.client = votingpowerv1.NewVotingPowerProviderServiceClient(conn)
		c.providers[providerID] = entry
	}

	return c, nil
}

func dialProvider(cfg ProviderConfig) (*grpc.ClientConn, error) {
	dialOpts := []grpc.DialOption{}

	if cfg.Secure {
		tlsConfig, err := buildTLSConfig(cfg)
		if err != nil {
			return nil, err
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(cfg.URL, dialOpts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func buildTLSConfig(cfg ProviderConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if cfg.ServerName != "" {
		tlsConfig.ServerName = cfg.ServerName
	}

	if cfg.CACertFile == "" {
		return tlsConfig, nil
	}

	caPEM, err := os.ReadFile(cfg.CACertFile)
	if err != nil {
		return nil, errors.Errorf("read ca cert file: %w", err)
	}

	roots := x509.NewCertPool()
	if ok := roots.AppendCertsFromPEM(caPEM); !ok {
		return nil, errors.Errorf("invalid CA cert PEM in %s", cfg.CACertFile)
	}
	tlsConfig.RootCAs = roots
	return tlsConfig, nil
}

// GetVotingPowers fetches voting powers from the external provider identified by address prefix bytes [0:10].
func (c *Client) GetVotingPowers(
	ctx context.Context,
	address symbiotic.CrossChainAddress,
	timestamp symbiotic.Timestamp,
) ([]symbiotic.OperatorVotingPower, error) {
	providerID := providerIDFromAddress(address.Address)

	entry, ok := c.providers[providerID]
	if !ok {
		return nil, errors.Errorf("external provider %s is not configured", providerIDString(providerID))
	}
	cfg := entry.cfg

	reqCtx, span := tracing.StartClientSpan(ctx, "external_vp.get_voting_powers",
		attribute.String("provider.id", providerIDString(providerID)),
		attribute.Int64("timestamp", int64(timestamp)),
	)
	defer span.End()

	resp, err := c.fetchVotingPowers(reqCtx, providerID, cfg, &votingpowerv1.GetVotingPowersAtRequest{Timestamp: uint64(timestamp)})
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	result, err := convertVotingPowers(address, resp)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	return result, nil
}

func (c *Client) fetchVotingPowers(
	ctx context.Context,
	providerID ProviderID,
	cfg ProviderConfig,
	req *votingpowerv1.GetVotingPowersAtRequest,
) (*votingpowerv1.GetVotingPowersAtResponse, error) {
	entry, ok := c.providers[providerID]
	if !ok {
		return nil, errors.Errorf("external provider %s is not configured", providerIDString(providerID))
	}
	client := entry.client
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if len(cfg.Headers) > 0 {
		callCtx = metadata.NewOutgoingContext(callCtx, metadata.New(cfg.Headers))
	}

	resp, err := client.GetVotingPowersAt(callCtx, req)
	if err != nil {
		return nil, errors.Errorf("external provider %s %s failed: %w", providerIDString(providerID), methodGetVotingPowersAt, err)
	}
	return resp, nil
}

func checkProviderHealth(
	ctx context.Context,
	providerID ProviderID,
	cfg ProviderConfig,
	healthClient grpc_health_v1.HealthClient,
) error {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	hcCtx, span := tracing.StartClientSpan(ctx, "external_vp.health_check",
		attribute.String("provider.id", providerIDString(providerID)),
	)
	defer span.End()

	callCtx, cancel := context.WithTimeout(hcCtx, timeout)
	defer cancel()

	resp, err := healthClient.Check(callCtx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		tracing.RecordError(span, err)
		return err
	}

	if resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		err := errors.Errorf("provider not serving: %s", resp.GetStatus().String())
		tracing.RecordError(span, err)
		return err
	}
	return nil
}

func convertVotingPowers(
	providerAddress symbiotic.CrossChainAddress,
	resp *votingpowerv1.GetVotingPowersAtResponse,
) ([]symbiotic.OperatorVotingPower, error) {
	aggregated := make(map[common.Address]*big.Int, len(resp.GetVotingPowers()))

	for _, vp := range resp.GetVotingPowers() {
		if !common.IsHexAddress(vp.GetOperator()) {
			return nil, errors.Errorf("invalid operator address: %q", vp.GetOperator())
		}
		operator := common.HexToAddress(vp.GetOperator())

		value, ok := new(big.Int).SetString(vp.GetVotingPower(), 10)
		if !ok {
			return nil, errors.Errorf("invalid voting power for operator %s: %q", operator.Hex(), vp.GetVotingPower())
		}
		if value.Sign() < 0 {
			return nil, errors.Errorf("negative voting power for operator %s", operator.Hex())
		}

		if _, exists := aggregated[operator]; !exists {
			aggregated[operator] = big.NewInt(0)
		}
		aggregated[operator].Add(aggregated[operator], value)
	}

	operators := make([]common.Address, 0, len(aggregated))
	for operator := range aggregated {
		operators = append(operators, operator)
	}
	slices.SortFunc(operators, func(a, b common.Address) int {
		return a.Cmp(b)
	})

	result := make([]symbiotic.OperatorVotingPower, 0, len(operators))
	for _, operator := range operators {
		result = append(result, symbiotic.OperatorVotingPower{
			Operator: operator,
			Vaults: []symbiotic.VaultVotingPower{
				{
					Vault:       providerAddress.Address,
					VotingPower: symbiotic.ToVotingPower(new(big.Int).Set(aggregated[operator])),
				},
			},
		})
	}

	return result, nil
}

func providerIDFromAddress(addr common.Address) ProviderID {
	var providerID ProviderID
	copy(providerID[:], addr[:10])
	return providerID
}

func providerIDString(id ProviderID) string {
	return "0x" + hex.EncodeToString(id[:])
}

func ParseProviderID(input string) (ProviderID, error) {
	trimmed := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(input), "0x"), "0X")
	if len(trimmed) != 20 {
		return ProviderID{}, errors.Errorf("provider id must be 10 bytes (20 hex chars), got %d", len(trimmed))
	}

	decoded, err := hex.DecodeString(trimmed)
	if err != nil {
		return ProviderID{}, errors.Errorf("invalid hex: %w", err)
	}

	var providerID ProviderID
	copy(providerID[:], decoded)
	return providerID, nil
}

func (c *Client) closeAll() {
	for _, provider := range c.providers {
		if provider.conn != nil {
			_ = provider.conn.Close()
		}
	}
}

// Close closes all provider connections.
func (c *Client) Close() error {
	var firstErr error
	for providerID, provider := range c.providers {
		if provider.conn == nil {
			continue
		}
		if err := provider.conn.Close(); err != nil {
			slog.Warn("failed to close external voting power provider connection", "providerId", providerIDString(providerID), "error", err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
