import { NextResponse } from "next/server";
import { auth } from "@/auth";

const publicPaths = new Set([
  "/",
  "/login",
  "/signup",
  "/sign-in",
  "/sign-up",
  "/auth-callback",
]);

const isPublicPath = (pathname: string) => {
  if (publicPaths.has(pathname)) {
    return true;
  }
  if (pathname.startsWith("/verify/")) {
    return true;
  }
  if (pathname.startsWith("/onboarding/join")) {
    return true;
  }
  return false;
};

export default auth((req) => {
  const pathname = req.nextUrl.pathname;
  if (!req.auth && !isPublicPath(pathname)) {
    const searchParams = req.nextUrl.search;
    const callBackUrl = `${pathname}${searchParams}`;
    const newUrl = new URL("/login", req.nextUrl.origin);
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
    "/((?!_next/static|images|_next/image|favicon*|logout|ingest|unauthorized|manifest*|apple-icon*).*)",
  ],
};
