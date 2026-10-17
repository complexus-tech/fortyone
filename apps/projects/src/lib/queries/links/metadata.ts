export type LinkMetadata = {
  fetchedAt?: number;
  stale?: boolean;
  version?: string;
  title?: string;
  description?: string;
  image?: string;
  previewImage?: string;
};

export const METADATA_SUCCESS_TTL_MS = 60 * 60 * 1000;
export const METADATA_FAILURE_TTL_MS = 30 * 1000;

export const hasLinkMetadata = (metadata?: LinkMetadata) =>
  Boolean(
    metadata?.title ||
      metadata?.description ||
      metadata?.image ||
      metadata?.previewImage,
  );

// Fragments identify a position in the same document. Query strings can select
// different documents and must retain their values and order.
export const normalizeMetadataUrl = (input: string): string | undefined => {
  try {
    const url = new URL(input);
    if (
      (url.protocol !== "http:" && url.protocol !== "https:") ||
      url.username ||
      url.password
    )
      return undefined;
    url.hash = "";
    return url.href;
  } catch {
    return undefined;
  }
};
