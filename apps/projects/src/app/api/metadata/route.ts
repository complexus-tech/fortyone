import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { auth } from "@/auth";
import { MetadataCapacityError } from "./metadata-cache";
import { metadataService } from "./metadata-service";
import { MetadataFetchError, parseMetadataUrl } from "./target-policy";
import { allowPreviewRefresh } from "./refresh-limiter";

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
    const force = searchParams.get("refresh") === "1";
    const target = parseMetadataUrl(url);
    target.hash = "";
    if (force && !(await allowPreviewRefresh(session.user.id, target.href))) {
      return NextResponse.json(
        { error: "Please wait a moment before refreshing again." },
        {
          status: 429,
          headers: { "Retry-After": "5", "Cache-Control": "private, no-store" },
        },
      );
    }
    const metadata = await metadataService.get(target, request.signal, {
      force,
    });
    return NextResponse.json(metadata, {
      headers: { "Cache-Control": "private, no-store" },
    });
  } catch (error) {
    if (error instanceof MetadataCapacityError) {
      return NextResponse.json(
        { error: "Metadata service is busy" },
        { status: 503, headers: { "Retry-After": "1" } },
      );
    }
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
