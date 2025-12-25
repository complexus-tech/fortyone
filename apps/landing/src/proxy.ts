import { NextResponse } from "next/server";
import { auth } from "@/auth";

export default auth((req) => {
  const isLoggedIn = Boolean(req.auth);
  const pathname = req.nextUrl.pathname;
  if (!isLoggedIn && req.nextUrl.pathname !== "/login") {
    const searchParams = req.nextUrl.search;
    const callBackUrl = `${pathname}${searchParams}`;
    const newUrl = new URL("/login", req.nextUrl.origin);
    newUrl.searchParams.set("callbackUrl", callBackUrl);
    return NextResponse.redirect(newUrl);
  }
});

export const config = {
  matcher: [
    "/onboarding/account",
    "/onboarding/create",
    "/onboarding/invite",
    "/onboarding/welcome",
    "/((?!|login|[^/]+|[^/]+/[^/]+|_next/static|images|_next/image|favicon*).*)",
  ],
};
