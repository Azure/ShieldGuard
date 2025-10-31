package imageextractor

import (
	"fmt"
	"strings"

	"github.com/open-policy-agent/opa/ast"
)

// ImageReference represents an extracted image with its AST location
type ImageReference struct {
	// Image is the full image string
	Image string `json:"image"`
	// ASTPath is the path to this image in the AST
	ASTPath string `json:"ast_path"`
}

// ExtractImages traverses the AST and extracts all image field values
func ExtractImages(astValue ast.Value) []ImageReference {
	var images []ImageReference
	extractImagesRecursive(astValue, []string{}, &images)
	return images
}

// extractImagesRecursive recursively traverses AST nodes to find image fields
func extractImagesRecursive(value ast.Value, pathStack []string, images *[]ImageReference) {
	switch v := value.(type) {
	case ast.Object:
		// Traverse object fields
		v.Foreach(func(keyTerm, valueTerm *ast.Term) {
			if keyStr, ok := keyTerm.Value.(ast.String); ok {
				key := string(keyStr)
				newPath := append(pathStack, key)

				// If this is an "image" field and value is a string, record it
				// Skip FIPS compliance injected fields to avoid duplicate extraction
				if key == "image" && !containsFIPSCompliancePath(pathStack) {
					if valueStr, ok := valueTerm.Value.(ast.String); ok {
						imageRef := ImageReference{
							Image:   string(valueStr),
							ASTPath: strings.Join(newPath, "."),
						}
						*images = append(*images, imageRef)
					}
				}

				// Continue recursion for nested structures
				extractImagesRecursive(valueTerm.Value, newPath, images)
			}
		})

	case *ast.Array:
		// Traverse array elements
		for i := 0; i < v.Len(); i++ {
			item := v.Elem(i)
			if item != nil {
				indexPath := append(pathStack, fmt.Sprintf("[%d]", i))
				extractImagesRecursive(item.Value, indexPath, images)
			}
		}

	case ast.String:
		// Terminal string value - already handled in object case for "image" fields
		return

	default:
		// Other types (numbers, booleans, etc.) - no action needed
		return
	}
}

// containsFIPSCompliancePath checks if the current path contains FIPS compliance data
func containsFIPSCompliancePath(pathStack []string) bool {
	for _, pathPart := range pathStack {
		if pathPart == "fips_compliance" {
			return true
		}
	}
	return false
}

// ExtractImagesFromConfigurations extracts images from multiple AST configurations
func ExtractImagesFromConfigurations(configurations []ast.Value) []ImageReference {
	var allImages []ImageReference

	for i, config := range configurations {
		configImages := ExtractImages(config)

		// Prefix each image path with configuration index for uniqueness
		for _, img := range configImages {
			img.ASTPath = fmt.Sprintf("config[%d].%s", i, img.ASTPath)
			allImages = append(allImages, img)
		}
	}

	return allImages
}