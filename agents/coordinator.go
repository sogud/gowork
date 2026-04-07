// Package agents provides specialized agent implementations.
package agents

import (
	"errors"

	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	adkmodel "google.golang.org/adk/model"
	"google.golang.org/adk/tool"
)

// Coordinator agent configuration constants
const (
	coordinatorName        = "coordinator"
	coordinatorDescription = "智能任务路由协调器"
)

// Coordinator instruction for guiding the agent's behavior
const coordinatorInstruction = `你是一个智能任务协调器，负责分析任务需求并协调多个专业Agent协同工作。

你的主要职责包括：
1. 分析用户任务的具体需求
2. 确定需要哪些专业Agent参与
3. 规划Agent的执行顺序和协作方式
4. 协调各Agent之间的信息流转

可用的专业Agent：
- researcher（研究专员）：负责信息收集和整理
- analyst（分析专员）：负责数据分析和洞察提取
- writer（写作专员）：负责内容生成和结构化输出
- reviewer（审查专员）：负责质量审查和改进建议

工作原则：
- 根据任务需求选择合适的Agent组合
- 优化执行顺序以提高效率
- 确保信息在各Agent间正确流转
- 避免不必要的重复工作

工作流程类型：
1. 研究型任务：researcher → analyst → writer
2. 分析型任务：analyst → writer → reviewer
3. 创作型任务：writer → reviewer
4. 审查型任务：reviewer
5. 综合型任务：researcher → analyst → writer → reviewer

输出格式建议：
## 任务分析
简要描述任务的核心需求

## 执行计划
### 工作流类型
[研究型/分析型/创作型/审查型/综合型]

### Agent调用顺序
1. Agent名称 - 职责 - 输入来源
2. Agent名称 - 职责 - 输入来源
...

### 信息流转
描述各Agent之间如何传递信息

## 预期输出
描述最终的预期结果

## 注意事项
列出需要特别关注的问题或约束

请根据用户任务需求制定合理的执行计划。`

// NewCoordinator creates a new coordinator agent using adk-go's llmagent.
// The coordinator agent is specialized for task routing and coordination.
//
// Parameters:
//   - model: The LLM model to use for the agent. Must not be nil.
//   - registry: The agent registry for dynamic agent access. Can be nil if not needed.
//
// Returns:
//   - agent.Agent: The created coordinator agent
//   - error: An error if the model is nil or agent creation fails
func NewCoordinator(model adkmodel.LLM, registry *Registry) (adkagent.Agent, error) {
	if model == nil {
		return nil, errors.New("model cannot be nil")
	}

	// Note: registry parameter is accepted for future use when dynamic agent
	// invocation is implemented. Currently, the coordinator provides routing
	// recommendations but does not directly invoke agents.

	cfg := llmagent.Config{
		Name:         coordinatorName,
		Description:  coordinatorDescription,
		Model:        model,
		Instruction:  coordinatorInstruction,
		Tools:        []tool.Tool{}, // Tools will be added in Phase 4
		SubAgents:    []adkagent.Agent{}, // Coordinator has no sub-agents initially
	}

	agent, err := llmagent.New(cfg)
	if err != nil {
		return nil, err
	}

	return agent, nil
}