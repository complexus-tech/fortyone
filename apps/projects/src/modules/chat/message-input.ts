import type { JSONContent } from "@tiptap/core";
import type { ChatDraft } from "./chat-rich-text";

export const getMessageInput = (draft: ChatDraft) => {
  const mentionedUserIds = new Set<string>();
  const agentMentions = new Set<"maya">();
  const visit = (node: JSONContent, quoted = false) => {
    const excluded =
      quoted ||
      node.type === "blockquote" ||
      node.type === "codeBlock" ||
      node.marks?.some((mark) => mark.type === "code");
    if (
      !excluded &&
      node.type === "mention" &&
      typeof node.attrs?.id === "string"
    ) {
      if (node.attrs.mentionKind === "agent" && node.attrs.agentKey === "maya")
        agentMentions.add("maya");
      else if (node.attrs.mentionKind !== "agent")
        mentionedUserIds.add(node.attrs.id);
    }
    node.content?.forEach((child) => {
      visit(child, Boolean(excluded));
    });
  };
  if (draft.content) visit(draft.content);
  return {
    text: draft.text.trim(),
    richContent: draft.content,
    mentionedUserIds: Array.from(mentionedUserIds),
    agentMentions: Array.from(agentMentions),
    attachmentIds: draft.attachments?.map((attachment) => attachment.id) ?? [],
  };
};
