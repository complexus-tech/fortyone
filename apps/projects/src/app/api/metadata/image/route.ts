import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { auth } from "@/auth";
import { MetadataFetchError, parseMetadataUrl } from "../target-policy";
import { metadataImageCache } from "./image-cache";

export const runtime = "nodejs";

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
    const version = searchParams.get("v") ?? "";
    if (!/^[a-zA-Z0-9-]{0,64}$/.test(version)) {
      return NextResponse.json(
        { error: "Invalid image version" },
        { status: 400 },
      );
    }
    const image = await metadataImageCache.get(
      validUrl.href,
      version,
      request.signal,
    );

    return new NextResponse(new Uint8Array(image.body), {
      headers: {
        "cache-control": "private, max-age=86400",
        "content-length": String(image.body.byteLength),
        "content-security-policy": "sandbox; default-src 'none'",
        "content-type": image.contentType,
        "cross-origin-resource-policy": "same-origin",
        "x-content-type-options": "nosniff",
      },
      status: 200,
    });
  } catch (error) {
    if (
      error instanceof MetadataFetchError &&
      (error.code === "invalid-url" || error.code === "unsafe-target")
    ) {
      return NextResponse.json({ error: "Invalid image URL" }, { status: 400 });
    }
    if (error instanceof MetadataFetchError && error.code === "timeout") {
      return NextResponse.json(
        { error: "Metadata image request timed out" },
        { status: 504 },
      );
    }
    return NextResponse.json(
      { error: "Failed to fetch metadata image" },
      { status: 502 },
    );
  }
}
