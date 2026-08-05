import { NextResponse, type NextRequest } from "next/server";
import { getSimilarPublicFeedback } from "@/modules/public-portal/query";

type RouteProps = {
  params: Promise<{ portalSlug: string }>;
};

export const GET = async (request: NextRequest, { params }: RouteProps) => {
  const { portalSlug } = await params;
  const title = request.nextUrl.searchParams.get("title")?.trim() ?? "";
  if (title.length < 3) return NextResponse.json({ data: [] });

  const description =
    request.nextUrl.searchParams.get("description")?.trim() ?? "";
  const requestedLimit = Number(request.nextUrl.searchParams.get("limit"));
  const limit = Number.isFinite(requestedLimit) ? requestedLimit : 3;
  const items = await getSimilarPublicFeedback(portalSlug, {
    description,
    limit,
    title,
  });

  return NextResponse.json({ data: items });
};
