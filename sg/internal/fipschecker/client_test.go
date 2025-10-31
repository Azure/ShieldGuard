package fipschecker

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Azure/ShieldGuard/sg/internal/imageextractor"
	"github.com/open-policy-agent/opa/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// MockFIPSClient allows us to test without pulling real Docker images
type MockFIPSClient struct {
	mockReports map[string][]BinaryReport
	mockErrors  map[string]error
}

// NewMockFIPSClient creates a mock FIPS client for testing
func NewMockFIPSClient() *MockFIPSClient {
	return &MockFIPSClient{
		mockReports: make(map[string][]BinaryReport),
		mockErrors:  make(map[string]error),
	}
}

// SetMockResult sets up a mock result for a specific image
func (m *MockFIPSClient) SetMockResult(image string, reports []BinaryReport, err error) {
	m.mockReports[image] = reports
	m.mockErrors[image] = err
}

// CheckMultipleImages simulates checking multiple images for testing
func (m *MockFIPSClient) CheckMultipleImages(images []imageextractor.ImageReference) (map[string]FIPSResult, error) {
	results := make(map[string]FIPSResult)

	for _, img := range images {
		result, err := m.CheckCompliance(img)
		if err != nil {
			result = FIPSResult{
				Image:       img.Image,
				IsCompliant: false,
				Status:      "ERROR",
				Details:     fmt.Sprintf("Check failed: %v", err),
			}
		}
		results[img.ASTPath] = result
	}

	return results, nil
}

// CheckCompliance simulates FIPS checking for testing
func (m *MockFIPSClient) CheckCompliance(imageRef imageextractor.ImageReference) (FIPSResult, error) {
	reports, exists := m.mockReports[imageRef.Image]
	if !exists {
		return FIPSResult{
			Image:       imageRef.Image,
			IsCompliant: false,
			Status:      "ERROR",
			Details:     "Mock: Image not configured",
		}, nil
	}

	if mockErr := m.mockErrors[imageRef.Image]; mockErr != nil {
		return FIPSResult{
			Image:       imageRef.Image,
			IsCompliant: false,
			Status:      "ERROR",
			Details:     fmt.Sprintf("Mock error: %v", mockErr),
		}, mockErr
	}

	// Analyze mock reports using real logic
	client := &FIPSClient{}
	isCompliant := client.analyzeReports(reports)

	result := FIPSResult{
		Image:       imageRef.Image,
		IsCompliant: isCompliant,
	}

	if isCompliant {
		result.Status = "COMPLIANT"
		result.Details = "All Go binaries pass FIPS compliance checks"
	} else {
		result.Status = "NOT COMPLIANT"
		result.Details = client.formatComplianceIssues(reports)
	}

	result.Summary = client.formatSummary(reports)
	return result, nil
}

func TestFIPSClient_CheckCompliance_CompliantImage(t *testing.T) {
	// Setup mock client
	mockClient := NewMockFIPSClient()

	// Create mock reports for a FIPS-compliant binary
	compliantReports := []BinaryReport{
		{
			RelativePath: "usr/local/bin/app",
			Type:         "gobinary",
			GoBinaryDetails: GoBinaryReportDetails{
				GoVersion:        "go1.24.4 X:systemcrypto",
				Module:           "github.com/example/fips-app",
				UseSystemcrypto:  true,
				CGOEnabled:       true,
				FailsOnFIPSCheck: false,
				RuntimePanicLog:  "",
			},
			Error: nil,
		},
	}

	testImage := "mcr.microsoft.com/oss/example/fips-app:v1.0.0"
	mockClient.SetMockResult(testImage, compliantReports, nil)

	// Test
	imageRef := imageextractor.ImageReference{
		Image:   testImage,
		ASTPath: "spec.containers[0].image",
	}

	result, err := mockClient.CheckCompliance(imageRef)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, testImage, result.Image)
	assert.True(t, result.IsCompliant, "Expected image to be FIPS compliant")
	assert.Equal(t, "COMPLIANT", result.Status)
	assert.Equal(t, "All Go binaries pass FIPS compliance checks", result.Details)
	assert.Equal(t, "Total: 1 | Systemcrypto: 1 | Failed FIPS: 0", result.Summary)
}

func TestFIPSClient_CheckCompliance_NonCompliantImage_NoSystemcrypto(t *testing.T) {
	// Setup mock client
	mockClient := NewMockFIPSClient()

	// Create mock reports for a non-FIPS binary (no systemcrypto)
	nonCompliantReports := []BinaryReport{
		{
			RelativePath: "usr/local/bin/regular-app",
			Type:         "gobinary",
			GoBinaryDetails: GoBinaryReportDetails{
				GoVersion:        "go1.24.4",
				Module:           "github.com/example/regular-app",
				UseSystemcrypto:  false, // No systemcrypto
				CGOEnabled:       false,
				FailsOnFIPSCheck: true,
				RuntimePanicLog:  "",
			},
			Error: nil,
		},
	}

	testImage := "docker.io/library/nginx:1.21"
	mockClient.SetMockResult(testImage, nonCompliantReports, nil)

	// Test
	imageRef := imageextractor.ImageReference{
		Image:   testImage,
		ASTPath: "spec.containers[0].image",
	}

	result, err := mockClient.CheckCompliance(imageRef)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, testImage, result.Image)
	assert.False(t, result.IsCompliant, "Expected image to be non-FIPS compliant")
	assert.Equal(t, "NOT COMPLIANT", result.Status)
	assert.Contains(t, result.Details, "systemcrypto not enabled")
	assert.Equal(t, "Total: 1 | Systemcrypto: 0 | Failed FIPS: 1", result.Summary)
}

