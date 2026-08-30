import type { IncomingHttpHeaders, IncomingMessage } from "node:http";
import { request as httpRequest } from "node:http";
import { request as httpsRequest } from "node:https";
import type { LookupFunction } from "node:net";
import type { ResolvedTarget, ResolveHostname } from "./target-policy";
import {
  MetadataFetchError,
  parseMetadataUrl,
  resolveHostname as defaultResolveHostname,
  resolvePublicTarget,
} from "./target-policy";

export const MAX_METADATA_RESPONSE_BYTES = 1024 * 1024;
export const MAX_METADATA_IMAGE_RESPONSE_BYTES = 2 * 1024 * 1024;
export const MAX_METADATA_REDIRECTS = 3;
export const METADATA_FETCH_TIMEOUT_MS = 8_000;

const PINNED_ADDRESS_ATTEMPT_TIMEOUT_MS = 250;
const HTML_ACCEPT_HEADER = "text/html, application/xhtml+xml;q=0.9";
const IMAGE_ACCEPT_HEADER =
  "image/avif, image/webp, image/apng, image/png, image/jpeg, image/gif, image/x-icon, image/vnd.microsoft.icon, image/bmp;q=0.8";
const ALLOWED_HTML_CONTENT_TYPES = new Set([
  "application/xhtml+xml",
  "text/html",
]);
const ALLOWED_IMAGE_CONTENT_TYPES = new Set([
  "image/apng",
  "image/avif",
  "image/bmp",
  "image/gif",
  "image/jpeg",
  "image/png",
  "image/vnd.microsoft.icon",
  "image/webp",
  "image/x-icon",
]);
const REDIRECT_STATUS_CODES = new Set([301, 302, 303, 307, 308]);

type MetadataHttpResponse =
  | {
      kind: "document";
      body: Buffer;
      contentType?: string;
    }
  | {
      kind: "redirect";
      location: string;
    };

type RequestMetadataDocument = (
  target: ResolvedTarget,
  signal: AbortSignal,
) => Promise<MetadataHttpResponse>;

type MetadataResponsePolicy = {
  allowedContentTypes: ReadonlySet<string>;
  maxResponseBytes: number;
  resourceName: string;
};

type ReadableHttpResponse = AsyncIterable<Uint8Array> & {
  destroy: () => void;
  headers: IncomingHttpHeaders;
  statusCode?: number;
};

type FetchPublicHtmlOptions = {
  request?: RequestMetadataDocument;
  resolveHostname?: ResolveHostname;
  signal?: AbortSignal;
  timeoutMs?: number;
};

const getHeader = (
  headers: IncomingHttpHeaders,
  name: keyof IncomingHttpHeaders,
) => {
  const value = headers[name];
  return Array.isArray(value) ? value[0] : value;
};

const parseContentLength = (headers: IncomingHttpHeaders) => {
  const contentLength = getHeader(headers, "content-length");
  if (contentLength === undefined) return undefined;

  const parsedLength = Number(contentLength);
  if (!Number.isSafeInteger(parsedLength) || parsedLength < 0) {
    throw new MetadataFetchError(
      "unsupported-response",
      "Metadata response has an invalid content length.",
    );
  }
  return parsedLength;
};

const readPublicHttpResponse = async (
  response: ReadableHttpResponse,
  policy: MetadataResponsePolicy,
): Promise<MetadataHttpResponse> => {
  const statusCode = response.statusCode;
  if (statusCode && REDIRECT_STATUS_CODES.has(statusCode)) {
    const location = getHeader(response.headers, "location");
    response.destroy();
    if (!location) {
      throw new MetadataFetchError(
        "upstream-error",
        "Metadata redirect has no destination.",
      );
    }
    return { kind: "redirect", location };
  }

  if (!statusCode || statusCode < 200 || statusCode >= 300) {
    response.destroy();
    throw new MetadataFetchError(
      "upstream-error",
      "Metadata host returned an unsuccessful response.",
    );
  }

  const contentType = getHeader(response.headers, "content-type")
    ?.split(";", 1)[0]
    ?.trim()
    .toLowerCase();
  if (!contentType || !policy.allowedContentTypes.has(contentType)) {
    response.destroy();
    throw new MetadataFetchError(
      "unsupported-response",
      `Metadata host did not return a supported ${policy.resourceName}.`,
    );
  }

  const contentEncoding = getHeader(response.headers, "content-encoding")
    ?.trim()
    .toLowerCase();
  if (contentEncoding && contentEncoding !== "identity") {
    response.destroy();
    throw new MetadataFetchError(
      "unsupported-response",
      "Metadata response encoding is unsupported.",
    );
  }

  const declaredLength = parseContentLength(response.headers);
  if (
    declaredLength !== undefined &&
    declaredLength > policy.maxResponseBytes
  ) {
    response.destroy();
    throw new MetadataFetchError(
      "response-too-large",
      "Metadata response is too large.",
    );
  }

  const chunks: Buffer[] = [];
  let receivedBytes = 0;
  for await (const chunk of response) {
    const buffer = Buffer.from(chunk);
    receivedBytes += buffer.byteLength;
    if (receivedBytes > policy.maxResponseBytes) {
      response.destroy();
      throw new MetadataFetchError(
        "response-too-large",
        "Metadata response is too large.",
      );
    }
    chunks.push(buffer);
  }

  return {
    body: Buffer.concat(chunks, receivedBytes),
    contentType,
    kind: "document",
  };
};

