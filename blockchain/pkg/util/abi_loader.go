package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// HardhatArtifact represents the structure of a Hardhat compilation artifact
type HardhatArtifact struct {
	Format       string          `json:"_format"`
	ContractName string          `json:"contractName"`
	SourceName   string          `json:"sourceName"`
	ABI          json.RawMessage `json:"abi"`
	Bytecode     string          `json:"bytecode"`
	DeployedBytecode string      `json:"deployedBytecode,omitempty"`
	LinkReferences json.RawMessage `json:"linkReferences,omitempty"`
	DeployedLinkReferences json.RawMessage `json:"deployedLinkReferences,omitempty"`
}

// LoadABIFromHardhatArtifact loads an ABI from a Hardhat artifact JSON file
func LoadABIFromHardhatArtifact(filePath string) (*abi.ABI, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read artifact file: %w", err)
	}

	// Parse the Hardhat artifact
	var artifact HardhatArtifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		return nil, fmt.Errorf("failed to parse artifact JSON: %w", err)
	}

	// Check if ABI exists
	if len(artifact.ABI) == 0 {
		return nil, fmt.Errorf("ABI is empty in artifact file")
	}

	// Parse the ABI
	parsedABI, err := abi.JSON(bytes.NewReader(artifact.ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	return &parsedABI, nil
}

// LoadABIFromJSON loads an ABI from a plain JSON file (just the ABI array)
func LoadABIFromJSON(filePath string) (*abi.ABI, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read ABI file: %w", err)
	}

	// Parse the ABI
	parsedABI, err := abi.JSON(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	return &parsedABI, nil
}

// LoadABI attempts to load an ABI from either a Hardhat artifact or plain JSON
func LoadABI(filePath string) (*abi.ABI, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Try to parse as Hardhat artifact first
	var artifact HardhatArtifact
	if err := json.Unmarshal(data, &artifact); err == nil && len(artifact.ABI) > 0 {
		// It's a Hardhat artifact
		parsedABI, err := abi.JSON(bytes.NewReader(artifact.ABI))
		if err != nil {
			return nil, fmt.Errorf("failed to parse ABI from artifact: %w", err)
		}
		return &parsedABI, nil
	}

	// Try to parse as plain ABI JSON
	parsedABI, err := abi.JSON(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse as plain ABI JSON: %w", err)
	}

	return &parsedABI, nil
}

// GetContractInfo extracts contract metadata from a Hardhat artifact
func GetContractInfo(filePath string) (*HardhatArtifact, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read artifact file: %w", err)
	}

	var artifact HardhatArtifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		return nil, fmt.Errorf("failed to parse artifact JSON: %w", err)
	}

	return &artifact, nil
}
