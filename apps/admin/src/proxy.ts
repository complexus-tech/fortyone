import { AuthSessionLookupError, getSessionFromRequest } from "auth";
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

const authServiceUnavailable = () =>
  new NextResponse("Authentication service is temporarily unavailable.", {
    status: 503,
    headers: {
      "Cache-Control": "no-store",
      "Content-Type": "text/plain; charset=utf-8",
      "Retry-After": "5",
    },
  });

export default async function proxy(req: NextRequest) {
  if (isPublicPath(req.nextUrl.pathname)) {
    return NextResponse.next();
  }

  let session: Awaited<ReturnType<typeof getSessionFromRequest>>;
  try {
    session = await getSessionFromRequest(req);
  } catch (error) {
    if (error instanceof AuthSessionLookupError) {
      return authServiceUnavailable();
    }
    throw error;
  }
  if (!session) {
    return buildProjectsLoginRedirect(req.nextUrl.href);
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!api|_next/static|_next/image).*)"],
};
