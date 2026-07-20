import { NextResponse } from "next/server";
import {
  getPublicContributor,
  isPublicPortalNotFoundError,
} from "@/modules/public-portal/query";

type RouteProps = {
  params: Promise<{ authorId: string; portalSlug: string }>;
};

export const GET = async (_request: Request, { params }: RouteProps) => {
  const { authorId, portalSlug } = await params;

  try {
    const contributor = await getPublicContributor(portalSlug, authorId);
    return NextResponse.json({ data: contributor });
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
