# API Documentation

Complete reference for the Cowork Agent REST API.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

Currently, the API does not require authentication. This will be added in future versions.

## Content Type

All requests and responses use JSON:

```
Content-Type: application/json
```

---

## Endpoints

### Health Check

Check if the server is running and healthy.

**Endpoint**: `GET /health`

**Response**:

```json
{
  "status": "healthy"
}
```

**Status Codes**:

| Code | Description |
|------|-------------|
| 200 | Server is healthy |

**Example**:

```bash
curl http://localhost:8080/api/v1/health
```

---

### List Agents

Get a list of all available agents.

**Endpoint**: `GET /agents`

**Response**:

```json
{
  "agents": ["researcher", "analyst", "writer", "reviewer", "coordinator"]
}
```

**Status Codes**:

| Code | Description |
|------|-------------|
| 200 | Success |

**Example**:

```bash
curl http://localhost:8080/api/v1/agents
```

---

### Execute Workflow

Execute a workflow with specified agents and task.

**Endpoint**: `POST /workflow/execute`

**Request Body**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| task | string | Yes | The task description to execute |
| type | string | No | Workflow type: `sequential`, `parallel`, `loop`, `dynamic`. Default: `sequential` |
| agents | []string | Yes | List of agent names to execute |
| max_iter | int | No | Maximum iterations for loop/dynamic workflows. Default: 3 |

**Request Example**:

```json
{
  "task": "Research and explain quantum computing basics",
  "type": "sequential",
  "agents": ["researcher", "writer"],
  "max_iter": 3
}
```

**Response**:

| Field | Type | Description |
|-------|------|-------------|
| workflow_id | string | Unique identifier for the workflow execution |
| status | string | Execution status: `completed` or `failed` |
| agent_results | object | Results from each agent, keyed by agent name |
| execution_time_ms | int64 | Total execution time in milliseconds |
| output | string | Final aggregated output |
| iterations | int | Number of iterations performed |

**Response Example**:

```json
{
  "workflow_id": "wf-1712200000000000000",
  "status": "completed",
  "agent_results": {
    "researcher": {
      "output": "Quantum computing uses quantum bits (qubits)..."
    },
    "writer": {
      "output": "# Quantum Computing Basics\n\nQuantum computing represents..."
    }
  },
  "execution_time_ms": 2500,
  "output": "Final comprehensive output...",
  "iterations": 1
}
```

**Status Codes**:

| Code | Description |
|------|-------------|
| 200 | Workflow executed successfully |
| 400 | Invalid request (missing required fields, invalid workflow type) |
| 500 | Internal server error (workflow execution failed) |

**Example**:

```bash
curl -X POST http://localhost:8080/api/v1/workflow/execute \
  -H "Content-Type: application/json" \
  -d '{
    "task": "Explain the water cycle",
    "type": "sequential",
    "agents": ["researcher", "writer"]
  }'
```

---

## Workflow Types

### Sequential

Agents execute one after another in order. Each agent receives the output from the previous agent.

**Use Cases**:
- Research and writing pipelines
- Step-by-step analysis
- Content creation with review

**Example**:

```json
{
  "task": "Write a report on climate change",
  "type": "sequential",
  "agents": ["researcher", "analyst", "writer", "reviewer"]
}
```

Execution flow:

```
Researcher -> Analyst -> Writer -> Reviewer
```

### Parallel

All agents execute simultaneously. Results are aggregated at the end.

**Use Cases**:
- Multi-perspective analysis
- Independent research tasks
- Comparative studies

**Example**:

```json
{
  "task": "Analyze the pros and cons of remote work",
  "type": "parallel",
  "agents": ["researcher", "analyst"]
}
```

Execution flow:

```
Researcher --\
             +-> Aggregate Results
Analyst -----/
```

### Loop

Agents execute in a loop until a stopping condition is met or max iterations reached.

**Use Cases**:
- Iterative refinement
- Quality improvement cycles
- Content revision

**Example**:

```json
{
  "task": "Write a blog post about AI ethics",
  "type": "loop",
  "agents": ["writer", "reviewer"],
  "max_iter": 5
}
```

Execution flow:

```
Writer -> Reviewer -> Writer -> Reviewer -> ... (until max_iter or satisfied)
```

### Dynamic

The coordinator agent dynamically decides which agents to invoke based on the task.

**Use Cases**:
- Complex, multi-step tasks
- Uncertain execution paths
- Adaptive workflows

**Example**:

```json
{
  "task": "Create a comprehensive guide on machine learning",
  "type": "dynamic",
  "agents": ["coordinator"]
}
```

Execution flow:

```
Coordinator -> [decides] -> Researcher -> Writer -> Reviewer
```

