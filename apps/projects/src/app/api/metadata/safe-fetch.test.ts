/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { createServer } from "node:http";
import type { AddressInfo } from "node:net";
import { Readable } from "node:stream";
import {
  fetchPublicImage,
  fetchPublicHtml,
  MAX_METADATA_IMAGE_RESPONSE_BYTES,
  MAX_METADATA_REDIRECTS,
  MAX_METADATA_RESPONSE_BYTES,
  readMetadataHttpResponse,
  readMetadataImageHttpResponse,
  requestMetadataDocument,
} from "./safe-fetch";
import { isPublicIpAddress, parseMetadataUrl } from "./target-policy";

const PUBLIC_ADDRESS = {
  address: "93.184.216.34",
  family: 4 as const,
};

const createResponse = ({
  body = Buffer.from("<html></html>"),
  headers = { "content-type": "text/html; charset=utf-8" },
  statusCode = 200,
}: {
  body?: Buffer | Buffer[];
  headers?: Record<string, string>;
  statusCode?: number;
} = {}) => {
  const chunks = Array.isArray(body) ? body : [body];
  const response = Object.assign(Readable.from(chunks), {
    headers,
    statusCode,
  });
  const destroy = jest.spyOn(response, "destroy");
  return { destroy, response };
};

describe("metadata target policy", () => {
  it.each([
    "ftp://example.com/file",
    "https://user:password@example.com",
    "https://example.com:8443",
    "http://localhost",
    "http://service.localhost",
    "not a URL",
  ])("rejects unsafe URL %s", (url) => {
    expect(() => parseMetadataUrl(url)).toThrow();
  });

  it.each([
    "https://example.com",
    "https://example.com:443/path",
    "http://example.com:80/path",
  ])("accepts public HTTP URL %s", (url) => {
    expect(parseMetadataUrl(url).href).toBe(new URL(url).href);
  });

  it.each([
    "0.0.0.0",
    "10.0.0.1",
    "100.64.0.1",
    "127.0.0.1",
    "169.254.169.254",
    "172.16.0.1",
    "192.168.1.1",
    "192.0.2.1",
    "198.18.0.1",
    "198.51.100.1",
    "203.0.113.1",
    "224.0.0.1",
    "255.255.255.255",
    "::",
    "::1",
    "::ffff:127.0.0.1",
    "fc00::1",
    "fe80::1",
    "2001:db8::1",
    "2002:7f00:1::",
    "ff02::1",
  ])("classifies %s as non-public", (address) => {
    expect(isPublicIpAddress(address)).toBe(false);
  });

  it.each(["8.8.8.8", "93.184.216.34", "2606:4700:4700::1111"])(
    "classifies %s as public",
    (address) => {
      expect(isPublicIpAddress(address)).toBe(true);
    },
  );

  it("rejects a DNS answer set containing any non-public address", async () => {
    const request = jest.fn();

    await expect(
      fetchPublicHtml("https://example.com", {
        request,
        resolveHostname: async () => [
          PUBLIC_ADDRESS,
          { address: "127.0.0.1", family: 4 },
        ],
      }),
    ).rejects.toMatchObject({ code: "unsafe-target" });
    expect(request).not.toHaveBeenCalled();
  });

  it.each([
    "http://127.0.0.1",
    "http://127.1",
    "http://0x7f000001",
    "http://[::1]",
    "http://[::ffff:127.0.0.1]",
  ])("rejects a direct non-public IP target %s", async (url) => {
    const request = jest.fn();

    await expect(fetchPublicHtml(url, { request })).rejects.toMatchObject({
      code: "unsafe-target",
    });
    expect(request).not.toHaveBeenCalled();
  });

  it("retains every validated address for DNS pinning and failover", async () => {
    const request = jest.fn(async () => ({
      body: Buffer.from("<html></html>"),
      kind: "document" as const,
    }));
    const secondaryAddress = {
      address: "2606:4700:4700::1111",
      family: 6 as const,
    };

    await fetchPublicHtml("https://example.com", {
      request,
      resolveHostname: async () => [secondaryAddress, PUBLIC_ADDRESS],
    });

    expect(request).toHaveBeenCalledWith(
      expect.objectContaining({
        addresses: [secondaryAddress, PUBLIC_ADDRESS],
      }),
      expect.any(AbortSignal),
    );
  });

  it("fails over from an unreachable pinned family to another validated address", async () => {
    const server = createServer((_request, response) => {
      response.writeHead(200, { "content-type": "text/html" });
      response.end("<html><title>Fallback</title></html>");
    });
    await new Promise<void>((resolve, reject) => {
      server.once("error", reject);
      server.listen(0, "127.0.0.1", () => {
        server.off("error", reject);
        resolve();
      });
    });

    const { port } = server.address() as AddressInfo;
    const controller = new AbortController();
    const timeout = setTimeout(() => {
      controller.abort(new Error("Pinned address failover test timed out."));
    }, 2_000);

    try {
      await expect(
        requestMetadataDocument(
          {
            addresses: [
              { address: "::1", family: 6 },
              { address: "127.0.0.1", family: 4 },
            ],
            url: new URL(`http://metadata-failover.test:${port}`),
          },
          controller.signal,
        ),
      ).resolves.toEqual({
        body: Buffer.from("<html><title>Fallback</title></html>"),
        kind: "document",
      });
    } finally {
      clearTimeout(timeout);
      await new Promise<void>((resolve, reject) => {
        server.close((error) => {
          if (error) reject(error);
          else resolve();
        });
      });
    }
  });

  it("leaves the final pinned address under the overall timeout budget", async () => {
    const body = Buffer.from("<html><title>Slow but valid</title></html>");
    const server = createServer((_request, response) => {
      setTimeout(() => {
        response.writeHead(200, { "content-type": "text/html" });
        response.end(body);
      }, 350);
    });
    await new Promise<void>((resolve, reject) => {
      server.once("error", reject);
      server.listen(0, "127.0.0.1", () => {
        server.off("error", reject);
        resolve();
      });
    });

    const { port } = server.address() as AddressInfo;
    const controller = new AbortController();
    const timeout = setTimeout(() => {
      controller.abort(new Error("Single-address timeout test timed out."));
    }, 1_000);

    try {
      await expect(
        requestMetadataDocument(
          {
            addresses: [{ address: "127.0.0.1", family: 4 }],
            url: new URL(`http://metadata-single.test:${port}`),
          },
          controller.signal,
        ),
      ).resolves.toEqual({ body, kind: "document" });
    } finally {
      clearTimeout(timeout);
      await new Promise<void>((resolve, reject) => {
        server.close((error) => {
          if (error) reject(error);
          else resolve();
        });
      });
    }
  });

  it("re-resolves and rejects a redirect to a private host", async () => {
    const request = jest.fn(async () => ({
      kind: "redirect" as const,
      location: "https://private.example/metadata",
    }));
    const resolveHostname = jest.fn(async (hostname: string) =>
      hostname === "example.com"
        ? [PUBLIC_ADDRESS]
        : [{ address: "10.0.0.10", family: 4 as const }],
    );

    await expect(
      fetchPublicHtml("https://example.com", {
        request,
        resolveHostname,
      }),
    ).rejects.toMatchObject({ code: "unsafe-target" });
    expect(resolveHostname).toHaveBeenCalledTimes(2);
    expect(request).toHaveBeenCalledTimes(1);
  });

  it("rejects redirects directly targeting link-local metadata services", async () => {
    const request = jest.fn(async () => ({
      kind: "redirect" as const,
      location: "http://169.254.169.254/latest/meta-data",
    }));

    await expect(
      fetchPublicHtml("https://example.com", {
        request,
        resolveHostname: async () => [PUBLIC_ADDRESS],
      }),
    ).rejects.toMatchObject({ code: "unsafe-target" });
    expect(request).toHaveBeenCalledTimes(1);
  });

  it("enforces the redirect limit", async () => {
    const request = jest.fn(async () => ({
      kind: "redirect" as const,
      location: "/again",
    }));

    await expect(
      fetchPublicHtml("https://example.com", {
        request,
        resolveHostname: async () => [PUBLIC_ADDRESS],
      }),
    ).rejects.toMatchObject({ code: "redirect-limit" });
    expect(request).toHaveBeenCalledTimes(MAX_METADATA_REDIRECTS + 1);
  });

  it("applies one timeout to the complete redirect and request flow", async () => {
    const request = jest.fn(
      async (_target: unknown, signal: AbortSignal) =>
        new Promise<never>((_resolve, reject) => {
          signal.addEventListener(
            "abort",
            () => {
              reject(
                signal.reason instanceof Error
                  ? signal.reason
                  : new Error("Metadata request aborted."),
              );
            },
            { once: true },
          );
        }),
    );

    await expect(
      fetchPublicHtml("https://example.com", {
        request,
        resolveHostname: async () => [PUBLIC_ADDRESS],
        timeoutMs: 5,
      }),
    ).rejects.toMatchObject({ code: "timeout" });
  });
});

