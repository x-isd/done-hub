package tools

import (
	"context"
	"done-hub/mcp/tools/available_model"
	"done-hub/mcp/tools/calculator"
	"done-hub/mcp/tools/current_time"
	"done-hub/mcp/tools/dashboard"
	"github.com/ThinkInAIXYZ/go-mcp/protocol"
)

type McpTool interface {
	GetTool() *protocol.Tool
	HandleRequest(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error)
}

var McpTools = make(map[string]McpTool)

func init() {
	McpTools[calculator.NAME] = &calculator.Calculator{}
	McpTools[available_model.NAME] = &available_model.AvailableModel{}
	McpTools[dashboard.NAME] = &dashboard.Dashboard{}
	McpTools[current_time.NAME] = &current_time.CurrentTime{}
}
