export type StoryCommentPayload = {
  comment: string;
  parentId?: string | null;
  mentions: string[];
  agentMentions?: "maya"[];
  creationId?: string;
};
