package fipschecker

import (
	"fmt"
	"os"
	"testing"

	"github.com/Azure/ShieldGuard/sg/internal/imageextractor"
	"github.com/open-policy-agent/opa/ast"
	"gopkg.in/yaml.v3"
)

func TestConfigFileIntegration(t *testing.T) {
	// Read the config file
	data, err := os.ReadFile("testdata/test_config.yaml")
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	// Parse YAML to generic map
	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	// Convert to AST value
	astValue, err := ast.InterfaceToValue(yamlData)
	if err != nil {
		t.Fatalf("Failed to convert to AST: %v", err)
	}

	// Extract images from the AST
	images := imageextractor.ExtractImages(astValue)

	fmt.Printf("Found %d images in config file:\n", len(images))
	for i, img := range images {
		fmt.Printf("  [%d] %s (at %s)\n", i+1, img.Image, img.ASTPath)
	}

	// Create FIPS client
	client := NewFIPSClient("/tmp/fips-check")

	if !client.IsAvailable() {
		t.Skip("FIPS checker is not available - ensure fips-check repository is cloned to /tmp/fips-check")
	}

	// Check FIPS compliance for all images
	fmt.Printf("\n=== FIPS Compliance Results ===\n")

	for _, img := range images {
		fmt.Printf("\n--- Checking: %s ---\n", img.Image)

		result, err := client.CheckCompliance(img)
		if err != nil {
			t.Errorf("Error checking image %s: %v", img.Image, err)
			continue
		}

		fmt.Printf("Image: %s\n", result.Image)
		fmt.Printf("Is Compliant: %t\n", result.IsCompliant)
		fmt.Printf("Status: %s\n", result.Status)
		fmt.Printf("Details: %s\n", result.Details)

		// Show full summary of binary scan results
		fmt.Printf("Summary: %s\n", result.Summary)
	}
}