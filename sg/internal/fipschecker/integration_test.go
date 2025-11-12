package fipschecker

import (
	"fmt"
	"testing"

	"github.com/Azure/ShieldGuard/sg/internal/imageextractor"
)

func TestSingleImageIntegration(t *testing.T) {
	// Create FIPS client with shell script approach
	client := NewFIPSClient("/tmp/fips-check")

	// Test image reference
	imageRef := imageextractor.ImageReference{
		Image:   "mcr.microsoft.com/oss/v2/istio/proxyv2-fips:v1.26.4-1",
		ASTPath: "test.image",
	}

	fmt.Printf("Testing FIPS compliance for image: %s\n", imageRef.Image)

	// Check if FIPS checker is available
	if !client.IsAvailable() {
		t.Skip("FIPS checker is not available - ensure fips-check repository is cloned to /tmp/fips-check")
	}

	// Run FIPS compliance check
	result, err := client.CheckCompliance(imageRef)
	if err != nil {
		t.Fatalf("Error checking FIPS compliance: %v", err)
	}

	// Display results
	fmt.Printf("\n=== FIPS Compliance Results ===\n")
	fmt.Printf("Image: %s\n", result.Image)
	fmt.Printf("Is Compliant: %t\n", result.IsCompliant)
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Details: %s\n", result.Details)
	fmt.Printf("Summary: %s\n", result.Summary)
}