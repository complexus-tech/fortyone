import { NextResponse } from "next/server";
import { auth } from "@/auth";

function extractSubdomain(host: string): string | null {
  const ROOT_DOMAIN = process.env.NEXT_PUBLIC_DOMAIN || 'fortyone.lc';

  // Remove port if present
  const [hostname] = host.split(':');

  // Remove www prefix if present
  const cleanHost = hostname.replace(/^www\./, '');

  // If it's exactly the root domain, no subdomain
  if (cleanHost === ROOT_DOMAIN) {
    return null;
  }

  // Extract subdomain from domain
  const parts = cleanHost.split('.');
  const rootParts = ROOT_DOMAIN.split('.');

  // If the domain doesn't end with our root domain, it's not a subdomain
  if (!cleanHost.endsWith(ROOT_DOMAIN)) {
    return null;
  }

  // Calculate subdomain parts
  const subdomainParts = parts.slice(0, parts.length - rootParts.length);

  // If no subdomain parts, it's the root domain
  if (subdomainParts.length === 0) {
    return null;
  }

  return subdomainParts.join('.');
}

export default auth((req) => {
  const host = req.headers.get('host') || '';
  const url = req.nextUrl.clone();

  // Extract subdomain from host
  const subdomain = extractSubdomain(host);

  // If subdomain exists, rewrite to internal workspace route
  if (subdomain) {
    // Transform subdomain.workspace.com/path to /workspace/path
    url.pathname = `/${subdomain}${url.pathname === '/' ? '' : url.pathname}`;

    // Add workspace context headers for downstream use
    const response = NextResponse.rewrite(url);
    response.headers.set('x-workspace-slug', subdomain);

    // Handle auth redirect for subdomains
    if (!req.auth && req.nextUrl.pathname !== "/") {
      const pathname = req.nextUrl.pathname;
      const searchParams = req.nextUrl.search;
      const callBackUrl = `${pathname}${searchParams}`;
      const newUrl = new URL("/", req.nextUrl.origin);
      newUrl.searchParams.set("callbackUrl", callBackUrl);
      return NextResponse.redirect(newUrl);
    }

    return response;
  }

  // Original auth logic for root domain
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
