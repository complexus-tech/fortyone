/* eslint-disable no-nested-ternary -- ok */
import * as cheerio from "cheerio";
import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { auth } from "@/auth";
import { parseFigmaUrl } from "@/lib/utils/figma";

export type LinkMetadata = {
  title?: string;
  description?: string;
  image?: string;
};

type FigmaOEmbedResponse = {
  title?: string;
  author_name?: string;
  provider_name?: string;
  thumbnail_url?: string;
};

const DAY_IN_SECONDS = 60 * 60 * 24;

async function fetchFigmaMetadata(url: string): Promise<LinkMetadata | null> {
  const endpoint = new URL("https://www.figma.com/oembed");
  endpoint.searchParams.set("url", url);

  try {
    const response = await fetch(endpoint.toString(), {
      next: { revalidate: DAY_IN_SECONDS },
      headers: {
        Accept: "application/json",
      },
    });

    if (!response.ok) {
      return null;
    }

    const payload = (await response.json()) as FigmaOEmbedResponse;
    const descriptor = [payload.provider_name, payload.author_name]
      .filter(Boolean)
      .join(" • ");

    return {
      title: payload.title,
      description: descriptor || "Figma preview",
      image: payload.thumbnail_url,
    };
  } catch {
    return null;
  }
}

async function fetchMetadata(url: string): Promise<LinkMetadata | null> {
  const session = await auth();
  if (!session) {
    return null;
  }

  const parsedFigma = parseFigmaUrl(url);
  if (parsedFigma) {
    const figmaPreview = await fetchFigmaMetadata(parsedFigma.canonicalUrl);
    if (figmaPreview?.title || figmaPreview?.image) {
      return figmaPreview;
    }
  }

  try {
    // eslint-disable-next-line import/namespace -- ok
    const $ = await cheerio.fromURL(url);
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
      ? new URL(favicon, url).href
      : image
        ? new URL(image, url).href
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
    const validUrl = new URL(url);
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