---

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error Type",
  "message": "Detailed error message"
}
```

### Common Errors

#### 400 Bad Request - Missing Task

```json
{
  "error": "Bad Request",
  "message": "task: task is required"
}
```

#### 400 Bad Request - Missing Agents

```json
{
  "error": "Bad Request",
  "message": "agents: at least one agent is required"
}
```

#### 400 Bad Request - Invalid Workflow Type

```json
{
  "error": "Bad Request",
  "message": "type: invalid workflow type, must be one of: sequential, parallel, loop, dynamic"
}
```

#### 500 Internal Server Error

```json
{
  "error": "Internal Server Error",
  "message": "workflow execution failed: context deadline exceeded"
}
```

---

## Agents Reference

### Researcher

**Role**: Information gathering and research

**Capabilities**:
- Searches for relevant information
- Collects data from multiple sources
- Summarizes findings

**Best For**:
- Initial research on a topic
- Fact-finding tasks
- Information synthesis

### Analyst

**Role**: Analysis and insights

**Capabilities**:
- Analyzes data and information
- Identifies patterns and trends
- Provides recommendations

**Best For**:
- Data analysis
- Comparative studies
- Strategic recommendations

### Writer

**Role**: Content creation

**Capabilities**:
- Writes clear, structured content
- Adapts tone and style
- Creates various content types

**Best For**:
- Article writing
- Report generation
- Documentation

### Reviewer

**Role**: Quality assurance

**Capabilities**:
- Reviews content quality
- Provides constructive feedback
- Identifies improvements

**Best For**:
- Content review
- Quality checks
- Feedback generation

### Coordinator

**Role**: Workflow orchestration

**Capabilities**:
- Plans agent sequences
- Delegates tasks
- Manages dependencies

**Best For**:
- Complex workflows
- Dynamic task routing
- Multi-agent coordination

---

## Rate Limiting

Currently, there is no rate limiting. This will be added in future versions.

## Timeouts

| Operation | Default Timeout | Configurable |
|-----------|-----------------|--------------|
| HTTP Request | 15s read, 15s write | Yes (config.yaml) |
| Workflow Execution | 300s | Yes (config.yaml) |
| Model Inference | 60s | Yes (config.yaml) |

## CORS

Currently, CORS is not configured. This will be added in future versions.

## Versioning

The API is versioned via the URL path (`/api/v1`). Future versions will maintain backward compatibility where possible.

## OpenAPI Specification

A formal OpenAPI specification will be provided in future versions.

---

## Examples

### Python Client

```python
import requests

BASE_URL = "http://localhost:8080/api/v1"

def execute_workflow(task, agents, workflow_type="sequential"):
    response = requests.post(
        f"{BASE_URL}/workflow/execute",
        json={
            "task": task,
            "type": workflow_type,
            "agents": agents
        }
    )
    return response.json()

# Example usage
result = execute_workflow(
    task="Explain the benefits of containerization",
    agents=["researcher", "writer"],
    workflow_type="sequential"
)
print(result["output"])
```

### JavaScript Client

```javascript
const BASE_URL = 'http://localhost:8080/api/v1';

async function executeWorkflow(task, agents, type = 'sequential') {
  const response = await fetch(`${BASE_URL}/workflow/execute`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      task,
      type,
      agents,
    }),
  });
  return response.json();
}

// Example usage
const result = await executeWorkflow(
  'Explain microservices architecture',
  ['researcher', 'analyst', 'writer'],
  'sequential'
);
console.log(result.output);
```

### Go Client

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type WorkflowRequest struct {
    Task    string   `json:"task"`
    Type    string   `json:"type"`
    Agents  []string `json:"agents"`
    MaxIter int      `json:"max_iter,omitempty"`
}

type WorkflowResponse struct {
    WorkflowID      string                 `json:"workflow_id"`
    Status          string                 `json:"status"`
    AgentResults    map[string]interface{} `json:"agent_results"`
    ExecutionTimeMs int64                  `json:"execution_time_ms"`
    Output          string                 `json:"output"`
    Iterations      int                    `json:"iterations"`
}

func executeWorkflow(task string, agents []string, workflowType string) (*WorkflowResponse, error) {
    req := WorkflowRequest{
        Task:   task,
        Type:   workflowType,
        Agents: agents,
    }

    body, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    resp, err := http.Post(
        "http://localhost:8080/api/v1/workflow/execute",
        "application/json",
        bytes.NewReader(body),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result WorkflowResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &result, nil
}

func main() {
    result, err := executeWorkflow(
        "Explain the SOLID principles",
        []string{"researcher", "writer"},
        "sequential",
    )
    if err != nil {
        panic(err)
    }
    fmt.Println(result.Output)
}
```