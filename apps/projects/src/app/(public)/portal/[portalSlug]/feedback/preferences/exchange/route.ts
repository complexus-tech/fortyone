import { NextResponse, type NextRequest } from "next/server";
import { exchangeFeedbackPreferenceTokenAction } from "@/modules/public-portal/actions";
import { getFeedbackPreferencesPath } from "@/modules/public-portal/utils";

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
    getFeedbackPreferencesPath(portalSlug),
    request.url,
  );
  if (!token) {
    destination.searchParams.set("preferenceLink", "invalid");
    return redirectWithoutTokenHistory(destination);
  }

  const response = await exchangeFeedbackPreferenceTokenAction({
    portalSlug,
    token,
  });
  destination.searchParams.set(
    "preferenceLink",
    response.data ? "accepted" : "invalid",
  );
  return redirectWithoutTokenHistory(destination);
};
