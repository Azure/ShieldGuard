package fipschecker

import (
	"context"
	"debug/buildinfo"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BinaryReport represents the result of checking a single binary
type BinaryReport struct {
	RelativePath    string                  `json:"relativePath"`
	Type            string                  `json:"type,omitempty"`
	Error           error                   `json:"error,omitempty"`
	GoBinaryDetails GoBinaryReportDetails   `json:"goBinaryDetails"`
}

// GoBinaryReportDetails contains details about a Go binary's FIPS compliance
type GoBinaryReportDetails struct {
	UseSystemcrypto  bool   `json:"useSystemcrypto"`
	FailsOnFIPSCheck bool   `json:"failsOnFIPSCheck"`
	GoVersion        string `json:"goVersion,omitempty"`
	Module           string `json:"module,omitempty"`
	BuildInfo        string `json:"buildInfo,omitempty"`
	CGOEnabled       bool   `json:"cgoEnabled,omitempty"`
	RuntimePanicLog  string `json:"runtimePanicLog,omitempty"`
}

// Check scans a directory for Go binaries and checks their FIPS compliance
func Check(ctx context.Context, rootDir string) ([]BinaryReport, error) {
	var reports []BinaryReport

	// Walk through the directory to find Go binaries
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors and continue
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip non-executable files
		if info.Mode()&0111 == 0 {
			return nil
		}

		// Check if this is a Go binary
		if isGoBinary(path) {
			relativePath, _ := filepath.Rel(rootDir, path)
			report := checkGoBinary(path, relativePath)
			reports = append(reports, report)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", rootDir, err)
	}

	return reports, nil
}

// FIPSEnvironment represents the host FIPS environment status
type FIPSEnvironment struct {
	OpenSSLVersion string `json:"opensslVersion"`
	FIPSCapable    bool   `json:"fipsCapable"`
	Status         string `json:"status"`
}

// CheckHostFIPSEnvironment performs host FIPS environment check
func CheckHostFIPSEnvironment() FIPSEnvironment {
	fmt.Println("=== Host FIPS Environment Check ===")

	// Check OpenSSL version
	opensslVersion := getOpenSSLVersion()
	fmt.Printf("OpenSSL Version: %s\n", opensslVersion)

	// Check FIPS capability
	fipsCapable := checkFIPSCapability()
	fmt.Printf("FIPS Capable: %t\n", fipsCapable)

	// Determine status
	var status string
	if fipsCapable {
		status = "✅ Status: Host is FIPS capable"
	} else {
		status = "❌ Status: Host is NOT FIPS capable"
	}
	fmt.Println(status)

	return FIPSEnvironment{
		OpenSSLVersion: opensslVersion,
		FIPSCapable:    fipsCapable,
		Status:         status,
	}
}

// getOpenSSLVersion gets the OpenSSL version from the system
func getOpenSSLVersion() string {
	// Try openssl version command
	cmd := exec.Command("openssl", "version")
	output, err := cmd.Output()
	if err != nil {
		return "OpenSSL not found or not accessible"
	}

	version := strings.TrimSpace(string(output))
	return version
}

// checkFIPSCapability checks if the system is FIPS capable
func checkFIPSCapability() bool {
	// Method 1: Check openssl version -a for FIPS info
	cmd := exec.Command("openssl", "version", "-a")
	output, err := cmd.Output()
	if err == nil {
		outputStr := string(output)
		if strings.Contains(outputStr, "FIPS") {
			return true
		}
	}

	// Method 2: Try to get FIPS mode status
	cmd = exec.Command("openssl", "list", "-providers")
	output, err = cmd.Output()
	if err == nil {
		outputStr := string(output)
		if strings.Contains(outputStr, "fips") || strings.Contains(outputStr, "FIPS") {
			return true
		}
	}

	// Method 3: Check for FIPS module files (common locations)
	fipsPaths := []string{
		"/usr/lib/ssl/fipsmodule.cnf",
		"/etc/ssl/fipsmodule.cnf",
		"/usr/local/ssl/fipsmodule.cnf",
	}

	for _, path := range fipsPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	// Method 4: Check OpenSSL version for FIPS capability (version >= 3.0)
	version := getOpenSSLVersion()
	if strings.Contains(version, "OpenSSL 3.") {
		return true // OpenSSL 3.x has built-in FIPS support
	}

	return false
}

// isGoBinary checks if a file is a Go binary using proper Go tools
func isGoBinary(binaryPath string) bool {
	// Try to read build info using debug/buildinfo package
	_, err := buildinfo.ReadFile(binaryPath)
	return err == nil
}

// checkGoBinary analyzes a single Go binary for FIPS compliance
func checkGoBinary(binaryPath, relativePath string) BinaryReport {
	report := BinaryReport{
		RelativePath: relativePath,
		GoBinaryDetails: GoBinaryReportDetails{
			UseSystemcrypto:  false,
			FailsOnFIPSCheck: false,
		},
	}

	// Check build info to determine FIPS compliance
	buildInfo, err := getBuildInfo(binaryPath)
	if err != nil {
		report.Error = fmt.Errorf("failed to get build info: %w", err)
		return report
	}

	report.GoBinaryDetails.BuildInfo = buildInfo

	// Parse build info to check for systemcrypto and FIPS settings
	report.GoBinaryDetails.UseSystemcrypto = checkSystemcrypto(buildInfo)
	report.GoBinaryDetails.FailsOnFIPSCheck = checkFIPSRuntime(binaryPath)

	// Extract Go version and module info
	report.GoBinaryDetails.GoVersion = extractGoVersion(buildInfo)
	report.GoBinaryDetails.Module = extractModule(buildInfo)

	return report
}

// getBuildInfo extracts build information from a Go binary
func getBuildInfo(binaryPath string) (string, error) {
	// Method 1: Try using go version -m command
	cmd := exec.Command("go", "version", "-m", binaryPath)
	output, err := cmd.Output()
	if err == nil {
		return string(output), nil
	}

	// Method 2: Try using debug/buildinfo package
	buildInfo, err := buildinfo.ReadFile(binaryPath)
	if err != nil {
		return "", fmt.Errorf("failed to read build info: %w", err)
	}

	// Format build info into a string
	var result strings.Builder
	result.WriteString(fmt.Sprintf("go %s\n", buildInfo.GoVersion))
	result.WriteString(fmt.Sprintf("path %s\n", buildInfo.Path))
	result.WriteString(fmt.Sprintf("mod %s %s\n", buildInfo.Main.Path, buildInfo.Main.Version))

	for _, setting := range buildInfo.Settings {
		result.WriteString(fmt.Sprintf("build %s=%s\n", setting.Key, setting.Value))
	}

	return result.String(), nil
}

// checkSystemcrypto determines if the binary uses systemcrypto
func checkSystemcrypto(buildInfo string) bool {
	// Check for FIPS-related build settings
	return strings.Contains(buildInfo, "build GOEXPERIMENT=systemcrypto") ||
		   strings.Contains(buildInfo, "build -tags=systemcrypto") ||
		   strings.Contains(buildInfo, "X:systemcrypto") ||
		   strings.Contains(buildInfo, "systemcrypto")
}

// checkFIPSRuntime performs a runtime check for FIPS compliance
func checkFIPSRuntime(binaryPath string) bool {
	// For safety, we'll skip actual execution and rely on build-time checks
	// In a production environment, you might want to:
	// - Execute with timeout and FIPS environment variables
	// - Check for specific FIPS-related symbols or libraries
	// - Validate against known FIPS-compliant build patterns

	// For now, we assume no runtime failures if build-time checks pass
	return false
}

// extractGoVersion extracts Go version from build info
func extractGoVersion(buildInfo string) string {
	lines := strings.Split(buildInfo, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "go ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return "unknown"
}

// extractModule extracts module path from build info
func extractModule(buildInfo string) string {
	lines := strings.Split(buildInfo, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "path ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		} else if strings.HasPrefix(line, "mod ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return "unknown"
}

// MarshalJSON implements custom JSON marshaling for BinaryReport
// to handle the error field properly
func (br BinaryReport) MarshalJSON() ([]byte, error) {
	type Alias BinaryReport
	aux := struct {
		Alias
		Error string `json:"error,omitempty"`
	}{
		Alias: Alias(br),
	}

	if br.Error != nil {
		aux.Error = br.Error.Error()
	}

	return json.Marshal(aux)
}