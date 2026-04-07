# Architecture Overview

## Layers

1. **Model Layer** - Ollama Gemma 4 adapter
2. **Agents Layer** - Specialized agents + coordinator
3. **Workflow Layer** - Orchestration patterns
4. **Memory Layer** - State management
5. **Tools Layer** - Extensible tool registry
6. **Server Layer** - REST API + CLI

## Data Flow

User Request -> Server -> Workflow Engine -> Agents -> Tools -> Model -> Response