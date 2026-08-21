package figma

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractTextContent(t *testing.T) {
	t.Parallel()

	payload := json.RawMessage(`{
		"type":"FRAME",
		"children":[
			{"type":"TEXT","characters":" Checkout "},
			{"type":"FRAME","children":[
				{"type":"TEXT","characters":"Pay now"},
				{"type":"TEXT","characters":"Pay now"},
				{"type":"TEXT","characters":"   "}
			]}
		]
	}`)

	require.Equal(t, []string{"Checkout", "Pay now"}, extractTextContent(payload))
}

func TestExtractTextContentLimitsLargeFrames(t *testing.T) {
	t.Parallel()

	children := make([]figmaNode, 0, maxImportedTextItems+5)
	for index := range maxImportedTextItems + 5 {
		children = append(children, figmaNode{
			Type:       "TEXT",
			Characters: fmt.Sprintf("Label %d", index),
		})
	}

	require.Len(t, collectTextContent(figmaNode{Type: "FRAME", Children: children}), maxImportedTextItems)
}