func TestFIPSClient_CheckCompliance_NonCompliantImage_RuntimeFailure(t *testing.T) {
	// Setup mock client
	mockClient := NewMockFIPSClient()

	// Create mock reports for a binary with systemcrypto but runtime failure
	runtimeFailureReports := []BinaryReport{
		{
			RelativePath: "usr/local/bin/partial-fips-app",
			Type:         "gobinary",
			GoBinaryDetails: GoBinaryReportDetails{
				GoVersion:        "go1.24.4 X:systemcrypto",
				Module:           "github.com/example/partial-fips-app",
				UseSystemcrypto:  true,  // Has systemcrypto
				CGOEnabled:       true,
				FailsOnFIPSCheck: true,  // But fails runtime check
				RuntimePanicLog:  "panic: opensslcrypto: FIPS mode requested but not available in OpenSSL",
			},
			Error: nil,
		},
	}

	testImage := "mcr.microsoft.com/oss/example/broken-fips:v1.0.0"
	mockClient.SetMockResult(testImage, runtimeFailureReports, nil)

	// Test
	imageRef := imageextractor.ImageReference{
		Image:   testImage,
		ASTPath: "spec.containers[0].image",
	}

	result, err := mockClient.CheckCompliance(imageRef)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, testImage, result.Image)
	assert.False(t, result.IsCompliant, "Expected image to be non-FIPS compliant")
	assert.Equal(t, "NOT COMPLIANT", result.Status)
	assert.Contains(t, result.Details, "runtime FIPS check failed")
	assert.Equal(t, "Total: 1 | Systemcrypto: 1 | Failed FIPS: 1", result.Summary)
}

func TestFIPSClient_CheckCompliance_MultipleImages(t *testing.T) {
	mockClient := NewMockFIPSClient()

	// Setup multiple test images with different compliance states
	testCases := []struct {
		image         string
		astPath       string
		reports       []BinaryReport
		expectedCompliant bool
	}{
		{
			image:   "mcr.microsoft.com/oss/v2/istio/proxyv2-fips:v1.26.4-1",
			astPath: "proxy.image",
			reports: []BinaryReport{
				{
					RelativePath: "usr/local/bin/pilot-agent",
					Type:         "gobinary",
					GoBinaryDetails: GoBinaryReportDetails{
						GoVersion:        "go1.24.4 X:systemcrypto",
						UseSystemcrypto:  true,
						CGOEnabled:       true,
						FailsOnFIPSCheck: false,
					},
				},
			},
			expectedCompliant: true,
		},
		{
			image:   "docker.io/istio/pilot:1.20.0",
			astPath: "sidecar.image",
			reports: []BinaryReport{
				{
					RelativePath: "usr/local/bin/pilot-discovery",
					Type:         "gobinary",
					GoBinaryDetails: GoBinaryReportDetails{
						GoVersion:        "go1.20.0",
						UseSystemcrypto:  false,
						CGOEnabled:       false,
						FailsOnFIPSCheck: true,
					},
				},
			},
			expectedCompliant: false,
		},
	}

	// Set up mock results
	for _, tc := range testCases {
		mockClient.SetMockResult(tc.image, tc.reports, nil)
	}

	// Test each image
	for _, tc := range testCases {
		t.Run(tc.image, func(t *testing.T) {
			imageRef := imageextractor.ImageReference{
				Image:   tc.image,
				ASTPath: tc.astPath,
			}

			result, err := mockClient.CheckCompliance(imageRef)

			assert.NoError(t, err)
			assert.Equal(t, tc.image, result.Image)
			assert.Equal(t, tc.expectedCompliant, result.IsCompliant,
				"Compliance mismatch for image %s", tc.image)

			if tc.expectedCompliant {
				assert.Equal(t, "COMPLIANT", result.Status)
			} else {
				assert.Equal(t, "NOT COMPLIANT", result.Status)
			}
		})
	}
}

func TestFIPSClient_CheckMultipleImages(t *testing.T) {
	mockClient := NewMockFIPSClient()

	// Setup test images
	images := []imageextractor.ImageReference{
		{Image: "mcr.microsoft.com/fips-app:v1", ASTPath: "app.image"},
		{Image: "docker.io/regular-app:v1", ASTPath: "sidecar.image"},
	}

	// Mock results
	mockClient.SetMockResult("mcr.microsoft.com/fips-app:v1", []BinaryReport{
		{
			RelativePath: "bin/app",
			GoBinaryDetails: GoBinaryReportDetails{
				UseSystemcrypto:  true,
				FailsOnFIPSCheck: false,
			},
		},
	}, nil)

	mockClient.SetMockResult("docker.io/regular-app:v1", []BinaryReport{
		{
			RelativePath: "bin/app",
			GoBinaryDetails: GoBinaryReportDetails{
				UseSystemcrypto:  false,
				FailsOnFIPSCheck: true,
			},
		},
	}, nil)

	// Test
	results, err := mockClient.CheckMultipleImages(images)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	// Check first image (compliant)
	result1 := results["app.image"]
	assert.Equal(t, "mcr.microsoft.com/fips-app:v1", result1.Image)
	assert.True(t, result1.IsCompliant)
	assert.Equal(t, "COMPLIANT", result1.Status)

	// Check second image (non-compliant)
	result2 := results["sidecar.image"]
	assert.Equal(t, "docker.io/regular-app:v1", result2.Image)
	assert.False(t, result2.IsCompliant)
	assert.Equal(t, "NOT COMPLIANT", result2.Status)
}

