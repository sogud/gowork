// Package main provides the entry point for the multi-agent workflow system.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sogud/gowork/agents"
	"github.com/sogud/gowork/memory"
	"github.com/sogud/gowork/model"
	"github.com/sogud/gowork/server"
	"github.com/sogud/gowork/tools"
	"github.com/sogud/gowork/workflow"
	adkagent "google.golang.org/adk/agent"
	adkmodel "google.golang.org/adk/model"
)

// CLI flags
var (
	configPath = flag.String("config", "config.yaml", "Path to configuration file")
	mode       = flag.String("mode", "api", "Running mode: api or cli")
	task       = flag.String("task", "", "Task description for CLI mode")
)

// Application holds all the main application components.
type Application struct {
	config       *AppConfig
	model        adkmodel.LLM
	registry     *agents.Registry
	memoryService memory.Service
	toolRegistry  tools.ToolRegistry
	apiServer    *server.APIServer
}

// AppConfig holds the application configuration from YAML file.
type AppConfig struct {
	Server  ServerConfig  `yaml:"server"`
	Ollama  OllamaConfig  `yaml:"ollama"`
	Agents  []AgentConfig `yaml:"agents"`
	Workflow WorkflowConfig `yaml:"workflow"`
	Memory  MemoryConfig  `yaml:"memory"`
	Tools   ToolsConfig   `yaml:"tools"`
}

// ServerConfig holds server configuration.
type ServerConfig struct {
	Port     int    `yaml:"port"`
	Mode     string `yaml:"mode"`
	LogLevel string `yaml:"log_level"`
}

// OllamaConfig holds Ollama model configuration.
type OllamaConfig struct {
	BaseURL   string        `yaml:"base_url"`
	Model     string        `yaml:"model"`
	Timeout   time.Duration `yaml:"timeout"`
	MaxRetries int          `yaml:"max_retries"`
}

// AgentConfig holds individual agent configuration.
type AgentConfig struct {
	Name    string `yaml:"name"`
	Enabled bool   `yaml:"enabled"`
}

// WorkflowConfig holds workflow configuration.
type WorkflowConfig struct {
	DefaultType string        `yaml:"default_type"`
	Timeout     time.Duration `yaml:"timeout"`
}

// MemoryConfig holds memory service configuration.
type MemoryConfig struct {
	Type string `yaml:"type"`
}

// ToolsConfig holds tools configuration.
type ToolsConfig struct {
	WebSearch       ToolConfig `yaml:"web_search"`
	FileOperations  ToolConfig `yaml:"file_operations"`
}

// ToolConfig holds individual tool configuration.
type ToolConfig struct {
	Enabled bool   `yaml:"enabled"`
	Mode    string `yaml:"mode"`
}

func main() {
	flag.Parse()

	// Load configuration
	appConfig, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize application
	app, err := initializeApp(ctx, appConfig)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run based on mode
	switch *mode {
	case "api":
		runAPIMode(app, sigChan, cancel)
	case "cli":
		if *task == "" {
			log.Fatal("CLI mode requires --task flag")
		}
		runCLIMode(ctx, app, *task)
	default:
		log.Fatalf("Unknown mode: %s. Use 'api' or 'cli'", *mode)
	}
}

// loadConfig loads configuration from YAML file.
// Currently uses hardcoded defaults; YAML parsing will be added later.
func loadConfig(path string) (*AppConfig, error) {
	// For now, return default configuration
	// TODO: Implement proper YAML parsing using yaml.v3 or similar
	config := &AppConfig{
		Server: ServerConfig{
			Port:     8080,
			Mode:     "api",
			LogLevel: "info",
		},
		Ollama: OllamaConfig{
			BaseURL:   "http://localhost:11434",
			Model:     "gemma4",
			Timeout:   60 * time.Second,
			MaxRetries: 3,
		},
		Agents: []AgentConfig{
			{Name: "researcher", Enabled: true},
			{Name: "analyst", Enabled: true},
			{Name: "writer", Enabled: true},
			{Name: "reviewer", Enabled: true},
			{Name: "coordinator", Enabled: true},
		},
		Workflow: WorkflowConfig{
			DefaultType: "sequential",
			Timeout:     300 * time.Second,
		},
		Memory: MemoryConfig{
			Type: "inmemory",
		},
		Tools: ToolsConfig{
			WebSearch:      ToolConfig{Enabled: true, Mode: "mock"},
			FileOperations: ToolConfig{Enabled: true, Mode: "mock"},
		},
	}

	log.Printf("Using configuration (defaults, YAML parsing pending)")
	return config, nil
}

// initializeApp initializes all application components.
func initializeApp(ctx context.Context, config *AppConfig) (*Application, error) {
	app := &Application{
		config: config,
	}

	var err error

	// 1. Create Ollama model adapter
	log.Println("Initializing Ollama model adapter...")
	modelConfig := &model.Config{
		BaseURL:   config.Ollama.BaseURL,
		ModelName: config.Ollama.Model,
		Timeout:   config.Ollama.Timeout,
	}

	app.model, err = model.NewOllamaModel(ctx, modelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama model: %w", err)
	}
	log.Printf("Model initialized: %s", app.model.Name())

	// 2. Initialize agent registry
	log.Println("Initializing agent registry...")
	app.registry = agents.NewRegistry()

	// 3. Create and register all agents
	log.Println("Creating and registering agents...")
	if err := registerAgents(app.model, app.registry, config.Agents); err != nil {
		return nil, fmt.Errorf("failed to register agents: %w", err)
	}
	log.Printf("Registered agents: %v", app.registry.List())

	// 4. Initialize memory service
	log.Println("Initializing memory service...")
	app.memoryService = memory.NewInMemoryService()

	// 5. Initialize tool registry and register tools
	log.Println("Initializing tool registry...")
	app.toolRegistry = tools.NewRegistry()
	if err := registerTools(app.toolRegistry, config.Tools); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}
	log.Printf("Registered tools: %v", app.toolRegistry.List())

	// 6. Create API server
	log.Println("Creating API server...")
	app.apiServer, err = server.NewAPIServer(
		app.registry,
		server.WithPort(config.Server.Port),
		server.WithShutdownTimeout(30*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create API server: %w", err)
	}

	log.Println("Application initialized successfully")
	return app, nil
}

