package tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/go-errors/errors"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var globalTestEnv *TestEnvironment

func TestMain(m *testing.M) {
	// Setup test environment before all tests
	var err error
	globalTestEnv, err = setupGlobalTestEnvironment()
	if err != nil {
		fmt.Printf("Failed to setup test environment: %v\n", err)
		os.Exit(1)
	}

	// Run all tests
	code := m.Run()

	// Cleanup after all tests
	cleanupGlobalTestEnvironment(globalTestEnv)

	os.Exit(code)
}

type RelaySidecarConfig struct {
	Keys          string
	DataDir       string
	ContainerName string
}

type TestEnvironment struct {
	Containers     map[int]testcontainers.Container
	ContainerPorts map[int]string
	SidecarConfigs []RelaySidecarConfig
}

func generateSidecarConfigs(env EnvInfo) []RelaySidecarConfig {
	const (
		basePrivateKey = 1000000000000000000
		swarmKey       = "0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364140"
	)

	configs := make([]RelaySidecarConfig, env.Operators)

	for i := int64(0); i < env.Operators; i++ {
		keyIndex := i
		symbPrivateKeyDecimal := basePrivateKey + keyIndex
		symbPrivateKeyHex := fmt.Sprintf("%064x", symbPrivateKeyDecimal)
		symbSecondaryPrivateKeyDecimal := basePrivateKey + keyIndex + 10_000
		symbSecondaryPrivateKeyHex := fmt.Sprintf("%064x", symbSecondaryPrivateKeyDecimal)

		// Generate key string in the same format as generate_network.sh
		keys := []string{
			fmt.Sprintf("symb/0/15/0x%s", symbPrivateKeyHex),
			fmt.Sprintf("symb/0/11/0x%s", symbSecondaryPrivateKeyHex),
			fmt.Sprintf("symb/1/0/0x%s", symbPrivateKeyHex),
			fmt.Sprintf("evm/1/31337/0x%s", symbPrivateKeyHex),
			fmt.Sprintf("evm/1/31338/0x%s", symbPrivateKeyHex),
			fmt.Sprintf("p2p/1/0/%s", swarmKey),
			fmt.Sprintf("p2p/1/1/%s", symbPrivateKeyHex),
		}
		keysString := strings.Join(keys, ",")

		configs[i] = RelaySidecarConfig{
			Keys:          keysString,
			DataDir:       fmt.Sprintf("/app/data-%02d", i+1),
			ContainerName: fmt.Sprintf("relay-sidecar-%d", i+1),
		}
	}

	return configs
}

