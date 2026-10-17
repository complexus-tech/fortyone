import { getApiUrl } from "@/lib/api-url";

export const getChatAttachmentUrl = (
  workspaceSlug: string,
  conversationId: string,
  attachmentId: string,
) =>
  `${getApiUrl()}/workspaces/${encodeURIComponent(workspaceSlug)}/chat/conversations/${encodeURIComponent(conversationId)}/attachments/${encodeURIComponent(attachmentId)}`;

export const formatFileSize = (bytes: number) =>
  bytes < 1_048_576
    ? `${Math.max(1, Math.round(bytes / 1024))} KB`
    : `${(bytes / 1_048_576).toFixed(1)} MB`;

export const getChatPdfPreviewUrl = (
  workspaceSlug: string,
  conversationId: string,
  attachmentId: string,
) =>
  `/api/chat/${encodeURIComponent(workspaceSlug)}/conversations/${encodeURIComponent(conversationId)}/attachments/${encodeURIComponent(attachmentId)}/preview`;

export const getChatAudioPreviewUrl = (
  workspaceSlug: string,
  conversationId: string,
  attachmentId: string,
) =>
  `${getChatPdfPreviewUrl(workspaceSlug, conversationId, attachmentId)}?media=audio`;

export const getChatAttachmentThumbnailUrl = (
  workspaceSlug: string,
  conversationId: string,
  attachmentId: string,
) =>
  getChatPdfPreviewUrl(workspaceSlug, conversationId, attachmentId).replace(
    /\/preview$/,
    "/thumbnail",
  );
