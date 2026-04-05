package engine

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// YAMLWorkflow represents the YAML file structure
type YAMLWorkflow struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Version     string                 `yaml:"version"`
	Input       map[string]InputField  `yaml:"input"`
	Steps       []YAMLStep             `yaml:"steps"`
	Output      map[string]string      `yaml:"output"`
	RetryPolicy YAMLRetryPolicy        `yaml:"retry_policy"`
	Timeout     YAMLTimeout            `yaml:"timeout"`
}

type InputField struct {
	Type        string   `yaml:"type"`
	Required    bool     `yaml:"required"`
	Description string   `yaml:"description"`
	Values      []string `yaml:"values"`
	Default     interface{} `yaml:"default"`
	Min         int      `yaml:"min"`
	Max         int      `yaml:"max"`
}

type YAMLStep struct {
	Name      string                 `yaml:"name"`
	Activity  string                 `yaml:"activity"`
	Input     map[string]interface{} `yaml:"input"`
	Output    string                 `yaml:"output"`
	OnSuccess []YAMLStep             `yaml:"on_success"`
	Condition string                 `yaml:"condition"`
	Parallel  bool                   `yaml:"parallel"`
}

type YAMLRetryPolicy struct {
	InitialInterval    string  `yaml:"initial_interval"`
	BackoffCoefficient float64 `yaml:"backoff_coefficient"`
	MaximumInterval    string  `yaml:"maximum_interval"`
	MaximumAttempts    int32   `yaml:"maximum_attempts"`
}

type YAMLTimeout struct {
	Activity string `yaml:"activity"`
	Workflow string `yaml:"workflow"`
}

// LoadWorkflowFromYAML parses a YAML file and returns WorkflowDefinition
func LoadWorkflowFromYAML(filePath string) (*WorkflowDefinition, error) {
	// Read YAML file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	// Parse YAML
	var yamlWorkflow YAMLWorkflow
	err = yaml.Unmarshal(data, &yamlWorkflow)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Convert to WorkflowDefinition
	definition := &WorkflowDefinition{
		Name:        yamlWorkflow.Name,
		Description: yamlWorkflow.Description,
		Steps:       convertSteps(yamlWorkflow.Steps),
		RetryPolicy: RetryPolicyConfig{
			InitialInterval:    parseDuration(yamlWorkflow.RetryPolicy.InitialInterval),
			BackoffCoefficient: yamlWorkflow.RetryPolicy.BackoffCoefficient,
			MaximumInterval:    parseDuration(yamlWorkflow.RetryPolicy.MaximumInterval),
			MaximumAttempts:    yamlWorkflow.RetryPolicy.MaximumAttempts,
		},
		Timeout: TimeoutConfig{
			Activity: parseDuration(yamlWorkflow.Timeout.Activity),
			Workflow: parseDuration(yamlWorkflow.Timeout.Workflow),
		},
	}

	return definition, nil
}

// convertSteps converts YAML steps to StepDefinition
func convertSteps(yamlSteps []YAMLStep) []StepDefinition {
	steps := make([]StepDefinition, len(yamlSteps))
	for i, yamlStep := range yamlSteps {
		steps[i] = StepDefinition{
			Name:      yamlStep.Name,
			Activity:  yamlStep.Activity,
			Input:     yamlStep.Input,
			Output:    yamlStep.Output,
			OnSuccess: convertSteps(yamlStep.OnSuccess),
			Parallel:  yamlStep.Parallel,
			Condition: yamlStep.Condition,
		}
	}
	return steps
}

// parseDuration converts string like "1s", "30s", "5m" to time.Duration
func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 30 * time.Second // default fallback
	}
	return d
}
