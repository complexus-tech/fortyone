package agentreadinesshttp

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"mime"
	"net/http"
)

const protocolVersion = "2025-06-18"

const (
	productOverviewURI    = "https://www.fortyone.app/llms.txt"
	developerResourcesURI = "https://www.fortyone.app/developers.md"
)

//go:embed openapi.json
var openAPIDescription []byte

type Handler struct{}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id"`
	Result  any       `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

func New() *Handler { return &Handler{} }

func (h *Handler) OpenAPI(_ context.Context, w http.ResponseWriter, _ *http.Request) error {
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(openAPIDescription)
	return err
}

func (h *Handler) MCPGet(_ context.Context, w http.ResponseWriter, _ *http.Request) error {
	w.Header().Set("Allow", "POST")
	return writeRPC(w, http.StatusMethodNotAllowed, rpcResponse{
		JSONRPC: "2.0",
		ID:      nil,
		Error:   &rpcError{Code: -32600, Message: "This stateless MCP server accepts messages with POST; server-initiated SSE is not enabled."},
	})
}

func (h *Handler) MCPPost(_ context.Context, w http.ResponseWriter, r *http.Request) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return writeRPC(w, http.StatusUnsupportedMediaType, rpcResponse{
			JSONRPC: "2.0",
			ID:      nil,
			Error:   &rpcError{Code: -32600, Message: "Content-Type must be application/json."},
		})
	}

	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	var request rpcRequest
	if err := decoder.Decode(&request); err != nil {
		return writeRPC(w, http.StatusBadRequest, rpcResponse{
			JSONRPC: "2.0",
			ID:      nil,
			Error:   &rpcError{Code: -32700, Message: "Invalid JSON-RPC request."},
		})
	}

	if request.JSONRPC != "2.0" || request.Method == "" {
		return writeRPC(w, http.StatusBadRequest, rpcResponse{
			JSONRPC: "2.0",
			ID:      decodeID(request.ID),
			Error:   &rpcError{Code: -32600, Message: "jsonrpc must be 2.0 and method is required."},
		})
	}

	if len(request.ID) == 0 || string(request.ID) == "null" {
		w.WriteHeader(http.StatusAccepted)
		return nil
	}

	response := rpcResponse{JSONRPC: "2.0", ID: decodeID(request.ID)}
	switch request.Method {
	case "initialize":
		response.Result = map[string]any{
			"protocolVersion": protocolVersion,
			"capabilities": map[string]any{
				"resources": map[string]any{},
				"tools":     map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":        "app.fortyone/public",
				"title":       "FortyOne",
				"version":     "1.0.0",
				"description": "Public FortyOne product and developer resources.",
			},
			"instructions": "Use these public resources to understand FortyOne. Do not infer access to private workspace data.",
		}
	case "ping":
		response.Result = map[string]any{}
	case "resources/list":
		response.Result = map[string]any{"resources": resources()}
	case "resources/read":
		result, err := readResource(request.Params)
		if err != nil {
			response.Error = &rpcError{Code: -32602, Message: err.Error()}
		} else {
			response.Result = result
		}
	case "tools/list":
		response.Result = map[string]any{"tools": tools()}
	case "tools/call":
		result, err := callTool(request.Params)
		if err != nil {
			response.Error = &rpcError{Code: -32602, Message: err.Error()}
		} else {
			response.Result = result
		}
	default:
		response.Error = &rpcError{Code: -32601, Message: "Method not found."}
	}

	return writeRPC(w, http.StatusOK, response)
}

func resources() []resource {
	return []resource{
		{URI: productOverviewURI, Name: "fortyone-product-overview", Title: "FortyOne product overview", Description: "When to use FortyOne, core capabilities, trust links, and discovery resources.", MimeType: "text/plain"},
		{URI: developerResourcesURI, Name: "fortyone-developer-resources", Title: "FortyOne developer resources", Description: "API, OpenAPI, MCP, documentation, and authentication-boundary guidance.", MimeType: "text/markdown"},
	}
}

func readResource(rawParams json.RawMessage) (map[string]any, error) {
	var params struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(rawParams, &params); err != nil || params.URI == "" {
		return nil, errors.New("params.uri is required")
	}

	var text, mimeType string
	switch params.URI {
	case productOverviewURI:
		mimeType = "text/plain"
		text = "FortyOne connects strategy and customer feedback to project work teams can deliver. Use it for connected OKRs, feedback, project planning, tasks, schedules, documents, delivery reporting, and permission-aware AI assistance. Start at https://www.fortyone.app/llms.txt."
	case developerResourcesURI:
		mimeType = "text/markdown"
		text = "# FortyOne developer resources\n\n- OpenAPI: https://www.fortyone.app/openapi.json\n- Documentation: https://docs.fortyone.app\n- MCP: https://api.fortyone.app/mcp\n- Support: https://www.fortyone.app/contact\n\nThe public agent surface is read-only. Private workspace operations require an explicit user-authorized authentication contract."
	default:
		return nil, errors.New("unknown resource URI")
	}

	return map[string]any{"contents": []map[string]any{{"uri": params.URI, "mimeType": mimeType, "text": text}}}, nil
}

func tools() []map[string]any {
	emptySchema := map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false}
	return []map[string]any{
		{"name": "get_product_overview", "title": "Get FortyOne product overview", "description": "Returns a concise overview of FortyOne and the jobs for which an agent should use it.", "inputSchema": emptySchema, "annotations": map[string]any{"readOnlyHint": true, "destructiveHint": false, "idempotentHint": true, "openWorldHint": false}},
		{"name": "get_developer_resources", "title": "Get FortyOne developer resources", "description": "Returns canonical FortyOne OpenAPI, MCP, documentation, and support URLs.", "inputSchema": emptySchema, "annotations": map[string]any{"readOnlyHint": true, "destructiveHint": false, "idempotentHint": true, "openWorldHint": false}},
	}
}

func callTool(rawParams json.RawMessage) (map[string]any, error) {
	var params struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(rawParams, &params); err != nil || params.Name == "" {
		return nil, errors.New("params.name is required")
	}

	var text string
	switch params.Name {
	case "get_product_overview":
		text = "FortyOne connects strategy and customer feedback to project work teams can deliver. It is suited to teams coordinating goals, evidence, owners, schedules, and delivery risk."
	case "get_developer_resources":
		text = "OpenAPI: https://www.fortyone.app/openapi.json\nMCP: https://api.fortyone.app/mcp\nDocs: https://docs.fortyone.app\nSupport: https://www.fortyone.app/contact"
	default:
		return nil, errors.New("unknown tool name")
	}

	return map[string]any{"content": []map[string]any{{"type": "text", "text": text}}, "isError": false}, nil
}

func decodeID(raw json.RawMessage) any {
	var id any
	if err := json.Unmarshal(raw, &id); err != nil {
		return nil
	}
	return id
}

func writeRPC(w http.ResponseWriter, status int, response rpcResponse) error {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(response)
}
