import { NextResponse } from "next/server";
import { auth } from "@/auth";

export default auth((req) => {
  const pathname = req.nextUrl.pathname;
  const searchParams = req.nextUrl.search;
  const host = req.headers.get("host")?.split(":")[0];
  const isFortyoneHost = host?.endsWith(".fortyone.app");
  const reservedSubdomains = new Set(["cloud"]);
  const cloudOnlyPrefixes = new Set([
    "/signup",
    "/auth-callback",
    "/verify",
    "/onboarding",
    "/unauthorized",
    "/logout",
  ]);
  const isCloudOnlyPath =
    pathname === "/" ||
    Array.from(cloudOnlyPrefixes).some((prefix) => pathname.startsWith(prefix));

  if (host && isFortyoneHost) {
    const subdomain = host.replace(".fortyone.app", "");
    const isSubdomain =
      subdomain.length > 0 && !reservedSubdomains.has(subdomain);

    if (isSubdomain && isCloudOnlyPath) {
      const redirectUrl = new URL(pathname, "https://cloud.fortyone.app");
      redirectUrl.search = searchParams;

      if (pathname === "/") {
        const callbackUrl = `/${subdomain}/my-work`;
        redirectUrl.searchParams.set("callbackUrl", callbackUrl);
      }

      return NextResponse.redirect(redirectUrl);
    }

    if (isSubdomain && !pathname.startsWith(`/${subdomain}`)) {
      const nextPath =
        pathname === "/" ? `/${subdomain}/my-work` : `/${subdomain}${pathname}`;
      const rewriteUrl = new URL(nextPath, req.nextUrl);
      rewriteUrl.search = searchParams;
      return NextResponse.rewrite(rewriteUrl);
    }
  }

  if (!req.auth && !isCloudOnlyPath) {
    const callBackUrl = `${pathname}${searchParams}`;
    const newUrl = new URL("/", "https://cloud.fortyone.app");
    newUrl.searchParams.set("callbackUrl", callBackUrl);
    return NextResponse.redirect(newUrl);
  }

  return NextResponse.next();
});

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     */
    "/((?!_next/static|images|_next/image|favicon*|ingest|manifest*|apple-icon*).*)",
  ],
};
