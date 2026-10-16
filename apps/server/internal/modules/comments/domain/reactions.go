package commentsdomain

// SupportsReaction keeps persisted reaction identities finite and canonical.
// The database constraint and clients use this same curated catalog.
func SupportsReaction(emoji string) bool {
	switch emoji {
	// Expressions.
	case "😀", "😃", "😄", "😁", "😆", "😅", "😂", "🤣", "😊", "😍", "🥰", "😎", "🤔", "🤨", "😐", "😬", "🙄", "😮", "😱", "😢", "😭", "😴", "🤯", "🥳", "🤩":
		return true
	// Feedback.
	case "👍", "👎", "👏", "🙌", "🙏", "🤝", "💪", "👌", "✌️", "🤞", "🫡", "👀", "✅", "❌", "⚠️", "❤️", "💚", "💙", "💜", "💯":
		return true
	// Celebration.
	case "🎉", "🎊", "🎈", "🎂", "🎁", "🏆", "🥇", "⭐", "🌟", "✨", "🔥", "💥", "🥂", "🍾", "🎶":
		return true
	// Work.
	case "🚀", "💡", "🎯", "📌", "📍", "📝", "📋", "📚", "🔍", "🔎", "🛠️", "🔧", "⚙️", "🧪", "🐛", "💻", "📊", "📈", "⏳", "☕":
		return true
	default:
		return false
	}
}
