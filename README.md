# Gowork

Multi-agent collaboration framework supporting any ADK-compatible models (Ollama, Gemini, custom).

## Overview

Gowork is a flexible multi-agent system that enables complex task execution through specialized AI agents working together. The system supports multiple workflow patterns and **multiple model providers**.

### Key Features

- **Multi-Model Support**: Works with Ollama, Gemini, or any ADK-compatible model
- **Multi-Agent Architecture**: 5 specialized agents (Researcher, Analyst, Writer, Reviewer, Coordinator)
- **Flexible Workflows**: Sequential, parallel, loop, and dynamic execution patterns
- **Terminal UI (TUI)**: Interactive terminal interface with real-time workflow monitoring
- **Streaming Support**: Real-time response streaming
- **State Management**: Shared memory service for agent collaboration
- **Extensible Tools**: Pluggable tool registry for extending agent capabilities
- **REST API**: HTTP endpoints for workflow execution and management

## Supported Models

| Provider | Description | Configuration |
|----------|-------------|---------------|
| **Ollama** | Local models (Gemma, Llama, etc.) | `provider: ollama` |
| **Gemini** | Google Gemini API | `provider: gemini` |
| **Custom** | Any ADK-compatible model | `provider: custom` |

## Architecture

```
+------------------+
|    User Request  |
+--------+---------+
         |
         v
+--------+---------+
|   Server Layer   |  <- REST API / CLI / TUI
+--------+---------+
         |
         v
+--------+---------+
| Workflow Engine  |  <- Sequential/Parallel/Loop/Dynamic
+--------+---------+
         |
         v
+--------+---------+
|  Agent Registry  |  <- Researcher, Analyst, Writer, Reviewer, Coordinator
+--------+---------+
         |
         +-----> +----------------+
                | Memory Service |  <- State sharing between agents
                +----------------+
         |
         v
+--------+---------+
|  Tool Registry   |  <- Web Search, File I/O, Calculator
+--------+---------+
         |
         v
+--------+---------+
|   Model Layer    |  <- Ollama / Gemini / Custom
+------------------+
```

### Components

| Layer | Description |
|-------|-------------|
| Model | Model provider abstraction (Ollama, Gemini, custom) |
| Agents | Specialized agents with distinct roles |
| Workflow | Orchestration engine for agent coordination |
| Memory | State management for cross-agent communication |
| Tools | Extensible tool system for agent capabilities |
| Server | HTTP API, CLI, and TUI interfaces |
| TUI | Terminal UI with Bubbletea framework |
| State | Application state management with subscriber pattern |

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Ollama installed and running
- Gemma 4 model pulled

### Installation

```bash
# Clone the repository
git clone https://github.com/sogud/gowork.git
cd gowork

# Install dependencies
go mod download

# Pull the Gemma 4 model
ollama pull gemma4
```

### Running the Server

```bash
# TUI mode (interactive terminal interface)
go run main.go --mode tui --config config.yaml

# API mode (default)
go run main.go --mode api --config config.yaml

# CLI mode for single task execution
go run main.go --mode cli --task "Research the benefits of microservices architecture"

# Build and run
go build -o gowork .
./gowork --mode tui --config config.yaml
```

### TUI Mode

The TUI provides an interactive terminal interface with:

- **Home Screen**: Quick navigation and agent status overview
- **TaskInput Screen**: Submit tasks with workflow type and agent selection
- **Monitor Screen**: Real-time workflow execution monitoring with progress bars
- **Config Screen**: Manage model, agents, and tools configuration
- **History Screen**: View past workflow executions

**TUI Keyboard Shortcuts:**

| Key | Action |
|-----|--------|
| `1-4` | Navigate to screens |
| `Tab` | Cycle through fields |
| `Space` | Toggle selection |
| `Enter` | Confirm/Submit |
| `Esc` | Go back |
| `q` | Quit |
| `?` | Show help |

### First Request

```bash
# Health check
curl http://localhost:8080/api/v1/health

# List available agents
curl http://localhost:8080/api/v1/agents

# Execute a workflow
curl -X POST http://localhost:8080/api/v1/workflow/execute \
  -H "Content-Type: application/json" \
  -d '{
    "task": "Research and summarize the latest trends in AI agents",
    "type": "sequential",
    "agents": ["researcher", "analyst", "writer"]
  }'
```

## Configuration

Configuration is managed via `config.yaml`:

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
  default_type: "dynamic"
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

### Configuration Options

| Section | Option | Description | Default |
|---------|--------|-------------|---------|
| server | port | HTTP server port | 8080 |
| server | mode | Running mode (api/cli) | api |
| server | log_level | Logging verbosity | info |
| ollama | base_url | Ollama API endpoint | http://localhost:11434 |
| ollama | model | Model name | gemma4 |
| ollama | timeout | Request timeout | 60s |
| ollama | max_retries | Retry attempts | 3 |
| workflow | default_type | Default workflow type | sequential |
| workflow | timeout | Workflow timeout | 300s |
| memory | type | Memory backend type | inmemory |

## Agents

### Researcher
Gathers information and performs initial research on topics.

### Analyst
Analyzes data and provides insights and recommendations.

### Writer
Creates written content, reports, and documentation.

