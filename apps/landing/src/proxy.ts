import { NextResponse } from "next/server";
import { auth } from "@/auth";

// export default auth((req) => {
//   const isLoggedIn = Boolean(req.auth);
//   const pathname = req.nextUrl.pathname;
//   if (!isLoggedIn && req.nextUrl.pathname !== "/login") {
//     const searchParams = req.nextUrl.search;
//     const callBackUrl = `${pathname}${searchParams}`;
//     const newUrl = new URL("/login", req.nextUrl.origin);
//     newUrl.searchParams.set("callbackUrl", callBackUrl);
//     return NextResponse.redirect(newUrl);
//   }
// });

export default auth((req) => {
  const { pathname, search } = req.nextUrl;
  const isLoggedIn = Boolean(req.auth);

  /**
   * 1. Docs proxy — MUST run first
   * Docs are public and should never trigger auth redirects
   */
  if (pathname.startsWith("/docs")) {
    const rewrittenPath = pathname.replace(/^\/docs/, "") || "/";

    const url = new URL(rewrittenPath, "https://docs.fortyone.app");
    url.search = search;

    const response = NextResponse.rewrite(url);
    response.headers.set("x-from-proxy", "fortyone-docs");
    return response;
  }

  /**
   * 2. Auth gate
   */
  if (!isLoggedIn && pathname !== "/login") {
    const callbackUrl = `${pathname}${search}`;
    const loginUrl = new URL("/login", req.nextUrl.origin);
    loginUrl.searchParams.set("callbackUrl", callbackUrl);
    return NextResponse.redirect(loginUrl);
  }

  /**
   * 3. Everything else
   */
  return NextResponse.next();
});

export const config = {
  matcher: [
    "/docs/:path*",

    "/onboarding/account",
    "/onboarding/create",
    "/onboarding/invite",
    "/onboarding/welcome",

    "/((?!|login|[^/]+|[^/]+/[^/]+|_next/static|images|_next/image|favicon*).*)",
  ],
};
