# Declarative Workflows - Separating DAG from Implementation

## The Problem

In `ticket_lifecycle_workflow.go`, the DAG structure and activity implementations are tightly coupled:

```go
// ❌ Hardcoded: Orchestration + Implementation mixed
func TicketLifecycleWorkflow(ctx workflow.Context, input models.TicketInput) {
    // DAG structure hardcoded in Go
    workflow.ExecuteActivity(ctx, ticketActivities.CreateTicket, input)
    workflow.ExecuteActivity(ctx, ticketActivities.AssignTicket, ticketResponse.ID, input.AgentID)
    workflow.ExecuteActivity(ctx, ticketActivities.UpdateTicket, ticketResponse.ID, "RESOLVED")
    // ... every step is explicitly coded
}
```

**Problems:**
- Every new workflow = new Go file with duplicated orchestration logic
- Changing workflow structure requires code changes + recompilation + redeployment
- Business users can't modify workflows - need developers
- No visual representation of DAG - it's hidden in code

## The Solution: Declarative YAML + Generic Engine

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    YAML Template (DAG)                      │
│  templates/ticket-lifecycle.yaml                            │
│  - Defines: steps, sequence, conditions, parallel execution│
│  - No implementation details                                │
└────────────────┬────────────────────────────────────────────┘
                 │
                 │ Parsed by
                 ↓
┌─────────────────────────────────────────────────────────────┐
│              Workflow Engine (Generic)                      │
│  engine/workflow_engine.go                                  │
│  - Reads YAML definition                                    │
│  - Executes steps in order defined by YAML                  │
│  - Handles: sequencing, parallelism, conditions             │
└────────────────┬────────────────────────────────────────────┘
                 │
                 │ Calls
                 ↓
┌─────────────────────────────────────────────────────────────┐
│            Activity Implementations                         │
│  activities/ticket_activities.go                            │
│  activities/notification_activities.go                      │
│  - Pure implementation: API calls, business logic           │
│  - Reusable across workflows                                │
└─────────────────────────────────────────────────────────────┘
```

### YAML Template (Declarative DAG)

```yaml
# templates/ticket-lifecycle.yaml
name: ticket-lifecycle
description: Complete ticket workflow

steps:
  - name: create-ticket
    activity: CreateTicket
    input:
      title: $input.title
      description: $input.description
      priority: $input.priority
    output: ticket              # Store result as $ticket
    on_success:                 # Run after successful creation
      - activity: SendTicketCreatedEmail
        input:
          ticketId: $ticket.id
          title: $ticket.title

  - name: assign-ticket
    activity: AssignTicket
    input:
      ticketId: $ticket.id
      agentId: $input.agentId
    output: ticket
    condition: $ticket.priority == 'HIGH'  # Only if HIGH priority
    on_success:
      - activity: SendTicketAssignedEmail
        input:
          ticketId: $ticket.id
          agentId: $ticket.assignedAgentId

  # ... more steps
