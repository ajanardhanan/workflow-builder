package engine

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/ajanardhanan/workflow-builder/activities"
	"github.com/ajanardhanan/workflow-builder/models"
)

// WorkflowDefinition represents the parsed YAML structure
type WorkflowDefinition struct {
	Name        string
	Description string
	Steps       []StepDefinition
	RetryPolicy RetryPolicyConfig
	Timeout     TimeoutConfig
}

type StepDefinition struct {
	Name       string
	Activity   string
	Input      map[string]interface{} // Input mapping from YAML
	Output     string                 // Variable name to store result
	OnSuccess  []StepDefinition       // Nested steps to run on success
	Parallel   bool                   // Execute in parallel with next step?
	Condition  string                 // Optional condition (e.g., "$ticket.priority == 'HIGH'")
}

type RetryPolicyConfig struct {
	InitialInterval    time.Duration
	BackoffCoefficient float64
	MaximumInterval    time.Duration
	MaximumAttempts    int32
}

type TimeoutConfig struct {
	Activity time.Duration
	Workflow time.Duration
}

// GenericWorkflow executes any workflow defined by WorkflowDefinition
func GenericWorkflow(ctx workflow.Context, definition WorkflowDefinition, input map[string]interface{}) (map[string]interface{}, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Generic Workflow", "name", definition.Name)

	// Configure activity options from definition
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: definition.Timeout.Activity,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    definition.RetryPolicy.InitialInterval,
			BackoffCoefficient: definition.RetryPolicy.BackoffCoefficient,
			MaximumInterval:    definition.RetryPolicy.MaximumInterval,
			MaximumAttempts:    definition.RetryPolicy.MaximumAttempts,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Activity registry
	ticketActivities := &activities.TicketActivities{}
	notificationActivities := &activities.NotificationActivities{}

	// Workflow context - stores variables like $ticket, $rating
	workflowContext := make(map[string]interface{})
	workflowContext["input"] = input

	// Execute steps sequentially or in parallel based on definition
	for _, step := range definition.Steps {
		err := executeStep(ctx, step, workflowContext, ticketActivities, notificationActivities)
		if err != nil {
			return nil, fmt.Errorf("step %s failed: %w", step.Name, err)
		}
	}

	return workflowContext, nil
}

// executeStep runs a single step and handles parallel/sequential execution
func executeStep(
	ctx workflow.Context,
	step StepDefinition,
	workflowContext map[string]interface{},
	ticketActivities *activities.TicketActivities,
	notificationActivities *activities.NotificationActivities,
) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Executing step", "name", step.Name, "activity", step.Activity)

	// Evaluate condition if present
	if step.Condition != "" {
		shouldExecute := evaluateCondition(step.Condition, workflowContext)
		if !shouldExecute {
			logger.Info("Skipping step due to condition", "condition", step.Condition)
			return nil
		}
	}

	// Resolve input parameters (e.g., $input.title, $ticket.id)
	resolvedInput := resolveInputs(step.Input, workflowContext)

	// Execute the activity based on activity name
	var result interface{}
	var err error

	switch step.Activity {
	case "CreateTicket":
		// Build TicketInput from resolved inputs
		ticketInput := models.TicketInput{
			Title:       getString(resolvedInput, "title"),
			Description: getString(resolvedInput, "description"),
			Priority:    getString(resolvedInput, "priority"),
		}
		var ticketResponse models.TicketResponse
		err = workflow.ExecuteActivity(ctx, ticketActivities.CreateTicket, ticketInput).Get(ctx, &ticketResponse)
		result = ticketResponse

	case "AssignTicket":
		ticketID := getString(resolvedInput, "ticketId")
		agentID := getString(resolvedInput, "agentId")
		var ticketResponse models.TicketResponse
		err = workflow.ExecuteActivity(ctx, ticketActivities.AssignTicket, ticketID, agentID).Get(ctx, &ticketResponse)
		result = ticketResponse

	case "UpdateTicket":
		ticketID := getString(resolvedInput, "ticketId")
		status := getString(resolvedInput, "status")
		var ticketResponse models.TicketResponse
		err = workflow.ExecuteActivity(ctx, ticketActivities.UpdateTicket, ticketID, status).Get(ctx, &ticketResponse)
		result = ticketResponse

	case "CloseTicket":
		ticketID := getString(resolvedInput, "ticketId")
		var ticketResponse models.TicketResponse
		err = workflow.ExecuteActivity(ctx, ticketActivities.CloseTicket, ticketID).Get(ctx, &ticketResponse)
		result = ticketResponse

	case "RateTicket":
		ticketID := getString(resolvedInput, "ticketId")
		score := getInt(resolvedInput, "score")
		var ratingResponse models.RatingResponse
		err = workflow.ExecuteActivity(ctx, ticketActivities.RateTicket, ticketID, score).Get(ctx, &ratingResponse)
		result = ratingResponse

	case "SendTicketCreatedEmail":
		ticketID := getString(resolvedInput, "ticketId")
		title := getString(resolvedInput, "title")
		err = workflow.ExecuteActivity(ctx, notificationActivities.SendTicketCreatedEmail, ticketID, title).Get(ctx, nil)

	case "SendTicketAssignedEmail":
		ticketID := getString(resolvedInput, "ticketId")
		title := getString(resolvedInput, "title")
		agentID := getString(resolvedInput, "agentId")
		err = workflow.ExecuteActivity(ctx, notificationActivities.SendTicketAssignedEmail, ticketID, title, agentID).Get(ctx, nil)

	case "SendTicketUpdatedEmail":
		ticketID := getString(resolvedInput, "ticketId")
		title := getString(resolvedInput, "title")
		status := getString(resolvedInput, "status")
		err = workflow.ExecuteActivity(ctx, notificationActivities.SendTicketUpdatedEmail, ticketID, title, status).Get(ctx, nil)

	case "SendTicketClosedEmail":
		ticketID := getString(resolvedInput, "ticketId")
		title := getString(resolvedInput, "title")
		err = workflow.ExecuteActivity(ctx, notificationActivities.SendTicketClosedEmail, ticketID, title).Get(ctx, nil)

	case "SendTicketRatedEmail":
		ticketID := getString(resolvedInput, "ticketId")
		title := getString(resolvedInput, "title")
		score := getInt(resolvedInput, "score")
		err = workflow.ExecuteActivity(ctx, notificationActivities.SendTicketRatedEmail, ticketID, title, score).Get(ctx, nil)

	default:
		return fmt.Errorf("unknown activity: %s", step.Activity)
	}

	if err != nil {
		return err
	}

	// Store result in workflow context (e.g., $ticket, $rating)
	if step.Output != "" {
		workflowContext[step.Output] = result
		logger.Info("Stored result", "variable", step.Output)
	}

	// Execute on_success hooks (e.g., send email notifications)
	for _, successStep := range step.OnSuccess {
		err := executeStep(ctx, successStep, workflowContext, ticketActivities, notificationActivities)
		if err != nil {
			logger.Warn("On-success step failed", "step", successStep.Name, "error", err)
			// Don't fail main workflow for notification failures
		}
	}

	return nil
}

