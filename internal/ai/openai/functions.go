package openai

// FunctionRegistry manages function definitions for OpenAI function calling
type FunctionRegistry struct {
	functions map[string]FunctionDefinition
}

// NewFunctionRegistry creates a new function registry
func NewFunctionRegistry() *FunctionRegistry {
	return &FunctionRegistry{
		functions: make(map[string]FunctionDefinition),
	}
}

// Register registers a function definition
func (r *FunctionRegistry) Register(name string, def FunctionDefinition) {
	r.functions[name] = def
}

// Get retrieves a function definition by name
func (r *FunctionRegistry) Get(name string) (FunctionDefinition, bool) {
	def, ok := r.functions[name]
	return def, ok
}

// GetAll returns all registered function definitions
func (r *FunctionRegistry) GetAll() []FunctionDefinition {
	defs := make([]FunctionDefinition, 0, len(r.functions))
	for _, def := range r.functions {
		defs = append(defs, def)
	}
	return defs
}

// GetTools returns all functions as Tool objects
func (r *FunctionRegistry) GetTools() []Tool {
	tools := make([]Tool, 0, len(r.functions))
	for _, def := range r.functions {
		tools = append(tools, Tool{
			Type:     "function",
			Function: def,
		})
	}
	return tools
}

// HelmAnalysisFunction returns a function definition for analyzing Helm charts
func HelmAnalysisFunction() FunctionDefinition {
	return FunctionDefinition{
		Name:        "analyze_helm_chart",
		Description: "Analyze a Helm chart for issues, compatibility, and best practices",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"chart_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the Helm chart to analyze",
				},
				"chart_version": map[string]interface{}{
					"type":        "string",
					"description": "Version of the Helm chart",
				},
				"values": map[string]interface{}{
					"type":        "object",
					"description": "Helm values to analyze",
				},
				"check_types": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
						"enum": []string{
							"security",
							"compatibility",
							"best-practices",
							"performance",
							"resources",
						},
					},
					"description": "Types of checks to perform",
				},
			},
			"required": []string{"chart_name", "chart_version"},
		},
	}
}

// CompatibilityCheckFunction returns a function definition for checking version compatibility
func CompatibilityCheckFunction() FunctionDefinition {
	return FunctionDefinition{
		Name:        "check_compatibility",
		Description: "Check compatibility between Helm chart versions, Kubernetes versions, and dependencies",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"chart_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the Helm chart",
				},
				"current_version": map[string]interface{}{
					"type":        "string",
					"description": "Current chart version",
				},
				"target_version": map[string]interface{}{
					"type":        "string",
					"description": "Target chart version for upgrade",
				},
				"kubernetes_version": map[string]interface{}{
					"type":        "string",
					"description": "Target Kubernetes version",
				},
				"dependencies": map[string]interface{}{
					"type":        "array",
					"description": "List of chart dependencies",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name": map[string]interface{}{
								"type": "string",
							},
							"version": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
			},
			"required": []string{"chart_name", "current_version", "target_version"},
		},
	}
}

// UpgradeStrategyFunction returns a function definition for generating upgrade strategies
func UpgradeStrategyFunction() FunctionDefinition {
	return FunctionDefinition{
		Name:        "generate_upgrade_strategy",
		Description: "Generate a detailed upgrade strategy for a Helm chart",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"chart_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the Helm chart",
				},
				"current_version": map[string]interface{}{
					"type":        "string",
					"description": "Current chart version",
				},
				"target_version": map[string]interface{}{
					"type":        "string",
					"description": "Target chart version",
				},
				"environment": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"dev", "staging", "production"},
					"description": "Target environment",
				},
				"constraints": map[string]interface{}{
					"type":        "array",
					"description": "Upgrade constraints",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"rollback_enabled": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether rollback capability is required",
					"default":     true,
				},
			},
			"required": []string{"chart_name", "current_version", "target_version", "environment"},
		},
	}
}

// RiskAssessmentFunction returns a function definition for assessing upgrade risks
func RiskAssessmentFunction() FunctionDefinition {
	return FunctionDefinition{
		Name:        "assess_upgrade_risk",
		Description: "Assess the risk level of a Helm chart upgrade",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"chart_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the Helm chart",
				},
				"current_version": map[string]interface{}{
					"type":        "string",
					"description": "Current chart version",
				},
				"target_version": map[string]interface{}{
					"type":        "string",
					"description": "Target chart version",
				},
				"breaking_changes": map[string]interface{}{
					"type":        "array",
					"description": "Known breaking changes",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"deprecations": map[string]interface{}{
					"type":        "array",
					"description": "Deprecated features or APIs",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"environment": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"dev", "staging", "production"},
					"description": "Target environment",
				},
			},
			"required": []string{"chart_name", "current_version", "target_version"},
		},
	}
}

// DefaultFunctionRegistry returns a registry with default Helm-related functions
func DefaultFunctionRegistry() *FunctionRegistry {
	registry := NewFunctionRegistry()

	registry.Register("analyze_helm_chart", HelmAnalysisFunction())
	registry.Register("check_compatibility", CompatibilityCheckFunction())
	registry.Register("generate_upgrade_strategy", UpgradeStrategyFunction())
	registry.Register("assess_upgrade_risk", RiskAssessmentFunction())

	return registry
}
