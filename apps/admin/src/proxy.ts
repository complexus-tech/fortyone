import { getSessionFromRequest } from "auth";
import { type NextRequest, NextResponse } from "next/server";

const PROJECTS_URL =
  process.env.NEXT_PUBLIC_PROJECTS_URL ?? "http://localhost:3000";

const PUBLIC_PREFIXES = new Set([
  "/_next",
  "/favicon",
  "/manifest",
  "/apple-icon",
]);

const isPublicPath = (pathname: string) =>
  Array.from(PUBLIC_PREFIXES).some((prefix) => pathname.startsWith(prefix));

const buildProjectsLoginRedirect = (requestUrl: string) => {
  const url = new URL("/", PROJECTS_URL);
  url.searchParams.set("callbackUrl", requestUrl);
  return NextResponse.redirect(url);
};

export default async function proxy(req: NextRequest) {
  if (isPublicPath(req.nextUrl.pathname)) {
    return NextResponse.next();
  }

  const session = await getSessionFromRequest(req);
  if (!session) {
    return buildProjectsLoginRedirect(req.nextUrl.href);
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!api|_next/static|_next/image).*)"],
};
