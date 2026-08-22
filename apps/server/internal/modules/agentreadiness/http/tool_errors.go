package agentreadinesshttp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type toolInputError struct {
	message string
}

func (e *toolInputError) Error() string {
	return e.message
}

func invalidToolInput(message string) error {
	return &toolInputError{message: message}
}

func invalidToolInputf(format string, args ...any) error {
	return invalidToolInput(fmt.Sprintf(format, args...))
}

func addSafeTool[In any](h *Handler, server *mcp.Server, definition *mcp.Tool, handler mcp.ToolHandlerFor[In, any]) {
	mcp.AddTool(server, definition, safeToolHandler(h, definition.Name, handler))
}

func safeToolHandler[In any](h *Handler, name string, handler mcp.ToolHandlerFor[In, any]) mcp.ToolHandlerFor[In, any] {
	return func(ctx context.Context, request *mcp.CallToolRequest, input In) (*mcp.CallToolResult, any, error) {
		result, output, err := handler(ctx, request, input)
		if err == nil {
			return result, output, nil
		}

		var inputErr *toolInputError
		if errors.As(err, &inputErr) {
			h.logToolInputError(ctx, name, inputErr)
			return nil, nil, inputErr
		}

		reference := uuid.NewString()
		h.logToolFailure(ctx, name, reference, err)
		return nil, nil, fmt.Errorf("FortyOne couldn't complete this request right now. Please try again. Reference: %s", reference)
	}
}

func (h *Handler) logToolInputError(ctx context.Context, toolName string, err error) {
	if h.cfg.Log == nil {
		return
	}
	h.cfg.Log.Warn(ctx, "MCP tool request rejected", "tool", toolName, "error", err)
}

func (h *Handler) logToolFailure(ctx context.Context, toolName, reference string, err error) {
	if h.cfg.Log == nil {
		return
	}
	h.cfg.Log.Error(ctx, "MCP tool failed", "tool", toolName, "reference", reference, "error", err)
}
