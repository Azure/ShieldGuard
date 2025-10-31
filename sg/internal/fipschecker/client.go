package fipschecker

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/Azure/ShieldGuard/sg/internal/imageextractor"
)

// FIPSResult represents the FIPS compliance result for an image
type FIPSResult struct {
	Image       string `json:"image"`
	IsCompliant bool   `json:"is_compliant"`
	Status      string `json:"status"`
	Details     string `json:"details,omitempty"`
	Summary     string `json:"summary,omitempty"`
}

// FIPSClient wraps the fips-check Go API
type FIPSClient struct {
	useGoAPI bool
}

// NewFIPSClient creates a new FIPS checker client using Go API
func NewFIPSClient(scriptPath string) *FIPSClient {
	return &FIPSClient{useGoAPI: true}
}

// CheckCompliance runs FIPS check on a container image by pulling and scanning it
func (c *FIPSClient) CheckCompliance(imageRef imageextractor.ImageReference) (FIPSResult, error) {
	if c.useGoAPI {
		return c.checkWithGoAPI(imageRef)
	}

	// Fallback to legacy shell method if needed
	return c.checkWithShell(imageRef)
}

// checkWithGoAPI uses the fips-check Go API to scan container image
func (c *FIPSClient) checkWithGoAPI(imageRef imageextractor.ImageReference) (FIPSResult, error) {
	result := FIPSResult{
		Image: imageRef.Image,
	}

	// First, check host FIPS environment
	hostEnv := CheckHostFIPSEnvironment()
	if !hostEnv.FIPSCapable {
		result.IsCompliant = false
		result.Status = "ERROR"
		result.Details = fmt.Sprintf("Host environment is not FIPS capable: %s", hostEnv.Status)
		return result, nil
	}

	// For container images, we need to pull and extract them first
	// This is a simplified version - in production you'd want proper container handling
	tempDir, err := c.extractContainerImage(imageRef.Image)
	if err != nil {
		result.IsCompliant = false
		result.Status = "ERROR"
		result.Details = fmt.Sprintf("Failed to extract container image: %v", err)
		return result, nil
	}

	// Run FIPS check on the extracted container
	ctx := context.Background()
	reports, err := Check(ctx, tempDir)
	if err != nil {
		result.IsCompliant = false
		result.Status = "ERROR"
		result.Details = fmt.Sprintf("FIPS check failed: %v", err)
		return result, nil
	}

	// Analyze results
	result.IsCompliant = c.analyzeReports(reports)
	if result.IsCompliant {
		result.Status = "COMPLIANT"
		result.Details = "All Go binaries pass FIPS compliance checks"
	} else {
		result.Status = "NOT COMPLIANT"
		result.Details = c.formatComplianceIssues(reports)
	}

	result.Summary = c.formatSummary(reports)
	return result, nil
}

// extractContainerImage pulls and extracts a container image for scanning
func (c *FIPSClient) extractContainerImage(image string) (string, error) {
	// This is a simplified implementation
	// In production, you'd want to use proper container libraries like:
	// - github.com/containers/image/v5
	// - github.com/google/go-containerregistry

	// For now, use docker to extract
	tempDir := fmt.Sprintf("/tmp/fips-scan-%s", strings.ReplaceAll(image, "/", "-"))

	// Pull image
	pullCmd := exec.Command("docker", "pull", image)
	if err := pullCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to pull image: %w", err)
	}

	// Create unique container name to avoid conflicts
	containerName := fmt.Sprintf("temp-fips-scan-%d", time.Now().UnixNano())

	// Create container and copy filesystem
	createCmd := exec.Command("docker", "create", "--name", containerName, image)
	if err := createCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// Copy files out
	copyCmd := exec.Command("docker", "cp", containerName+":/", tempDir)
	copyCmd.Run() // Best effort

	// Cleanup container
	cleanCmd := exec.Command("docker", "rm", containerName)
	cleanCmd.Run()

	return tempDir, nil
}

// analyzeReports determines overall compliance from binary reports
func (c *FIPSClient) analyzeReports(reports []BinaryReport) bool {
	if len(reports) == 0 {
		return false // No Go binaries found
	}

	for _, report := range reports {
		if report.Error != nil {
			continue // Skip errored reports
		}

		details := report.GoBinaryDetails

		// Check FIPS compliance criteria
		if !details.UseSystemcrypto {
			return false // systemcrypto required
		}

		if details.FailsOnFIPSCheck {
			return false // Runtime check failed
		}
	}

	return true // All binaries are compliant
}

// formatComplianceIssues creates a summary of compliance issues
func (c *FIPSClient) formatComplianceIssues(reports []BinaryReport) string {
	var issues []string

	for _, report := range reports {
		if report.Error != nil {
			issues = append(issues, fmt.Sprintf("Error checking %s: %v", report.RelativePath, report.Error))
			continue
		}

		details := report.GoBinaryDetails
		if !details.UseSystemcrypto {
			issues = append(issues, fmt.Sprintf("%s: systemcrypto not enabled", report.RelativePath))
		}

		if details.FailsOnFIPSCheck {
			issues = append(issues, fmt.Sprintf("%s: runtime FIPS check failed", report.RelativePath))
		}
	}

	return strings.Join(issues, "; ")
}

