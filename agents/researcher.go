// Package agents provides specialized agent implementations.
package agents

import (
	"errors"

	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	adkmodel "google.golang.org/adk/model"
	"google.golang.org/adk/tool"
)

// Researcher agent configuration constants
const (
	researcherName        = "researcher"
	researcherDescription = "信息收集与整理专员"
)

// Researcher instruction for guiding the agent's behavior
const researcherInstruction = `你是一个专业的研究助手，负责收集和整理信息。

你的主要职责包括：
1. 收集与用户问题相关的信息
2. 整理和总结搜索结果
3. 提供结构化的研究报告
4. 识别信息缺口并提出后续搜索建议

工作原则：
- 保持客观和准确性
- 引用可靠的信息来源
- 使用清晰的结构化格式呈现结果
- 在信息不确定时明确说明

输出格式建议：
## 研究摘要
简要总结主要发现

## 详细信息
列出关键信息和数据

## 信息来源
记录信息来源和引用

## 后续建议
如有必要，提出进一步研究方向

请根据用户需求进行信息收集和整理工作。`

// NewResearcher creates a new researcher agent using adk-go's llmagent.
// The researcher agent is specialized for information gathering and organization.
//
// Parameters:
//   - model: The LLM model to use for the agent. Must not be nil.
//
// Returns:
//   - agent.Agent: The created researcher agent
//   - error: An error if the model is nil or agent creation fails
func NewResearcher(model adkmodel.LLM) (adkagent.Agent, error) {
	if model == nil {
		return nil, errors.New("model cannot be nil")
	}

	cfg := llmagent.Config{
		Name:         researcherName,
		Description:  researcherDescription,
		Model:        model,
		Instruction:  researcherInstruction,
		Tools:        []tool.Tool{}, // Tools will be added in Phase 4
		SubAgents:    []adkagent.Agent{}, // Researcher has no sub-agents
	}

	agent, err := llmagent.New(cfg)
	if err != nil {
		return nil, err
	}

	return agent, nil
}