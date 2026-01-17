import { NextResponse } from "next/server";
import { auth } from "@/auth";

export default auth((req) => {
  if (!req.auth && req.nextUrl.pathname !== "/") {
    const pathname = req.nextUrl.pathname;
    const searchParams = req.nextUrl.search;
    const callBackUrl = `${pathname}${searchParams}`;
    const newUrl = new URL("/", req.nextUrl.origin);
    newUrl.searchParams.set("callbackUrl", callBackUrl);
    return NextResponse.redirect(newUrl);
  }

  return NextResponse.next();
});

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - login (Auth routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     */
    "/((?!verify|auth-callback|signup|api/auth|_next/static|images|_next/image|favicon*|ingest|unauthorized|manifest*|apple-icon*).*)",
  ],
};
