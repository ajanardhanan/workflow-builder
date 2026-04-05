package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/ajanardhanan/workflow-builder/activities"
	"github.com/ajanardhanan/workflow-builder/models"
)

// TicketLifecycleWorkflow orchestrates the complete ticket lifecycle
// Steps: Create → Assign → Update → Close → Rate
func TicketLifecycleWorkflow(ctx workflow.Context, input models.TicketInput) (models.TicketWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Ticket Lifecycle Workflow", "title", input.Title)

	// Configure activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    1 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    10 * time.Second,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var ticketActivities *activities.TicketActivities
	var notificationActivities *activities.NotificationActivities

	result := models.TicketWorkflowResult{}

	// Step 1: Create Ticket
	logger.Info("Step 1: Creating ticket")
	var ticketResponse models.TicketResponse
	err := workflow.ExecuteActivity(ctx, ticketActivities.CreateTicket, input).Get(ctx, &ticketResponse)
	if err != nil {
		logger.Error("Failed to create ticket", "error", err)
		return result, fmt.Errorf("failed to create ticket: %w", err)
	}
	result.TicketID = ticketResponse.ID
	result.Status = ticketResponse.Status
	result.WorkflowStep = "CREATED"

	logger.Info("Ticket created successfully", "ticketID", ticketResponse.ID)

	// Send email notification for ticket creation
	err = workflow.ExecuteActivity(ctx, notificationActivities.SendTicketCreatedEmail,
		ticketResponse.ID, ticketResponse.Title).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to send ticket created email", "error", err)
		// Don't fail workflow on email error
	}

	// Step 2: Assign Ticket
	logger.Info("Step 2: Assigning ticket to agent", "agentID", input.AgentID)
	err = workflow.ExecuteActivity(ctx, ticketActivities.AssignTicket, ticketResponse.ID, input.AgentID).Get(ctx, &ticketResponse)
	if err != nil {
		logger.Error("Failed to assign ticket", "error", err)
		return result, fmt.Errorf("failed to assign ticket: %w", err)
	}
	result.AgentID = ticketResponse.AssignedAgentID
	result.Status = ticketResponse.Status
	result.WorkflowStep = "ASSIGNED"

	logger.Info("Ticket assigned successfully", "agentID", ticketResponse.AssignedAgentID)

	// Send email notification for ticket assignment
	err = workflow.ExecuteActivity(ctx, notificationActivities.SendTicketAssignedEmail,
		ticketResponse.ID, ticketResponse.Title, ticketResponse.AssignedAgentID).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to send ticket assigned email", "error", err)
	}

	// Step 3: Update Ticket Status to RESOLVED
	logger.Info("Step 3: Updating ticket to RESOLVED")
	err = workflow.ExecuteActivity(ctx, ticketActivities.UpdateTicket, ticketResponse.ID, "RESOLVED").Get(ctx, &ticketResponse)
	if err != nil {
		logger.Error("Failed to update ticket", "error", err)
		return result, fmt.Errorf("failed to update ticket: %w", err)
	}
	result.Status = ticketResponse.Status
	result.WorkflowStep = "UPDATED"

	logger.Info("Ticket updated successfully", "status", ticketResponse.Status)

	// Send email notification for ticket update
	err = workflow.ExecuteActivity(ctx, notificationActivities.SendTicketUpdatedEmail,
		ticketResponse.ID, ticketResponse.Title, ticketResponse.Status).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to send ticket updated email", "error", err)
	}

	// Step 4: Close Ticket
	logger.Info("Step 4: Closing ticket")
	err = workflow.ExecuteActivity(ctx, ticketActivities.CloseTicket, ticketResponse.ID).Get(ctx, &ticketResponse)
	if err != nil {
		logger.Error("Failed to close ticket", "error", err)
		return result, fmt.Errorf("failed to close ticket: %w", err)
	}
	result.Status = ticketResponse.Status
	result.WorkflowStep = "CLOSED"

	logger.Info("Ticket closed successfully")

	// Send email notification for ticket closure
	err = workflow.ExecuteActivity(ctx, notificationActivities.SendTicketClosedEmail,
		ticketResponse.ID, ticketResponse.Title).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to send ticket closed email", "error", err)
	}

	// Step 5: Rate Ticket
	logger.Info("Step 5: Rating ticket", "score", input.Rating)
	var ratingResponse models.RatingResponse
	err = workflow.ExecuteActivity(ctx, ticketActivities.RateTicket, ticketResponse.ID, input.Rating).Get(ctx, &ratingResponse)
	if err != nil {
		logger.Error("Failed to rate ticket", "error", err)
		return result, fmt.Errorf("failed to rate ticket: %w", err)
	}
	result.Rating = ratingResponse.Score
	result.RatingID = ratingResponse.ID
	result.WorkflowStep = "RATED"

	logger.Info("Ticket rated successfully", "ratingID", ratingResponse.ID, "score", ratingResponse.Score)

	// Send email notification for ticket rating
	err = workflow.ExecuteActivity(ctx, notificationActivities.SendTicketRatedEmail,
		ticketResponse.ID, ticketResponse.Title, ratingResponse.Score).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to send ticket rated email", "error", err)
	}

	logger.Info("Ticket Lifecycle Workflow completed successfully",
		"ticketID", result.TicketID,
		"status", result.Status,
		"rating", result.Rating)

	return result, nil
}
