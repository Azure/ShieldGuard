package fipschecker

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"

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

// FIPSClient wraps the fips-check shell script
type FIPSClient struct {
	fipsCheckPath string // Path to the fips-check repository directory
	cache         sync.Map // Cache for FIPS results by image name
}

// NewFIPSClient creates a new FIPS checker client using the shell script
func NewFIPSClient(fipsCheckPath string) *FIPSClient {
	if fipsCheckPath == "" {
		fipsCheckPath = "/tmp/fips-check" // Default location
	}

	return &FIPSClient{
		fipsCheckPath: fipsCheckPath,
		cache:         sync.Map{},
	}
}

// CheckCompliance runs FIPS check on a container image using shell script with caching
func (c *FIPSClient) CheckCompliance(imageRef imageextractor.ImageReference) (FIPSResult, error) {
	// Check cache first
	if cached, found := c.cache.Load(imageRef.Image); found {
		result := cached.(FIPSResult)
		// Update the ASTPath for this specific reference
		result.Image = imageRef.Image // Ensure consistency
		return result, nil
	}

	// Run the check and cache the result
	result, err := c.checkWithShellScript(imageRef)
	if err != nil {
		return result, err
	}

	// Cache the result by image name
	c.cache.Store(imageRef.Image, result)
	return result, nil
}






// checkWithShellScript runs FIPS check using the shell script
func (c *FIPSClient) checkWithShellScript(imageRef imageextractor.ImageReference) (FIPSResult, error) {
	result := FIPSResult{
		Image: imageRef.Image,
	}

	// Execute the fips-check script from the correct directory
	scriptPath := fmt.Sprintf("%s/build-and-check.sh", c.fipsCheckPath)
	cmd := exec.Command("bash", scriptPath, imageRef.Image)
	cmd.Dir = c.fipsCheckPath // Set working directory to fips-check directory
	output, err := cmd.CombinedOutput()

	outputStr := string(output)

	// Parse output to determine compliance
	if err != nil {
		result.IsCompliant = false
		result.Status = "NOT COMPLIANT"
		result.Details = fmt.Sprintf("FIPS check failed: %v", err)
	} else {
		// Check for actual compliance in output
		if strings.Contains(outputStr, "❌ FIPS Status: NOT COMPLIANT") ||
		   strings.Contains(outputStr, "Host is NOT FIPS capable") ||
		   strings.Contains(outputStr, "Failed FIPS:") && !strings.Contains(outputStr, "Failed FIPS: 0") {
			result.IsCompliant = false
			result.Status = "NOT COMPLIANT"
			result.Details = "Image does not meet FIPS compliance requirements"
		} else if strings.Contains(outputStr, "✅ FIPS Status: COMPLIANT") &&
		          strings.Contains(outputStr, "Failed FIPS: 0") &&
		          strings.Contains(outputStr, "Host is FIPS capable") {
			result.IsCompliant = true
			result.Status = "COMPLIANT"
			result.Details = "All Go binaries pass FIPS compliance checks"
		} else {
			result.IsCompliant = false
			result.Status = "NOT COMPLIANT"
			result.Details = "Unable to determine FIPS compliance from output"
		}
	}

	result.Summary = outputStr
	return result, nil
}

// CheckMultipleImages runs FIPS check on multiple images in parallel with caching
func (c *FIPSClient) CheckMultipleImages(images []imageextractor.ImageReference) (map[string]FIPSResult, error) {
	results := make(map[string]FIPSResult)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Use a semaphore to limit concurrent FIPS checks (Docker pulls can be resource intensive)
	semaphore := make(chan struct{}, 3) // Limit to 3 concurrent checks

	for _, img := range images {
		wg.Add(1)
		go func(imageRef imageextractor.ImageReference) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result, err := c.CheckCompliance(imageRef)
			if err != nil {
				result = FIPSResult{
					Image:       imageRef.Image,
					IsCompliant: false,
					Status:      "ERROR",
					Details:     fmt.Sprintf("Check failed: %v", err),
				}
			}

			mu.Lock()
			results[imageRef.ASTPath] = result
			mu.Unlock()
		}(img)
	}

	wg.Wait()
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

// IsAvailable checks if fips-check script is available
func (c *FIPSClient) IsAvailable() bool {
	scriptPath := fmt.Sprintf("%s/build-and-check.sh", c.fipsCheckPath)
	_, err := exec.LookPath("bash")
	if err != nil {
		return false
	}
	// Check if script exists
	if _, err := exec.Command("test", "-f", scriptPath).Output(); err != nil {
		return false
	}
	return true
}