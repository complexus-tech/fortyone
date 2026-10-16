package commentsdomain

import "testing"

func TestReactionCatalogUsesCanonicalFiniteEmoji(t *testing.T) {
	for _, emoji := range []string{"👍", "✅", "👀", "❤️", "🎉", "😂", "🙏", "🚀", "💡", "🔥", "👏", "🙌", "🤔", "😊", "😢", "👎", "💯", "🤯", "🫡", "🛠️", "⚙️", "🥂", "☕"} {
		if !SupportsReaction(emoji) {
			t.Errorf("catalog reaction %q is unsupported", emoji)
		}
	}
	for _, emoji := range []string{"", "hello", ":heart:", "❤", "👍🏽", "👍👍", " 👀", "🚀 ", "<img src=x>", "🛠", "⚙", "⚠", "✌"} {
		if SupportsReaction(emoji) {
			t.Errorf("non-canonical reaction %q was accepted", emoji)
		}
	}
}
