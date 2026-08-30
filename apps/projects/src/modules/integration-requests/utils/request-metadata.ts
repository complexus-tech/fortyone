export type RequestAttachment = {
  name: string;
  url?: string;
};

export type RequestExternalLink = {
  title: string;
  url: string;
};

export const getRequestExternalLinks = (
  metadata: Record<string, unknown>,
): RequestExternalLink[] => {
  const raw = metadata.links ?? metadata.urls ?? metadata.external_links;
  if (!Array.isArray(raw)) return [];

  return raw.flatMap((item): RequestExternalLink[] => {
    if (typeof item === "string" && item.trim()) {
      return [{ title: item, url: item }];
    }
    if (!item || typeof item !== "object") {
      return [];
    }
    const record = item as Record<string, unknown>;
    const url = record.url ?? record.href;
    if (typeof url !== "string" || !url.trim()) {
      return [];
    }
    const title = record.title ?? record.name ?? record.label ?? url;
    return [
      {
        title: typeof title === "string" && title.trim() ? title : url,
        url,
      },
    ];
  });
};

export const getRequestAttachments = (
  metadata: Record<string, unknown>,
): RequestAttachment[] => {
  const raw = metadata.attachments ?? metadata.files;
  if (!Array.isArray(raw)) return [];

  return raw.flatMap((item): RequestAttachment[] => {
    if (typeof item === "string") {
      return [{ name: item, url: item }];
    }
    if (!item || typeof item !== "object") {
      return [];
    }
    const record = item as Record<string, unknown>;
    const name = record.name ?? record.filename ?? record.title ?? record.url;
    if (typeof name !== "string" || !name.trim()) {
      return [];
    }
    return [
      {
        name,
        url: typeof record.url === "string" ? record.url : undefined,
      },
    ];
  });
};
