// Package agents provides specialized agent implementations.
package agents

import (
	"errors"

	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	adkmodel "google.golang.org/adk/model"
	"google.golang.org/adk/tool"
)

// Analyst agent configuration constants
const (
	analystName        = "analyst"
	analystDescription = "数据分析与洞察提取专员"
)

// Analyst instruction for guiding the agent's behavior
const analystInstruction = `你是一个专业的数据分析师，负责分析收集到的信息并提取关键洞察。

你的主要职责包括：
1. 深度分析收集到的信息和数据
2. 识别关键模式、趋势和异常
3. 提取有价值的洞察和结论
4. 将复杂信息转化为易于理解的分析报告

工作原则：
- 基于事实和数据进行客观分析
- 识别信息中的关键模式和关联
- 提供可操作的洞察和建议
- 清晰解释分析方法和结论

输出格式建议：
## 核心洞察
列出最重要的发现和洞察（3-5条）

## 详细分析
提供深入的分析内容：
- 数据趋势和模式
- 关键发现和证据
- 异常点或特殊情况

## 数据解读
解释数据的含义和影响

## 结论与建议
基于分析得出的结论和后续行动建议

请基于提供的信息进行深度分析。`

// NewAnalyst creates a new analyst agent using adk-go's llmagent.
// The analyst agent is specialized for data analysis and insights extraction.
//
// Parameters:
//   - model: The LLM model to use for the agent. Must not be nil.
//
// Returns:
//   - agent.Agent: The created analyst agent
//   - error: An error if the model is nil or agent creation fails
func NewAnalyst(model adkmodel.LLM) (adkagent.Agent, error) {
	if model == nil {
		return nil, errors.New("model cannot be nil")
	}

	cfg := llmagent.Config{
		Name:         analystName,
		Description:  analystDescription,
		Model:        model,
		Instruction:  analystInstruction,
		Tools:        []tool.Tool{}, // Tools will be added in Phase 4
		SubAgents:    []adkagent.Agent{}, // Analyst has no sub-agents
	}

	agent, err := llmagent.New(cfg)
	if err != nil {
		return nil, err
	}

	return agent, nil
}