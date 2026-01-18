import { NextResponse } from "next/server";
import { auth } from "@/auth";

const AUTH_HOST = "auth.fortyone.app";
const DOMAIN_SUFFIX = ".fortyone.app";
const RESERVED_SUBDOMAINS = new Set(["auth"]);
const AUTH_ONLY_PREFIXES = new Set([
  "/signup",
  "/auth-callback",
  "/verify",
  "/onboarding",
  "/unauthorized",
  "/api/auth",
]);
const AUTH_ONLY_SEGMENTS = new Set([
  "signup",
  "auth-callback",
  "verify",
  "onboarding",
  "unauthorized",
  "api",
]);

const getHost = (req: Request) => req.headers.get("host")?.split(":")[0];

const getSubdomain = (host?: string) =>
  host?.endsWith(DOMAIN_SUFFIX) ? host.replace(DOMAIN_SUFFIX, "") : undefined;

const isAuthOnlyPath = (pathname: string) =>
  pathname === "/" ||
  Array.from(AUTH_ONLY_PREFIXES).some((prefix) => pathname.startsWith(prefix));

const getPathSegments = (pathname: string) =>
  pathname.split("/").filter(Boolean);

const getAuthWorkspaceSlug = (host: string, pathSegments: string[]) => {
  if (host !== AUTH_HOST || pathSegments.length === 0) {
    return undefined;
  }

  return AUTH_ONLY_SEGMENTS.has(pathSegments[0]) ? undefined : pathSegments[0];
};

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
  protocol: string,
  host: string,
  callbackUrl: string,
) => {
  const baseUrl = `${protocol}//${host}`;
  const url = new URL("/", baseUrl);
  url.searchParams.set("callbackUrl", callbackUrl);
  return NextResponse.redirect(url);
};

const buildSubdomainUrl = (
  subdomain: string,
  restPath: string,
  searchParams: string,
) => {
  const url = new URL(restPath, `https://${subdomain}${DOMAIN_SUFFIX}`);
  url.search = searchParams;
  return url;
};

export default auth((req) => {
  const pathname = req.nextUrl.pathname;
  const searchParams = req.nextUrl.search;
  const host = getHost(req);

  if (!host || !host.endsWith(DOMAIN_SUFFIX)) {
    if (!req.auth && !isAuthOnlyPath(pathname)) {
      const redirectHost = host ?? AUTH_HOST;
      return buildHostRedirect(
        req.nextUrl.protocol,
        redirectHost,
        `${pathname}${searchParams}`,
      );
    }

    return NextResponse.next();
  }

  const subdomain = getSubdomain(host);
  const isAuthHost = host === AUTH_HOST;
  const isSubdomain = !!subdomain && !RESERVED_SUBDOMAINS.has(subdomain);
  const pathSegments = getPathSegments(pathname);
  const authWorkspaceSlug = getAuthWorkspaceSlug(host, pathSegments);

  if (isAuthHost && authWorkspaceSlug) {
    const restPath = pathname.replace(`/${authWorkspaceSlug}`, "") || "/";
    const subdomainUrl = buildSubdomainUrl(
      authWorkspaceSlug,
      restPath,
      searchParams,
    );

    if (!req.auth) {
      const callbackUrl = `https://${authWorkspaceSlug}${DOMAIN_SUFFIX}${restPath}${searchParams}`;
      return buildAuthRedirect(callbackUrl);
    }

    return NextResponse.redirect(subdomainUrl);
  }

  if (isSubdomain && isAuthOnlyPath(pathname)) {
    const redirectUrl = buildAuthUrl(pathname, searchParams);

    if (pathname === "/") {
      redirectUrl.searchParams.set(
        "callbackUrl",
        `https://${subdomain}${DOMAIN_SUFFIX}/my-work`,
      );
    }

    return NextResponse.redirect(redirectUrl);
  }

  if (isSubdomain && !pathname.startsWith(`/${subdomain}`)) {
    if (!req.auth) {
      const callbackUrl = `https://${subdomain}${DOMAIN_SUFFIX}${pathname}${searchParams}`;
      return buildAuthRedirect(callbackUrl);
    }

    const nextPath =
      pathname === "/" ? `/${subdomain}/my-work` : `/${subdomain}${pathname}`;
    const rewriteUrl = new URL(nextPath, req.nextUrl);
    rewriteUrl.search = searchParams;
    return NextResponse.rewrite(rewriteUrl);
  }

  if (!req.auth && !isAuthOnlyPath(pathname)) {
    return buildAuthRedirect(`${pathname}${searchParams}`);
  }

  return NextResponse.next();
});

export const config = {
  matcher: [
    "/((?!verify|auth-callback|signup|api/auth|_next/static|images|_next/image|favicon*|ingest|unauthorized|manifest*|apple-icon*).*)",
  ],
};
