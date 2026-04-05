# Workflow Builder with Temporal

A workflow orchestration system using Temporal to automate ticket management workflows.

## Architecture

```
workflow-builder/
├── temporal-worker/           # Go Temporal worker
│   ├── workflows/            # Workflow definitions
│   ├── activities/           # Activities (API calls, emails)
│   ├── models/              # Data models
│   ├── trigger/             # CLI to trigger workflows
│   └── main.go              # Worker entry point
├── docker-compose.yml        # Temporal server + PostgreSQL
└── templates/               # YAML workflow templates
```

## Features

### Ticket Lifecycle Workflow

Automates the complete JIRA-like ticket lifecycle:

1. **Create Ticket** → Email notification sent
2. **Assign to Agent** → Email notification sent
3. **Update Status (RESOLVED)** → Email notification sent
4. **Close Ticket** → Email notification sent
5. **Rate Ticket** → Email notification sent

All email notifications are sent to: `akjanardhan@gmail.com` (console logging for now)

## Prerequisites

- **Go 1.21+**: `brew install go`
- **Docker & Docker Compose**: For Temporal server
- **Running ticket-api-server**: Backend API must be accessible

## Setup

### 1. Install Go

```bash
brew install go
```

### 2. Start Temporal Server

```bash
# Start Temporal server, PostgreSQL, and Temporal UI
docker-compose up -d

# Verify services are running
docker-compose ps
```

Services:
- **Temporal Server**: localhost:7233
- **Temporal UI**: http://localhost:8080
- **PostgreSQL**: localhost:5432

### 3. Install Go Dependencies

```bash
cd temporal-worker
go mod tidy
```

### 4. Get Agent ID

You need an agent ID to assign tickets. Get it from your agents:

```bash
curl -s http://localhost:3000/api/agents | jq '.[] | {id, name}'
```

Copy an agent ID (e.g., `pGFNuC5OVDzmLfp80DPD`)

## Running the Worker

### Start the Temporal Worker

In one terminal:

```bash
cd temporal-worker
go run main.go
```

You should see:
```
🚀 Temporal Worker started
   Task Queue: ticket-workflow-queue
   Registered Workflows:
     - TicketLifecycleWorkflow
   Registered Activities:
     - CreateTicket, AssignTicket, UpdateTicket, CloseTicket, RateTicket
     - SendEmail notifications

   Waiting for workflows...
```

Keep this running!

## Triggering Workflows

### Trigger Ticket Lifecycle Workflow

In another terminal:

```bash
cd temporal-worker

# Basic usage (with your agent ID)
go run trigger/main.go \
  --title "Fix login bug" \
  --description "Users cannot login with Google OAuth" \
  --priority HIGH \
  --agent pGFNuC5OVDzmLfp80DPD \
  --rating 5
```

### Options

```
--title       Ticket title (default: "Sample Ticket")
--description Ticket description
--priority    LOW | MEDIUM | HIGH | URGENT (default: MEDIUM)
--agent       Agent ID (REQUIRED)
--rating      Rating score 1-5 (default: 5)
```

### Example Workflows

**High Priority Bug:**
```bash
go run trigger/main.go \
  --title "Critical: Database connection timeout" \
  --description "Production database timing out after deployment" \
  --priority URGENT \
  --agent pGFNuC5OVDzmLfp80DPD \
  --rating 5
```

**Feature Request:**
```bash
go run trigger/main.go \
  --title "Add dark mode to UI" \
  --description "Users requesting dark theme option" \
  --priority LOW \
  --agent yOJeRw9444Bkq8Y2ppxm \
  --rating 4
```

## Monitoring Workflows

### Temporal UI

View running and completed workflows:

```
http://localhost:8080/namespaces/default/workflows
```

Features:
- Workflow execution history
- Activity status and retries
- Workflow timeline visualization
- Error details

### Console Logs

The worker logs all steps and email notifications:

