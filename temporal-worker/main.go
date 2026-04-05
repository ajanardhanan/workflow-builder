package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/ajanardhanan/workflow-builder/activities"
	"github.com/ajanardhanan/workflow-builder/workflows"
)

const (
	TaskQueueName = "ticket-workflow-queue"
)

func main() {
	// Create Temporal client
	c, err := client.Dial(client.Options{
		HostPort: "localhost:7233",
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()

	// Create worker
	w := worker.New(c, TaskQueueName, worker.Options{})

	// Register workflows
	w.RegisterWorkflow(workflows.TicketLifecycleWorkflow)
	w.RegisterWorkflow(workflows.TicketLifecycleDeclarativeWorkflow)

	// Register activities
	ticketActivities := activities.NewTicketActivities()
	w.RegisterActivity(ticketActivities.CreateTicket)
	w.RegisterActivity(ticketActivities.AssignTicket)
	w.RegisterActivity(ticketActivities.UpdateTicket)
	w.RegisterActivity(ticketActivities.CloseTicket)
	w.RegisterActivity(ticketActivities.RateTicket)

	notificationActivities := activities.NewNotificationActivities()
	w.RegisterActivity(notificationActivities.SendEmail)
	w.RegisterActivity(notificationActivities.SendTicketCreatedEmail)
	w.RegisterActivity(notificationActivities.SendTicketAssignedEmail)
	w.RegisterActivity(notificationActivities.SendTicketUpdatedEmail)
	w.RegisterActivity(notificationActivities.SendTicketClosedEmail)
	w.RegisterActivity(notificationActivities.SendTicketRatedEmail)

	log.Println("🚀 Temporal Worker started")
	log.Printf("   Task Queue: %s\n", TaskQueueName)
	log.Println("   Registered Workflows:")
	log.Println("     - TicketLifecycleWorkflow (hardcoded)")
	log.Println("     - TicketLifecycleDeclarativeWorkflow (YAML-driven)")
	log.Println("   Registered Activities:")
	log.Println("     - CreateTicket, AssignTicket, UpdateTicket, CloseTicket, RateTicket")
	log.Println("     - SendEmail notifications")
	log.Println("\n   Waiting for workflows...\n")

	// Start worker
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