// evaluateCondition checks if a condition string is true
// Example: "$ticket.priority == 'HIGH'" or "$input.rating > 3"
func evaluateCondition(condition string, context map[string]interface{}) bool {
	// Simple implementation - production would use a proper expression parser
	// For now, just return true (execute all steps)
	return true
}

// resolveInputs replaces variables like $input.title with actual values
// Input: {"title": "$input.title", "priority": "$input.priority"}
// Context: {"input": {"title": "Bug fix", "priority": "HIGH"}}
// Output: {"title": "Bug fix", "priority": "HIGH"}
func resolveInputs(inputs map[string]interface{}, context map[string]interface{}) map[string]interface{} {
	resolved := make(map[string]interface{})

	for key, value := range inputs {
		// Check if value is a string starting with $
		if strValue, ok := value.(string); ok && len(strValue) > 0 && strValue[0] == '$' {
			// Extract the variable path (e.g., "$input.title" -> ["input", "title"])
			resolvedValue := resolveVariablePath(strValue[1:], context)
			resolved[key] = resolvedValue
		} else {
			resolved[key] = value
		}
	}

	return resolved
}

// resolveVariablePath navigates nested maps to get value
// Path: "input.title", Context: {"input": {"title": "Bug"}}
// Returns: "Bug"
func resolveVariablePath(path string, context map[string]interface{}) interface{} {
	// Handle simple paths like "input.title"
	// Split by "." and navigate
	parts := []rune(path)
	var current interface{} = context

	startIdx := 0
	for i := 0; i <= len(parts); i++ {
		if i == len(parts) || parts[i] == '.' {
			if i > startIdx {
				key := string(parts[startIdx:i])
				if m, ok := current.(map[string]interface{}); ok {
					current = m[key]
				} else if ticket, ok := current.(models.TicketResponse); ok {
					// Handle TicketResponse fields
					switch key {
					case "id":
						return ticket.ID
					case "title":
						return ticket.Title
					case "status":
						return ticket.Status
					case "assignedAgentId":
						return ticket.AssignedAgentID
					case "priority":
						return ticket.Priority
					}
				} else if rating, ok := current.(models.RatingResponse); ok {
					// Handle RatingResponse fields
					switch key {
					case "id":
						return rating.ID
					case "score":
						return rating.Score
					}
				} else {
					return nil
				}
			}
			startIdx = i + 1
		}
	}

	return current
}

// getString extracts a string value from the input map
func getString(input map[string]interface{}, key string) string {
	if val, ok := input[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getInt extracts an int value from the input map
func getInt(input map[string]interface{}, key string) int {
	if val, ok := input[key]; ok {
		if i, ok := val.(int); ok {
			return i
		}
		// Handle float64 (from JSON unmarshaling)
		if f, ok := val.(float64); ok {
			return int(f)
		}
	}
	return 0
}
