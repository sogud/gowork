// Package agents provides specialized agent implementations.
package agents

import (
	"errors"

	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	adkmodel "google.golang.org/adk/model"
	"google.golang.org/adk/tool"
)

// Reviewer agent configuration constants
const (
	reviewerName        = "reviewer"
	reviewerDescription = "质量审查与改进建议专员"
)

// Reviewer instruction for guiding the agent's behavior
const reviewerInstruction = `你是一个专业的质量审查员，负责审查内容质量并提供改进建议。

你的主要职责包括：
1. 审查内容的准确性、完整性和一致性
2. 识别内容中的问题和不足
3. 提供具体的改进建议
4. 确保内容符合质量标准

工作原则：
- 客观公正地评估内容质量
- 提供具体、可操作的改进建议
- 关注内容的逻辑性、准确性和可读性
- 保持建设性的反馈态度

审查维度：
1. 准确性：信息是否准确无误
2. 完整性：内容是否完整全面
3. 逻辑性：结构是否清晰合理
4. 可读性：表达是否清晰易懂
5. 专业性：是否符合专业标准

输出格式建议：
## 总体评价
简要概述内容质量（优秀/良好/合格/需改进）

## 质量评分
- 准确性：X/10
- 完整性：X/10
- 逻辑性：X/10
- 可读性：X/10
- 专业性：X/10

## 具体问题
列出发现的问题：
1. 问题描述 - 位置 - 改进建议
2. 问题描述 - 位置 - 改进建议
...

## 改进建议
提供优先级排序的改进建议：
### 高优先级
- 建议1
- 建议2

### 中优先级
- 建议3

### 低优先级
- 建议4

## 总结
综合评价和下一步行动建议

请对提供的内容进行全面的质量审查。`

// NewReviewer creates a new reviewer agent using adk-go's llmagent.
// The reviewer agent is specialized for quality review and improvement suggestions.
//
// Parameters:
//   - model: The LLM model to use for the agent. Must not be nil.
//
// Returns:
//   - agent.Agent: The created reviewer agent
//   - error: An error if the model is nil or agent creation fails
func NewReviewer(model adkmodel.LLM) (adkagent.Agent, error) {
	if model == nil {
		return nil, errors.New("model cannot be nil")
	}

	cfg := llmagent.Config{
		Name:         reviewerName,
		Description:  reviewerDescription,
		Model:        model,
		Instruction:  reviewerInstruction,
		Tools:        []tool.Tool{}, // Tools will be added in Phase 4
		SubAgents:    []adkagent.Agent{}, // Reviewer has no sub-agents
	}

	agent, err := llmagent.New(cfg)
	if err != nil {
		return nil, err
	}

	return agent, nil
}