import * as cheerio from "cheerio";
import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { auth } from "@/auth";
import { fetchPublicHtml } from "./safe-fetch";
import { MetadataFetchError, parseMetadataUrl } from "./target-policy";

export const runtime = "nodejs";

export type LinkMetadata = {
  title?: string;
  description?: string;
  image?: string;
};

const toMetadataImageProxyPath = (
  values: readonly (string | undefined)[],
  baseUrl: URL,
) => {
  for (const value of values) {
    if (!value) continue;
    try {
      const imageUrl = parseMetadataUrl(new URL(value, baseUrl));
      return `/api/metadata/image?${new URLSearchParams({
        url: imageUrl.href,
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

  return {
    title,
    description,
    image: toMetadataImageProxyPath([favicon, image], finalUrl),
  };
}

export async function GET(request: NextRequest) {
  const session = await auth();
  if (!session?.user) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  const { searchParams } = new URL(request.url);
  const url = searchParams.get("url");

  if (!url) {
    return NextResponse.json(
      { error: "URL parameter is required" },
      { status: 400 },
    );
  }

  try {
    const validUrl = parseMetadataUrl(url);
    const metadata = await fetchMetadata(validUrl, request.signal);
    return NextResponse.json(metadata);
  } catch (error) {
    if (
      error instanceof MetadataFetchError &&
      (error.code === "invalid-url" || error.code === "unsafe-target")
    ) {
      return NextResponse.json({ error: "Invalid URL" }, { status: 400 });
    }
    if (error instanceof MetadataFetchError && error.code === "timeout") {
      return NextResponse.json(
        { error: "Metadata request timed out" },
        { status: 504 },
      );
    }
    return NextResponse.json(
      { error: "Failed to fetch metadata" },
      { status: 502 },
    );
  }
}
