package workflows

import (
	"path/filepath"

	"go.temporal.io/sdk/workflow"

	"github.com/ajanardhanan/workflow-builder/engine"
	"github.com/ajanardhanan/workflow-builder/models"
)

// TicketLifecycleDeclarativeWorkflow is a YAML-driven version of the ticket workflow
// The DAG structure is defined in templates/ticket-lifecycle.yaml
// This function just loads the YAML and executes it
func TicketLifecycleDeclarativeWorkflow(ctx workflow.Context, input models.TicketInput) (models.TicketWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Declarative Ticket Lifecycle Workflow", "title", input.Title)

	// Load workflow definition from YAML
	// Use absolute path from project root
	yamlPath := filepath.Join("..", "templates", "ticket-lifecycle.yaml")
	definition, err := engine.LoadWorkflowFromYAML(yamlPath)
	if err != nil {
		logger.Error("Failed to load workflow definition", "error", err)
		return models.TicketWorkflowResult{}, err
	}

	logger.Info("Loaded workflow definition", "name", definition.Name)

	// Convert input struct to map for generic engine
	inputMap := map[string]interface{}{
		"title":       input.Title,
		"description": input.Description,
		"priority":    input.Priority,
		"agentId":     input.AgentID,
		"rating":      input.Rating,
	}

	// Execute the workflow using the generic engine
	resultContext, err := engine.GenericWorkflow(ctx, *definition, inputMap)
	if err != nil {
		logger.Error("Workflow execution failed", "error", err)
		return models.TicketWorkflowResult{}, err
	}

	// Extract results from context
	result := models.TicketWorkflowResult{}

	// Try to extract ticket information
	if ticketData, ok := resultContext["ticket"]; ok {
		if ticket, ok := ticketData.(models.TicketResponse); ok {
			result.TicketID = ticket.ID
			result.Status = ticket.Status
			result.AgentID = ticket.AssignedAgentID
		}
	}

	// Try to extract rating information
	if ratingData, ok := resultContext["rating"]; ok {
		if rating, ok := ratingData.(models.RatingResponse); ok {
			result.Rating = rating.Score
			result.RatingID = rating.ID
		}
	}

	result.WorkflowStep = "RATED"

	logger.Info("Declarative Workflow completed successfully",
		"ticketID", result.TicketID,
		"status", result.Status,
		"rating", result.Rating)

	return result, nil
}
