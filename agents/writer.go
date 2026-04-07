// Package agents provides specialized agent implementations.
package agents

import (
	"errors"

	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	adkmodel "google.golang.org/adk/model"
	"google.golang.org/adk/tool"
)

// Writer agent configuration constants
const (
	writerName        = "writer"
	writerDescription = "内容生成与结构化输出专员"
)

// Writer instruction for guiding the agent's behavior
const writerInstruction = `你是一个专业的内容创作者，负责基于研究和分析生成高质量内容。

你的主要职责包括：
1. 基于研究和分析结果生成结构化内容
2. 组织信息并以清晰、逻辑性强的方式呈现
3. 确保内容的准确性和可读性
4. 根据不同需求调整内容风格和格式

工作原则：
- 内容准确、逻辑清晰
- 使用结构化的格式组织信息
- 语言简洁明了，避免冗余
- 根据受众调整表达方式

输出格式建议：
## 标题
清晰描述内容主题

## 概述
简要介绍主要内容（2-3句话）

## 主体内容
### 第一部分
详细内容...

### 第二部分
详细内容...

### 第三部分
详细内容...

## 关键要点
- 要点1
- 要点2
- 要点3

## 总结
简明扼要的总结

请根据研究和分析结果生成高质量的内容。`

// NewWriter creates a new writer agent using adk-go's llmagent.
// The writer agent is specialized for content generation and structured output.
//
// Parameters:
//   - model: The LLM model to use for the agent. Must not be nil.
//
// Returns:
//   - agent.Agent: The created writer agent
//   - error: An error if the model is nil or agent creation fails
func NewWriter(model adkmodel.LLM) (adkagent.Agent, error) {
	if model == nil {
		return nil, errors.New("model cannot be nil")
	}

	cfg := llmagent.Config{
		Name:         writerName,
		Description:  writerDescription,
		Model:        model,
		Instruction:  writerInstruction,
		Tools:        []tool.Tool{}, // Tools will be added in Phase 4
		SubAgents:    []adkagent.Agent{}, // Writer has no sub-agents
	}

	agent, err := llmagent.New(cfg)
	if err != nil {
		return nil, err
	}

	return agent, nil
}