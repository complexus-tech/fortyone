export type PreviewAttachment = {
  id: string;
  fileName: string;
  mediaType: string;
  sizeBytes: number;
  url: string;
};

export type PreviewMediaKind = "image" | "pdf" | "video" | "audio" | "file";

export const getPreviewMediaKind = (
  attachment: Pick<PreviewAttachment, "mediaType" | "fileName">,
): PreviewMediaKind => {
  const mediaType = attachment.mediaType.split(";")[0].trim().toLowerCase();
  if (mediaType.startsWith("audio/")) return "audio";
  if (mediaType.startsWith("image/")) return "image";
  if (mediaType.startsWith("video/")) return "video";
  if (mediaType === "application/pdf") return "pdf";
  if (!mediaType || mediaType === "application/octet-stream") {
    if (/\.(?:mp3|m4a|aac|wav|ogg|oga|opus|flac)$/i.test(attachment.fileName)) {
      return "audio";
    }
    if (/\.pdf$/i.test(attachment.fileName)) return "pdf";
  }
  return "file";
};

export const getPreviewAttachments = (attachments: PreviewAttachment[]) =>
  attachments.filter(
    (attachment) => getPreviewMediaKind(attachment) !== "audio",
  );

export const getAdjacentPreviewId = (
  attachments: PreviewAttachment[],
  selectedId: string | null,
  direction: -1 | 1,
) => {
  const index = attachments.findIndex(
    (attachment) => attachment.id === selectedId,
  );
  return index < 0 ? null : attachments[index + direction]?.id ?? null;
};

export const isPreviewNavigationTarget = (target: EventTarget | null) =>
  target instanceof HTMLElement &&
  Boolean(
    target.closest(
      'input, textarea, select, button, a, [role="slider"], [contenteditable]:not([contenteditable="false"]), [data-preview-local-keys]',
    ),
  );

export const getSafePreviewUrl = (value: string) => {
  if (!value) return null;
  try {
    const url = new URL(value, "https://preview.invalid");
    if (!["https:", "http:", "blob:"].includes(url.protocol)) return null;
    if (url.username || url.password) return null;
    return value;
  } catch {
    return null;
  }
};
