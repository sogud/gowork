// Package tools provides tool implementations for the agent system.
package tools

import (
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// CalculatorInput represents the input for the calculator tool.
type CalculatorInput struct {
	Expression string `json:"expression" description:"Mathematical expression to evaluate (e.g., '2+3*4')"`
}

// CalculatorOutput represents the output from the calculator tool.
type CalculatorOutput struct {
	Result     float64 `json:"result" description:"The result of the calculation"`
	Expression string  `json:"expression" description:"The original expression"`
}

// NewCalculatorTool creates a calculator tool for basic arithmetic operations.
// Supports +, -, *, / and parentheses.
func NewCalculatorTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input CalculatorInput) (CalculatorOutput, error) {
		result, err := evaluateExpression(input.Expression)
		if err != nil {
			return CalculatorOutput{}, err
		}

		return CalculatorOutput{
			Result:     result,
			Expression: input.Expression,
		}, nil
	}

	return functiontool.New(functiontool.Config{
		Name:        "calculator",
		Description: "Evaluate basic mathematical expressions with +, -, *, / and parentheses",
	}, handler)
}

// evaluateExpression evaluates a simple mathematical expression.
// This is a basic implementation supporting +, -, *, / and parentheses.
func evaluateExpression(expr string) (float64, error) {
	// Remove spaces
	expr = strings.ReplaceAll(expr, " ", "")

	if expr == "" {
		return 0, fmt.Errorf("empty expression")
	}

	// Use a simple recursive descent parser with position tracking
	parser := &exprParser{expr: expr}
	result, err := parser.parseExpression()
	if err != nil {
		return 0, err
	}

	if parser.pos < len(parser.expr) {
		return 0, fmt.Errorf("unexpected character at position %d: '%c'", parser.pos, parser.expr[parser.pos])
	}

	return result, nil
}

// exprParser is a simple expression parser.
type exprParser struct {
	expr string
	pos  int
}

// parseExpression parses: expression = term (('+' | '-') term)*
func (p *exprParser) parseExpression() (float64, error) {
	result, err := p.parseTerm()
	if err != nil {
		return 0, err
	}

	for p.pos < len(p.expr) {
		op := p.expr[p.pos]
		if op != '+' && op != '-' {
			break
		}
		p.pos++

		right, err := p.parseTerm()
		if err != nil {
			return 0, err
		}

		if op == '+' {
			result += right
		} else {
			result -= right
		}
	}

	return result, nil
}

// parseTerm parses: term = factor ((' * ' | '/') factor)*
func (p *exprParser) parseTerm() (float64, error) {
	result, err := p.parseFactor()
	if err != nil {
		return 0, err
	}

	for p.pos < len(p.expr) {
		op := p.expr[p.pos]
		if op != '*' && op != '/' {
			break
		}
		p.pos++

		right, err := p.parseFactor()
		if err != nil {
			return 0, err
		}

		if op == '*' {
			result *= right
		} else {
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			result /= right
		}
	}

	return result, nil
}

// parseFactor parses: factor = number | '(' expression ')'
func (p *exprParser) parseFactor() (float64, error) {
	if p.pos >= len(p.expr) {
		return 0, fmt.Errorf("unexpected end of expression")
	}

	// Check for parentheses
	if p.expr[p.pos] == '(' {
		p.pos++ // skip '('
		result, err := p.parseExpression()
		if err != nil {
			return 0, err
		}
		if p.pos >= len(p.expr) || p.expr[p.pos] != ')' {
			return 0, fmt.Errorf("missing closing parenthesis")
		}
		p.pos++ // skip ')'
		return result, nil
	}

	// Parse number (including negative numbers)
	start := p.pos

	// Handle negative sign
	if p.expr[p.pos] == '-' {
		p.pos++
	}

	// Parse digits and decimal point
	for p.pos < len(p.expr) && (p.expr[p.pos] >= '0' && p.expr[p.pos] <= '9' || p.expr[p.pos] == '.') {
		p.pos++
	}

	if start == p.pos {
		return 0, fmt.Errorf("expected number at position %d", start)
	}

	numStr := p.expr[start:p.pos]
	result, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number '%s': %w", numStr, err)
	}

	return result, nil
}