```

### Generic Workflow (Interprets YAML)

```go
// workflows/declarative_workflow.go
func TicketLifecycleDeclarativeWorkflow(ctx workflow.Context, input models.TicketInput) {
    // ✅ Load DAG definition from YAML
    definition, _ := engine.LoadWorkflowFromYAML("../templates/ticket-lifecycle.yaml")
    
    // ✅ Execute using generic engine
    result, _ := engine.GenericWorkflow(ctx, *definition, inputMap)
    
    return result
}
```

The workflow function is now **4 lines** instead of 145 lines!

### Activity Implementations (Unchanged)

```go
// activities/ticket_activities.go
// ✅ Pure implementation - no orchestration logic
func (a *TicketActivities) CreateTicket(ctx context.Context, input models.TicketInput) (models.TicketResponse, error) {
    // Just the implementation
    resp, err := http.Post(API_BASE_URL+"/tickets", ...)
    return ticketResponse, err
}
```

## Benefits

### 1. **Separation of Concerns**

| Concern | Location | Who Modifies |
|---------|----------|--------------|
| **DAG Structure** | YAML template | Product/Business team |
| **Orchestration Engine** | `engine/workflow_engine.go` | Platform team (once) |
| **Activity Logic** | `activities/*.go` | Backend developers |

### 2. **Easy Workflow Modifications**

**Change workflow order:**
```yaml
# Before: Create → Assign → Update → Close
# After: Create → Update → Assign → Close
# Just reorder steps in YAML, no code changes!
```

**Add conditional logic:**
```yaml
- name: escalate-urgent
  activity: EscalateTicket
  condition: $ticket.priority == 'URGENT'
  input:
    ticketId: $ticket.id
```

**Add parallel execution:**
```yaml
- name: notify-all
  parallel: true
  steps:
    - activity: SendEmail
    - activity: SendSlack  
    - activity: SendSMS
```

### 3. **Reusable Activities**

Activities become building blocks usable in multiple workflows:

```yaml
# Workflow 1: ticket-lifecycle.yaml
steps:
  - activity: CreateTicket
  - activity: SendEmail

# Workflow 2: bug-triage.yaml  
steps:
  - activity: CreateTicket      # Same activity!
  - activity: AssignToOnCall
  - activity: SendSlack
```

### 4. **Visual DAG Representation**

YAML is much easier to visualize than Go code. You can generate diagrams:

```
templates/ticket-lifecycle.yaml
        ↓
   [Diagram Tool]
        ↓
CreateTicket
    ↓
AssignTicket ──→ SendEmail (parallel)
    ↓
UpdateTicket
    ↓
CloseTicket
    ↓
RateTicket
```

## Comparison: Hardcoded vs Declarative

### Hardcoded Approach (Current)

**File:** `workflows/ticket_lifecycle_workflow.go` (145 lines)

```go
func TicketLifecycleWorkflow(ctx workflow.Context, input models.TicketInput) {
    // Configure options
    activityOptions := workflow.ActivityOptions{...}
    
    // Step 1
    var ticketResponse models.TicketResponse
    err := workflow.ExecuteActivity(ctx, ticketActivities.CreateTicket, input).Get(ctx, &ticketResponse)
    if err != nil { return err }
    
    // Email notification
    err = workflow.ExecuteActivity(ctx, notificationActivities.SendTicketCreatedEmail, ...).Get(ctx, nil)
    if err != nil { logger.Warn(...) }
    
    // Step 2
    err = workflow.ExecuteActivity(ctx, ticketActivities.AssignTicket, ...).Get(ctx, &ticketResponse)
    if err != nil { return err }
    
    // ... 100 more lines
}
```

**Pros:**
- Full Go type safety
- IDE autocomplete
- Compile-time error checking

**Cons:**
- Verbose (145 lines for 5 steps)
- Orchestration logic mixed with activity calls
- Requires code changes for workflow modifications
- Hard to visualize DAG structure
- Duplication across similar workflows

### Declarative Approach (New)

**File 1:** `templates/ticket-lifecycle.yaml` (115 lines)

```yaml
steps:
  - name: create-ticket
    activity: CreateTicket
    output: ticket
  - name: assign-ticket  
    activity: AssignTicket
    output: ticket
  # ... clean, readable structure
```

**File 2:** `workflows/declarative_workflow.go` (25 lines)

```go
func TicketLifecycleDeclarativeWorkflow(ctx workflow.Context, input models.TicketInput) {
    definition, _ := engine.LoadWorkflowFromYAML("../templates/ticket-lifecycle.yaml")
    result, _ := engine.GenericWorkflow(ctx, *definition, inputMap)
    return result
}
```

**Pros:**
- Clear separation: DAG (YAML) vs Implementation (Go)
- Easy to modify workflow structure
- Generic engine works for all workflows
- Visual representation from YAML
- Business users can read/modify YAML

**Cons:**
- Runtime YAML parsing overhead (can cache)
- Less type safety (YAML is stringly-typed)
- Variable resolution is dynamic (e.g., `$ticket.id`)
- Requires building generic engine once

## Advanced DAG Features (Future)

The generic engine can support:

### 1. Parallel Execution

```yaml
- name: notify-stakeholders
  parallel: true
  steps:
    - activity: SendEmailToCustomer
    - activity: SendEmailToManager
    - activity: SendSlackMessage
  # All 3 execute simultaneously
```

### 2. Conditional Branching

```yaml
- name: handle-by-priority
  condition: $ticket.priority
  branches:
    URGENT:
      - activity: EscalateToSenior
      - activity: NotifyManager
    HIGH:
      - activity: AssignToAvailable
    default:
      - activity: AddToQueue
```

### 3. Loops (Dynamic DAG)

```yaml
- name: create-subtasks
  for_each: $input.components
  activity: CreateSubtask
  input:
    component: $item
  output: subtasks  # Array of results
```

### 4. Wait Conditions

```yaml
- name: wait-for-approval
  activity: WaitForSignal
  input:
    signalName: approval-received
  timeout: 24h
```

## Implementation Status

**✅ Implemented:**
- YAML parser (`engine/yaml_parser.go`)
- Generic workflow engine (`engine/workflow_engine.go`)
- Sequential step execution
- Variable resolution (`$input.title`, `$ticket.id`)
- On-success hooks (email notifications)
- Declarative workflow wrapper

**🚧 Not Yet Implemented:**
- Parallel execution (YAML defines it, engine doesn't execute yet)
- Condition evaluation (structure exists, always returns true)
- Advanced variable paths (nested objects)
- Error handling strategies from YAML
- Loop/iteration support

## How to Use

### Option 1: Keep Hardcoded Workflow (Type-Safe)

```go
// Use existing ticket_lifecycle_workflow.go
c.ExecuteWorkflow(ctx, workflowOptions, workflows.TicketLifecycleWorkflow, input)
```

### Option 2: Switch to Declarative Workflow

```go
// Register declarative workflow in main.go
w.RegisterWorkflow(workflows.TicketLifecycleDeclarativeWorkflow)

// Trigger it
c.ExecuteWorkflow(ctx, workflowOptions, workflows.TicketLifecycleDeclarativeWorkflow, input)
```

Both produce the same result, but declarative approach is more maintainable long-term.

## Next Steps

To fully leverage declarative workflows:

1. **Add proper variable resolution** - Handle nested paths like `$ticket.assignedAgent.name`
2. **Implement condition parser** - Evaluate expressions like `$ticket.priority == 'HIGH'`
3. **Support parallel execution** - Execute multiple activities concurrently
4. **Add workflow validation** - Check YAML syntax and activity names at startup
5. **Build visual DAG renderer** - Generate Mermaid diagrams from YAML
6. **Cache YAML definitions** - Load once at worker startup instead of per-workflow

## Conclusion

The declarative approach **separates DAG structure (YAML) from activity implementation (Go)**, making workflows:
- **Easier to understand** - YAML is more readable than 145 lines of Go
- **Faster to modify** - Change YAML, not code
- **More maintainable** - Generic engine handles all orchestration
- **Reusable** - Activities work across workflows

The trade-off is less compile-time safety and a more complex engine, but for workflow-heavy systems, the benefits outweigh the costs.
