# Cowork Agent - 多 Agent 协作框架设计文档

**日期**: 2026-04-05
**项目**: adk-go 多 Agent 协作框架 Demo
**模型**: Gemma 4 (本地 Ollama, 模型名称: `gemma4`)

---

## 1. 项目概述

### 1.1 目标

构建一个完整的、模块化的多 Agent 协作框架，展示 adk-go 的核心特性：
- 多种 workflow 编排模式（顺序、并行、循环、动态路由）
- 跨 Agent 状态共享与记忆管理
- 可扩展的工具生态系统
- RESTful API + CLI 双入口
- 本地模型集成（Ollama Gemma 4）

### 1.2 核心特性

1. **自定义 Ollama 模型适配器** - 实现 model.LLM 接口，对接本地 Gemma 4 (通过 `ollama pull gemma4`)
2. **专业领域 Agents** - Researcher、Analyst、Writer、Reviewer + Coordinator
3. **智能协调器** - 动态路由任务，根据请求类型分配给合适的 agents
4. **多模式 Workflow** - Sequential/Parallel/Loop/Dynamic 四种编排模式
5. **状态共享机制** - Memory Service 支持跨 agent 状态传递
6. **工具扩展系统** - 可注册、可分配的工具集
7. **双入口服务** - API server + CLI launcher

---

## 2. 系统架构

### 2.1 模块化分层架构

```
cowork-agent/
├── model/          # 模型适配层
│   ├── ollama.go         # Ollama Gemma 4 适配器
│   └── config.go         # 模型配置
│
├── agents/         # Agent 定义层
│   ├── researcher.go     # 信息收集 agent
│   ├── analyst.go        # 数据分析 agent
│   ├── writer.go         # 内容生成 agent
│   ├── reviewer.go       # 质量审查 agent
│   ├── coordinator.go    # 智能路由协调器
│   └── registry.go       # Agent 注册管理
│
├── workflow/       # 协作编排层
│   ├── engine.go         # Workflow 执行引擎
│   ├── sequential.go     # 顺序编排（扩展原生）
│   ├── parallel.go       # 并行编排（扩展原生）
│   ├── loop.go           # 循环编排（扩展原生）
│   ├── dynamic.go        # 动态路由编排（新增）
│   └── config.go         # Workflow 配置
│
├── memory/         # 状态管理层
│   ├── service.go        # Memory service 接口（扩展）
│   ├── inmemory.go       # 内存存储（继承原生）
│   ├── persistent.go     # 持久化存储（可选）
│   └── state.go          # 状态管理器（新增）
│
├── tools/          # 工具集成层
│   ├── base.go           # 工具基础接口与注册
│   ├── search.go         # Web 搜索工具
│   ├── file.go           # 文件操作工具
│   ├── calculator.go     # 计算工具
│   ├── code.go           # 代码分析工具
│   └── agent_tools.go    # Agent 调用工具
│
├── server/         # 运行服务层
│   ├── api.go            # RESTful API 服务
│   ├── launcher.go       # CLI 启动器（扩展原生）
│   ├── middleware.go     # 中间件（日志、认证、限流）
│   └ health.go           # 健康检查与监控
│   └── config.go         # 服务配置
│
├── config.yaml     # 主配置文件
├── main.go         # 入口文件
└── README.md       # 使用文档
```

### 2.2 系统协作流程

```
用户请求 → [CLI/API] → Server Layer
                           ↓
                    Workflow Engine
                           ↓
                  选择编排模式 (Sequential/Parallel/Dynamic)
                           ↓
                    Coordinator Agent (动态路由)
                           ↓
                  分配任务给专业 Agents
                           ↓
            [Researcher → Analyst → Writer → Reviewer]
                           ↓
                    Memory & State 共享
                           ↓
                    Tools Layer 执行具体操作
                           ↓
                    Ollama Model (Gemma 4)
                           ↓
                    结果汇总 → 返回用户
```

---

## 3. 核心模块详细设计