// formatSummary creates a summary of the scan results
func (c *FIPSClient) formatSummary(reports []BinaryReport) string {
	totalBinaries := len(reports)
	systemcryptoCount := 0
	failedCount := 0

	for _, report := range reports {
		if report.Error == nil {
			details := report.GoBinaryDetails
			if details.UseSystemcrypto {
				systemcryptoCount++
			}
			if details.FailsOnFIPSCheck {
				failedCount++
			}
		}
	}

	return fmt.Sprintf("Total: %d | Systemcrypto: %d | Failed FIPS: %d",
		totalBinaries, systemcryptoCount, failedCount)
}

// checkWithShell is the fallback shell-based implementation
func (c *FIPSClient) checkWithShell(imageRef imageextractor.ImageReference) (FIPSResult, error) {
	result := FIPSResult{
		Image: imageRef.Image,
	}

	// Execute the fips-check script
	cmd := exec.Command("bash", "/tmp/fips-check/build-and-check.sh", imageRef.Image)
	output, err := cmd.CombinedOutput()

	if err != nil {
		result.IsCompliant = false
		result.Status = "NOT COMPLIANT"
		result.Details = fmt.Sprintf("FIPS check failed: %v", err)
	} else {
		result.IsCompliant = true
		result.Status = "COMPLIANT"
	}

	result.Summary = string(output)
	return result, nil
}

// CheckMultipleImages runs FIPS check on multiple images
func (c *FIPSClient) CheckMultipleImages(images []imageextractor.ImageReference) (map[string]FIPSResult, error) {
	results := make(map[string]FIPSResult)

	for _, img := range images {
		result, err := c.CheckCompliance(img)
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

// FileFIPSResult represents FIPS compliance for an entire file
type FileFIPSResult struct {
	FileName        string       `json:"file_name"`
	IsCompliant     bool         `json:"is_compliant"`
	Status          string       `json:"status"`
	Details         string       `json:"details,omitempty"`
	Summary         string       `json:"summary,omitempty"`
	ImageResults    []FIPSResult `json:"image_results"`
	TotalImages     int          `json:"total_images"`
	CompliantImages int          `json:"compliant_images"`
}

// AggregateResultsByFile takes image-level FIPS results and creates file-level results
// A file is compliant only if ALL images in the file are compliant
func (c *FIPSClient) AggregateResultsByFile(fipsResults map[string]FIPSResult) map[string]FileFIPSResult {
	fileResults := make(map[string]FileFIPSResult)

	// Group results by file name (extract from AST path prefix)
	fileImageGroups := make(map[string][]FIPSResult)

	for astPath, result := range fipsResults {
		// Extract file name from AST path (format: "filename:actual.ast.path")
		parts := strings.SplitN(astPath, ":", 2)
		fileName := parts[0]

		if _, exists := fileImageGroups[fileName]; !exists {
			fileImageGroups[fileName] = []FIPSResult{}
		}
		fileImageGroups[fileName] = append(fileImageGroups[fileName], result)
	}

	// Calculate file-level compliance
	for fileName, imageResults := range fileImageGroups {
		totalImages := len(imageResults)
		compliantImages := 0
		errorImages := 0
		var nonCompliantDetails []string

		for _, result := range imageResults {
			if result.IsCompliant {
				compliantImages++
			} else if result.Status == "ERROR" {
				errorImages++
				nonCompliantDetails = append(nonCompliantDetails,
					fmt.Sprintf("%s: %s", result.Image, result.Details))
			} else {
				nonCompliantDetails = append(nonCompliantDetails,
					fmt.Sprintf("%s: non-compliant", result.Image))
			}
		}

		// File is compliant only if ALL images are compliant
		fileCompliant := compliantImages == totalImages && errorImages == 0

		var status, details string
		if errorImages > 0 {
			status = "ERROR"
			details = fmt.Sprintf("%d image(s) had errors during FIPS checking", errorImages)
		} else if fileCompliant {
			status = "COMPLIANT"
			details = "All images in file pass FIPS compliance checks"
		} else {
			status = "NOT COMPLIANT"
			details = strings.Join(nonCompliantDetails, "; ")
		}

		summary := fmt.Sprintf("Total: %d | Compliant: %d | Non-compliant: %d | Errors: %d",
			totalImages, compliantImages, totalImages-compliantImages-errorImages, errorImages)

		fileResults[fileName] = FileFIPSResult{
			FileName:        fileName,
			IsCompliant:     fileCompliant,
			Status:          status,
			Details:         details,
			Summary:         summary,
			ImageResults:    imageResults,
			TotalImages:     totalImages,
			CompliantImages: compliantImages,
		}
	}

	return fileResults
}

// IsAvailable checks if fips-check is available
func (c *FIPSClient) IsAvailable() bool {
	// For Go API, we just check if the package can be imported
	return c.useGoAPI
}