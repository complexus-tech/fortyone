import { NextResponse, type NextRequest } from "next/server";
import { confirmFeedbackVerificationAction } from "@/modules/public-portal/actions";
import { getPortalPathBySlug } from "@/modules/public-portal/utils";

const redirectWithoutTokenHistory = (destination: URL) => {
  const response = NextResponse.redirect(destination, 303);
  response.headers.set("Cache-Control", "no-store");
  response.headers.set("Referrer-Policy", "no-referrer");
  return response;
};

export const GET = async (
  request: NextRequest,
  { params }: { params: Promise<{ portalSlug: string }> },
) => {
  const { portalSlug } = await params;
  const token = request.nextUrl.searchParams.get("token")?.trim();
  const destination = new URL(
    getPortalPathBySlug(portalSlug, "feedback"),
    request.url,
  );

  if (!token) {
    destination.searchParams.set("feedbackVerification", "invalid");
    return redirectWithoutTokenHistory(destination);
  }

  const response = await confirmFeedbackVerificationAction({
    portalSlug,
    token,
  });
  destination.searchParams.set(
    "feedbackVerification",
    response.data?.participant ? "success" : "invalid",
  );
  return redirectWithoutTokenHistory(destination);
};
