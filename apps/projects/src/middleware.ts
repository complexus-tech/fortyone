import { NextResponse } from "next/server";
import { auth } from "@/auth";
import { changeWorkspace } from "@/components/shared/sidebar/actions";

export default auth(async (req) => {
  const hostname = req.headers.get("host") || "";
  const subdomain = hostname.split(".")[0];

  if (!req.auth && req.nextUrl.pathname !== "/login") {
    const pathname = req.nextUrl.pathname;
    const searchParams = req.nextUrl.search;
    const callBackUrl = `${pathname}${searchParams}`;
    const newUrl = new URL("/login", req.nextUrl.origin);
    newUrl.searchParams.set("callbackUrl", callBackUrl);
    return NextResponse.redirect(newUrl);
  }

  // Add workspace access validation
  if (req.auth && req.nextUrl.pathname !== "/login") {
    const workspace = req.auth.workspaces.find(
      (w) => w.name.toLowerCase() === subdomain.toLowerCase(),
    );

    // If user doesn't have access to this workspace subdomain
    if (!workspace) {
      // Redirect to last used workspace
      // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition -- ok for now
      if (req.auth.activeWorkspace) {
        const redirectUrl = new URL(
          req.nextUrl.pathname,
          `https://${req.auth.activeWorkspace.name}.${process.env.NEXT_PUBLIC_DOMAIN}`,
        );
        redirectUrl.search = req.nextUrl.search;
        return NextResponse.redirect(redirectUrl);
      }

      // If no active workspace, use first available workspace
      if (req.auth.workspaces.length > 0) {
        const redirectUrl = new URL(
          req.nextUrl.pathname,
          `https://${req.auth.workspaces[0].name}.${process.env.NEXT_PUBLIC_DOMAIN}`,
        );
        redirectUrl.search = req.nextUrl.search;
        return NextResponse.redirect(redirectUrl);
      }

      // If user has no workspaces, redirect to error page
      return NextResponse.redirect(new URL("/no-access", req.nextUrl.origin));
    }

    // If workspace exists but is different from active workspace
    if (workspace.id !== req.auth.activeWorkspace.id) {
      await changeWorkspace(workspace.id);
      return NextResponse.next();
    }
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
    "/((?!login|_next/static|images|_next/image|favicon*|signup).*)",
  ],
};
