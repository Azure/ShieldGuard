package imageextractor

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/open-policy-agent/opa/ast"
	"github.com/stretchr/testify/assert"
)

func TestExtractImages(t *testing.T) {
	// Test data matching your example
	jsonStr := `{
		"proxy": {
			"autoInject": "enabled",
			"clusterDomain": "cluster.local",
			"componentLogLevel": "misc:error",
			"excludeIPRanges": "",
			"excludeInboundPorts": "",
			"excludeOutboundPorts": "",
			"image": "mcr.microsoft.com/oss/v2/istio/proxyv2-fips:v1.26.4-1"
		}
	}`

	// Parse JSON to Go interface{} first
	var jsonData interface{}
	err := json.Unmarshal([]byte(jsonStr), &jsonData)
	assert.NoError(t, err)

	// Convert to AST Value
	value, err := ast.InterfaceToValue(jsonData)
	assert.NoError(t, err)

	// Extract images
	images := ExtractImages(value)

	// Verify results
	assert.Len(t, images, 1)
	assert.Equal(t, "mcr.microsoft.com/oss/v2/istio/proxyv2-fips:v1.26.4-1", images[0].Image)
	assert.Equal(t, "proxy.image", images[0].ASTPath)

	// IMPORTANT: Verify that the AST path can navigate back to the same image value
	// This ensures the image reference is correctly associated with the original AST
	t.Run("VerifyASTPathPointsToSameImage", func(t *testing.T) {
		// Navigate to the image using the extracted path
		foundValue := navigateASTPath(value, images[0].ASTPath)
		assert.NotNil(t, foundValue, "AST path should navigate to a valid value")

		// Verify the found value matches the original image string
		if stringValue, ok := foundValue.(ast.String); ok {
			assert.Equal(t, images[0].Image, string(stringValue),
				"AST path should navigate to the exact same image value")
		} else {
			t.Errorf("Expected ast.String but got %T", foundValue)
		}
	})
}

func TestExtractImages_ComplexAST(t *testing.T) {
	// More complex test with arrays and multiple images
	jsonStr := `{
		"proxy": {
			"image": "mcr.microsoft.com/oss/v2/istio/proxyv2-fips:v1.26.4-1"
		},
		"containers": [
			{
				"name": "app",
				"image": "nginx:1.21"
			},
			{
				"name": "sidecar",
				"image": "envoy:v1.18.0"
			}
		],
		"initContainers": [
			{
				"image": "busybox:1.35"
			}
		]
	}`

	// Parse JSON to AST Value
	var jsonData interface{}
	err := json.Unmarshal([]byte(jsonStr), &jsonData)
	assert.NoError(t, err)

	value, err := ast.InterfaceToValue(jsonData)
	assert.NoError(t, err)

	// Extract all images
	images := ExtractImages(value)

	// Should find 4 images total
	assert.Len(t, images, 4, "Should extract all 4 image references")

	// Verify each image reference can navigate back to the correct AST location
	for _, imgRef := range images {
		t.Run("VerifyAST_"+imgRef.ASTPath, func(t *testing.T) {
			// Navigate using the AST path
			foundValue := navigateASTPath(value, imgRef.ASTPath)
			assert.NotNil(t, foundValue, "Path %s should navigate to a valid value", imgRef.ASTPath)

			// Verify it's a string value matching the expected image
			if stringValue, ok := foundValue.(ast.String); ok {
				assert.Equal(t, imgRef.Image, string(stringValue),
					"Path %s should navigate to image %s", imgRef.ASTPath, imgRef.Image)
			} else {
				t.Errorf("Path %s should navigate to ast.String, got %T", imgRef.ASTPath, foundValue)
			}
		})
	}

	// Verify specific expected paths and images
	expectedImages := map[string]string{
		"proxy.image":             "mcr.microsoft.com/oss/v2/istio/proxyv2-fips:v1.26.4-1",
		"containers.[0].image":    "nginx:1.21",
		"containers.[1].image":    "envoy:v1.18.0",
		"initContainers.[0].image": "busybox:1.35",
	}

	for expectedPath, expectedImage := range expectedImages {
		found := false
		for _, imgRef := range images {
			if imgRef.ASTPath == expectedPath && imgRef.Image == expectedImage {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find image %s at path %s", expectedImage, expectedPath)
	}
}

// navigateASTPath navigates through an AST using a dot-separated path
// This helper function verifies that the extracted AST path correctly points
// to the original image location in the AST structure
func navigateASTPath(value ast.Value, path string) ast.Value {
	if path == "" {
		return value
	}

	// Parse path: "proxy.image" -> ["proxy", "image"]
	// Also handle array notation: "containers[0].image" -> ["containers", "0", "image"]
	pathParts := parseTestPath(path)

	current := value
	for _, part := range pathParts {
		switch v := current.(type) {
		case ast.Object:
			// Look for the key in the object
			var found ast.Value
			v.Foreach(func(keyTerm, valueTerm *ast.Term) {
				if keyStr, ok := keyTerm.Value.(ast.String); ok && string(keyStr) == part {
					found = valueTerm.Value
				}
			})
			if found == nil {
				return nil // Path not found
			}
			current = found

		case *ast.Array:
			// Parse array index
			index, err := strconv.Atoi(part)
			if err != nil || index >= v.Len() || index < 0 {
				return nil // Invalid array index
			}
			elem := v.Elem(index)
			if elem == nil {
				return nil
			}
			current = elem.Value

		default:
			// Cannot navigate further
			return nil
		}
	}

	return current
}

// parseTestPath converts "proxy.image" or "containers[0].image" to path parts
func parseTestPath(path string) []string {
	// Handle array notation: containers[0].image -> containers.0.image
	path = strings.ReplaceAll(path, "[", ".")
	path = strings.ReplaceAll(path, "]", "")

	parts := strings.Split(path, ".")
	var result []string
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}