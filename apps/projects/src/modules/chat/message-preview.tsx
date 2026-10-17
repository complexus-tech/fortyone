import type { JSONContent } from "@tiptap/core";
import type { ReactNode } from "react";
import { Fragment } from "react";
import { ChatMention } from "./chat-rich-text";

const BLOCK_NODES = new Set([
  "paragraph",
  "heading",
  "listItem",
  "blockquote",
  "codeBlock",
]);

const hasPreviewContent = (node: JSONContent): boolean =>
  node.type === "mention" ||
  Boolean(node.type === "text" && node.text?.trim()) ||
  Boolean(node.content?.some(hasPreviewContent));

const renderPreview = (node: JSONContent): ReactNode => {
  if (node.type === "text") return node.text;
  if (node.type === "hardBreak") return " ";
  if (node.type === "mention") {
    return (
      <ChatMention>
        @{typeof node.attrs?.label === "string" ? node.attrs.label : "Member"}
      </ChatMention>
    );
  }

  return (
    <>
      {node.content?.map((child, index) => (
        <Fragment key={index}>{renderPreview(child)}</Fragment>
      ))}
      {node.type && BLOCK_NODES.has(node.type) ? " " : null}
    </>
  );
};

// Keep previews inline and non-interactive inside their conversation button.
export const MessagePreview = ({
  text,
  content,
}: {
  text: string;
  content: JSONContent | null;
}) => (content && hasPreviewContent(content) ? renderPreview(content) : text);
