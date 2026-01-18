import { NextResponse } from "next/server";
import { auth } from "@/auth";

const CLOUD_HOST = "cloud.fortyone.app";
const DOMAIN_SUFFIX = ".fortyone.app";
const RESERVED_SUBDOMAINS = new Set(["cloud"]);
const CLOUD_ONLY_PREFIXES = new Set([
  "/signup",
  "/auth-callback",
  "/verify",
  "/onboarding",
  "/unauthorized",
  "/api/auth",
]);
const CLOUD_ONLY_SEGMENTS = new Set([
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

const isCloudOnlyPath = (pathname: string) =>
  pathname === "/" ||
  Array.from(CLOUD_ONLY_PREFIXES).some((prefix) => pathname.startsWith(prefix));

const getPathSegments = (pathname: string) =>
  pathname.split("/").filter(Boolean);

const getCloudWorkspaceSlug = (host: string, pathSegments: string[]) => {
  if (host !== CLOUD_HOST || pathSegments.length === 0) {
    return undefined;
  }

  return CLOUD_ONLY_SEGMENTS.has(pathSegments[0]) ? undefined : pathSegments[0];
};

const buildCloudUrl = (pathname: string, searchParams: string) => {
  const url = new URL(pathname, `https://${CLOUD_HOST}`);
  url.search = searchParams;
  return url;
};

const buildCloudRedirect = (callbackUrl: string) => {
  const url = new URL("/", `https://${CLOUD_HOST}`);
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
    if (!req.auth && !isCloudOnlyPath(pathname)) {
      const redirectHost = host ?? CLOUD_HOST;
      return buildHostRedirect(
        req.nextUrl.protocol,
        redirectHost,
        `${pathname}${searchParams}`,
      );
    }

    return NextResponse.next();
  }

  const subdomain = getSubdomain(host);
  const isCloudHost = host === CLOUD_HOST;
  const isSubdomain = !!subdomain && !RESERVED_SUBDOMAINS.has(subdomain);
  const pathSegments = getPathSegments(pathname);
  const cloudWorkspaceSlug = getCloudWorkspaceSlug(host, pathSegments);

  if (isCloudHost && cloudWorkspaceSlug) {
    const restPath = pathname.replace(`/${cloudWorkspaceSlug}`, "") || "/";
    const subdomainUrl = buildSubdomainUrl(
      cloudWorkspaceSlug,
      restPath,
      searchParams,
    );

    if (!req.auth) {
      const callbackUrl = `https://${cloudWorkspaceSlug}${DOMAIN_SUFFIX}${restPath}${searchParams}`;
      return buildCloudRedirect(callbackUrl);
    }

    return NextResponse.redirect(subdomainUrl);
  }

  if (isSubdomain && isCloudOnlyPath(pathname)) {
    const redirectUrl = buildCloudUrl(pathname, searchParams);

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
      return buildCloudRedirect(callbackUrl);
    }

    const nextPath =
      pathname === "/" ? `/${subdomain}/my-work` : `/${subdomain}${pathname}`;
    const rewriteUrl = new URL(nextPath, req.nextUrl);
    rewriteUrl.search = searchParams;
    return NextResponse.rewrite(rewriteUrl);
  }

  if (!req.auth && !isCloudOnlyPath(pathname)) {
    return buildCloudRedirect(`${pathname}${searchParams}`);
  }

  return NextResponse.next();
});

export const config = {
  matcher: [
    "/((?!verify|auth-callback|signup|api/auth|_next/static|images|_next/image|favicon*|ingest|unauthorized|manifest*|apple-icon*).*)",
  ],
};
