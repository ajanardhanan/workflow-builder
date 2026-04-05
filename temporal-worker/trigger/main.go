package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"go.temporal.io/sdk/client"

	"github.com/ajanardhanan/workflow-builder/models"
	"github.com/ajanardhanan/workflow-builder/workflows"
)

const TaskQueueName = "ticket-workflow-queue"

func main() {
	// Command line flags
	title := flag.String("title", "Sample Ticket", "Ticket title")
	description := flag.String("description", "This is a sample ticket created via workflow", "Ticket description")
	priority := flag.String("priority", "MEDIUM", "Ticket priority (LOW, MEDIUM, HIGH, URGENT)")
	agentID := flag.String("agent", "", "Agent ID to assign ticket to (required)")
	rating := flag.Int("rating", 5, "Rating score (1-5)")
	declarative := flag.Bool("declarative", true, "Use declarative YAML-driven workflow (default: true)")
	flag.Parse()

	if *agentID == "" {
		log.Fatal("Error: --agent flag is required")
	}

	// Create Temporal client
	c, err := client.Dial(client.Options{
		HostPort: "localhost:7233",
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal client:", err)
	}
	defer c.Close()

	// Prepare workflow input
	input := models.TicketInput{
		Title:       *title,
		Description: *description,
		Priority:    *priority,
		AgentID:     *agentID,
		Rating:      *rating,
	}

	// Generate unique workflow ID
	workflowID := fmt.Sprintf("ticket-lifecycle-%d", time.Now().Unix())

	workflowType := "Hardcoded"
	if *declarative {
		workflowType = "Declarative (YAML-driven)"
	}

	log.Println("🚀 Starting Ticket Lifecycle Workflow")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("Workflow Type: %s\n", workflowType)
	log.Printf("Workflow ID: %s\n", workflowID)
	log.Printf("Input:\n")
	inputJSON, _ := json.MarshalIndent(input, "  ", "  ")
	fmt.Printf("  %s\n", string(inputJSON))
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// Start workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: TaskQueueName,
	}

	var we client.WorkflowRun
	if *declarative {
		we, err = c.ExecuteWorkflow(context.Background(), workflowOptions, workflows.TicketLifecycleDeclarativeWorkflow, input)
	} else {
		we, err = c.ExecuteWorkflow(context.Background(), workflowOptions, workflows.TicketLifecycleWorkflow, input)
	}
	if err != nil {
		log.Fatalln("Unable to execute workflow:", err)
	}

	log.Printf("✅ Workflow started successfully!\n")
	log.Printf("   Workflow ID: %s\n", we.GetID())
	log.Printf("   Run ID: %s\n", we.GetRunID())
	log.Printf("\n📊 View workflow in Temporal UI: http://localhost:8080/namespaces/default/workflows/%s\n", workflowID)
	log.Println("\nWaiting for workflow to complete...\n")

	// Wait for workflow completion
	var result models.TicketWorkflowResult
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Workflow execution failed:", err)
	}

	// Display result
	log.Println("\n🎉 Workflow completed successfully!")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("Result:\n")
	resultJSON, _ := json.MarshalIndent(result, "  ", "  ")
	fmt.Printf("  %s\n", string(resultJSON))
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("\n✨ Ticket ID: %s\n", result.TicketID)
	log.Printf("✨ Final Status: %s\n", result.Status)
	log.Printf("✨ Rating: %d/5\n", result.Rating)
}