// registerAgents creates and registers all enabled agents.
func registerAgents(model adkmodel.LLM, registry *agents.Registry, agentConfigs []AgentConfig) error {
	for _, agentCfg := range agentConfigs {
		if !agentCfg.Enabled {
			log.Printf("Agent %s is disabled, skipping", agentCfg.Name)
			continue
		}

		var agent adkagent.Agent
		var err error

		switch agentCfg.Name {
		case "researcher":
			agent, err = agents.NewResearcher(model)
		case "analyst":
			agent, err = agents.NewAnalyst(model)
		case "writer":
			agent, err = agents.NewWriter(model)
		case "reviewer":
			agent, err = agents.NewReviewer(model)
		case "coordinator":
			agent, err = agents.NewCoordinator(model, registry)
		default:
			return fmt.Errorf("unknown agent name: %s", agentCfg.Name)
		}

		if err != nil {
			return fmt.Errorf("failed to create agent %s: %w", agentCfg.Name, err)
		}

		// Wrap agent as ExecutableAgent for registry
		execAgent := agents.NewExecutableAgent(agent)
		if err := registry.Register(execAgent); err != nil {
			return fmt.Errorf("failed to register agent %s: %w", agentCfg.Name, err)
		}

		log.Printf("Agent %s created and registered", agentCfg.Name)
	}

	return nil
}

// registerTools creates and registers all enabled tools.
func registerTools(registry tools.ToolRegistry, toolsCfg ToolsConfig) error {
	var err error

	if toolsCfg.WebSearch.Enabled {
		searchTool, err := tools.NewWebSearchTool(tools.WebSearchToolConfig{})
		if err != nil {
			return fmt.Errorf("failed to create web_search tool: %w", err)
		}
		if err := registry.Register(searchTool); err != nil {
			return fmt.Errorf("failed to register web_search tool: %w", err)
		}
		log.Println("Tool web_search registered")
	}

	if toolsCfg.FileOperations.Enabled {
		fileReaderTool, err := tools.NewFileReaderTool(tools.FileToolConfig{})
		if err != nil {
			return fmt.Errorf("failed to create file_reader tool: %w", err)
		}
		if err := registry.Register(fileReaderTool); err != nil {
			return fmt.Errorf("failed to register file_reader tool: %w", err)
		}
		log.Println("Tool file_reader registered")

		fileWriterTool, err := tools.NewFileWriterTool(tools.FileToolConfig{})
		if err != nil {
			return fmt.Errorf("failed to create file_writer tool: %w", err)
		}
		if err := registry.Register(fileWriterTool); err != nil {
			return fmt.Errorf("failed to register file_writer tool: %w", err)
		}
		log.Println("Tool file_writer registered")
	}

	// Always register calculator tool
	calcTool, err := tools.NewCalculatorTool()
	if err != nil {
		return fmt.Errorf("failed to create calculator tool: %w", err)
	}
	if err := registry.Register(calcTool); err != nil {
		return fmt.Errorf("failed to register calculator tool: %w", err)
	}
	log.Println("Tool calculator registered")

	return nil
}

// runAPIMode starts the API server and handles graceful shutdown.
func runAPIMode(app *Application, sigChan chan os.Signal, cancel context.CancelFunc) {
	log.Printf("Starting API server on port %d...", app.config.Server.Port)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := app.apiServer.Start(); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	case err := <-errChan:
		log.Printf("Server error: %v", err)
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := app.apiServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}

	log.Println("Server stopped")
}

// runCLIMode executes a single task and outputs the result.
func runCLIMode(ctx context.Context, app *Application, taskDescription string) {
	log.Printf("Running CLI mode with task: %s", taskDescription)

	// Create workflow configuration
	workflowCfg := &workflow.Config{
		Type:    workflow.ParseType(app.config.Workflow.DefaultType),
		Agents:  app.registry.List(),
		Task:    taskDescription,
		MaxIter: 3,
	}

	// Create workflow engine with registry and memory service
	engine, err := workflow.NewEngine(
		workflowCfg,
		workflow.WithRegistry(app.registry),
		workflow.WithSessionService(app.memoryService),
	)
	if err != nil {
		log.Fatalf("Failed to create workflow engine: %v", err)
	}

	// Execute workflow
	log.Println("Executing workflow...")
	startTime := time.Now()
	result, err := engine.Execute(ctx)
	executionTime := time.Since(startTime)

	if err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}

	// Print results
	fmt.Println("\n=== Workflow Results ===")
	statusStr := "FAILED"
	if result.Success {
		statusStr = "SUCCESS"
	}
	fmt.Printf("Status: %s\n", statusStr)
	fmt.Printf("Type: %s\n", result.Type)
	fmt.Printf("Iterations: %d\n", result.Iterations)
	fmt.Printf("Execution Time: %v\n", executionTime)
	fmt.Println("\n=== Agent Results ===")
	for _, ar := range result.AgentResults {
		fmt.Printf("\n[%s]:\n", ar.AgentName)
		if ar.Error != nil {
			fmt.Printf("  Error: %v\n", ar.Error)
		} else {
			fmt.Printf("  Output: %s\n", ar.Output)
		}
	}
	fmt.Println("\n=== Final Output ===")
	fmt.Println(result.Output)
}