### 3.1 Model 层 - Ollama Gemma 4 适配器

**职责**: 实现 adk-go 的 `model.LLM` 接口，连接本地 Ollama 服务

**核心接口**:
```go
type ollamaModel struct {
    client    *http.Client
    baseURL   string  // "http://localhost:11434"
    modelName string  // "gemma4"
}

// 实现 model.LLM 接口
func NewOllamaModel(baseURL, modelName string) (model.LLM, error)
func (m *ollamaModel) Name() string
func (m *ollamaModel) GenerateContent(ctx, req, stream) iter.Seq2[*LLMResponse, error]
```

**技术要点**:
- HTTP 客户端连接本地 Ollama API
- 请求格式转换: `genai.Content` → Ollama chat format
- 流式响应支持: SSE (Server-Sent Events) parsing
- 健康检查: Ollama 服务可用性 + 模型加载状态
- 超时控制: 可配置请求超时时间

**配置示例**:
```yaml
ollama:
  base_url: "http://localhost:11434"
  model: "gemma4"
  timeout: 60s
  max_retries: 3
```

### 3.2 Agents 层 - 专业 Agents + 协调器

**职责**: 定义专业领域 agents 和智能协调器

**Agent 角色设计**:

| Agent | 职责 | 工具 | Instruction |
|-------|------|------|-------------|
| Researcher | 信息收集整理 | WebSearch, FileReader | 收集准确全面的信息 |
| Analyst | 数据分析洞察 | Calculator, CodeAnalyzer | 提取关键洞察和数据模式 |
| Writer | 内容生成 | FileWriter, Template | 生成结构化高质量内容 |
| Reviewer | 质量审查 | CodeAnalyzer, QualityCheck | 审查质量提出改进建议 |
| Coordinator | 任务路由 | AgentCall, WorkflowQuery | 分析请求分配任务 |

**Agent 配置示例**:
```go
researcherAgent := llmagent.New(llmagent.Config{
    Name:        "researcher",
    Model:       ollamaModel,
    Description: "负责信息收集和整理",
    Instruction: "你的任务是收集准确、全面的信息...",
    Tools:       []tool.Tool{webSearchTool, fileReaderTool},
})
```

**Coordinator 动态路由逻辑**:
1. 分析任务关键词和语义
2. 确定需要哪些专业 agents
3. 决定执行顺序（依赖关系分析）
4. 构建动态 workflow 配置
5. 监控执行进度并调整策略

### 3.3 Workflow 层 - 协作编排引擎

**职责**: 管理 agents 的执行流程和协作模式

**Workflow Engine 核心**:
```go
type WorkflowEngine struct {
    memoryService memory.Service
    sessionStore  session.Service
    agentRegistry map[string]agent.Agent
}

func (e *WorkflowEngine) Execute(ctx, workflowConfig) (*WorkflowResult, error)
```

**编排模式**:

1. **Sequential Workflow** - 顺序执行
```go
// 适用场景: 有明确依赖关系的任务流
// 例如: 研究 → 分析 → 写作 → 审查
sequentialWorkflow := sequentialagent.New(sequentialagent.Config{
    AgentConfig: agent.Config{
        SubAgents: []agent.Agent{researcher, analyst, writer, reviewer},
    },
    StateSharing: true,  // 启用状态传递
})
```

2. **Parallel Workflow** - 并行执行
```go
// 适用场景: 多个独立分析任务
// 例如: 安全检查 + 性能分析 + 风格审查
parallelWorkflow := parallelagent.New(parallelagent.Config{
    AgentConfig: agent.Config{
        SubAgents: []agent.Agent{securityReviewer, performanceAnalyzer, styleChecker},
    },
    Timeout: 30*time.Second,
    FailStrategy: "partial",  // 部分失败继续执行
})
```

