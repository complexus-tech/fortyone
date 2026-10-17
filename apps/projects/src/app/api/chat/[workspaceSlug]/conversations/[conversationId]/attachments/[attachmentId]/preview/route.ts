import { headers } from "next/headers";
import { auth } from "@/auth";
import { getApiUrl } from "@/lib/api-url";

export const runtime = "nodejs";

type RouteProps = {
  params: Promise<{
    attachmentId: string;
    conversationId: string;
    workspaceSlug: string;
  }>;
};

type ByteRange = { start: number; end: number | null } | { suffix: number };

const UUID_PATTERN =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
const WORKSPACE_SLUG_PATTERN = /^[a-z0-9-]{1,255}$/;
const MAX_PDF_BYTES = 25 * 1024 * 1024;
const MAX_AUDIO_BYTES = 20 * 1024 * 1024;
const FETCH_TIMEOUT_MS = 60_000;
const PRIVATE_HEADERS = {
  "Cache-Control": "private, no-store",
  "Content-Security-Policy": "sandbox; default-src 'none'",
  "Cross-Origin-Resource-Policy": "same-origin",
  "Referrer-Policy": "no-referrer",
  "X-Content-Type-Options": "nosniff",
};

const failure = (status: number) =>
  new Response(null, { status, headers: PRIVATE_HEADERS });

const fetchFailureStatus = (request: AbortSignal, upstream: AbortSignal) => {
  if (request.aborted) return 499;
  if (upstream.aborted) return 504;
  return 502;
};

const parseByteRange = (value: string): ByteRange | null => {
  // eslint-disable-next-line prefer-named-capture-group -- The app's ES2017 target requires numbered capture groups.
  const match = /^bytes=(\d*)-(\d*)$/.exec(value);
  if (!match || (!match[1] && !match[2])) return null;

  if (!match[1]) {
    const suffix = Number(match[2]);
    return Number.isSafeInteger(suffix) && suffix > 0 ? { suffix } : null;
  }

  const start = Number(match[1]);
  const end = match[2] ? Number(match[2]) : null;
  if (
    !Number.isSafeInteger(start) ||
    (end !== null && (!Number.isSafeInteger(end) || end < start))
  ) {
    return null;
  }
  return { start, end };
};

const storageTarget = (location: string | null) => {
  if (!location) return null;
  try {
    const url = new URL(location);
    if (url.username || url.password || url.hash) return null;
    if (url.protocol === "https:") return url;

    // Only the authenticated API can supply a target. Local S3-compatible
    // development may use loopback HTTP; remote storage always requires TLS.
    if (
      process.env.NODE_ENV !== "production" &&
      url.protocol === "http:" &&
      ["localhost", "127.0.0.1", "[::1]"].includes(url.hostname)
    ) {
      return url;
    }
  } catch {
    return null;
  }
  return null;
};

const responseLength = (
  response: Response,
  requestedRange: ByteRange | null,
  maxBytes: number,
) => {
  const rawLength = response.headers.get("content-length");
  const length = rawLength === null ? null : Number(rawLength);
  if (
    rawLength !== null &&
    (!/^\d+$/.test(rawLength) ||
      !Number.isSafeInteger(length) ||
      length === null ||
      length <= 0 ||
      length > maxBytes)
  ) {
    return null;
  }

  if (response.status === 200) return { length };
  const contentRange = response.headers.get("content-range");
  // eslint-disable-next-line prefer-named-capture-group -- The app's ES2017 target requires numbered capture groups.
  const match = /^bytes (\d+)-(\d+)\/(\d+)$/.exec(contentRange ?? "");
  if (!requestedRange || !match) return null;

  const start = Number(match[1]);
  const end = Number(match[2]);
  const total = Number(match[3]);
  if (
    !Number.isSafeInteger(total) ||
    total <= 0 ||
    total > maxBytes ||
    start > end ||
    end >= total
  ) {
    return null;
  }
  const expectedStart =
    "suffix" in requestedRange
      ? Math.max(0, total - requestedRange.suffix)
      : requestedRange.start;
  const expectedEnd =
    "suffix" in requestedRange
      ? total - 1
      : Math.min(requestedRange.end ?? total - 1, total - 1);
  const expectedLength = end - start + 1;
  if (
    start !== expectedStart ||
    end !== expectedEnd ||
    (length !== null && length !== expectedLength)
  ) {
    return null;
  }
  return { length: expectedLength, contentRange: contentRange ?? undefined };
};