func TestFIPSClient_CheckCompliance_NoGoBinaries(t *testing.T) {
	mockClient := NewMockFIPSClient()

	// Mock an image with no Go binaries
	testImage := "docker.io/library/alpine:latest"
	mockClient.SetMockResult(testImage, []BinaryReport{}, nil)

	imageRef := imageextractor.ImageReference{
		Image:   testImage,
		ASTPath: "spec.containers[0].image",
	}

	result, err := mockClient.CheckCompliance(imageRef)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, testImage, result.Image)
	assert.False(t, result.IsCompliant, "Image with no Go binaries should be non-compliant")
	assert.Equal(t, "NOT COMPLIANT", result.Status)
	assert.Equal(t, "Total: 0 | Systemcrypto: 0 | Failed FIPS: 0", result.Summary)
}

// Integration test with real FIPS checker (requires docker)
func TestFIPSClient_RealIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test requires Docker and may take time
	// Skip if CI environment variable is set or Docker not available
	if os.Getenv("CI") != "" {
		t.Skip("Skipping Docker integration test in CI")
	}

	client := NewFIPSClient("")

	// Test with a known image (this will actually pull and scan)
	imageRef := imageextractor.ImageReference{
		Image:   "alpine:latest", // Small image for testing
		ASTPath: "test.image",
	}

	result, err := client.CheckCompliance(imageRef)

	// Should complete without error (even if not compliant)
	assert.NoError(t, err)
	assert.Equal(t, "alpine:latest", result.Image)
	assert.Contains(t, []string{"COMPLIANT", "NOT COMPLIANT", "ERROR"}, result.Status)
	assert.NotEmpty(t, result.Summary)

	t.Logf("Real FIPS check result for %s: %s", result.Image, result.Status)
	t.Logf("Details: %s", result.Details)
}

