/* eslint-disable no-nested-ternary -- ok */
import ky from "ky";
import * as cheerio from "cheerio";
import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";

export type LinkMetadata = {
  title?: string;
  description?: string;
  image?: string;
};

async function fetchMetadata(url: string): Promise<LinkMetadata | null> {
  try {
    const html = await ky
      .get(url, {
        timeout: 10000,
        headers: {
          "User-Agent": "Mozilla/5.0 (compatible; LinkMetadataBot/1.0)",
          Accept:
            "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
          "Accept-Language": "en-US,en;q=0.5",
          "Accept-Encoding": "gzip, deflate",
          Connection: "keep-alive",
          "Cache-Control": "public, max-age=86400, stale-while-revalidate=60", // 24 hours, 1 minute stale-while-revalidate
        },
      })
      .text();
    const $ = cheerio.load(html);
    const title =
      $('meta[property="og:title"]').attr("content") ||
      $('meta[name="twitter:title"]').attr("content") ||
      $("title").text().trim() ||
      "No Title Found";
    const description =
      $('meta[name="description"]').attr("content") ||
      $('meta[property="og:description"]').attr("content") ||
      $('meta[name="twitter:description"]').attr("content") ||
      "No Description Found";
    const image =
      $('meta[property="og:image"]').attr("content") ||
      $('meta[name="twitter:image"]').attr("content") ||
      $('meta[name="twitter:image:src"]').attr("content");
    const favicon =
      $('link[rel="shortcut icon"]').attr("href") ||
      $('link[rel="icon"]').attr("href") ||
      $('link[rel="apple-touch-icon"]').attr("href") ||
      $('link[rel="apple-touch-icon-precomposed"]').attr("href");

    const finalImage = favicon
      ? new URL(favicon as string, url).href
      : image
        ? new URL(image as string, url).href
        : undefined;

    return {
      title: title === "No Title Found" ? undefined : title,
      description:
        description === "No Description Found" ? undefined : description,
      image: finalImage,
    };
  } catch (error) {
    // Return null on error
    return null;
  }
}

export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url);
  const url = searchParams.get("url");

  if (!url) {
    return NextResponse.json(
      { error: "URL parameter is required" },
      { status: 400 },
    );
  }

  try {
    // Validate URL
    const validUrl = new URL(url);
    // Use the validated URL
    const metadata = await fetchMetadata(validUrl.href);

    if (!metadata) {
      return NextResponse.json(
        { error: "Failed to fetch metadata" },
        { status: 500 },
      );
    }

    return NextResponse.json(metadata);
  } catch {
    return NextResponse.json({ error: "Invalid URL" }, { status: 400 });
  }
}