describe("metadata response policy", () => {
  it("accepts bounded HTML responses", async () => {
    const body = Buffer.from("<html><title>Example</title></html>");
    const { response } = createResponse({ body });

    await expect(readMetadataHttpResponse(response)).resolves.toEqual({
      body,
      kind: "document",
    });
  });

  it("returns redirects without consuming their body", async () => {
    const { destroy, response } = createResponse({
      headers: { location: "/next" },
      statusCode: 302,
    });

    await expect(readMetadataHttpResponse(response)).resolves.toEqual({
      kind: "redirect",
      location: "/next",
    });
    expect(destroy).toHaveBeenCalled();
  });

  it.each([
    [{ "content-type": "application/json" }, "unsupported-response"],
    [
      { "content-type": "text/html", "content-encoding": "gzip" },
      "unsupported-response",
    ],
    [
      {
        "content-length": String(MAX_METADATA_RESPONSE_BYTES + 1),
        "content-type": "text/html",
      },
      "response-too-large",
    ],
  ])("rejects an unsafe response", async (headers, code) => {
    const { destroy, response } = createResponse({ headers });

    await expect(readMetadataHttpResponse(response)).rejects.toMatchObject({
      code,
    });
    expect(destroy).toHaveBeenCalled();
  });

  it("enforces the byte limit while streaming without a content length", async () => {
    const { destroy, response } = createResponse({
      body: [
        Buffer.alloc(MAX_METADATA_RESPONSE_BYTES),
        Buffer.from("one byte too many"),
      ],
    });

    await expect(readMetadataHttpResponse(response)).rejects.toMatchObject({
      code: "response-too-large",
    });
    expect(destroy).toHaveBeenCalled();
  });
});

describe("metadata image policy", () => {
  it("accepts bounded raster images and preserves their content type", async () => {
    const body = Buffer.from("raster-image");
    const { response } = createResponse({
      body,
      headers: { "content-type": "image/png" },
    });

    await expect(readMetadataImageHttpResponse(response)).resolves.toEqual({
      body,
      contentType: "image/png",
      kind: "document",
    });
  });

  it.each([
    [{ "content-type": "image/svg+xml" }, "unsupported-response"],
    [
      {
        "content-length": String(MAX_METADATA_IMAGE_RESPONSE_BYTES + 1),
        "content-type": "image/png",
      },
      "response-too-large",
    ],
  ])("rejects an unsafe image response", async (headers, code) => {
    const { destroy, response } = createResponse({ headers });

    await expect(readMetadataImageHttpResponse(response)).rejects.toMatchObject(
      {
        code,
      },
    );
    expect(destroy).toHaveBeenCalled();
  });

  it("applies the public-target policy to proxied images", async () => {
    const request = jest.fn();

    await expect(
      fetchPublicImage("http://169.254.169.254/latest/meta-data", { request }),
    ).rejects.toMatchObject({ code: "unsafe-target" });
    expect(request).not.toHaveBeenCalled();
  });
});