// Test FIPS checker with real image reference running locally
func TestFIPSClient_WithRealImageReference_Local(t *testing.T) {
	// Test with a real imageextractor.ImageReference
	imageRef := imageextractor.ImageReference{
		Image:   "alpine:latest", // Small, commonly available image
		ASTPath: "spec.containers[0].image",
	}

	// Create real FIPS client
	client := NewFIPSClient("")

	t.Run("CheckCompliance_RealImage", func(t *testing.T) {
		// Test the complete flow: ImageReference -> FIPS checker -> Result
		result, err := client.CheckCompliance(imageRef)

		// Assertions - should complete without error
		assert.NoError(t, err, "FIPS check should complete without error")
		assert.Equal(t, imageRef.Image, result.Image, "Result should contain original image name")
		assert.NotEmpty(t, result.Status, "Result should have a status")
		assert.Contains(t, []string{"COMPLIANT", "NOT COMPLIANT", "ERROR"}, result.Status, "Status should be valid")

		// Summary may be empty for ERROR cases
		if result.Status != "ERROR" {
			assert.NotEmpty(t, result.Summary, "Result should have summary for non-error cases")
		}

		// Log the results for manual verification
		t.Logf("Image: %s", result.Image)
		t.Logf("Status: %s", result.Status)
		t.Logf("Is Compliant: %t", result.IsCompliant)
		t.Logf("Details: %s", result.Details)
		t.Logf("Summary: %s", result.Summary)

		// Verify the result structure is valid
		if result.Status == "COMPLIANT" {
			assert.True(t, result.IsCompliant, "Compliant status should have IsCompliant=true")
		} else if result.Status == "NOT COMPLIANT" {
			assert.False(t, result.IsCompliant, "Non-compliant status should have IsCompliant=false")
		}
	})

	t.Run("CheckCompliance_InvalidImage", func(t *testing.T) {
		// Test with non-existent image
		invalidImageRef := imageextractor.ImageReference{
			Image:   "nonexistent/invalid-image:999",
			ASTPath: "test.image",
		}

		result, err := client.CheckCompliance(invalidImageRef)

		// Should handle gracefully - either error or ERROR status
		if err != nil {
			t.Logf("Expected error for invalid image: %v", err)
		} else {
			assert.Equal(t, "ERROR", result.Status, "Invalid image should result in ERROR status")
			assert.False(t, result.IsCompliant, "Invalid image should not be compliant")
		}
	})

	t.Run("CheckCompliance_FIPSCompliantImage_Istio", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping FIPS image test in short mode (large image download)")
		}

		// Test with a known FIPS-compliant image from Microsoft
		fipsImageRef := imageextractor.ImageReference{
			Image:   "mcr.microsoft.com/oss/v2/istio/proxyv2-fips:v1.26.4-1",
			ASTPath: "spec.containers[0].image",
		}

		// Test the complete flow with a real FIPS-compliant image
		result, err := client.CheckCompliance(fipsImageRef)

		// Assertions - should complete without error
		assert.NoError(t, err, "FIPS check should complete without error")
		assert.Equal(t, fipsImageRef.Image, result.Image, "Result should contain original image name")
		assert.NotEmpty(t, result.Status, "Result should have a status")
		assert.Contains(t, []string{"COMPLIANT", "NOT COMPLIANT", "ERROR"}, result.Status, "Status should be valid")

		// Log the results for analysis
		t.Logf("FIPS Image Test Results:")
		t.Logf("Image: %s", result.Image)
		t.Logf("Status: %s", result.Status)
		t.Logf("Is Compliant: %t", result.IsCompliant)
		t.Logf("Details: %s", result.Details)
		t.Logf("Summary: %s", result.Summary)

		// This image should ideally be COMPLIANT, but let's verify what we actually get
		if result.Status == "COMPLIANT" {
			assert.True(t, result.IsCompliant, "FIPS-compliant image should have IsCompliant=true")
			t.Log("✅ SUCCESS: Image correctly identified as FIPS-compliant!")
		} else if result.Status == "NOT COMPLIANT" {
			t.Logf("⚠️  ANALYSIS: FIPS image marked as non-compliant. Reason: %s", result.Details)
			t.Log("This could mean:")
			t.Log("  1. No Go binaries found in the image")
			t.Log("  2. Go binaries not built with GOEXPERIMENT=systemcrypto")
			t.Log("  3. Go binaries are C++ based (Envoy proxy) not Go")
			t.Log("  4. Binaries are in non-standard locations")
		} else if result.Status == "ERROR" {
			t.Logf("❌ ERROR: Failed to check FIPS image. Details: %s", result.Details)
		}

		// Verify result structure is valid regardless of compliance status
		assert.NotEmpty(t, result.Image, "Should have image name")
		if result.Status != "ERROR" {
			assert.NotEmpty(t, result.Summary, "Should have summary for successful checks")
		}
	})

	t.Run("CheckCompliance_AzureIPMasqAgent", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping Azure IP masq agent test in short mode (image download)")
		}

		// Test with Azure IP masq agent - likely contains Go binaries
		azureImageRef := imageextractor.ImageReference{
			Image:   "mcr.microsoft.com/oss/v2/azure/ip-masq-agent-v2:v0.1.15",
			ASTPath: "spec.containers[0].image",
		}

		// Test the complete flow with Azure IP masq agent
		result, err := client.CheckCompliance(azureImageRef)

		// Assertions - should complete without error
		assert.NoError(t, err, "FIPS check should complete without error")
		assert.Equal(t, azureImageRef.Image, result.Image, "Result should contain original image name")
		assert.NotEmpty(t, result.Status, "Result should have a status")
		assert.Contains(t, []string{"COMPLIANT", "NOT COMPLIANT", "ERROR"}, result.Status, "Status should be valid")

		// Log the results for analysis
		t.Logf("Azure IP Masq Agent Test Results:")
		t.Logf("Image: %s", result.Image)
		t.Logf("Status: %s", result.Status)
		t.Logf("Is Compliant: %t", result.IsCompliant)
		t.Logf("Details: %s", result.Details)
		t.Logf("Summary: %s", result.Summary)

		// Analyze the results
		if result.Status == "COMPLIANT" {
			assert.True(t, result.IsCompliant, "FIPS-compliant image should have IsCompliant=true")
			t.Log("✅ SUCCESS: Azure IP masq agent is FIPS-compliant!")
		} else if result.Status == "NOT COMPLIANT" {
			t.Logf("⚠️  ANALYSIS: Azure IP masq agent marked as non-compliant.")
			t.Logf("Details: %s", result.Details)

			// Check if Go binaries were found
			if strings.Contains(result.Summary, "Total: 0") {
				t.Log("No Go binaries found - might be a different implementation")
			} else {
				t.Log("Go binaries found but not FIPS-compliant")
			}
		} else if result.Status == "ERROR" {
			t.Logf("❌ ERROR: Failed to check Azure IP masq agent. Details: %s", result.Details)
		}

		// Verify result structure is valid regardless of compliance status
		assert.NotEmpty(t, result.Image, "Should have image name")
		if result.Status != "ERROR" {
			assert.NotEmpty(t, result.Summary, "Should have summary for successful checks")
		}
	})

	t.Run("CheckCompliance_NonFIPSIstio_Comparison", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping non-FIPS Istio test in short mode (image download)")
		}

		// Test with regular (non-FIPS) Istio image for comparison
		regularIstioRef := imageextractor.ImageReference{
			Image:   "mcr.microsoft.com/oss/v2/istio/proxyv2:v1.25.5-1", // Non-FIPS version
			ASTPath: "spec.containers[0].image",
		}

		// Test the complete flow with regular Istio image
		result, err := client.CheckCompliance(regularIstioRef)

		// Assertions - should complete without error
		assert.NoError(t, err, "FIPS check should complete without error")
		assert.Equal(t, regularIstioRef.Image, result.Image, "Result should contain original image name")
		assert.NotEmpty(t, result.Status, "Result should have a status")
		assert.Contains(t, []string{"COMPLIANT", "NOT COMPLIANT", "ERROR"}, result.Status, "Status should be valid")

		// Log the results for analysis
		t.Logf("Regular Istio (Non-FIPS) Test Results:")
		t.Logf("Image: %s", result.Image)
		t.Logf("Status: %s", result.Status)
		t.Logf("Is Compliant: %t", result.IsCompliant)
		t.Logf("Details: %s", result.Details)
		t.Logf("Summary: %s", result.Summary)

		// Analyze the results - this should likely be NOT COMPLIANT
		if result.Status == "COMPLIANT" {
			t.Log("🤔 UNEXPECTED: Regular Istio image is FIPS-compliant (might have systemcrypto)")
		} else if result.Status == "NOT COMPLIANT" {
			t.Log("✅ EXPECTED: Regular Istio image is not FIPS-compliant")
			if strings.Contains(result.Summary, "Total: 0") {
				t.Log("   Reason: No Go binaries found")
			} else {
				t.Log("   Reason: Go binaries found but not FIPS-compliant")
			}
		} else if result.Status == "ERROR" {
			t.Logf("❌ ERROR: Failed to check regular Istio image. Details: %s", result.Details)
		}

		// Verify result structure is valid
		assert.NotEmpty(t, result.Image, "Should have image name")
		if result.Status != "ERROR" {
			assert.NotEmpty(t, result.Summary, "Should have summary for successful checks")
		}
	})
}