export const readMetadataHttpResponse = async (
  response: ReadableHttpResponse,
): Promise<MetadataHttpResponse> => {
  const result = await readPublicHttpResponse(response, {
    allowedContentTypes: ALLOWED_HTML_CONTENT_TYPES,
    maxResponseBytes: MAX_METADATA_RESPONSE_BYTES,
    resourceName: "HTML document",
  });
  return result.kind === "redirect"
    ? result
    : { body: result.body, kind: "document" };
};

export const readMetadataImageHttpResponse = (response: ReadableHttpResponse) =>
  readPublicHttpResponse(response, {
    allowedContentTypes: ALLOWED_IMAGE_CONTENT_TYPES,
    maxResponseBytes: MAX_METADATA_IMAGE_RESPONSE_BYTES,
    resourceName: "raster image",
  });

const requestPinnedDocument = (
  target: ResolvedTarget,
  address: ResolvedTarget["addresses"][number],
  signal: AbortSignal,
  readResponse: (
    response: ReadableHttpResponse,
  ) => Promise<MetadataHttpResponse>,
  accept: string,
  attemptTimeoutMs?: number,
): Promise<MetadataHttpResponse> => {
  const { url } = target;
  const request = url.protocol === "https:" ? httpsRequest : httpRequest;
  const pinnedLookup: LookupFunction = (_hostname, options, callback) => {
    if (options.all) {
      callback(null, [address]);
      return;
    }
    callback(null, address.address, address.family);
  };

  return new Promise<MetadataHttpResponse>((resolve, reject) => {
    let attemptTimeout: ReturnType<typeof setTimeout> | undefined;
    let responseStarted = false;
    const outgoingRequest = request(
      url,
      {
        agent: false,
        family: address.family,
        headers: {
          accept,
          "accept-encoding": "identity",
          connection: "close",
          "user-agent": "FortyOne-Metadata/1.0",
        },
        lookup: pinnedLookup,
        signal,
      },
      (response: IncomingMessage) => {
        responseStarted = true;
        if (attemptTimeout) clearTimeout(attemptTimeout);
        void readResponse(response).then(resolve, reject);
      },
    );
    if (attemptTimeoutMs !== undefined) {
      attemptTimeout = setTimeout(() => {
        if (!responseStarted) {
          outgoingRequest.destroy(
            new Error(
              `Metadata address did not respond within ${attemptTimeoutMs}ms.`,
            ),
          );
        }
      }, attemptTimeoutMs);
    }
    outgoingRequest.once("error", (error) => {
      if (attemptTimeout) clearTimeout(attemptTimeout);
      reject(error);
    });
    outgoingRequest.end();
  });
};

const requestPublicDocument = async (
  target: ResolvedTarget,
  signal: AbortSignal,
  readResponse: (
    response: ReadableHttpResponse,
  ) => Promise<MetadataHttpResponse>,
  accept: string,
): Promise<MetadataHttpResponse> => {
  if (target.addresses.length === 0) {
    throw new MetadataFetchError(
      "upstream-error",
      "Metadata host did not resolve.",
    );
  }

  const requestAddress = async (
    index: number,
    lastError?: Error,
  ): Promise<MetadataHttpResponse> => {
    if (index >= target.addresses.length) {
      throw new MetadataFetchError(
        "upstream-error",
        "Every resolved metadata address failed.",
        { cause: lastError },
      );
    }
    const address = target.addresses[index];
    try {
      return await requestPinnedDocument(
        target,
        address,
        signal,
        readResponse,
        accept,
        index < target.addresses.length - 1
          ? PINNED_ADDRESS_ATTEMPT_TIMEOUT_MS
          : undefined,
      );
    } catch (error) {
      if (signal.aborted) {
        throw signal.reason instanceof Error
          ? signal.reason
          : new MetadataFetchError("upstream-error", "Metadata fetch aborted.");
      }
      if (error instanceof MetadataFetchError) throw error;
      const addressError =
        error instanceof Error
          ? error
          : new Error("Metadata address request failed.", { cause: error });
      return requestAddress(index + 1, addressError);
    }
  };

  return requestAddress(0);
};

