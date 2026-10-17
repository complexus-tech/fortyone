import type { JSONContent } from "@tiptap/core";
import type { ChatDraft } from "./chat-rich-text";

const getContentText = (node: JSONContent): string => {
  if (node.type === "text") return node.text ?? "";
  if (node.type === "hardBreak") return "\n";
  if (node.type === "mention") {
    const label = node.attrs?.label;
    return typeof label === "string" ? `@${label}` : "";
  }
  const separator = [
    "doc",
    "bulletList",
    "orderedList",
    "listItem",
    "blockquote",
  ].includes(node.type ?? "")
    ? "\n"
    : "";
  return node.content?.map(getContentText).join(separator) ?? "";
};

const getDraftText = (draft: ChatDraft) => {
  if (draft.text.trim()) return draft.text;
  return draft.content ? getContentText(draft.content) : "";
};

const getBlocks = (draft: ChatDraft, text: string): JSONContent[] =>
  draft.content?.content && getContentText(draft.content).trim()
    ? draft.content.content
    : text.split("\n").map((line) => ({
        type: "paragraph",
        content: line ? [{ type: "text", text: line }] : [],
      }));

/** Keep the recipient's unfinished message when moving a new-message draft. */
export const mergeChatDrafts = (
  existing: ChatDraft,
  incoming: ChatDraft,
): ChatDraft => {
  const existingText = getDraftText(existing);
  const incomingText = getDraftText(incoming);
  const hasExistingText = Boolean(existingText.trim());
  const hasIncomingText = Boolean(incomingText.trim());
  const attachments = new Map(
    existing.attachments?.map((attachment) => [attachment.id, attachment]),
  );
  for (const attachment of incoming.attachments ?? []) {
    if (!attachments.has(attachment.id))
      attachments.set(attachment.id, attachment);
  }
  let text = hasExistingText ? existingText : incomingText;
  let content = hasExistingText ? existing.content : incoming.content;

  if (hasExistingText && hasIncomingText) {
    text = `${existingText.trimEnd()}\n\n${incomingText.trimStart()}`;
    content = {
      type: "doc",
      content: [
        ...getBlocks(existing, existingText),
        { type: "paragraph" },
        ...getBlocks(incoming, incomingText),
      ],
    };
  }

  return {
    ...existing,
    text,
    content,
    ...(attachments.size || existing.attachments || incoming.attachments
      ? { attachments: Array.from(attachments.values()) }
      : {}),
  };
};