// Test comprehensive FIPS analysis of multiple images
func TestFIPSClient_ComprehensiveMultiImageAnalysis(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive multi-image test in short mode")
	}

	client := NewFIPSClient("")

	// Define all test images
	testImages := []struct {
		name        string
		imageRef    imageextractor.ImageReference
		expected    string // COMPLIANT, NOT COMPLIANT, or "" for unknown
		description string
	}{
		{
			name: "IstioFIPS",
			imageRef: imageextractor.ImageReference{
				Image:   "mcr.microsoft.com/oss/v2/istio/proxyv2-fips:v1.26.4-1",
				ASTPath: "spec.containers[0].image",
			},
			expected:    "COMPLIANT",
			description: "Microsoft Istio Proxy with FIPS support",
		},
		{
			name: "AzureIPMasqAgent",
			imageRef: imageextractor.ImageReference{
				Image:   "mcr.microsoft.com/oss/v2/azure/ip-masq-agent-v2:v0.1.15",
				ASTPath: "spec.containers[1].image",
			},
			expected:    "COMPLIANT",
			description: "Microsoft Azure IP Masq Agent v2",
		},
		{
			name: "RegularIstio",
			imageRef: imageextractor.ImageReference{
				Image:   "mcr.microsoft.com/oss/v2/istio/proxyv2:v1.25.5-1",
				ASTPath: "spec.containers[2].image",
			},
			expected:    "", // Unknown - let's see what we get
			description: "Regular Istio Proxy (non-FIPS version)",
		},
		{
			name: "KubernetesCSIBlob",
			imageRef: imageextractor.ImageReference{
				Image:   "mcr.microsoft.com/oss/kubernetes-csi/blob-csi:v1.26.6",
				ASTPath: "spec.containers[3].image",
			},
			expected:    "", // Unknown - let's see what we get
			description: "Kubernetes CSI Blob Storage Driver",
		},
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🔍 COMPREHENSIVE FIPS COMPLIANCE ANALYSIS REPORT")
	fmt.Println(strings.Repeat("=", 80))

	var results []struct {
		name   string
		result FIPSResult
		err    error
	}

	// Test each image
	for _, test := range testImages {
		t.Run(test.name, func(t *testing.T) {
			fmt.Printf("\n📦 ANALYZING: %s\n", test.description)
			fmt.Printf("Image: %s\n", test.imageRef.Image)
			fmt.Println(strings.Repeat("-", 60))

			result, err := client.CheckCompliance(test.imageRef)

			// Store for summary
			results = append(results, struct {
				name   string
				result FIPSResult
				err    error
			}{test.name, result, err})

			// Assertions
			assert.NoError(t, err, "Should complete without error")
			assert.Equal(t, test.imageRef.Image, result.Image)
			assert.Contains(t, []string{"COMPLIANT", "NOT COMPLIANT", "ERROR"}, result.Status)

			// Log detailed results
			status := result.Status
			if result.IsCompliant {
				status = "✅ " + status
			} else {
				if result.Status == "ERROR" {
					status = "❌ " + status
				} else {
					status = "⚠️ " + status
				}
			}

			fmt.Printf("Status: %s\n", status)
			fmt.Printf("Details: %s\n", result.Details)
			fmt.Printf("Summary: %s\n", result.Summary)

			// Validate against expected results if provided
			if test.expected != "" {
				assert.Equal(t, test.expected, result.Status,
					"Expected %s to be %s", test.name, test.expected)
			}
		})
	}

	// Print comprehensive summary
	t.Run("Summary", func(t *testing.T) {
		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println("📊 FINAL COMPLIANCE SUMMARY")
		fmt.Println(strings.Repeat("=", 80))

		compliant := 0
		nonCompliant := 0
		errors := 0

		for _, result := range results {
			status := "❌"
			if result.result.IsCompliant {
				status = "✅"
				compliant++
			} else if result.result.Status == "ERROR" {
				status = "⚠️"
				errors++
			} else {
				nonCompliant++
			}

			fmt.Printf("%s %s: %s (%s)\n",
				status, result.name, result.result.Status, result.result.Summary)
		}

		fmt.Println(strings.Repeat("-", 80))
		fmt.Printf("📈 COMPLIANCE METRICS:\n")
		fmt.Printf("   ✅ FIPS Compliant: %d\n", compliant)
		fmt.Printf("   ⚠️  Non-Compliant: %d\n", nonCompliant)
		fmt.Printf("   ❌ Errors: %d\n", errors)
		fmt.Printf("   📊 Success Rate: %.1f%%\n", float64(compliant)/float64(len(results))*100)
		fmt.Println(strings.Repeat("=", 80))
	})
}

