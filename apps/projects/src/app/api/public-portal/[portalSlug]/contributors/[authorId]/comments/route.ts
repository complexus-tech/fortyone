import { NextResponse, type NextRequest } from "next/server";
import {
  getPublicContributorComments,
  isPublicPortalNotFoundError,
} from "@/modules/public-portal/query";

type RouteProps = {
  params: Promise<{ authorId: string; portalSlug: string }>;
};

const parsePositiveInteger = (value: string | null, fallback: number) => {
  const parsed = Number(value);
  return Number.isInteger(parsed) && parsed > 0 ? parsed : fallback;
};

export const GET = async (request: NextRequest, { params }: RouteProps) => {
  const { authorId, portalSlug } = await params;
  const page = parsePositiveInteger(
    request.nextUrl.searchParams.get("page"),
    1,
  );
  const pageSize = parsePositiveInteger(
    request.nextUrl.searchParams.get("pageSize"),
    20,
  );

  try {
    const comments = await getPublicContributorComments(
      portalSlug,
      authorId,
      page,
      pageSize,
    );
    return NextResponse.json({ data: comments });
  } catch (error) {
    if (isPublicPortalNotFoundError(error)) {
      return NextResponse.json(
        { error: { message: "Contributor not found" } },
        { status: 404 },
      );
    }

    throw error;
  }
};