export const requestMetadataDocument: RequestMetadataDocument = (
  target,
  signal,
) =>
  requestPublicDocument(
    target,
    signal,
    readMetadataHttpResponse,
    HTML_ACCEPT_HEADER,
  );

const requestMetadataImage: RequestMetadataDocument = (target, signal) =>
  requestPublicDocument(
    target,
    signal,
    readMetadataImageHttpResponse,
    IMAGE_ACCEPT_HEADER,
  );

const waitForAbort = <T>(promise: Promise<T>, signal: AbortSignal) =>
  new Promise<T>((resolve, reject) => {
    const handleAbort = () => {
      reject(
        signal.reason instanceof Error
          ? signal.reason
          : new MetadataFetchError("upstream-error", "Metadata fetch aborted."),
      );
    };
    if (signal.aborted) {
      handleAbort();
      return;
    }

    signal.addEventListener("abort", handleAbort, { once: true });
    promise.then(
      (value) => {
        signal.removeEventListener("abort", handleAbort);
        resolve(value);
      },
      (error: unknown) => {
        signal.removeEventListener("abort", handleAbort);
        reject(
          error instanceof Error
            ? error
            : new MetadataFetchError(
                "upstream-error",
                "Metadata fetch failed.",
                { cause: error },
              ),
        );
      },
    );
  });

const fetchPublicResource = async (
  input: string | URL,
  options: FetchPublicHtmlOptions,
  defaultRequest: RequestMetadataDocument,
) => {
  const request = options.request ?? defaultRequest;
  const resolveHostname = options.resolveHostname ?? defaultResolveHostname;
  const timeoutMs = options.timeoutMs ?? METADATA_FETCH_TIMEOUT_MS;
  const controller = new AbortController();
  const handleCallerAbort = () => {
    const reason = options.signal?.reason;
    controller.abort(
      reason instanceof Error
        ? reason
        : new MetadataFetchError("upstream-error", "Metadata fetch aborted."),
    );
  };
  if (options.signal?.aborted) handleCallerAbort();
  else
    options.signal?.addEventListener("abort", handleCallerAbort, {
      once: true,
    });

  const timeout = setTimeout(() => {
    controller.abort(
      new MetadataFetchError("timeout", "Metadata fetch timed out."),
    );
  }, timeoutMs);

  try {
    const fetchRedirect = async (
      currentUrl: URL,
      redirectCount: number,
    ): Promise<{ body: Buffer; contentType?: string; finalUrl: URL }> => {
      const target = await waitForAbort(
        resolvePublicTarget(currentUrl, resolveHostname),
        controller.signal,
      );
      const result = await waitForAbort(
        request(target, controller.signal),
        controller.signal,
      );
      if (result.kind === "document") {
        return {
          body: result.body,
          contentType: result.contentType,
          finalUrl: currentUrl,
        };
      }
      if (redirectCount >= MAX_METADATA_REDIRECTS) {
        throw new MetadataFetchError(
          "redirect-limit",
          "Metadata host redirected too many times.",
        );
      }

      const nextUrl = parseMetadataUrl(new URL(result.location, currentUrl));
      return fetchRedirect(nextUrl, redirectCount + 1);
    };

    return await fetchRedirect(parseMetadataUrl(input), 0);
  } finally {
    clearTimeout(timeout);
    options.signal?.removeEventListener("abort", handleCallerAbort);
  }
};

export const fetchPublicHtml = async (
  input: string | URL,
  options: FetchPublicHtmlOptions = {},
) => {
  const { body, finalUrl } = await fetchPublicResource(
    input,
    options,
    requestMetadataDocument,
  );
  return { body, finalUrl };
};

export const fetchPublicImage = async (
  input: string | URL,
  options: FetchPublicHtmlOptions = {},
) => {
  const result = await fetchPublicResource(
    input,
    options,
    requestMetadataImage,
  );
  if (!result.contentType) {
    throw new MetadataFetchError(
      "unsupported-response",
      "Metadata host did not return a supported raster image.",
    );
  }
  return {
    body: result.body,
    contentType: result.contentType,
    finalUrl: result.finalUrl,
  };
};