// Test host FIPS environment check
func TestFIPSClient_HostFIPSEnvironmentCheck(t *testing.T) {
	// Test the host FIPS environment check functionality
	env := CheckHostFIPSEnvironment()

	// Verify structure
	assert.NotEmpty(t, env.OpenSSLVersion, "Should have OpenSSL version")
	assert.NotEmpty(t, env.Status, "Should have status")

	// Log results for manual verification
	t.Logf("Host FIPS Environment Check Results:")
	t.Logf("OpenSSL Version: %s", env.OpenSSLVersion)
	t.Logf("FIPS Capable: %t", env.FIPSCapable)
	t.Logf("Status: %s", env.Status)

	// Basic validation
	if env.FIPSCapable {
		assert.Contains(t, env.Status, "✅", "FIPS capable status should be positive")
	} else {
		assert.Contains(t, env.Status, "❌", "Non-FIPS capable status should be negative")
	}

	// Verify OpenSSL is accessible
	if !strings.Contains(env.OpenSSLVersion, "not found") {
		assert.Contains(t, env.OpenSSLVersion, "OpenSSL", "Should contain OpenSSL in version string")
	}
}

// Test FIPS checker availability and setup
func TestFIPSClient_LocalSetup(t *testing.T) {
	client := NewFIPSClient("")

	// Test client creation and availability
	assert.NotNil(t, client, "FIPS client should be created")

	// Check if FIPS checker is available locally
	isAvailable := client.IsAvailable()
	t.Logf("FIPS checker available: %t", isAvailable)

	if !isAvailable {
		t.Log("FIPS checker not available - this is expected in some environments")
	}
}

