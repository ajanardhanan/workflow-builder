package activities

import (
	"context"
	"fmt"
	"log"
)

const EMAIL_RECIPIENT = "akjanardhan@gmail.com"

type NotificationActivities struct{}

func NewNotificationActivities() *NotificationActivities {
	return &NotificationActivities{}
}

// SendEmail simulates sending an email (console logging for now)
func (a *NotificationActivities) SendEmail(ctx context.Context, subject string, body string) error {
	log.Printf("📧 EMAIL NOTIFICATION")
	log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("To: %s", EMAIL_RECIPIENT)
	log.Printf("Subject: %s", subject)
	log.Printf("Body:")
	log.Printf("%s", body)
	log.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	return nil
}

// SendTicketCreatedEmail sends notification when ticket is created
func (a *NotificationActivities) SendTicketCreatedEmail(ctx context.Context, ticketID string, title string) error {
	subject := fmt.Sprintf("Ticket #%s Created", ticketID)
	body := fmt.Sprintf(`Hi,

A new ticket has been created in the system.

Ticket ID: %s
Title: %s
Status: OPEN

You can view this ticket in the ticket management system.

Best regards,
Ticket Management System`, ticketID, title)

	return a.SendEmail(ctx, subject, body)
}

// SendTicketAssignedEmail sends notification when ticket is assigned
func (a *NotificationActivities) SendTicketAssignedEmail(ctx context.Context, ticketID string, title string, agentID string) error {
	subject := fmt.Sprintf("Ticket #%s Assigned", ticketID)
	body := fmt.Sprintf(`Hi,

Ticket #%s has been assigned to an agent.

Ticket ID: %s
Title: %s
Assigned To: Agent %s
Status: IN_PROGRESS

The agent will begin working on this ticket.

Best regards,
Ticket Management System`, ticketID, ticketID, title, agentID)

	return a.SendEmail(ctx, subject, body)
}

// SendTicketUpdatedEmail sends notification when ticket is updated
func (a *NotificationActivities) SendTicketUpdatedEmail(ctx context.Context, ticketID string, title string, status string) error {
	subject := fmt.Sprintf("Ticket #%s Updated", ticketID)
	body := fmt.Sprintf(`Hi,

Ticket #%s has been updated.

Ticket ID: %s
Title: %s
New Status: %s

Please check the ticket management system for more details.

Best regards,
Ticket Management System`, ticketID, ticketID, title, status)

	return a.SendEmail(ctx, subject, body)
}

// SendTicketClosedEmail sends notification when ticket is closed
func (a *NotificationActivities) SendTicketClosedEmail(ctx context.Context, ticketID string, title string) error {
	subject := fmt.Sprintf("Ticket #%s Closed", ticketID)
	body := fmt.Sprintf(`Hi,

Ticket #%s has been closed.

Ticket ID: %s
Title: %s
Status: CLOSED

Thank you for using our ticket management system.

Best regards,
Ticket Management System`, ticketID, ticketID, title)

	return a.SendEmail(ctx, subject, body)
}

// SendTicketRatedEmail sends notification when ticket is rated
func (a *NotificationActivities) SendTicketRatedEmail(ctx context.Context, ticketID string, title string, score int) error {
	subject := fmt.Sprintf("Ticket #%s Rated", ticketID)

	stars := ""
	for i := 0; i < score; i++ {
		stars += "⭐"
	}

	body := fmt.Sprintf(`Hi,

Ticket #%s has been rated.

Ticket ID: %s
Title: %s
Rating: %s (%d/5)

Thank you for your feedback!

Best regards,
Ticket Management System`, ticketID, ticketID, title, stars, score)

	return a.SendEmail(ctx, subject, body)
}
