package gitops

import (
	"errors"
	"strings"

	"gopkg.in/yaml.v2"
)

type DeploymentYaml struct {
	Image string `yaml:"image,omitempty"`
}

// Change image in a deployment.yaml file
func ChangeImageVersion(fileContent []byte, newImageTag string) (string, error) {
	var deploymentYaml DeploymentYaml
	err := yaml.Unmarshal(fileContent, &deploymentYaml)
	if err != nil {
		return string(fileContent), err // Return origin file content with error
	}
	if deploymentYaml.Image == "" {
		err = errors.New("gitops: deployment image not found")
		return string(fileContent), err // Return origin file content with error
	}

	parsedContent := strings.ReplaceAll(string(fileContent), deploymentYaml.Image, newImageTag)
	return parsedContent, nil
}
