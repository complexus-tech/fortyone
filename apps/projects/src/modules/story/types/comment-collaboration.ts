export const COMMENT_REACTION_GROUPS = [
  "Expressions",
  "Feedback",
  "Celebrations",
  "Work",
] as const;

export const COMMENT_REACTIONS = [
  { emoji: "😀", label: "Grinning", group: "Expressions" },
  { emoji: "😃", label: "Happy", group: "Expressions" },
  { emoji: "😄", label: "Big smile", group: "Expressions" },
  { emoji: "😁", label: "Beaming", group: "Expressions" },
  { emoji: "😆", label: "Laughing", group: "Expressions" },
  { emoji: "😅", label: "Relieved", group: "Expressions" },
  { emoji: "😂", label: "Laugh", group: "Expressions" },
  { emoji: "🤣", label: "Rolling with laughter", group: "Expressions" },
  { emoji: "😊", label: "Smile", group: "Expressions" },
  { emoji: "😍", label: "Love it", group: "Expressions" },
  { emoji: "🥰", label: "Feeling loved", group: "Expressions" },
  { emoji: "😎", label: "Cool", group: "Expressions" },
  { emoji: "🤔", label: "Thinking", group: "Expressions" },
  { emoji: "🤨", label: "Skeptical", group: "Expressions" },
  { emoji: "😐", label: "Neutral", group: "Expressions" },
  { emoji: "😬", label: "Grimacing", group: "Expressions" },
  { emoji: "🙄", label: "Eye roll", group: "Expressions" },
  { emoji: "😮", label: "Surprised", group: "Expressions" },
  { emoji: "😱", label: "Shocked", group: "Expressions" },
  { emoji: "😢", label: "Sad", group: "Expressions" },
  { emoji: "😭", label: "Crying", group: "Expressions" },
  { emoji: "😴", label: "Sleepy", group: "Expressions" },
  { emoji: "🤯", label: "Mind blown", group: "Expressions" },
  { emoji: "🥳", label: "Party", group: "Expressions" },
  { emoji: "🤩", label: "Star struck", group: "Expressions" },
  { emoji: "👍", label: "Thumbs up", group: "Feedback" },
  { emoji: "👎", label: "Thumbs down", group: "Feedback" },
  { emoji: "👏", label: "Applause", group: "Feedback" },
  { emoji: "🙌", label: "Raised hands", group: "Feedback" },
  { emoji: "🙏", label: "Thank you", group: "Feedback" },
  { emoji: "🤝", label: "Handshake", group: "Feedback" },
  { emoji: "💪", label: "Strong", group: "Feedback" },
  { emoji: "👌", label: "OK", group: "Feedback" },
  { emoji: "✌️", label: "Peace", group: "Feedback" },
  { emoji: "🤞", label: "Fingers crossed", group: "Feedback" },
  { emoji: "🫡", label: "Salute", group: "Feedback" },
  { emoji: "👀", label: "Eyes", group: "Feedback" },
  { emoji: "✅", label: "Done", group: "Feedback" },
  { emoji: "❌", label: "No", group: "Feedback" },
  { emoji: "⚠️", label: "Warning", group: "Feedback" },
  { emoji: "❤️", label: "Heart", group: "Feedback" },
  { emoji: "💚", label: "Green heart", group: "Feedback" },
  { emoji: "💙", label: "Blue heart", group: "Feedback" },
  { emoji: "💜", label: "Purple heart", group: "Feedback" },
  { emoji: "💯", label: "One hundred", group: "Feedback" },
  { emoji: "🎉", label: "Celebrate", group: "Celebrations" },
  { emoji: "🎊", label: "Confetti", group: "Celebrations" },
  { emoji: "🎈", label: "Balloon", group: "Celebrations" },
  { emoji: "🎂", label: "Birthday", group: "Celebrations" },
  { emoji: "🎁", label: "Gift", group: "Celebrations" },
  { emoji: "🏆", label: "Trophy", group: "Celebrations" },
  { emoji: "🥇", label: "Gold medal", group: "Celebrations" },
  { emoji: "⭐", label: "Star", group: "Celebrations" },
  { emoji: "🌟", label: "Glowing star", group: "Celebrations" },
  { emoji: "✨", label: "Sparkles", group: "Celebrations" },
  { emoji: "🔥", label: "Fire", group: "Celebrations" },
  { emoji: "💥", label: "Boom", group: "Celebrations" },
  { emoji: "🥂", label: "Cheers", group: "Celebrations" },
  { emoji: "🍾", label: "Champagne", group: "Celebrations" },
  { emoji: "🎶", label: "Music", group: "Celebrations" },
  { emoji: "🚀", label: "Rocket", group: "Work" },
  { emoji: "💡", label: "Idea", group: "Work" },
  { emoji: "🎯", label: "Target", group: "Work" },
  { emoji: "📌", label: "Pin", group: "Work" },
  { emoji: "📍", label: "Location", group: "Work" },
  { emoji: "📝", label: "Note", group: "Work" },
  { emoji: "📋", label: "Checklist", group: "Work" },
  { emoji: "📚", label: "Books", group: "Work" },
  { emoji: "🔍", label: "Search", group: "Work" },
  { emoji: "🔎", label: "Investigate", group: "Work" },
  { emoji: "🛠️", label: "Tools", group: "Work" },
  { emoji: "🔧", label: "Wrench", group: "Work" },
  { emoji: "⚙️", label: "Settings", group: "Work" },
  { emoji: "🧪", label: "Experiment", group: "Work" },
  { emoji: "🐛", label: "Bug", group: "Work" },
  { emoji: "💻", label: "Computer", group: "Work" },
  { emoji: "📊", label: "Chart", group: "Work" },
  { emoji: "📈", label: "Growth", group: "Work" },
  { emoji: "⏳", label: "Waiting", group: "Work" },
  { emoji: "☕", label: "Coffee", group: "Work" },
] as const;

export type CommentEmoji = (typeof COMMENT_REACTIONS)[number]["emoji"];

export type CommentReaction = {
  emoji: CommentEmoji;
  count: number;
  reacted: boolean;
};

export type CommentState = {
  commentId: string;
  reactions: CommentReaction[];
  resolvedAt: string | null;
  resolvedBy: string | null;
};

export type StoryCommentState = {
  lastReadAt: string | null;
  observedAt: string;
  unreadCount: number;
  latestCommentId: string | null;
  comments: CommentState[];
  threadReads?: Record<string, string>;
};

export type CommentReadInput = {
  observedAt: string;
  threadId?: string;
};
