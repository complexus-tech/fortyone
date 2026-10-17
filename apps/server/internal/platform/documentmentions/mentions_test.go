package documentmentions

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOnlyActiveMarkdownLinksCarryMentionAuthority(t *testing.T) {
	id := uuid.New()
	link := "[@Name](mention:" + id.String() + ")"
	for _, example := range []string{"`" + link + "`", "    " + link, "```\n" + link + "\n```", "> " + link, "<div>" + link + "</div>", "<span>" + link + "</span>", "\\" + link} {
		require.Empty(t, Candidates(example).UserIDs, example)
	}
	require.Equal(t, []uuid.UUID{id}, Candidates("Please "+link+" and "+link).UserIDs)
	require.True(t, Candidates("Please [@Other](agent:maya)").Maya)
	require.False(t, Candidates("> [@Maya](agent:maya)").Maya)
}

func TestCanonicalNamesEscapeMarkdownAndUnknownIDsBecomePlainText(t *testing.T) {
	id, unknown := uuid.New(), uuid.New()
	result := Canonicalize("[@Forged](mention:"+id.String()+") [@Unknown](mention:"+unknown.String()+") [@Bad](mention:not-a-uuid)", map[uuid.UUID]string{id: "Jo [admin](x)"})
	require.Equal(t, []uuid.UUID{id}, result.UserIDs)
	require.Contains(t, result.Body, `[@Jo \[admin\]\(x\)](mention:`+id.String()+")")
	require.Contains(t, result.Body, "@Unknown @Bad")
	require.Equal(t, []uuid.UUID{id}, Candidates(result.Body).UserIDs)
}

func TestMentionRecipientsAreBoundedAndCodeIndentationIsPreserved(t *testing.T) {
	names := map[uuid.UUID]string{}
	var links []string
	for i := 0; i < 12; i++ {
		id := uuid.New()
		names[id] = fmt.Sprintf("User%d", i)
		links = append(links, "[@Someone](mention:"+id.String()+")")
	}
	result := Canonicalize(strings.Join(links, " "), names)
	require.Len(t, result.UserIDs, MaximumRecipients)
	require.Len(t, Candidates(result.Body).UserIDs, MaximumRecipients)
	code := "    " + links[0] + "\n"
	require.Equal(t, code, Canonicalize(code, names).Body)
	require.Empty(t, Canonicalize(code, names).UserIDs)
}
