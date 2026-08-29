package links

import linksdomain "github.com/complexus-tech/projects-api/internal/modules/links/domain"

// Compatibility aliases keep the existing service and HTTP API stable while
// the persistence adapter depends only on transport-neutral domain values.
type CoreLink = linksdomain.Link
type CoreNewLink = linksdomain.CreateLink
type CoreUpdateLink = linksdomain.UpdateLink
