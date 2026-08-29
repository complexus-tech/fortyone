import { NextResponse, type NextRequest } from "next/server";
import {
  getPublicPortal,
  isPublicPortalNotFoundError,
  type PublicPortalQuery,
} from "@/modules/public-portal/query";
import { parsePublicPortalFilters } from "@/modules/public-portal/query-params";

type RouteProps = {
  params: Promise<{ portalSlug: string }>;
};

const getQuery = (request: NextRequest): PublicPortalQuery => {
  const searchParams = request.nextUrl.searchParams;
  const page = Number(searchParams.get("page") ?? "1");
  const pageSize = Number(searchParams.get("pageSize") ?? "20");

  return {
    ...parsePublicPortalFilters(searchParams),
    authorId: searchParams.get("authorId") ?? undefined,
    page: Number.isFinite(page) ? page : 1,
    pageSize: Number.isFinite(pageSize) ? pageSize : 20,
  };
};

export const GET = async (request: NextRequest, { params }: RouteProps) => {
  const { portalSlug } = await params;
  try {
    const portal = await getPublicPortal(portalSlug, getQuery(request));

    return NextResponse.json({ data: portal });
  } catch (error) {
    if (isPublicPortalNotFoundError(error)) {
      return NextResponse.json(
        { error: { message: "Feedback portal not found" } },
        { status: 404 },
      );
    }

    throw error;
  }
};
