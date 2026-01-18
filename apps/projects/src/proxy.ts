import { NextResponse } from "next/server";
import { auth } from "@/auth";

const AUTH_HOST = "cloud.fortyone.app";
const DOMAIN_SUFFIX = ".fortyone.app";
const RESERVED_SUBDOMAINS = new Set(["cloud"]);
const AUTH_ONLY_PREFIXES = new Set([
  "/signup",
  "/auth-callback",
  "/verify",
  "/onboarding",
  "/unauthorized",
]);

const getHost = (req: Request) => req.headers.get("host")?.split(":")[0];

const getSubdomain = (host?: string) =>
  host?.endsWith(DOMAIN_SUFFIX) ? host.replace(DOMAIN_SUFFIX, "") : undefined;

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
  protocol: string,
  host: string,
  callbackUrl: string,
) => {
  const baseUrl = `${protocol}//${host}`;
  const url = new URL("/", baseUrl);
  url.searchParams.set("callbackUrl", callbackUrl);
  return NextResponse.redirect(url);
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
  const isSubdomain = !!subdomain && !RESERVED_SUBDOMAINS.has(subdomain);

  if (isSubdomain && isAuthOnlyPath(pathname)) {
    if (pathname === "/" && req.auth) {
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

  if (isSubdomain && !req.auth) {
    const callbackUrl = `https://${subdomain}${DOMAIN_SUFFIX}${pathname}${searchParams}`;
    return buildAuthRedirect(callbackUrl);
  }

  if (!req.auth && !isAuthOnlyPath(pathname)) {
    return buildAuthRedirect(`${pathname}${searchParams}`);
  }

  return NextResponse.next();
});

export const config = {
  matcher: [
    "/((?!api/auth|_next/static|images|_next/image|favicon*|ingest|manifest*|apple-icon*).*)",
  ],
};
