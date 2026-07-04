import { NextResponse, type NextRequest } from "next/server";
import {
  getPublicPortal,
  isPublicPortalNotFoundError,
  type PublicPortalQuery,
} from "@/modules/public-portal/query";
import type { PublicRequestStatus } from "@/modules/public-portal/types";

type RouteProps = {
  params: Promise<{ portalSlug: string }>;
};

const getQuery = (request: NextRequest): PublicPortalQuery => {
  const searchParams = request.nextUrl.searchParams;
  const page = Number(searchParams.get("page") ?? "1");
  const pageSize = Number(searchParams.get("pageSize") ?? "20");
  const sort = searchParams.get("sort");
  const status = searchParams.get("status");

  return {
    boardId: searchParams.get("boardId") ?? undefined,
    page: Number.isFinite(page) ? page : 1,
    pageSize: Number.isFinite(pageSize) ? pageSize : 20,
    search: searchParams.get("search") ?? undefined,
    sort:
      sort === "newest" || sort === "oldest" || sort === "top" ? sort : "top",
    status: isPublicRequestStatus(status) ? status : undefined,
  };
};

const isPublicRequestStatus = (
  status: string | null,
): status is PublicRequestStatus =>
  status === "pending" ||
  status === "reviewing" ||
  status === "planned" ||
  status === "in_progress" ||
  status === "completed" ||
  status === "closed";

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