func setupGlobalTestEnvironment() (*TestEnvironment, error) {
	ctx := context.Background()

	// Use the existing docker-compose network
	networkName := "temp-network_symbiotic-network"

	deploymentData, err := loadDeploymentData(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to load deployment data: %v", err)
	}

	// Generate sidecar configurations based on environment variables
	sidecarConfigs := generateSidecarConfigs(deploymentData.Env)

	containers := make(map[int]testcontainers.Container)
	containerPorts := make(map[int]string)

	// Get project root directory (assuming we're in e2e/tests/)
	projectRoot, err := filepath.Abs("../../")
	if err != nil {
		return nil, errors.Errorf("failed to get project root: %v", err)
	}

	tempNetworkDir := filepath.Join(projectRoot, "e2e", "temp-network")

	// Clean up existing data directories before starting containers
	for i := range sidecarConfigs {
		dataHostDir := filepath.Join(tempNetworkDir, fmt.Sprintf("data-%02d", i+1))
		if err := os.RemoveAll(dataHostDir); err != nil {
			fmt.Printf("Warning: failed to clean data directory %s: %v\n", dataHostDir, err)
		}
		// Create fresh directory for bind mount
		if err := os.MkdirAll(dataHostDir, 0755); err != nil {
			return nil, errors.Errorf("failed to create data directory %s: %v", dataHostDir, err)
		}
	}

	// Start each relay sidecar container concurrently
	type containerResult struct {
		index     int
		container testcontainers.Container
		port      string
	}

	var wg sync.WaitGroup
	errorChan := make(chan error, len(sidecarConfigs))
	containerChan := make(chan containerResult, len(sidecarConfigs))

	for i, config := range sidecarConfigs {
		wg.Add(1)
		go func(i int, config RelaySidecarConfig) {
			defer wg.Done()

			fmt.Printf("Starting container: %s\n", config.ContainerName)

			// Create data directory path
			dataHostDir := filepath.Join(tempNetworkDir, fmt.Sprintf("data-%02d", i+1))
			deployDataDir := filepath.Join(tempNetworkDir, "deploy-data")

			opts := []string{
				"--config /tmp/sidecar.yaml",
				fmt.Sprintf("--driver.address %s", deploymentData.Driver.Addr),
				fmt.Sprintf("--storage-dir %s", config.DataDir),
				fmt.Sprintf("--secret-keys %s", config.Keys),
			}
			// Build the command to start the sidecar
			startCommand := fmt.Sprintf("./relay_sidecar %s", strings.Join(opts, " "))

			req := testcontainers.ContainerRequest{
				Image:        "relay_sidecar:dev",
				Name:         config.ContainerName,
				ExposedPorts: []string{"8080/tcp"},
				Cmd:          []string{"sh", "-c", startCommand},
				Files: []testcontainers.ContainerFile{{
					HostFilePath:      filepath.Join(projectRoot, "e2e", "tests", "sidecar.yaml"),
					ContainerFilePath: "/tmp/sidecar.yaml",
					FileMode:          0644,
				}},
				HostConfigModifier: func(hostConfig *container.HostConfig) {
					hostConfig.Mounts = []mount.Mount{
						{
							Type:   mount.TypeBind,
							Source: dataHostDir,
							Target: config.DataDir,
						},
						{
							Type:   mount.TypeBind,
							Source: deployDataDir,
							Target: "/deploy-data",
						},
					}
				},
				Networks: []string{networkName},
				WaitingFor: wait.ForAll(
					wait.ForHTTP("/healthz").WithPort("8080/tcp").WithStartupTimeout(30 * time.Second),
				),
			}

			containerInstance, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
				ContainerRequest: req,
				Started:          true,
			})
			if err != nil {
				errorChan <- errors.Errorf("failed to start container %s: %v", config.ContainerName, err)
				return
			}

			// Get the mapped port
			mappedPort, err := containerInstance.MappedPort(ctx, "8080")
			if err != nil {
				errorChan <- errors.Errorf("failed to get mapped port for %s: %v", config.ContainerName, err)
				return
			}

			containerChan <- containerResult{
				index:     i,
				container: containerInstance,
				port:      mappedPort.Port(),
			}

			fmt.Printf("Container %s started on port %s\n", config.ContainerName, mappedPort.Port())
		}(i, config)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errorChan)
	close(containerChan)

	// Check for errors
	for err := range errorChan {
		return nil, err
	}

	// Collect results
	for result := range containerChan {
		containers[result.index] = result.container
		containerPorts[result.index] = result.port
	}

	return &TestEnvironment{
		Containers:     containers,
		ContainerPorts: containerPorts,
		SidecarConfigs: sidecarConfigs,
	}, nil
}

func cleanupGlobalTestEnvironment(env *TestEnvironment) {
	if env == nil {
		return
	}

	ctx := context.Background()

	// Stop and remove containers
	for i, containerInstance := range env.Containers {
		fmt.Printf("Stopping container: %d\n", i)
		if err := containerInstance.Terminate(ctx); err != nil {
			fmt.Printf("Error stopping container %d: %v\n", i, err)
		}
	}

	// Note: We don't remove the network since it might be used by docker-compose
	// The network will be cleaned up when docker-compose is stopped
}

func (env *TestEnvironment) GetContainerPort(i int) string {
	return env.ContainerPorts[i]
}

// Helper function to get container port
func (env *TestEnvironment) GetHealthEndpoint(i int) string {
	return fmt.Sprintf("http://localhost:%s/healthz", globalTestEnv.GetContainerPort(i))
}

func (env *TestEnvironment) GetGRPCAddress(index int) string {
	return fmt.Sprintf("localhost:%s", env.GetContainerPort(index))
}