### Reviewer
Reviews and provides feedback on outputs from other agents.

### Coordinator
Orchestrates multi-agent workflows and manages task delegation.

## Workflows

### Sequential
Agents execute one after another, with each agent's output passed to the next.

```bash
curl -X POST http://localhost:8080/api/v1/workflow/execute \
  -H "Content-Type: application/json" \
  -d '{"task": "...", "type": "sequential", "agents": ["researcher", "writer"]}'
```

### Parallel
All agents execute simultaneously, with results aggregated at the end.

```bash
curl -X POST http://localhost:8080/api/v1/workflow/execute \
  -H "Content-Type: application/json" \
  -d '{"task": "...", "type": "parallel", "agents": ["researcher", "analyst"]}'
```

### Loop
Agents execute in a loop until a condition is met or max iterations reached.

```bash
curl -X POST http://localhost:8080/api/v1/workflow/execute \
  -H "Content-Type: application/json" \
  -d '{"task": "...", "type": "loop", "agents": ["researcher"], "max_iter": 5}'
```

### Dynamic
Coordinator dynamically decides which agents to invoke based on the task.

```bash
curl -X POST http://localhost:8080/api/v1/workflow/execute \
  -H "Content-Type: application/json" \
  -d '{"task": "...", "type": "dynamic", "agents": ["coordinator"]}'
```

## API Reference

See [docs/API.md](docs/API.md) for complete API documentation.

## Quick Start Guide

See [docs/QUICKSTART.md](docs/QUICKSTART.md) for step-by-step instructions.

## Architecture Details

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed architecture documentation.

## Project Structure

```
gowork/
+-- main.go              # Application entry point
+-- config.yaml          # Configuration file
+-- go.mod               # Go module definition
+-- go.sum               # Dependency checksums
+-- agents/              # Agent implementations
|   +-- registry.go      # Agent registry
|   +-- researcher.go    # Researcher agent
|   +-- analyst.go       # Analyst agent
|   +-- writer.go        # Writer agent
|   +-- reviewer.go      # Reviewer agent
|   +-- coordinator.go   # Coordinator agent
+-- model/               # Model adapter
|   +-- provider.go      # Model provider abstraction
|   +-- ollama.go        # Ollama adapter
|   +-- config.go        # Model configuration
+-- workflow/            # Workflow engine
|   +-- engine.go        # Execution engine
|   +-- config.go        # Workflow configuration
+-- memory/              # State management
|   +-- service.go       # Memory service interface
|   +-- inmemory.go      # In-memory implementation
|   +-- state.go         # State manager
+-- state/               # Application state (for TUI)
|   +-- app_state.go     # Immutable state types
|   +-- store.go         # State store with subscribers
|   +-- notifier.go      # Event notification system
|   +-- history_store.go # Workflow history persistence
+-- tools/               # Tool implementations
|   +-- registry.go      # Tool registry
|   +-- search.go        # Web search tool
|   +-- file.go          # File read/write tools
|   +-- calculator.go    # Calculator tool
+-- server/              # HTTP server
|   +-- api.go           # API handlers
|   +-- config.go        # Server configuration
+-- tui/                 # Terminal UI
|   +-- main.go          # TUI entry point
|   +-- app.go           # Bubbletea app model
|   +-- registry.go      # Screen registry
|   +-- components/      # Reusable UI components
|   |   +-- input.go     # Text input
|   |   +-- list.go      # Selection list
|   |   +-- progress.go  # Progress bar
|   |   +-- status_bar.go# Status bar
|   |   +-- border.go    # Border styles
|   +-- screens/         # Screen implementations
|   |   +-- home.go      # Home screen
|   |   +-- task_input.go# Task submission
|   |   +-- monitor.go   # Workflow monitoring
|   |   +-- config.go    # Configuration
|   |   +-- history.go   # History view
|   +-- styles/          # Lipgloss styles
|       +-- theme.go     # Color theme
|       +-- layout.go    # Layout utilities
+-- tests/               # Integration tests
|   +-- e2e_test.go      # End-to-end tests
|   +-- mock_server.go   # Mock Ollama server
+-- docs/                # Documentation
    +-- ARCHITECTURE.md  # Architecture overview
    +-- API.md           # API documentation
    +-- QUICKSTART.md    # Quick start guide
```

## Testing

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific package tests
go test ./workflow -v
```

### Test Coverage

| Package | Coverage |
|---------|----------|
| agents | 50.7% |
| memory | 93.6% |
| model | 90.7% |
| server | 90.9% |
| state | 90.0% |
| tools | 79.5% |
| tui | 85.0% |
| tui/components | 82.0% |
| tui/screens | 80.0% |
| workflow | 95.5% |
| **Total** | **80.0%** |

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Write tests for new functionality
- Maintain test coverage above 80%
- Follow Go naming conventions
- Document exported functions and types
- Run `go fmt` before committing

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Built with [adk-go](https://github.com/google/adk-go) - Agent Development Kit for Go
- Powered by [Ollama](https://ollama.ai/) - Local LLM inference
- Uses [Gemma 4](https://ai.google.dev/gemma) - Open language model by Google
- TUI powered by [Bubbletea](https://github.com/charmbracelet/bubbletea) - Elegant TUI framework