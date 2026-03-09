import { type NextRequest, NextResponse } from "next/server";
import { getSessionFromRequest } from "auth";

const AUTH_HOST = process.env.NEXT_PUBLIC_AUTH_HOST ?? "cloud.fortyone.app";
const DOMAIN_SUFFIX = ".fortyone.app";
const RESERVED_SUBDOMAINS = new Set(["cloud"]);
const AUTH_ONLY_PREFIXES = new Set([
  "/signup",
  "/auth-callback",
  "/verify",
  "/onboarding",
  "/unauthorized",
]);

const getHostname = (req: NextRequest) => req.nextUrl.hostname;

const getSubdomain = (hostname: string) =>
  hostname.endsWith(DOMAIN_SUFFIX)
    ? hostname.replace(DOMAIN_SUFFIX, "")
    : undefined;

const isFortyOneHost = (hostname: string) =>
  hostname === "fortyone.app" || hostname.endsWith(DOMAIN_SUFFIX);

const isAuthOnlyPath = (pathname: string) =>
  pathname === "/" ||
  Array.from(AUTH_ONLY_PREFIXES).some((prefix) => pathname.startsWith(prefix));

const buildAuthUrl = (pathname: string, searchParams: string) => {
  const url = new URL(pathname, `https://${AUTH_HOST}`);
  url.search = searchParams;
  return url;
};

const buildAuthRedirect = (callbackUrl: string) => {
  const url = new URL("/", `https://${AUTH_HOST}`);
  url.searchParams.set("callbackUrl", callbackUrl);
  return NextResponse.redirect(url);
};

const buildHostRedirect = (
  requestUrl: string,
  callbackUrl: string,
) => {
  const url = new URL("/", requestUrl);
  url.searchParams.set("callbackUrl", callbackUrl);
  return NextResponse.redirect(url);
};

export default async function proxy(req: NextRequest) {
  const user = await getSessionFromRequest(req);
  const pathname = req.nextUrl.pathname;
  const searchParams = req.nextUrl.search;
  const hostname = getHostname(req);
  const isAuthenticated = !!user;

  if (!isFortyOneHost(hostname)) {
    if (!isAuthenticated && !isAuthOnlyPath(pathname)) {
      return buildHostRedirect(req.url, `${pathname}${searchParams}`);
    }

    return NextResponse.next();
  }

  const subdomain = getSubdomain(hostname);
  const isWorkspaceSubdomain =
    !!subdomain && !RESERVED_SUBDOMAINS.has(subdomain);

  if (isWorkspaceSubdomain && isAuthOnlyPath(pathname)) {
    if (pathname === "/" && isAuthenticated) {
      const rewriteUrl = new URL(`/${subdomain}/my-work`, req.nextUrl);
      rewriteUrl.search = searchParams;
      return NextResponse.rewrite(rewriteUrl);
    }

    const redirectUrl = buildAuthUrl(pathname, searchParams);

    if (pathname === "/") {
      redirectUrl.searchParams.set(
        "callbackUrl",
        `https://${subdomain}${DOMAIN_SUFFIX}/my-work`,
      );
    }

    return NextResponse.redirect(redirectUrl);
  }

  if (isWorkspaceSubdomain && !isAuthenticated) {
    const callbackUrl = `https://${subdomain}${DOMAIN_SUFFIX}${pathname}${searchParams}`;
    return buildAuthRedirect(callbackUrl);
  }

  if (isWorkspaceSubdomain && !pathname.startsWith(`/${subdomain}`)) {
    const nextPath =
      pathname === "/" ? `/${subdomain}/my-work` : `/${subdomain}${pathname}`;
    const rewriteUrl = new URL(nextPath, req.nextUrl);
    rewriteUrl.search = searchParams;
    return NextResponse.rewrite(rewriteUrl);
  }

  if (!isAuthenticated && !isAuthOnlyPath(pathname)) {
    return buildAuthRedirect(`${pathname}${searchParams}`);
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    "/((?!api|_next/static|images|_next/image|favicon*|ingest|manifest*|apple-icon*).*)",
  ],
};
