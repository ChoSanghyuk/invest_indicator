package util

import (
	"testing"
)

func TestLoadABIFromHardhatArtifact(t *testing.T) {
	// Test with RouterV2 artifact
	artifactPath := "../../blackholedex-contracts/artifacts/contracts/RouterV2.sol/RouterV2.json"

	abi, err := LoadABIFromHardhatArtifact(artifactPath)
	if err != nil {
		t.Skipf("Could not load RouterV2 artifact (may not exist): %v", err)
	}

	// Check that ABI was loaded
	if abi == nil {
		t.Fatal("ABI is nil")
	}

	// Check that some methods exist
	if len(abi.Methods) == 0 {
		t.Fatal("No methods found in ABI")
	}

	t.Logf("Loaded ABI with %d methods, %d events", len(abi.Methods), len(abi.Events))

	// Check for specific RouterV2 methods
	expectedMethods := []string{"swapExactTokensForTokens", "addLiquidity", "removeLiquidity"}
	for _, methodName := range expectedMethods {
		if _, exists := abi.Methods[methodName]; !exists {
			t.Errorf("Expected method %s not found in ABI", methodName)
		} else {
			t.Logf("Found method: %s", methodName)
		}
	}
}

func TestLoadABI(t *testing.T) {
	// Test with RouterV2 artifact (auto-detect)
	artifactPath := "../../blackholedex-contracts/artifacts/contracts/RouterV2.sol/RouterV2.json"

	abi, err := LoadABI(artifactPath)
	if err != nil {
		t.Skipf("Could not load artifact (may not exist): %v", err)
	}

	if abi == nil {
		t.Fatal("ABI is nil")
	}

	t.Logf("Auto-detected and loaded ABI with %d methods", len(abi.Methods))
}

func TestGetContractInfo(t *testing.T) {
	artifactPath := "../../blackholedex-contracts/artifacts/contracts/RouterV2.sol/RouterV2.json"

	artifact, err := GetContractInfo(artifactPath)
	if err != nil {
		t.Skipf("Could not load artifact (may not exist): %v", err)
	}

	if artifact.ContractName != "RouterV2" {
		t.Errorf("Expected contract name 'RouterV2', got '%s'", artifact.ContractName)
	}

	if artifact.SourceName != "contracts/RouterV2.sol" {
		t.Errorf("Expected source name 'contracts/RouterV2.sol', got '%s'", artifact.SourceName)
	}

	if artifact.Format != "hh-sol-artifact-1" {
		t.Errorf("Expected format 'hh-sol-artifact-1', got '%s'", artifact.Format)
	}

	t.Logf("Contract Info - Name: %s, Source: %s, Format: %s",
		artifact.ContractName, artifact.SourceName, artifact.Format)
}

func TestLoadMultipleArtifacts(t *testing.T) {
	contracts := []struct {
		path         string
		contractName string
		methods      []string
	}{
		{
			path:         "../../blackholedex-contracts/artifacts/contracts/RouterV2.sol/RouterV2.json",
			contractName: "RouterV2",
			methods:      []string{"swapExactTokensForTokens", "addLiquidity"},
		},
		{
			path:         "../../blackholedex-contracts/artifacts/contracts/VotingEscrow.sol/VotingEscrow.json",
			contractName: "VotingEscrow",
			methods:      []string{"create_lock", "increase_amount"},
		},
		{
			path:         "../../blackholedex-contracts/artifacts/contracts/Black.sol/Black.json",
			contractName: "Black",
			methods:      []string{"approve", "transfer"},
		},
	}

	for _, tc := range contracts {
		t.Run(tc.contractName, func(t *testing.T) {
			abi, err := LoadABI(tc.path)
			if err != nil {
				t.Skipf("Could not load %s: %v", tc.contractName, err)
			}

			for _, methodName := range tc.methods {
				if method, exists := abi.Methods[methodName]; exists {
					t.Logf("%s.%s found with %d inputs", tc.contractName, methodName, len(method.Inputs))
				} else {
					t.Logf("%s.%s not found (may be expected)", tc.contractName, methodName)
				}
			}
		})
	}
}
