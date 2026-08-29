package search

import searchdomain "github.com/complexus-tech/projects-api/internal/modules/search/domain"

// Service-facing aliases preserve the existing public module contract while
// keeping the canonical types in the dependency-neutral domain package.
type (
	CoreSearchStory     = searchdomain.CoreSearchStory
	CoreSearchObjective = searchdomain.CoreSearchObjective
	CoreSearchResult    = searchdomain.CoreSearchResult
	CoreSimilarStory    = searchdomain.CoreSimilarStory
	SearchType          = searchdomain.SearchType
	SortOption          = searchdomain.SortOption
	SearchParams        = searchdomain.SearchParams
)

const (
	SearchTypeAll        = searchdomain.SearchTypeAll
	SearchTypeStories    = searchdomain.SearchTypeStories
	SearchTypeObjectives = searchdomain.SearchTypeObjectives
	SortByRelevance      = searchdomain.SortByRelevance
	SortByUpdated        = searchdomain.SortByUpdated
	SortByCreated        = searchdomain.SortByCreated
)
