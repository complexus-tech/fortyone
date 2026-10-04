package taskhandlers

import (
	"strings"

	"github.com/complexus-tech/projects-api/pkg/mailer"
)

const routineEmailDetailLimit = 10

// Keep the complete email bounded, rather than spending a fresh detail budget
// on every source. Source-specific links retain access to omitted details.
func limitRoutineEmailSections(sections []mailer.Digest, workspaceURL string) []mailer.Digest {
	result := make([]mailer.Digest, 0, len(sections))
	remaining := routineEmailDetailLimit
	for _, section := range sections {
		bounded := mailer.Digest{Intro: section.Intro}
		omitted := false
		budgetOmitted := false
		var sourceMore *mailer.DigestRow
		moreURL := workspaceURL
		for _, row := range section.Rows {
			if row.More {
				sourceMore = &row
				omitted = true
				if row.URL != "" {
					moreURL = row.URL
				}
				continue
			}
			if remaining == 0 {
				budgetOmitted = true
				omitted = true
				continue
			}
			bounded.Rows = append(bounded.Rows, row)
			remaining--
		}
		if omitted {
			more := mailer.DigestRow{More: true, Text: "Review remaining updates →", URL: moreURL}
			if sourceMore != nil && !budgetOmitted {
				more = *sourceMore
			}
			bounded.Rows = append(bounded.Rows, more)
		}
		result = append(result, bounded)
	}
	return result
}

func routineSectionsPlainText(sections []mailer.Digest) string {
	var result strings.Builder
	for _, section := range sections {
		result.WriteString(section.Intro + "\n")
		for _, row := range section.Rows {
			result.WriteString(strings.TrimSpace(row.Label+" "+row.Text) + "\n")
			if row.Detail != "" {
				result.WriteString(row.Detail + "\n")
			}
			if row.URL != "" {
				result.WriteString(row.URL + "\n")
			}
			result.WriteByte('\n')
		}
	}
	return result.String()
}