export const GET = async (request: Request, { params }: RouteProps) => {
  const session = await auth();
  if (!session?.user) return failure(401);

  const { attachmentId, conversationId, workspaceSlug } = await params;
  if (
    !UUID_PATTERN.test(attachmentId) ||
    !UUID_PATTERN.test(conversationId) ||
    !WORKSPACE_SLUG_PATTERN.test(workspaceSlug)
  ) {
    return failure(400);
  }

  const previewMedia = new URL(request.url).searchParams.get("media");
  if (previewMedia !== null && previewMedia !== "audio") return failure(400);
  const audioPreview = previewMedia === "audio";
  const maxBytes = audioPreview ? MAX_AUDIO_BYTES : MAX_PDF_BYTES;
  const allowedMediaTypes = audioPreview
    ? ["audio/mp4", "audio/webm"]
    : ["application/pdf"];

  const rawRange = request.headers.get("range");
  const range = rawRange !== null ? parseByteRange(rawRange) : null;
  if (rawRange !== null && !range) return failure(416);
  if (request.signal.aborted) return failure(499);

  const signal = AbortSignal.any([
    request.signal,
    AbortSignal.timeout(FETCH_TIMEOUT_MS),
  ]);

  try {
    const cookie = (await headers()).get("cookie");
    const authorization = await fetch(
      new URL(
        `/workspaces/${encodeURIComponent(workspaceSlug)}/chat/conversations/${conversationId}/attachments/${attachmentId}`,
        getApiUrl(),
      ),
      {
        cache: "no-store",
        headers: cookie ? { cookie } : undefined,
        redirect: "manual",
        signal,
      },
    );
    await authorization.body?.cancel();

    if (authorization.status !== 307) {
      return failure(
        [401, 403, 404].includes(authorization.status)
          ? authorization.status
          : 502,
      );
    }

    const target = storageTarget(authorization.headers.get("location"));
    if (!target) return failure(502);

    // Do not forward application cookies, authorization, or upstream redirects
    // to object storage. The short-lived URL is the only storage credential.
    const upstream = await fetch(target, {
      cache: "no-store",
      credentials: "omit",
      headers: {
        "Accept-Encoding": "identity",
        ...(rawRange ? { Range: rawRange } : {}),
      },
      redirect: "manual",
      signal,
    });

    if (upstream.status === 416 && range) {
      await upstream.body?.cancel();
      const response = failure(416);
      const contentRange = upstream.headers.get("content-range");
      // eslint-disable-next-line prefer-named-capture-group -- The app's ES2017 target requires numbered capture groups.
      const total = /^bytes \*\/(\d+)$/.exec(contentRange ?? "")?.[1];
      if (total && Number(total) <= maxBytes) {
        response.headers.set("Content-Range", `bytes */${total}`);
      }
      return response;
    }

    const mediaType = upstream.headers
      .get("content-type")
      ?.split(";", 1)[0]
      .trim()
      .toLowerCase();
    const encoding = upstream.headers.get("content-encoding");
    const size = responseLength(upstream, range, maxBytes);
    if (
      ![200, 206].includes(upstream.status) ||
      !upstream.body ||
      !mediaType ||
      !allowedMediaTypes.includes(mediaType) ||
      (encoding !== null && encoding !== "identity") ||
      !size
    ) {
      await upstream.body?.cancel();
      return failure(502);
    }

    let received = 0;
    const body = upstream.body.pipeThrough(
      new TransformStream<Uint8Array, Uint8Array>({
        transform(chunk, controller) {
          received += chunk.byteLength;
          if (received > (size.length ?? maxBytes)) {
            throw new Error("Attachment preview exceeded its allowed size.");
          }
          controller.enqueue(chunk);
        },
        flush() {
          if (size.length !== null && received !== size.length) {
            throw new Error(
              "Attachment preview ended before the file was complete.",
            );
          }
        },
      }),
      { signal },
    );
    const fileName = audioPreview
      ? `recording.${mediaType === "audio/mp4" ? "m4a" : "webm"}`
      : "document.pdf";
    const responseHeaders = new Headers({
      ...PRIVATE_HEADERS,
      "Content-Type": mediaType,
      "Content-Disposition": `inline; filename="${fileName}"`,
    });
    if (size.length !== null)
      responseHeaders.set("Content-Length", String(size.length));
    if (size.contentRange)
      responseHeaders.set("Content-Range", size.contentRange);
    if (upstream.headers.get("accept-ranges") === "bytes")
      responseHeaders.set("Accept-Ranges", "bytes");

    return new Response(body, {
      headers: responseHeaders,
      status: upstream.status,
    });
  } catch {
    return failure(fetchFailureStatus(request.signal, signal));
  }
};
