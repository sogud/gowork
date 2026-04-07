# Quick Start Guide

This guide will help you get Cowork Agent up and running quickly.

## Prerequisites

Before starting, ensure you have the following installed:

| Requirement | Version | Purpose |
|-------------|---------|---------|
| Go | 1.21+ | Runtime environment |
| Ollama | Latest | Local LLM inference |
| Gemma 4 | Latest | Language model |

### Installing Ollama

```bash
# macOS
brew install ollama

# Linux
curl -fsSL https://ollama.ai/install.sh | sh

# Start Ollama service
ollama serve
```

### Pulling the Model

```bash
# Pull Gemma 4 model
ollama pull gemma4

# Verify installation
ollama list
```

## Installation Steps

### Step 1: Clone the Repository

```bash
git clone https://github.com/sogud/gowork.git
cd gowork
```

### Step 2: Install Dependencies

```bash
go mod download
```

### Step 3: Configure the Application

Create or modify `config.yaml`:

```yaml
server:
  port: 8080
  mode: "api"
  log_level: "info"

ollama:
  base_url: "http://localhost:11434"
  model: "gemma4"
  timeout: 60s
  max_retries: 3

agents:
  - name: "researcher"
    enabled: true
  - name: "analyst"
    enabled: true
  - name: "writer"
    enabled: true
  - name: "reviewer"
    enabled: true
  - name: "coordinator"
    enabled: true

workflow:
  default_type: "sequential"
  timeout: 300s

memory:
  type: "inmemory"

tools:
  web_search:
    enabled: true
    mode: "mock"
  file_operations:
    enabled: true
```

### Step 4: Verify Ollama is Running

```bash
# Check if Ollama is accessible
curl http://localhost:11434/api/tags

# Expected output: JSON list of available models
```

## Running the Application

### API Mode

Start the HTTP server:

```bash
# Using go run
go run main.go --mode api --config config.yaml

# Or build and run
go build -o gowork .
./gowork --mode api --config config.yaml
```

You should see output like:

```
2024/04/04 10:00:00 Initializing Ollama model adapter...
2024/04/04 10:00:00 Model initialized: gemma4
2024/04/04 10:00:00 Initializing agent registry...
2024/04/04 10:00:00 Creating and registering agents...
2024/04/04 10:00:00 Agent researcher created and registered
2024/04/04 10:00:00 Agent analyst created and registered
2024/04/04 10:00:00 Agent writer created and registered
2024/04/04 10:00:00 Agent reviewer created and registered
2024/04/04 10:00:00 Agent coordinator created and registered
2024/04/04 10:00:00 Registered agents: [researcher analyst writer reviewer coordinator]
2024/04/04 10:00:00 Initializing memory service...
2024/04/04 10:00:00 Initializing tool registry...
2024/04/04 10:00:00 Tool web_search registered
2024/04/04 10:00:00 Tool file_reader registered
2024/04/04 10:00:00 Tool file_writer registered
2024/04/04 10:00:00 Tool calculator registered
2024/04/04 10:00:00 Creating API server...
2024/04/04 10:00:00 Application initialized successfully
2024/04/04 10:00:00 Starting API server on port 8080...
```

### CLI Mode

Execute a single task directly:

```bash
go run main.go --mode cli --task "Research the benefits of Go for building AI applications"
```

## First Workflow Execution

### Health Check

```bash
curl http://localhost:8080/api/v1/health
```

Response:

```json
{
  "status": "healthy"
}
```

### List Available Agents

```bash
curl http://localhost:8080/api/v1/agents
```

Response:

```json
{
  "agents": ["researcher", "analyst", "writer", "reviewer", "coordinator"]
}
```

### Execute Your First Workflow

#### Sequential Workflow

Execute agents in sequence, with output passed between them:

```bash
curl -X POST http://localhost:8080/api/v1/workflow/execute \
  -H "Content-Type: application/json" \
  -d '{
    "task": "Explain the concept of multi-agent systems in AI",
    "type": "sequential",
    "agents": ["researcher", "writer"]
  }'
```

Expected response:

```json
{
  "workflow_id": "wf-1712200000000000000",
  "status": "completed",
  "agent_results": {
    "researcher": {
      "output": "Research findings about multi-agent systems..."
    },
    "writer": {
      "output": "Written explanation based on research..."
    }
  },
  "execution_time_ms": 1500,
  "output": "Final aggregated output...",
  "iterations": 1
}
```

#### Parallel Workflow

Execute multiple agents simultaneously:

```bash
curl -X POST http://localhost:8080/api/v1/workflow/execute \
  -H "Content-Type: application/json" \
  -d '{
    "task": "Analyze the pros and cons of microservices architecture",
    "type": "parallel",
    "agents": ["researcher", "analyst"]
  }'
```

#### Dynamic Workflow

Let the coordinator decide the best agent sequence:

```bash
curl -X POST http://localhost:8080/api/v1/workflow/execute \
  -H "Content-Type: application/json" \
  -d '{
    "task": "Write a comprehensive report on the state of AI in 2024",
    "type": "dynamic",
    "agents": ["coordinator"]
  }'
```

## Common Scenarios

### Scenario 1: Research and Report

Create a research report on a topic:

```bash
curl -X POST http://localhost:8080/api/v1/workflow/execute \
  -H "Content-Type: application/json" \
  -d '{
    "task": "Research and write a report on sustainable energy solutions",
    "type": "sequential",
    "agents": ["researcher", "analyst", "writer", "reviewer"]
  }'
```

### Scenario 2: Quick Analysis

Get quick analysis from multiple perspectives:

```bash
curl -X POST http://localhost:8080/api/v1/workflow/execute \
  -H "Content-Type: application/json" \
  -d '{
    "task": "Compare React and Vue.js for enterprise applications",
    "type": "parallel",
    "agents": ["analyst"]
  }'
```

### Scenario 3: Iterative Refinement

Refine content through multiple iterations:

```bash
curl -X POST http://localhost:8080/api/v1/workflow/execute \
  -H "Content-Type: application/json" \
  -d '{
    "task": "Write a technical blog post about GraphQL",
    "type": "loop",
    "agents": ["writer", "reviewer"],
    "max_iter": 3
  }'
```

### Scenario 4: Complex Task Orchestration

Let the system decide the best approach:

```bash
curl -X POST http://localhost:8080/api/v1/workflow/execute \
  -H "Content-Type: application/json" \
  -d '{
    "task": "Create a comprehensive guide for setting up a Kubernetes cluster",
    "type": "dynamic",
    "agents": ["coordinator"]
  }'
```

## Troubleshooting

### Ollama Connection Error

```
Error: failed to create Ollama model: connection refused
```

**Solution**: Ensure Ollama is running:

```bash
ollama serve
```

### Model Not Found

```
Error: model 'gemma4' not found
```

**Solution**: Pull the model:

```bash
ollama pull gemma4
```

### Port Already in Use

```
Error: listen tcp :8080: bind: address already in use
```

**Solution**: Change the port in `config.yaml` or kill the existing process:

```bash
lsof -i :8080
kill -9 <PID>
```

### Timeout Errors

```
Error: context deadline exceeded
```

**Solution**: Increase timeout in `config.yaml`:

```yaml
ollama:
  timeout: 120s
workflow:
  timeout: 600s
```

## Next Steps

- Read the [API Documentation](API.md) for detailed endpoint information
- Explore the [Architecture](ARCHITECTURE.md) to understand the system design
- Check out the [Contributing Guidelines](../README.md#contributing) to contribute

## Getting Help

- Open an issue on GitHub for bugs or feature requests
- Check existing documentation in the `docs/` directory
- Review test files for usage examples