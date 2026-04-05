package models

// TicketInput is the input to start the ticket lifecycle workflow
type TicketInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"` // LOW, MEDIUM, HIGH, URGENT
	AgentID     string `json:"agentId"`
	Rating      int    `json:"rating"` // 1-5
}

// TicketWorkflowResult contains the final state of the ticket workflow
type TicketWorkflowResult struct {
	TicketID     string `json:"ticketId"`
	Status       string `json:"status"`
	AgentID      string `json:"agentId"`
	Rating       int    `json:"rating"`
	RatingID     string `json:"ratingId"`
	WorkflowStep string `json:"workflowStep"`
}

// CreateTicketRequest matches the API schema
type CreateTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
}

// AssignTicketRequest matches the API schema
type AssignTicketRequest struct {
	AgentID string `json:"agentId"`
}

// UpdateTicketRequest matches the API schema
type UpdateTicketRequest struct {
	Status string `json:"status"`
}

// RateTicketRequest matches the API schema
type RateTicketRequest struct {
	Score    int    `json:"score"`
	Feedback string `json:"feedback"`
}

// TicketResponse from API
type TicketResponse struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	Status          string `json:"status"`
	Priority        string `json:"priority"`
	AssignedAgentID string `json:"assignedAgentId"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
	ClosedAt        string `json:"closedAt"`
}

// RatingResponse from API
type RatingResponse struct {
	ID        string `json:"id"`
	TicketID  string `json:"ticketId"`
	AgentID   string `json:"agentId"`
	Score     int    `json:"score"`
	Feedback  string `json:"feedback"`
	CreatedAt string `json:"createdAt"`
}
