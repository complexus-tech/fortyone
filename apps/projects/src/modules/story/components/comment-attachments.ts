import type { Editor } from "@tiptap/core";
import { getApiUrl } from "@/lib/api-url";
import type { StoryAttachment } from "../types";

export const COMMENT_ATTACHMENT_NODE = "commentAttachment";
export const COMMENT_ATTACHMENT_ATTRIBUTE = "data-comment-attachment-id";
const UUID_PATTERN =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

export const commentAttachmentUrl = (
  workspaceSlug: string,
  storyId: string,
  attachmentId: string,
) =>
  `${getApiUrl()}/workspaces/${encodeURIComponent(workspaceSlug)}/stories/${encodeURIComponent(storyId)}/attachments/${encodeURIComponent(attachmentId)}`;

export const insertCommentAttachment = (
  editor: Editor,
  attachment: StoryAttachment,
) => {
  const { doc } = editor.state;
  const end = doc.content.size;
  const last = doc.lastChild;
  // Reuse the trailing typing paragraph instead of accumulating a blank block
  // for every upload. Removing the last attachment then restores an empty editor.
  const position =
    last?.type.name === "paragraph" && last.content.size === 0
      ? { from: end - last.nodeSize, to: end }
      : end;
  return editor.commands.insertContentAt(position, [
    { type: COMMENT_ATTACHMENT_NODE, attrs: attachment },
    { type: "paragraph" },
  ]);
};

export const readCommentAttachment = (
  element: Element,
): StoryAttachment | null => {
  const id = element.getAttribute(COMMENT_ATTACHMENT_ATTRIBUTE);
  const filename = element.getAttribute("data-filename");
  const url = element.getAttribute("data-url");
  const mimeType = element.getAttribute("data-mime-type");
  if (!id || !UUID_PATTERN.test(id) || !filename || !url || !mimeType)
    return null;
  try {
    const parsed = new URL(url);
    const api = new URL(getApiUrl());
    if (
      parsed.origin !== api.origin ||
      !["http:", "https:"].includes(parsed.protocol) ||
      parsed.username ||
      parsed.password ||
      parsed.search ||
      parsed.hash ||
      !parsed.pathname.startsWith(
        `${api.pathname.replace(/\/$/, "")}/workspaces/`,
      ) ||
      !/\/workspaces\/[^/]+\/stories\/[^/]+\/attachments\/[^/]+$/.test(
        parsed.pathname,
      ) ||
      !parsed.pathname.endsWith(`/attachments/${id}`)
    )
      return null;
  } catch {
    return null;
  }
  return {
    id,
    filename,
    url,
    mimeType,
    size: Number(element.getAttribute("data-size")) || 0,
    createdAt: element.getAttribute("data-created-at") ?? "",
    uploadedBy: element.getAttribute("data-uploaded-by") ?? "",
  };
};

export const splitCommentAttachments = (html: string) => {
  const document = new DOMParser().parseFromString(html, "text/html");
  const attachments: StoryAttachment[] = [];
  const seen = new Set<string>();
  document
    .querySelectorAll(`[${COMMENT_ATTACHMENT_ATTRIBUTE}]`)
    .forEach((element) => {
      const attachment = readCommentAttachment(element);
      if (!attachment) return;
      if (!seen.has(attachment.id)) {
        attachments.push(attachment);
        seen.add(attachment.id);
      }
      element.remove();
    });
  return { html: document.body.innerHTML, attachments };
};
