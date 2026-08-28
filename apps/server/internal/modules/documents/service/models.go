package documents

import documentdomain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"

type Visibility = documentdomain.Visibility

const (
	VisibilityWorkspace  = documentdomain.VisibilityWorkspace
	VisibilityRestricted = documentdomain.VisibilityRestricted
	VisibilityPrivate    = documentdomain.VisibilityPrivate
)

type RelationshipType = documentdomain.RelationshipType

const (
	RelationshipStory     = documentdomain.RelationshipStory
	RelationshipObjective = documentdomain.RelationshipObjective
)

// Compatibility aliases keep the existing HTTP and service contracts stable
// while persistence depends only on transport-neutral domain values.
type CoreDocumentMember = documentdomain.Member
type CoreRelatedWork = documentdomain.RelatedWork
type CoreDocument = documentdomain.Document
type CoreDocumentSummary = documentdomain.Summary
type CoreListInput = documentdomain.ListInput
type CoreCreateInput = documentdomain.CreateInput
type CoreUpdateInput = documentdomain.UpdateInput
type CoreAccessInput = documentdomain.AccessInput
type CoreRelationshipInput = documentdomain.RelationshipInput
type CoreMediaInput = documentdomain.MediaInput
