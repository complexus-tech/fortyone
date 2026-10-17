import { headers } from "next/headers";
import { auth } from "@/auth";
import { getApiUrl } from "@/lib/api-url";

export const runtime = "nodejs";

const MAX_BYTES = 2 * 1024 * 1024;
const PRIVATE_HEADERS = {
  "Cache-Control": "private, no-store",
  "Cross-Origin-Resource-Policy": "same-origin",
  "Referrer-Policy": "no-referrer",
  "X-Content-Type-Options": "nosniff",
};
const fail = (status: number) =>
  new Response(null, { status, headers: PRIVATE_HEADERS });
type RouteProps = {
  params: Promise<{
    attachmentId: string;
    conversationId: string;
    workspaceSlug: string;
  }>;
};

export const GET = async (request: Request, { params }: RouteProps) => {
  const session = await auth();
  if (!session?.user) return fail(401);
  const { attachmentId, conversationId, workspaceSlug } = await params;
  const uuid =
    /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
  if (
    !uuid.test(attachmentId) ||
    !uuid.test(conversationId) ||
    !/^[a-z0-9-]{1,255}$/.test(workspaceSlug)
  )
    return fail(400);
  const signal = AbortSignal.any([request.signal, AbortSignal.timeout(15_000)]);
  const cookie = (await headers()).get("cookie");
  const refresh = new URL(request.url).searchParams.get("refresh") === "1";
  const etag = request.headers.get("if-none-match");
  try {
    const access = await fetch(
      new URL(
        `/workspaces/${encodeURIComponent(workspaceSlug)}/chat/conversations/${conversationId}/attachments/${attachmentId}/thumbnail${refresh ? "?refresh=1" : ""}`,
        getApiUrl(),
      ),
      {
        cache: "no-store",
        redirect: "manual",
        signal,
        headers: {
          ...(cookie ? { cookie } : {}),
          ...(etag ? { "If-None-Match": etag } : {}),
        },
      },
    );
    await access.body?.cancel();
    if (access.status === 202)
      return new Response(null, {
        status: 202,
        headers: { ...PRIVATE_HEADERS, "Retry-After": "2" },
      });
    if (access.status === 304 && etag && access.headers.get("etag") === etag)
      return new Response(null, {
        status: 304,
        headers: { ...PRIVATE_HEADERS, ETag: etag },
      });
    if (access.status !== 307)
      return fail(
        [401, 403, 404, 422].includes(access.status) ? access.status : 502,
      );
    const location = access.headers.get("location");
    if (!location) return fail(502);
    const target = new URL(location);
    const local =
      process.env.NODE_ENV !== "production" &&
      target.protocol === "http:" &&
      ["localhost", "127.0.0.1", "[::1]"].includes(target.hostname);
    if (
      target.username ||
      target.password ||
      target.hash ||
      (target.protocol !== "https:" && !local)
    )
      return fail(502);
    const upstream = await fetch(target, {
      cache: "no-store",
      credentials: "omit",
      redirect: "manual",
      signal,
      headers: { "Accept-Encoding": "identity" },
    });
    const length = Number(upstream.headers.get("content-length"));
    if (
      upstream.status !== 200 ||
      upstream.headers.get("content-type")?.split(";", 1)[0].trim() !==
        "image/jpeg" ||
      !upstream.body ||
      (upstream.headers.has("content-length") &&
        (!Number.isSafeInteger(length) || length < 1 || length > MAX_BYTES)) ||
      (upstream.headers.has("content-encoding") &&
        upstream.headers.get("content-encoding") !== "identity")
    ) {
      await upstream.body?.cancel();
      return fail(502);
    }
    let received = 0;
    const body = upstream.body.pipeThrough(
      new TransformStream<Uint8Array, Uint8Array>({
        transform(chunk, controller) {
          received += chunk.byteLength;
          if (received > MAX_BYTES)
            throw new Error("Thumbnail exceeded its allowed size");
          controller.enqueue(chunk);
        },
        flush() {
          if (!received || (length && received !== length))
            throw new Error("Incomplete thumbnail");
        },
      }),
      { signal },
    );
    const responseHeaders = new Headers({
      ...PRIVATE_HEADERS,
      "Content-Type": "image/jpeg",
      "Content-Disposition": 'inline; filename="preview.jpg"',
    });
    const version = access.headers.get("etag");
    if (version && /^"[a-f0-9]{64}"$/.test(version))
      responseHeaders.set("ETag", version);
    return new Response(body, { headers: responseHeaders });
  } catch {
    if (request.signal.aborted) return fail(499);
    return fail(signal.aborted ? 504 : 502);
  }
};