3. **Loop Workflow** - 循环执行
```go
// 适用场景: 需要迭代优化的任务
// 例如: 生成内容 → 审查 → 修改 → 再审查
loopWorkflow := loopagent.New(loopagent.Config{
    AgentConfig: agent.Config{
        SubAgents: []agent.Agent{writer, reviewer},
    },
    MaxIterations: 5,
    TerminationCondition: "quality_score >= 80",
})
```

4. **Dynamic Workflow** - 动态路由（创新特性）
```go
// 适用场景: 复杂任务需要智能编排
// Coordinator 实时决定执行策略
dynamicWorkflow := NewDynamicWorkflow(DynamicConfig{
    Coordinator:      coordinatorAgent,
    AvailableAgents:  agentRegistry,
    DecisionModel:    ollamaModel,
    MemoryService:    memoryService,
})
```

### 3.4 Memory 层 - 状态管理与共享

**职责**: 跨 agents 的状态共享和历史记忆管理

**Memory Service 扩展接口**:
```go
type Service interface {
    // 基础功能（继承 adk-go）
    AddSessionToMemory(ctx, session) error
    SearchMemory(ctx, req) (*SearchResponse, error)

    // 新增功能 - 状态共享
    GetSharedState(ctx, appName, userID, key) (*StateEntry, error)
    SetSharedState(ctx, appName, userID, key, value) error
    UpdateAgentState(ctx, agentName, sessionID, state) error

    // 新增功能 - Workflow 状态
    GetWorkflowState(ctx, workflowID) (*WorkflowState, error)
    UpdateWorkflowProgress(ctx, workflowID, agentName, status) error
}
```

**StateManager 核心机制**:
```go
type StateManager struct {
    memoryService Service
    mu            sync.RWMutex
}

// Agents 间共享状态
func (s *StateManager) ShareBetweenAgents(ctx, fromAgent, toAgent, key, value) error

// Workflow 执行状态追踪
func (s *StateManager) TrackWorkflowProgress(ctx, workflowID, agents []string) error
```

**状态共享流程**:
```
Researcher 执行 → 存储研究结果到共享状态
                ↓
Analyst 读取共享状态 → 进行分析 → 存储分析洞察
                ↓
Writer 读取分析结果 → 生成内容 → 存储草稿
                ↓
Reviewer 访问全部历史 → 审查质量 → 存储审查意见
```

**持久化选项**（可选扩展）:
```yaml
memory:
  type: "persistent"
  backend: "redis"
  connection: "localhost:6379"
  ttl: 3600
  encryption: true
```

### 3.5 Tools 层 - 可扩展工具集

**职责**: 为 agents 提供丰富的工具能力

**工具实现策略**:
- Demo 版本：使用模拟工具（WebSearchTool 返回预定义数据）
- 生产版本：可扩展对接真实 API（Google Search、文件系统等）

**工具注册机制**:
```go
type ToolRegistry struct {
    tools map[string]tool.Tool
}

func (r *ToolRegistry) Register(name string, t tool.Tool) error
func (r *ToolRegistry) GetAvailableTools() []tool.Tool
```

**核心工具定义**:

| 工具 | 功能 | 参数 | 返回 |
|------|------|------|------|
| WebSearchTool | 模拟搜索（demo） | query | 预定义搜索结果 |
| FileReaderTool | 读取文件 | filepath | 文件内容 |
| FileWriterTool | 写入文件 | filepath, content | 成功/失败 |
| CalculatorTool | 数学计算 | expression | 计算结果 |
| CodeAnalyzerTool | 代码分析 | code | 分析报告 |
| AgentCallTool | Agent 调用 | agentName, input | Agent 响应 |

**Agent 工具分配策略**:
```go
var agentTools = map[string][]tool.Tool{
    "researcher":  {WebSearchTool{}, FileReaderTool{}},
    "analyst":     {CalculatorTool{}, CodeAnalyzerTool{}},
    "writer":      {FileWriterTool{}, TemplateTool{}},
    "reviewer":    {CodeAnalyzerTool{}, QualityCheckTool{}},
    "coordinator": {AgentCallTool{}, WorkflowQueryTool{}},
}
```

