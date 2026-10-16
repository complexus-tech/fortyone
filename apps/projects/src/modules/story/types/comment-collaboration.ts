export const COMMENT_REACTIONS = [
  { emoji: "👍", label: "Thumbs up" },
  { emoji: "✅", label: "Done" },
  { emoji: "👀", label: "Eyes" },
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
};
