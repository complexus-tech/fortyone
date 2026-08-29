import { type NextRequest, NextResponse } from "next/server";
import { getSessionFromRequest } from "auth";
import {
  getCanonicalPublicPath,
  getInternalPublicPath,
  isPublicPath,
} from "./public-portal-routes";
import { isSlackMinimalLinkPreview } from "./story-link-preview";
import {
  getFeedbackWidgetFrameAncestors,
  isValidFeedbackWidgetParent,
} from "./modules/feedback-widget/embed-security";

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

const buildHostRedirect = (requestUrl: string, callbackUrl: string) => {
  const url = new URL("/", requestUrl);
  url.searchParams.set("callbackUrl", callbackUrl);
  return NextResponse.redirect(url);
};

const FEEDBACK_WIDGET_EMBED_PREFIX = "/embed/feedback/";
const FEEDBACK_PORTAL_SLUG_PATTERN =
  /^(?=.{3,255}$)[a-z0-9](?:[a-z0-9-]*[a-z0-9])$/;

const protectFeedbackWidgetEmbed = (req: NextRequest) => {
  const encodedSlug = req.nextUrl.pathname.slice(
    FEEDBACK_WIDGET_EMBED_PREFIX.length,
  );
  let portalSlug = "";
  try {
    portalSlug = decodeURIComponent(encodedSlug);
  } catch {
    // The invalid slug is handled by the same fail-closed response below.
  }
  const parentOrigin = req.nextUrl.searchParams.get("parentOrigin");
  const deny = (status = 404) =>
    new NextResponse(status === 404 ? "Not found" : "Widget unavailable", {
      headers: {
        "Cache-Control": "no-store",
        "Content-Security-Policy": "frame-ancestors 'none'",
      },
      status,
    });

  if (
    !FEEDBACK_PORTAL_SLUG_PATTERN.test(portalSlug) ||
    !isValidFeedbackWidgetParent(parentOrigin)
  ) {
    return deny();
  }

  const next = NextResponse.next();
  next.headers.set("Cache-Control", "no-store");
  next.headers.set(
    "Content-Security-Policy",
    getFeedbackWidgetFrameAncestors(parentOrigin),
  );
  return next;
};

export default async function proxy(req: NextRequest) {
  const pathname = req.nextUrl.pathname;
  const searchParams = req.nextUrl.search;
  const hostname = getHostname(req);
  const userAgent = req.headers.get("user-agent");

  if (isSlackMinimalLinkPreview(pathname, userAgent)) {
    const previewUrl = req.nextUrl.clone();
    previewUrl.pathname = "/api/story-link-preview";
    previewUrl.search = "";
    return NextResponse.rewrite(previewUrl);
  }

  if (pathname.startsWith(FEEDBACK_WIDGET_EMBED_PREFIX)) {
    return protectFeedbackWidgetEmbed(req);
  }

  const user = await getSessionFromRequest(req);
  const isAuthenticated = Boolean(user);
  const subdomain = getSubdomain(hostname);
  const isWorkspaceSubdomain =
    typeof subdomain === "string" && !RESERVED_SUBDOMAINS.has(subdomain);

  if (isWorkspaceSubdomain) {
    const canonicalPublicPath = getCanonicalPublicPath(pathname, subdomain);
    if (canonicalPublicPath) {
      const canonicalUrl = req.nextUrl.clone();
      canonicalUrl.pathname = canonicalPublicPath;
      return NextResponse.redirect(canonicalUrl);
    }

    const internalPublicPath = getInternalPublicPath(pathname, subdomain);
    if (internalPublicPath) {
      const internalUrl = req.nextUrl.clone();
      internalUrl.pathname = internalPublicPath;
      return NextResponse.rewrite(internalUrl);
    }
  }

  if (isPublicPath(pathname)) {
    return NextResponse.next();
  }

  if (!isFortyOneHost(hostname)) {
    if (!isAuthenticated && !isAuthOnlyPath(pathname)) {
      return buildHostRedirect(req.url, `${pathname}${searchParams}`);
    }

    return NextResponse.next();
  }

  if (isWorkspaceSubdomain && isAuthOnlyPath(pathname)) {
    if (pathname === "/" && isAuthenticated) {
      const rewriteUrl = new URL(`/${subdomain}/my-work`, req.nextUrl);
      rewriteUrl.search = searchParams;
      return NextResponse.rewrite(rewriteUrl);
    }

    const redirectUrl = buildAuthUrl(pathname, searchParams);

    if (pathname === "/") {
      if (!redirectUrl.searchParams.has("callbackUrl")) {
        redirectUrl.searchParams.set(
          "callbackUrl",
          `https://${subdomain}${DOMAIN_SUFFIX}/my-work`,
        );
      }
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
    "/((?!api|_next/static|images|integrations|_next/image|favicon*|ingest|manifest*|apple-icon*).*)",
  ],
};
