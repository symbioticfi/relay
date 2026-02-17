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
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	votingpowerv1 "github.com/symbioticfi/relay/internal/gen/votingpower/v1"
	symbiotic "github.com/symbioticfi/relay/symbiotic/entity"
)

const (
	defaultTimeout = 5 * time.Second
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
	providers map[ProviderID]provider
}

type provider struct {
	cfg    ProviderConfig
	conn   *grpc.ClientConn
	client votingpowerv1.VotingPowerProviderServiceClient
}

// NewClient creates a new external voting power client and validates provider connectivity.
func NewClient(ctx context.Context, cfgs []ProviderConfig) (*Client, error) {
	c := &Client{
		providers: make(map[ProviderID]provider, len(cfgs)),
	}
	orderedIDs := make([]ProviderID, 0, len(cfgs))

	for _, cfg := range cfgs {
		id, err := ParseProviderID(cfg.ID)
		if err != nil {
			return nil, errors.Errorf("invalid provider id %q: %w", cfg.ID, err)
		}
		if cfg.URL == "" {
			return nil, errors.Errorf("provider %s: url is required", providerIDString(id))
		}
		if _, ok := c.providers[id]; ok {
			return nil, errors.Errorf("duplicate provider id: %s", providerIDString(id))
		}
		c.providers[id] = provider{cfg: cfg}
		orderedIDs = append(orderedIDs, id)
	}

	for _, id := range orderedIDs {
		p := c.providers[id]
		conn, err := dial(ctx, p.cfg)
		if err != nil {
			_ = c.Close()
			return nil, errors.Errorf("dial provider %s: %w", providerIDString(id), err)
		}

		c.providers[id] = provider{
			cfg:    p.cfg,
			conn:   conn,
			client: votingpowerv1.NewVotingPowerProviderServiceClient(conn),
		}
	}

	return c, nil
}

// Public API stays the same.
func (c *Client) GetVotingPowers(
	ctx context.Context,
	address symbiotic.CrossChainAddress,
	timestamp symbiotic.Timestamp,
) ([]symbiotic.OperatorVotingPower, error) {
	id := providerIDFromAddress(address.Address)
	p, ok := c.providers[id]
	if !ok {
		return nil, errors.Errorf("external provider id %s is not configured", providerIDString(id))
	}

	timeout := p.cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if len(p.cfg.Headers) > 0 {
		callCtx = metadata.NewOutgoingContext(callCtx, metadata.New(p.cfg.Headers))
	}

	resp, err := p.client.GetVotingPowersAt(callCtx, &votingpowerv1.GetVotingPowersAtRequest{
		Timestamp: uint64(timestamp),
	})
	if err != nil {
		return nil, errors.Errorf("external provider %s GetVotingPowersAt failed: %w", providerIDString(id), err)
	}

	agg := map[common.Address]*big.Int{}
	for _, vp := range resp.GetVotingPowers() {
		if !common.IsHexAddress(vp.GetOperator()) {
			return nil, errors.Errorf("invalid operator address: %q", vp.GetOperator())
		}
		op := common.HexToAddress(vp.GetOperator())

		v, parsed := new(big.Int).SetString(vp.GetVotingPower(), 10)
		if !parsed || v.Sign() < 0 {
			return nil, errors.Errorf("invalid voting power for operator %s: %q", op.Hex(), vp.GetVotingPower())
		}

		if agg[op] == nil {
			agg[op] = new(big.Int)
		}
		agg[op].Add(agg[op], v)
	}

	ops := make([]common.Address, 0, len(agg))
	for op := range agg {
		ops = append(ops, op)
	}
	slices.SortFunc(ops, func(a, b common.Address) int { return a.Cmp(b) })

	out := make([]symbiotic.OperatorVotingPower, 0, len(ops))
	for _, op := range ops {
		out = append(out, symbiotic.OperatorVotingPower{
			Operator: op,
			Vaults: []symbiotic.VaultVotingPower{{
				Vault:       address.Address,
				VotingPower: symbiotic.ToVotingPower(new(big.Int).Set(agg[op])),
			}},
		})
	}
	return out, nil
}

func (c *Client) Close() error {
	var firstErr error
	for id, p := range c.providers {
		if p.conn == nil {
			continue
		}
		if err := p.conn.Close(); err != nil && firstErr == nil {
			firstErr = err
			slog.Warn("failed to close external voting power provider connection",
				"providerId", providerIDString(id),
				"error", err,
			)
		}
	}
	return firstErr
}

func ParseProviderID(input string) (ProviderID, error) {
	s := strings.TrimSpace(input)
	s = strings.TrimPrefix(strings.TrimPrefix(s, "0x"), "0X")
	if len(s) != 20 {
		return ProviderID{}, errors.Errorf("provider id must be 10 bytes (20 hex chars), got %d", len(s))
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return ProviderID{}, errors.Errorf("invalid hex: %w", err)
	}
	var id ProviderID
	copy(id[:], b)
	return id, nil
}

func dial(ctx context.Context, cfg ProviderConfig) (*grpc.ClientConn, error) {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}
	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	if cfg.Secure {
		tlsCfg, err := buildTLSConfig(cfg)
		if err != nil {
			return nil, err
		}
		creds = grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg))
	}

	conn, err := grpc.NewClient(cfg.URL, creds)
	if err != nil {
		return nil, errors.Errorf("failed to create grpc client: %w", err)
	}

	conn.Connect()
	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			return conn, nil
		}

		if !conn.WaitForStateChange(dialCtx, state) {
			_ = conn.Close()
			return nil, errors.Errorf("failed to connect to external provider: %w", dialCtx.Err())
		}
	}
}

func buildTLSConfig(cfg ProviderConfig) (*tls.Config, error) {
	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if cfg.ServerName != "" {
		tlsCfg.ServerName = cfg.ServerName
	}
	if cfg.CACertFile == "" {
		return tlsCfg, nil
	}

	caPEM, err := os.ReadFile(cfg.CACertFile)
	if err != nil {
		return nil, errors.Errorf("read ca cert file: %w", err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(caPEM) {
		return nil, errors.Errorf("invalid CA cert PEM in %s", cfg.CACertFile)
	}
	tlsCfg.RootCAs = roots
	return tlsCfg, nil
}

func providerIDFromAddress(addr common.Address) ProviderID {
	var id ProviderID
	copy(id[:], addr[:10])
	return id
}

func providerIDString(id ProviderID) string {
	return "0x" + hex.EncodeToString(id[:])
}