**AgentCallTool 核心**（用于 Coordinator）:
```go
type AgentCallTool struct {
    agentRegistry map[string]agent.Agent
}

func (t *AgentCallTool) CallAgent(ctx, agentName, input) (*AgentResponse, error) {
    // 动态调用其他 agent
    // 支持跨 agent 通信
}
```

### 3.6 Server 层 - 运行服务与接口

**职责**: 提供 API 和 CLI 双入口，管理服务生命周期

**RESTful API 设计**:

| 端点 | 方法 | 功能 | 参数 |
|------|------|------|------|
| `/api/v1/workflow/execute` | POST | 执行 workflow | task, type, agents, config |
| `/api/v1/workflow/{id}/status` | GET | 查询状态 | workflow_id |
| `/api/v1/agent/{name}/invoke` | POST | 调用 agent | input |
| `/api/v1/agents` | GET | 获取 agents 列表 | - |
| `/api/v1/memory/search` | GET | 搜索记忆 | query |
| `/api/v1/memory/state` | POST | 更新状态 | key, value |
| `/api/v1/health` | GET | 健康检查 | - |

**API 请求/响应格式**:
```json
// Workflow 执行请求
{
  "task": "研究 Go 语言并发模式并生成技术报告",
  "type": "dynamic",
  "agents": [],
  "config": {}
}

// Workflow 执行响应
{
  "workflow_id": "wf-12345",
  "status": "completed",
  "results": {
    "researcher": "收集了 5 个并发模式案例...",
    "analyst": "识别出 3 种核心模式...",
    "writer": "生成报告草稿...",
    "reviewer": "审查完成"
  },
  "execution_time_ms": 45000,
  "shared_state": {
    "research_data": {...},
    "analysis_insights": {...}
  }
}
```

**CLI Launcher 命令**:
```bash
# 启动 API 服务
cowork-agent serve --port 8080 --config config.yaml

# 执行任务
cowork-agent run --task "研究 Go 并发模式"
cowork-agent run --workflow sequential --agents researcher,writer
cowork-agent run --workflow parallel --agents security,performance,style

# 状态查询
cowork-agent status --workflow-id wf-12345
cowork-agent list-agents

# 健康检查
cowork-agent health
```

**中间件功能**:
- **LoggingMiddleware**: 记录请求日志、执行时间、错误信息
- **RateLimitMiddleware**: 限流保护（可配置 RPS）
- **AuthMiddleware**: 认证验证（可选）

**健康检查机制**:
```go
type HealthChecker struct {
    ollamaClient   *http.Client
    memoryService  memory.Service
}

func (h *HealthChecker) Check(ctx) (*HealthStatus, error) {
    // 检查项:
    // 1. Ollama 服务连接
    // 2. Gemma 4 模型可用性
    // 3. Memory service 状态
    // 4. Agents 注册状态
    // 5. 系统 resource 使用率
}
```

---

## 4. 配置管理

### 4.1 主配置文件

```yaml
# config.yaml
server:
  port: 8080
  mode: "api"  # "api" or "cli"
  log_level: "info"

ollama:
  base_url: "http://localhost:11434"
  model: "gemma4"
  timeout: 60s
  max_retries: 3

agents:
  - name: "researcher"
    enabled: true
    model: "gemma4"
  - name: "analyst"
    enabled: true
    model: "gemma4"
  - name: "writer"
    enabled: true
    model: "gemma4"
  - name: "reviewer"
    enabled: true
    model: "gemma4"
  - name: "coordinator"
    enabled: true
    model: "gemma4"

workflow:
  default_type: "dynamic"
  timeout: 300s
  max_iterations: 5

memory:
  type: "inmemory"
  # 持久化配置（可选）
  persistent_backend: "redis"
  connection: "localhost:6379"
  ttl: 3600
  encryption: false

tools:
  web_search:
    enabled: true
    mode: "mock"  # demo 使用模拟数据，生产可切换为 "real"
    mock_data_dir: "./data/mock_search.json"
  file_operations:
    enabled: true
    allowed_paths: ["/tmp", "./data"]
```

