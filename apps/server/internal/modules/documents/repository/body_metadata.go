package documentsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	documentdomain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	"strings"

	documentssql "github.com/complexus-tech/projects-api/internal/modules/documents/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/net/html"
)

type bodyReference struct {
	EntityType string    `json:"entityType"`
	ID         uuid.UUID `json:"id"`
}

// Only structured editor atoms carry identity. Text that happens to name a
// person or contain a URL must not create relationships or notifications.
func documentBodyMetadata(content string) ([]bodyReference, []uuid.UUID, error) {
	root, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return nil, nil, fmt.Errorf("parse document identities: %w", err)
	}
	references := make([]bodyReference, 0)
	mentions := make([]uuid.UUID, 0)
	seenReferences := make(map[bodyReference]bool)
	seenMentions := make(map[uuid.UUID]bool)
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			switch node.Data {
			case "script", "style", "code", "pre":
				return
			}
			attrs := make(map[string]string, len(node.Attr))
			for _, attr := range node.Attr {
				attrs[attr.Key] = attr.Val
			}
			id, parseErr := uuid.Parse(attrs["data-id"])
			if node.Data == "span" && parseErr == nil && id != uuid.Nil {
				switch attrs["data-type"] {
				case "mention":
					if attrs["data-mention-kind"] == "user" && !seenMentions[id] {
						mentions = append(mentions, id)
						seenMentions[id] = true
					}
				case "document-reference":
					kind := attrs["data-entity-type"]
					ref := bodyReference{EntityType: kind, ID: id}
					if (kind == "story" || kind == "objective") && !seenReferences[ref] {
						references = append(references, ref)
						seenReferences[ref] = true
					}
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return references, mentions, nil
}

func syncDocumentBody(ctx context.Context, queries *documentssql.Queries, workspace, actor, document uuid.UUID, content string) error {
	references, mentions, err := documentBodyMetadata(content)
	if err != nil {
		return err
	}
	referenceData, err := json.Marshal(references)
	if err != nil {
		return fmt.Errorf("encode document references: %w", err)
	}
	mentionData, err := json.Marshal(mentions)
	if err != nil {
		return fmt.Errorf("encode document mentions: %w", err)
	}
	if err := queries.SyncDocumentBodyReferences(ctx, documentssql.SyncDocumentBodyReferencesParams{WorkspaceID: workspace, ActorID: actor, DocumentID: document, ReferenceData: referenceData}); err != nil {
		return bodyMetadataError("sync document references", err)
	}
	if err := queries.SyncDocumentBodyMentions(ctx, documentssql.SyncDocumentBodyMentionsParams{WorkspaceID: workspace, ActorID: actor, DocumentID: document, MentionData: mentionData}); err != nil {
		return bodyMetadataError("sync document mentions", err)
	}
	return nil
}

// Copies start private and without Yjs state. Retain readable labels but remove
// authority from chips whose targets the copier cannot currently access.
func filterDuplicateReferences(ctx context.Context, queries *documentssql.Queries, workspace, actor uuid.UUID, content string) (string, error) {
	references, _, err := documentBodyMetadata(content)
	if err != nil {
		return "", err
	}
	unavailable := map[uuid.UUID]bool{}
	for _, reference := range references {
		_, err := getRelatedWork(ctx, queries, workspace, actor, documentdomain.RelationshipType(reference.EntityType), reference.ID)
		if errors.Is(err, documentdomain.ErrNotFound) {
			unavailable[reference.ID] = true
		} else if err != nil {
			return "", err
		}
	}
	if len(unavailable) == 0 {
		return content, nil
	}
	root, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return "", err
	}
	var body *html.Node
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			if node.Data == "body" {
				body = node
			}
			kind, id := "", ""
			for _, attr := range node.Attr {
				if attr.Key == "data-type" {
					kind = attr.Val
				}
				if attr.Key == "data-id" {
					id = attr.Val
				}
			}
			parsed, err := uuid.Parse(id)
			if kind == "document-reference" && err == nil && unavailable[parsed] {
				node.Attr = nil
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	var output strings.Builder
	if body != nil {
		for child := body.FirstChild; child != nil; child = child.NextSibling {
			if err := html.Render(&output, child); err != nil {
				return "", err
			}
		}
	}
	return output.String(), nil
}

func bodyMetadataError(operation string, err error) error {
	var databaseError *pgconn.PgError
	if errors.As(err, &databaseError) {
		switch databaseError.Code {
		case "42501":
			return fmt.Errorf("%s: %w", operation, documentdomain.ErrForbidden)
		case "22023", "22P02":
			return fmt.Errorf("%s: %w", operation, documentdomain.ErrInvalidInput)
		}
	}
	return fmt.Errorf("%s: %w", operation, err)
}
