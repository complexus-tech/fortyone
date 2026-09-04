import { headers } from "next/headers";
import { auth } from "@/auth";
import { getApiUrl } from "@/lib/api-url";

type RouteProps = {
  params: Promise<{ referenceId: string; workspaceSlug: string }>;
};

const UUID_PATTERN =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
const MAX_WORKSPACE_SLUG_LENGTH = 255;
const WORKSPACE_SLUG_PATTERN = /^[a-z0-9-]+$/;

export const GET = async (_request: Request, { params }: RouteProps) => {
  const session = await auth();
  if (!session?.user) return new Response(null, { status: 401 });

  const { referenceId, workspaceSlug } = await params;
  if (
    !UUID_PATTERN.test(referenceId) ||
    workspaceSlug.length === 0 ||
    workspaceSlug.length > MAX_WORKSPACE_SLUG_LENGTH ||
    !WORKSPACE_SLUG_PATTERN.test(workspaceSlug)
  ) {
    return new Response(null, { status: 400 });
  }

  const requestHeaders = await headers();
  const cookie = requestHeaders.get("cookie");
  const upstreamURL = new URL(
    `/workspaces/${encodeURIComponent(workspaceSlug)}/google-drive/files/${encodeURIComponent(referenceId)}/preview`,
    getApiUrl(),
  );
  const upstream = await fetch(upstreamURL, {
    cache: "no-store",
    credentials: "include",
    headers: cookie ? { cookie } : undefined,
  });
  if (!upstream.ok || !upstream.body) {
    return new Response(null, {
      headers: { "Cache-Control": "private, no-store" },
      status: upstream.status,
    });
  }

  const contentType = upstream.headers.get("content-type");
  if (!contentType?.startsWith("image/")) {
    return new Response(null, { status: 502 });
  }

  return new Response(upstream.body, {
    headers: {
      "Cache-Control": "private, no-store",
      "Content-Security-Policy": "sandbox; default-src 'none'",
      "Content-Type": contentType,
      "Cross-Origin-Resource-Policy": "same-origin",
      "Referrer-Policy": "no-referrer",
      "X-Content-Type-Options": "nosniff",
    },
    status: 200,
  });
};