---

## 5. 使用场景示例

### 5.1 场景 1：技术报告生成（Sequential Workflow）

**任务**: "研究 Go 语言并发模式并生成技术报告"

**执行流程**:
1. Coordinator 分析 → 决定使用 sequential workflow
2. Researcher 收集资料 → 存入共享状态
3. Analyst 分析模式 → 提取关键洞察
4. Writer 生成报告 → 结构化内容
5. Reviewer 审查质量 → 提出改进建议

**CLI 执行**:
```bash
cowork-agent run --task "研究 Go 语言并发模式并生成技术报告"
```

**API 执行**:
```bash
curl -X POST http://localhost:8080/api/v1/workflow/execute \
  -H "Content-Type: application/json" \
  -d '{"task": "研究 Go 语言并发模式并生成技术报告", "type": "dynamic"}'
```

### 5.2 场景 2：代码审查（Parallel Workflow）

**任务**: "全面审查这段代码的安全性、性能和风格"

**执行流程**:
1. Coordinator 分析 → 决定使用 parallel workflow
2. SecurityReviewer、PerformanceAnalyzer、StyleChecker 并行执行
3. 结果汇总 → 综合审查报告

**CLI 执行**:
```bash
cowork-agent run \
  --workflow parallel \
  --agents security-reviewer,performance-analyzer,style-checker \
  --task "审查这段代码"
```

### 5.3 场景 3：迭代优化（Loop Workflow）

**任务**: "生成一篇高质量的技术文章，不断优化直到质量评分≥80"

**执行流程**:
1. Writer 生成初稿 → 存储草稿
2. Reviewer 审查评分 → 如果<80，提出改进建议
3. Writer 根据建议修改 → 再次审查
4. 循环直到评分达标或达到最大迭代次数

**CLI 执行**:
```bash
cowork-agent run \
  --workflow loop \
  --agents writer,reviewer \
  --task "生成高质量技术文章" \
  --max-iterations 5
```

---

## 6. 技术实现要点

### 6.1 Ollama 模型适配

**请求格式转换**:
```go
// adk-go LLMRequest → Ollama API format
{
  "model": "gemma4",
  "messages": [
    {"role": "system", "content": instruction},
    {"role": "user", "content": user_query}
  ],
  "stream": true,
  "options": {
    "temperature": 0.7,
    "top_p": 0.9,
    "num_ctx": 8192  // Gemma 4 上下文窗口
  }
}
```

**响应格式转换**:
```go
// Ollama response → adk-go LLMResponse
{
  "content": &genai.Content{
    Parts: []*genai.Part{
      {Text: response_text}
    }
  },
  "turn_complete": true,
  "usage_metadata": {
    "prompt_token_count": 100,
    "candidates_token_count": 150
  }
}
```

### 6.2 状态共享实现

**共享状态存储结构**:
```go
type SharedState struct {
    WorkflowID   string
    Key          string
    Value        map[string]any
    FromAgent    string
    Timestamp    time.Time
    Version      int  // 支持版本控制
}
```

**状态传递机制**:
```go
// Agent 执行时注入共享状态
func (a *Agent) Run(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
    // 1. 从 memory service 读取前序 agent 的状态
    // 2. 将共享状态加入 context
    // 3. 执行 agent 逻辑
    // 4. 将新状态写入 memory service
}
```

### 6.3 动态路由决策

**Coordinator 决策逻辑**:
```go
func (c *Coordinator) DecideWorkflow(ctx, task) (*WorkflowPlan, error) {
    // 1. 使用 LLM 分析任务语义
    analysis := c.analyzeTask(ctx, task)

    // 2. 确定需要的 agents
    requiredAgents := c.identifyAgents(analysis)

    // 3. 分析依赖关系
    dependencies := c.analyzeDependencies(requiredAgents)

    // 4. 构建 workflow 配置
    plan := c.buildWorkflowPlan(requiredAgents, dependencies)

    return plan, nil
}
```