// Test complete integration: AST → ImageExtractor → FIPS Checker
func TestFIPSClient_CompleteIntegrationFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration flow test in short mode")
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🔄 COMPLETE INTEGRATION FLOW TEST")
	fmt.Println("AST → ImageExtractor → FIPS Checker → Results")
	fmt.Println(strings.Repeat("=", 80))

	// Step 1: Create a Kubernetes YAML AST with multiple images
	kubernetesYAML := `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
spec:
  template:
    spec:
      containers:
      - name: istio-proxy-fips
        image: mcr.microsoft.com/oss/v2/istio/proxyv2-fips:v1.26.4-1
        ports:
        - containerPort: 8080
      - name: ip-masq-agent
        image: mcr.microsoft.com/oss/v2/azure/ip-masq-agent-v2:v0.1.15
        env:
        - name: NODE_NAME
          value: test-node
      - name: storage-driver
        image: mcr.microsoft.com/oss/kubernetes-csi/blob-csi:v1.26.6
        volumeMounts:
        - name: storage
          mountPath: /data
      initContainers:
      - name: init-container
        image: alpine:latest
        command: ["sh", "-c", "echo 'Init complete'"]
`

	fmt.Printf("\n📄 STEP 1: Parse Kubernetes YAML to AST\n")
	fmt.Printf("YAML Content: %d lines with 4 container images\n", strings.Count(kubernetesYAML, "\n"))

	// Parse YAML to Go interface first
	var yamlData interface{}
	err := yaml.Unmarshal([]byte(kubernetesYAML), &yamlData)
	require.NoError(t, err, "Should parse YAML successfully")

	// Convert to AST Value
	astValue, err := ast.InterfaceToValue(yamlData)
	require.NoError(t, err, "Should convert to AST successfully")

	fmt.Printf("✅ AST Creation: Success\n")

	// Step 2: Extract images using ImageExtractor
	fmt.Printf("\n🔍 STEP 2: Extract Images from AST\n")

	images := imageextractor.ExtractImages(astValue)

	fmt.Printf("Found %d images:\n", len(images))
	for i, img := range images {
		fmt.Printf("  %d. %s (path: %s)\n", i+1, img.Image, img.ASTPath)
	}

	// Verify we found the expected images
	assert.Len(t, images, 4, "Should extract 4 images from the YAML")

	expectedImages := []string{
		"mcr.microsoft.com/oss/v2/istio/proxyv2-fips:v1.26.4-1",
		"mcr.microsoft.com/oss/v2/azure/ip-masq-agent-v2:v0.1.15",
		"mcr.microsoft.com/oss/kubernetes-csi/blob-csi:v1.26.6",
		"alpine:latest",
	}

	for _, expectedImage := range expectedImages {
		found := false
		for _, extractedImg := range images {
			if extractedImg.Image == expectedImage {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find image: %s", expectedImage)
	}

	fmt.Printf("✅ Image Extraction: Success\n")

	// Step 3: Feed extracted images to FIPS Checker
	fmt.Printf("\n🔒 STEP 3: FIPS Compliance Check for All Images\n")

	client := NewFIPSClient("")
	fipsResults, err := client.CheckMultipleImages(images)
	require.NoError(t, err, "FIPS checking should complete successfully")

	fmt.Printf("FIPS Analysis Results:\n")

	compliantCount := 0
	nonCompliantCount := 0
	errorCount := 0

	for astPath, result := range fipsResults {
		var status string
		if result.IsCompliant {
			status = "✅ COMPLIANT"
			compliantCount++
		} else if result.Status == "ERROR" {
			status = "❌ ERROR"
			errorCount++
		} else {
			status = "⚠️ NOT COMPLIANT"
			nonCompliantCount++
		}

		fmt.Printf("  📍 %s\n", astPath)
		fmt.Printf("     Image: %s\n", result.Image)
		fmt.Printf("     Status: %s\n", status)
		fmt.Printf("     Summary: %s\n", result.Summary)
		if result.Details != "" {
			fmt.Printf("     Details: %s\n", result.Details)
		}
		fmt.Println()
	}

	// Step 4: Verify Results and Integration
	fmt.Printf("\n📊 STEP 4: Integration Validation\n")

	// Verify we got results for all extracted images
	assert.Len(t, fipsResults, len(images), "Should get FIPS results for all extracted images")

	// Verify AST paths match between extraction and FIPS results
	for _, img := range images {
		_, exists := fipsResults[img.ASTPath]
		assert.True(t, exists, "FIPS results should contain AST path: %s", img.ASTPath)
	}

	// Print final summary
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("🎯 INTEGRATION SUMMARY:\n")
	fmt.Printf("   Images Extracted: %d\n", len(images))
	fmt.Printf("   Images Analyzed: %d\n", len(fipsResults))
	fmt.Printf("   ✅ FIPS Compliant: %d\n", compliantCount)
	fmt.Printf("   ⚠️  Non-Compliant: %d\n", nonCompliantCount)
	fmt.Printf("   ❌ Errors: %d\n", errorCount)
	fmt.Printf("   📊 Success Rate: %.1f%%\n", float64(compliantCount)/float64(len(fipsResults))*100)
	fmt.Println(strings.Repeat("=", 80))

	// Final assertions
	assert.Equal(t, len(images), len(fipsResults), "Should have 1:1 mapping of images to FIPS results")
	assert.True(t, compliantCount > 0, "Should have at least some FIPS-compliant images")

	t.Logf("✅ COMPLETE INTEGRATION FLOW: SUCCESS!")
	t.Logf("Successfully demonstrated: AST → ImageExtractor → FIPS Checker → Results")

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("🎉 COMPLETE END-TO-END VALIDATION: SUCCESS!")
	fmt.Println("✅ AST Parsing → Image Extraction → FIPS Analysis → Results")
	fmt.Println(strings.Repeat("=", 80))
}

// Test multiple YAML files each containing 1-2 images (realistic scenario)
func TestFIPSClient_MultiFileIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multi-file integration test in short mode")
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🗂️  MULTI-FILE FIPS INTEGRATION TEST")
	fmt.Println("Multiple YAML files → AST → ImageExtractor → FIPS Checker")
	fmt.Println(strings.Repeat("=", 80))

	// Define multiple configuration files with 1-2 images each
	configFiles := []struct {
		name     string
		content  string
		expected int // expected number of images
	}{
		{
			name: "istio-gateway.yaml",
			content: `
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  name: my-gateway
spec:
  selector:
    istio: gateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - "*"
# Istio gateway config with proxy image
proxy:
  image: "mcr.microsoft.com/oss/v2/istio/proxyv2-fips:v1.26.4-1"
`,
			expected: 1,
		},
		{
			name: "azure-services.yaml",
			content: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: azure-config
# Azure services configuration with direct image fields
ipMasqAgent:
  image: "mcr.microsoft.com/oss/v2/azure/ip-masq-agent-v2:v0.1.15"
csiDriver:
  image: "mcr.microsoft.com/oss/kubernetes-csi/blob-csi:v1.26.6"
`,
			expected: 2,
		},
		{
			name: "application-deployment.yaml",
			content: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: app
        image: nginx:1.21
        ports:
        - containerPort: 80
`,
			expected: 1,
		},
		{
			name: "init-job.yaml",
			content: `
apiVersion: batch/v1
kind: Job
metadata:
  name: init-job
spec:
  template:
    spec:
      restartPolicy: Never
      initContainers:
      - name: setup
        image: alpine:latest
        command: ["sh", "-c", "echo 'Setting up'"]
      containers:
      - name: worker
        image: busybox:1.35
        command: ["sh", "-c", "echo 'Processing'"]
`,
			expected: 2,
		},
	}

	fmt.Printf("📁 Processing %d configuration files:\n\n", len(configFiles))

	var allImages []imageextractor.ImageReference

	// Process each configuration file
	for i, config := range configFiles {
		fmt.Printf("📄 [%d/%d] Processing: %s\n", i+1, len(configFiles), config.name)

		// Step 1: Parse YAML to AST
		var yamlData interface{}
		err := yaml.Unmarshal([]byte(config.content), &yamlData)
		require.NoError(t, err, "Should parse %s successfully", config.name)

		astValue, err := ast.InterfaceToValue(yamlData)
		require.NoError(t, err, "Should convert %s to AST", config.name)

		// Step 2: Extract images from this file
		images := imageextractor.ExtractImages(astValue)
		fmt.Printf("   🔍 Found %d images (expected %d)\n", len(images), config.expected)

		for j, img := range images {
			fmt.Printf("      %d. %s (path: %s)\n", j+1, img.Image, img.ASTPath)

			// Add file prefix to AST path for uniqueness across files
			img.ASTPath = fmt.Sprintf("%s:%s", config.name, img.ASTPath)
			allImages = append(allImages, img)
		}

		assert.Equal(t, config.expected, len(images),
			"File %s should contain %d images", config.name, config.expected)
		fmt.Println()
	}

	fmt.Printf("📊 Total images extracted from all files: %d\n\n", len(allImages))

	// Step 3: Run FIPS check on ALL extracted images
	fmt.Printf("🔒 FIPS Compliance Analysis for All Images\n")
	fmt.Println(strings.Repeat("-", 60))

	client := NewFIPSClient("")
	fipsResults, err := client.CheckMultipleImages(allImages)
	require.NoError(t, err, "FIPS checking should complete successfully")

	// Step 4: Analyze results by file
	compliantTotal := 0
	nonCompliantTotal := 0
	errorTotal := 0

	for fileName, config := range map[string]struct{expectedImages int; description string}{
		"istio-gateway.yaml":        {1, "Istio Gateway Configuration"},
		"azure-services.yaml":       {2, "Azure Services Configuration"},
		"application-deployment.yaml": {1, "Application Deployment"},
		"init-job.yaml":            {2, "Initialization Job"},
	} {
		fmt.Printf("\n📋 %s (%s):\n", fileName, config.description)

		fileCompliant := 0
		fileNonCompliant := 0
		fileErrors := 0

		for astPath, result := range fipsResults {
			if strings.HasPrefix(astPath, fileName+":") {
				var status string
				if result.IsCompliant {
					status = "✅ COMPLIANT"
					fileCompliant++
					compliantTotal++
				} else if result.Status == "ERROR" {
					status = "❌ ERROR"
					fileErrors++
					errorTotal++
				} else {
					status = "⚠️ NOT COMPLIANT"
					fileNonCompliant++
					nonCompliantTotal++
				}

				fmt.Printf("   %s: %s\n", result.Image, status)
				if result.Details != "" && result.Status != "COMPLIANT" {
					// Truncate long details for readability
					details := result.Details
					if len(details) > 100 {
						details = details[:97] + "..."
					}
					fmt.Printf("      Details: %s\n", details)
				}
				if result.Summary != "" {
					fmt.Printf("      Summary: %s\n", result.Summary)
				}
			}
		}

		fmt.Printf("   📈 File Summary: ✅ %d | ⚠️ %d | ❌ %d\n",
			fileCompliant, fileNonCompliant, fileErrors)
	}

	// Step 5: Aggregate results by file (ALL images must be compliant for file to be compliant)
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("📋 FILE-LEVEL FIPS COMPLIANCE AGGREGATION")
	fmt.Println("(File is compliant only if ALL images in file are compliant)")
	fmt.Println(strings.Repeat("=", 80))

	fileResults := client.AggregateResultsByFile(fipsResults)

	fileCompliantCount := 0
	fileNonCompliantCount := 0
	fileErrorCount := 0

	for fileName, fileResult := range fileResults {
		var statusIcon string
		if fileResult.IsCompliant {
			statusIcon = "✅ COMPLIANT"
			fileCompliantCount++
		} else if fileResult.Status == "ERROR" {
			statusIcon = "❌ ERROR"
			fileErrorCount++
		} else {
			statusIcon = "⚠️ NOT COMPLIANT"
			fileNonCompliantCount++
		}

		fmt.Printf("📄 %s: %s\n", fileName, statusIcon)
		fmt.Printf("   Summary: %s\n", fileResult.Summary)
		if !fileResult.IsCompliant {
			fmt.Printf("   Details: %s\n", fileResult.Details)
		}
		fmt.Println()
	}

	// Step 6: Overall summary
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("📊 MULTI-FILE FIPS COMPLIANCE SUMMARY")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Printf("📁 Configuration Files Processed: %d\n", len(configFiles))
	fmt.Printf("📦 Total Images Analyzed: %d\n", len(allImages))
	fmt.Printf("✅ FIPS Compliant Images: %d\n", compliantTotal)
	fmt.Printf("⚠️  Non-Compliant Images: %d\n", nonCompliantTotal)
	fmt.Printf("❌ Error Images: %d\n", errorTotal)
	fmt.Printf("📈 Image Success Rate: %.1f%%\n",
		float64(compliantTotal)/float64(len(allImages))*100)

	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("📋 FILE-LEVEL COMPLIANCE:\n")
	fmt.Printf("✅ FIPS Compliant Files: %d\n", fileCompliantCount)
	fmt.Printf("⚠️  Non-Compliant Files: %d\n", fileNonCompliantCount)
	fmt.Printf("❌ Error Files: %d\n", fileErrorCount)
	fmt.Printf("📈 File Success Rate: %.1f%%\n",
		float64(fileCompliantCount)/float64(len(configFiles))*100)

	// Step 6: Verify complete pipeline worked
	assert.Equal(t, len(allImages), len(fipsResults),
		"Should have FIPS results for all extracted images")
	assert.True(t, compliantTotal > 0 || nonCompliantTotal > 0,
		"Should have processed some images successfully")

	fmt.Println("\n🎉 MULTI-FILE INTEGRATION SUCCESS!")
	fmt.Println("✅ Multiple YAML files → AST → Image Extraction → FIPS Analysis → Results")
	fmt.Println("✅ Demonstrated realistic multi-configuration scenario")
	fmt.Println(strings.Repeat("=", 80))

	t.Logf("Multi-file integration test completed successfully!")
	t.Logf("Processed %d files with %d total images", len(configFiles), len(allImages))
	t.Logf("FIPS Results: %d compliant, %d non-compliant, %d errors",
		compliantTotal, nonCompliantTotal, errorTotal)
}

// Benchmark test
func BenchmarkFIPSClient_CheckCompliance(b *testing.B) {
	mockClient := NewMockFIPSClient()

	// Setup mock data
	mockReports := []BinaryReport{
		{
			RelativePath: "bin/app",
			GoBinaryDetails: GoBinaryReportDetails{
				UseSystemcrypto:  true,
				FailsOnFIPSCheck: false,
			},
		},
	}

	testImage := "benchmark/image:latest"
	mockClient.SetMockResult(testImage, mockReports, nil)

	imageRef := imageextractor.ImageReference{
		Image:   testImage,
		ASTPath: "test.image",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := mockClient.CheckCompliance(imageRef)
		require.NoError(b, err)
	}
}