```
2026-04-05 10:30:00 INFO Starting Ticket Lifecycle Workflow title=Fix login bug
2026-04-05 10:30:01 INFO Step 1: Creating ticket
2026-04-05 10:30:02 INFO Ticket created successfully ticketID=abc123
📧 EMAIL NOTIFICATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
To: akjanardhan@gmail.com
Subject: Ticket #abc123 Created
Body:
Hi,

A new ticket has been created in the system.

Ticket ID: abc123
Title: Fix login bug
Status: OPEN
...
```

## Workflow Execution Flow

```
┌─────────────────────────────────────────────────┐
│  Trigger Workflow (CLI)                         │
│  Input: title, description, priority, agent     │
└────────────────┬────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────┐
│  Step 1: Create Ticket                          │
│  API: POST /api/tickets                         │
│  Email: "Ticket #123 Created"                   │
└────────────────┬────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────┐
│  Step 2: Assign Ticket                          │
│  API: PUT /api/tickets/{id}/assign              │
│  Email: "Ticket #123 Assigned to Agent"         │
└────────────────┬────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────┐
│  Step 3: Update Ticket                          │
│  API: PUT /api/tickets/{id}                     │
│  Email: "Ticket #123 Updated: RESOLVED"         │
└────────────────┬────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────┐
│  Step 4: Close Ticket                           │
│  API: PUT /api/tickets/{id}/close               │
│  Email: "Ticket #123 Closed"                    │
└────────────────┬────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────┐
│  Step 5: Rate Ticket                            │
│  API: POST /api/tickets/{id}/rate               │
│  Email: "Ticket #123 Rated: ⭐⭐⭐⭐⭐ (5/5)"      │
└────────────────┬────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────┐
│  Workflow Complete                              │
│  Result: ticketID, status, rating               │
└─────────────────────────────────────────────────┘
```

## Error Handling & Retries

Temporal automatically handles:

- **Activity Retries**: Failed API calls retry with exponential backoff
- **Timeout Handling**: Activities timeout after 30 seconds
- **Workflow Durability**: Workflows survive crashes and restarts

Configuration (in `ticket_lifecycle_workflow.go`):
```go
RetryPolicy: &workflow.RetryPolicy{
    InitialInterval:    1 * time.Second,
    BackoffCoefficient: 2.0,
    MaximumInterval:    10 * time.Second,
    MaximumAttempts:    3,
}
```

## Testing

### Test Full Workflow

```bash
# Get an agent ID first
AGENT_ID=$(curl -s http://localhost:3000/api/agents | jq -r '.[0].id')

# Run workflow
go run trigger/main.go \
  --title "Test Workflow" \
  --description "Testing complete lifecycle" \
  --priority MEDIUM \
  --agent $AGENT_ID \
  --rating 5
```

Expected output:
- Worker logs showing each step
- 5 email notifications (console logs)
- New ticket visible in UI at http://localhost:3000
- Workflow visible in Temporal UI at http://localhost:8080

## Troubleshooting

### Worker Can't Connect to Temporal

```
Error: Unable to create Temporal client
```

**Solution**: Start Temporal server
```bash
docker-compose up -d
docker-compose ps  # Verify temporal is running
```

### API Call Failed

```
Error: failed to create ticket: 500 Internal Server Error
```

**Solution**: Verify ticket-api-server is running
```bash
curl http://localhost:3000/api/agents
```

### Agent Not Found

```
Error: failed to assign ticket: Agent not found
```

**Solution**: Use a valid agent ID
```bash
curl -s http://localhost:3000/api/agents | jq '.[] | {id, name}'
```

## Next Steps

- [ ] Add more workflows (escalation, SLA monitoring)
- [ ] Implement real SMTP email sending
- [ ] Build Next.js UI for workflow management
- [ ] Add workflow templates (YAML/JSON)
- [ ] Schedule workflows with cron
- [ ] Add webhook triggers

## Tech Stack

- **Temporal**: Workflow orchestration
- **Go**: Worker implementation
- **Docker**: Local Temporal deployment
- **ticket-api-server**: Backend API (Cloud Run)
- **PostgreSQL**: Temporal persistence