**决策输出示例**:
```json
{
  "workflow_type": "sequential",
  "agents": ["researcher", "analyst", "writer", "reviewer"],
  "execution_order": [
    {"agent": "researcher", "depends_on": []},
    {"agent": "analyst", "depends_on": ["researcher"]},
    {"agent": "writer", "depends_on": ["analyst"]},
    {"agent": "reviewer", "depends_on": ["writer"]}
  ],
  "state_sharing": true
}
```

---

## 7. 扩展性设计

### 7.1 新增 Agent

**步骤**:
1. 在 `agents/` 创建新 agent 文件
2. 配置 agent 的 instruction 和 tools
3. 在 `config.yaml` 注册新 agent
4. 更新 Coordinator 的路由逻辑

### 7.2 新增工具

**步骤**:
1. 在 `tools/` 实现新工具（遵循 tool.Tool 接口）
2. 在 ToolRegistry 注册
3. 分配给需要的 agents

### 7.3 新增 Workflow 模式

**步骤**:
1. 在 `workflow/` 实现新编排模式
2. 集成到 WorkflowEngine
3. 更新 API 和 CLI 支持新模式

### 7.4 持久化扩展

**支持的存储后端**:
- Redis（高性能，适合生产）
- PostgreSQL（关系型，适合复杂查询）
- File Storage（简单，适合单机）

---

## 8. 测试策略

### 8.1 单元测试

**覆盖模块**:
- Model 层: Ollama 适配器请求/响应转换
- Agents 层: 各 agent 的 instruction 解析
- Workflow 层: 编排逻辑正确性
- Memory 层: 状态存储和检索
- Tools 层: 工具执行正确性

### 8.2 集成测试

**测试场景**:
- Sequential workflow 完整流程
- Parallel workflow 并发执行
- Dynamic workflow 路由决策
- 状态共享机制验证
- API 端点集成测试

### 8.3 E2E 测试

**测试流程**:
1. 启动完整系统（Ollama + API server）
2. 执行真实任务
3. 验证输出结果质量
4. 检查状态共享正确性

---

## 9. 部署建议

### 9.1 本地开发环境

**前置条件**:
- Go 1.21+
- Ollama 已安装并运行
- Gemma 4 模型已下载 (`ollama pull gemma4`)

**启动步骤**:
```bash
# 1. 安装 Ollama
curl -fsSL https://ollama.ai/install.sh | sh

# 2. 下载 Gemma 4 模型
ollama pull gemma4

# 3. 启动 Ollama 服务
ollama serve

# 4. 准备模拟数据（WebSearchTool）
mkdir -p data
# 创建 mock_search.json（预定义搜索结果）

# 5. 运行 cowork-agent
go run main.go serve --config config.yaml
```

### 9.2 生产部署

**容器化部署**:
```dockerfile
FROM golang:1.21-alpine
# ... 构建 + Ollama 集成
```

**云原生部署**:
- Google Cloud Run
- Kubernetes Deployment
- 支持 Redis 作为 memory backend

---

## 10. 总结

本设计文档定义了一个完整的、模块化的多 Agent 协作框架，具备以下特点：

**核心优势**:
1. **模块化架构** - 清晰的分层设计，易于理解和扩展
2. **生产级特性** - 包含动态路由、状态共享、健康检查等完整功能
3. **灵活编排** - 支持多种 workflow 模式，适应不同场景
4. **本地模型支持** - Ollama Gemma 4 集成，降低 API 成本
5. **双入口服务** - API + CLI，满足不同使用场景

**创新特性**:
- Dynamic Workflow（智能路由编排）
- 跨 Agent 状态共享机制
- AgentCallTool（支持 agent 间调用）
- 完整的监控和健康检查

**扩展性**:
- 易于新增 agents、tools、workflow 模式
- 支持多种持久化后端
- 模块化设计便于维护

下一步：创建详细的实现计划，逐步构建各模块。