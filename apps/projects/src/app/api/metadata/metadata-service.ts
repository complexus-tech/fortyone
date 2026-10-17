import { randomUUID } from "node:crypto";
import { after } from "next/server";
import * as cheerio from "cheerio";
import type { LinkMetadata } from "@/lib/queries/links/metadata";
import { fetchPublicHtml } from "./safe-fetch";
import { parseMetadataUrl } from "./target-policy";
import { createMetadataCache } from "./metadata-cache";
import { sharedPreviewCache } from "./shared-cache";

const toMetadataImageProxyPath = (
  values: readonly (string | undefined)[],
  baseUrl: URL,
  version: string,
) => {
  for (const value of values) {
    if (!value || value.length > 8192) continue;
    try {
      const imageUrl = parseMetadataUrl(new URL(value, baseUrl));
      if (imageUrl.href.length > 8192) continue;
      return `/api/metadata/image?${new URLSearchParams({
        url: imageUrl.href,
        v: version,
      }).toString()}`;
    } catch {
      // A malformed or unsupported favicon should not hide a valid OG image.
    }
  }
  return undefined;
};

async function fetchMetadata(
  url: URL,
  signal: AbortSignal,
): Promise<LinkMetadata> {
  const { body, finalUrl } = await fetchPublicHtml(url, { signal });
  const $ = cheerio.load(body);
  const title =
    $('meta[property="og:title"]').attr("content") ||
    $('meta[name="twitter:title"]').attr("content") ||
    $("title").text().trim() ||
    undefined;
  const description =
    $('meta[name="description"]').attr("content") ||
    $('meta[property="og:description"]').attr("content") ||
    $('meta[name="twitter:description"]').attr("content") ||
    undefined;
  const image =
    $('meta[property="og:image"]').attr("content") ||
    $('meta[name="twitter:image"]').attr("content") ||
    $('meta[name="twitter:image:src"]').attr("content");
  const favicon =
    $('link[rel="shortcut icon"]').attr("href") ||
    $('link[rel="icon"]').attr("href") ||
    $('link[rel="apple-touch-icon"]').attr("href") ||
    $('link[rel="apple-touch-icon-precomposed"]').attr("href");

  const version = randomUUID();
  return {
    version,
    fetchedAt: Date.now(),
    title: title?.slice(0, 512),
    description: description?.slice(0, 2048),
    image: toMetadataImageProxyPath([favicon, image], finalUrl, version),
    previewImage: toMetadataImageProxyPath(
      [
        $('meta[property="og:image:secure_url"]').attr("content"),
        $('meta[property="og:image"]').attr("content"),
        $('meta[name="twitter:image"]').attr("content"),
        $('meta[name="twitter:image:src"]').attr("content"),
      ],
      finalUrl,
      version,
    ),
  };
}

export const metadataService = createMetadataCache(fetchMetadata, {
  shared: sharedPreviewCache,
  schedule: (work) => {
    after(work);
  },
});
