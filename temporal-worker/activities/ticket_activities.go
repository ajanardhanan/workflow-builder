package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ajanardhanan/workflow-builder/models"
)

const API_BASE_URL = "https://ticket-api-server-421581751516.us-central1.run.app/api"

type TicketActivities struct {
	httpClient *http.Client
}

func NewTicketActivities() *TicketActivities {
	return &TicketActivities{
		httpClient: &http.Client{},
	}
}

// CreateTicket creates a new ticket via API
func (a *TicketActivities) CreateTicket(ctx context.Context, input models.TicketInput) (models.TicketResponse, error) {
	reqBody := models.CreateTicketRequest{
		Title:       input.Title,
		Description: input.Description,
		Priority:    input.Priority,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return models.TicketResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", API_BASE_URL+"/tickets", bytes.NewBuffer(jsonData))
	if err != nil {
		return models.TicketResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return models.TicketResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return models.TicketResponse{}, fmt.Errorf("failed to create ticket: %s", string(body))
	}

	var ticket models.TicketResponse
	if err := json.NewDecoder(resp.Body).Decode(&ticket); err != nil {
		return models.TicketResponse{}, err
	}

	return ticket, nil
}

// AssignTicket assigns a ticket to an agent
func (a *TicketActivities) AssignTicket(ctx context.Context, ticketID string, agentID string) (models.TicketResponse, error) {
	reqBody := models.AssignTicketRequest{
		AgentID: agentID,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return models.TicketResponse{}, err
	}

	url := fmt.Sprintf("%s/tickets/%s/assign", API_BASE_URL, ticketID)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return models.TicketResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return models.TicketResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return models.TicketResponse{}, fmt.Errorf("failed to assign ticket: %s", string(body))
	}

	var ticket models.TicketResponse
	if err := json.NewDecoder(resp.Body).Decode(&ticket); err != nil {
		return models.TicketResponse{}, err
	}

	return ticket, nil
}

// UpdateTicket updates a ticket status
func (a *TicketActivities) UpdateTicket(ctx context.Context, ticketID string, status string) (models.TicketResponse, error) {
	reqBody := models.UpdateTicketRequest{
		Status: status,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return models.TicketResponse{}, err
	}

	url := fmt.Sprintf("%s/tickets/%s", API_BASE_URL, ticketID)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return models.TicketResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return models.TicketResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return models.TicketResponse{}, fmt.Errorf("failed to update ticket: %s", string(body))
	}

	var ticket models.TicketResponse
	if err := json.NewDecoder(resp.Body).Decode(&ticket); err != nil {
		return models.TicketResponse{}, err
	}

	return ticket, nil
}

// CloseTicket closes a ticket
func (a *TicketActivities) CloseTicket(ctx context.Context, ticketID string) (models.TicketResponse, error) {
	url := fmt.Sprintf("%s/tickets/%s/close", API_BASE_URL, ticketID)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, nil)
	if err != nil {
		return models.TicketResponse{}, err
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return models.TicketResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return models.TicketResponse{}, fmt.Errorf("failed to close ticket: %s", string(body))
	}

	var ticket models.TicketResponse
	if err := json.NewDecoder(resp.Body).Decode(&ticket); err != nil {
		return models.TicketResponse{}, err
	}

	return ticket, nil
}

// RateTicket rates a closed ticket
func (a *TicketActivities) RateTicket(ctx context.Context, ticketID string, score int) (models.RatingResponse, error) {
	reqBody := models.RateTicketRequest{
		Score:    score,
		Feedback: fmt.Sprintf("Automated rating: %d/5", score),
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return models.RatingResponse{}, err
	}

	url := fmt.Sprintf("%s/tickets/%s/rate", API_BASE_URL, ticketID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return models.RatingResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return models.RatingResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return models.RatingResponse{}, fmt.Errorf("failed to rate ticket: %s", string(body))
	}

	var rating models.RatingResponse
	if err := json.NewDecoder(resp.Body).Decode(&rating); err != nil {
		return models.RatingResponse{}, err
	}

	return rating, nil
}
