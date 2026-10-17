/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import ky from "ky";
import type { LinkMetadata } from "./get-metadata";
import { getLinkMetadata } from "./get-metadata";

jest.mock("ky", () => ({
  __esModule: true,
  default: { get: jest.fn() },
}));

const mockedGet = jest.mocked(ky.get);
type PendingRequest = {
  signal?: AbortSignal | null;
  resolve: (metadata: LinkMetadata) => void;
  reject: (reason: Error) => void;
};
let requests: PendingRequest[];

beforeEach(() => {
  requests = [];
  mockedGet.mockReset().mockImplementation((_url, options) => {
    const signal = options?.signal;
    return {
      json: () =>
        new Promise<LinkMetadata>((resolve, reject) => {
          const onAbort = () => {
            const reason: unknown = signal?.reason;
            reject(reason instanceof Error ? reason : new Error("Aborted"));
          };
          if (signal?.aborted) {
            onAbort();
            return;
          }
          signal?.addEventListener("abort", onAbort, { once: true });
          requests.push({
            signal,
            resolve: (metadata) => {
              signal?.removeEventListener("abort", onAbort);
              resolve(metadata);
            },
            reject: (reason) => {
              signal?.removeEventListener("abort", onAbort);
              reject(reason);
            },
          });
        }),
    } as never;
  });
});

const nextMicrotask = () =>
  new Promise<void>((resolve) => {
    queueMicrotask(resolve);
  });

describe("link metadata requests", () => {
  it("returns social preview metadata and disables transport retries", async () => {
    const controller = new AbortController();
    const result = getLinkMetadata(
      "https://example.com/article",
      controller.signal,
    );
    await nextMicrotask();
    const metadata = {
      title: "Article",
      image: "/api/metadata/image?url=icon",
      previewImage: "/api/metadata/image?url=preview",
    };
    requests[0].resolve(metadata);
    await expect(result).resolves.toEqual(metadata);
    expect(mockedGet).toHaveBeenCalledWith(
      "/api/metadata?url=https%3A%2F%2Fexample.com%2Farticle",
      { retry: 0, signal: controller.signal, timeout: 12_000 },
    );
  });

  it("limits concurrent metadata fetches to four and starts waiting work in order", async () => {
    const results = Array.from({ length: 6 }, (_, index) =>
      getLinkMetadata(`https://example.com/${index}`),
    );
    await nextMicrotask();
    expect(mockedGet).toHaveBeenCalledTimes(4);

    requests[0].resolve({ title: "First" });
    await results[0];
    await nextMicrotask();
    expect(mockedGet).toHaveBeenCalledTimes(5);
    expect(mockedGet.mock.calls[4][0]).toContain("%2F4");

    requests[1].resolve({});
    await results[1];
    await nextMicrotask();
    expect(mockedGet).toHaveBeenCalledTimes(6);
    expect(mockedGet.mock.calls[5][0]).toContain("%2F5");
    requests.slice(2).forEach((request) => {
      request.resolve({});
    });
    await Promise.all(results);
  });

  it("removes a canceled queued request without fetching it or blocking later work", async () => {
    const active = Array.from({ length: 4 }, (_, index) =>
      getLinkMetadata(`https://example.com/active-${index}`),
    );
    const controller = new AbortController();
    const canceled = getLinkMetadata(
      "https://example.com/canceled",
      controller.signal,
    );
    const aborted = expect(canceled).rejects.toMatchObject({
      name: "AbortError",
    });
    const next = getLinkMetadata("https://example.com/next");
    await nextMicrotask();

    controller.abort();
    await aborted;
    requests[0].resolve({});
    await active[0];
    await nextMicrotask();
    expect(mockedGet).toHaveBeenCalledTimes(5);
    expect(mockedGet.mock.calls[4][0]).toContain("%2Fnext");
    expect(
      mockedGet.mock.calls.some(([url]) => String(url).includes("canceled")),
    ).toBe(false);

    requests.slice(1).forEach((request) => {
      request.resolve({});
    });
    await Promise.all([...active, next]);
  });

  it("cancels an in-flight fetch and releases its slot", async () => {
    const controller = new AbortController();
    const canceled = getLinkMetadata(
      "https://example.com/canceled",
      controller.signal,
    );
    const aborted = expect(canceled).rejects.toMatchObject({
      name: "AbortError",
    });
    const active = Array.from({ length: 3 }, (_, index) =>
      getLinkMetadata(`https://example.com/active-${index}`),
    );
    const next = getLinkMetadata("https://example.com/next");
    await nextMicrotask();
    expect(requests[0].signal).toBe(controller.signal);

    controller.abort();
    await aborted;
    await nextMicrotask();
    expect(mockedGet).toHaveBeenCalledTimes(5);
    requests.slice(1).forEach((request) => {
      request.resolve({});
    });
    await Promise.all([...active, next]);
  });

  it("does not start an already canceled request", async () => {
    const controller = new AbortController();
    controller.abort();
    await expect(
      getLinkMetadata("https://example.com", controller.signal),
    ).rejects.toMatchObject({ name: "AbortError" });
    expect(mockedGet).not.toHaveBeenCalled();
  });

  it("falls back to a plain link when metadata fails", async () => {
    const result = getLinkMetadata("https://example.com");
    await nextMicrotask();
    requests[0].reject(new Error("Upstream unavailable"));
    await expect(result).resolves.toEqual({});
  });

  it("forces server refresh and propagates errors instead of replacing a good card", async () => {
    const result = getLinkMetadata("https://example.com", undefined, {
      refresh: true,
    });
    await nextMicrotask();
    expect(mockedGet.mock.calls[0][0]).toBe(
      "/api/metadata?url=https%3A%2F%2Fexample.com%2F&refresh=1",
    );
    requests[0].reject(new Error("Refresh unavailable"));
    await expect(result).rejects.toThrow("Refresh unavailable");
  });

  it("strips fragments from requests without altering query strings", async () => {
    const result = getLinkMetadata(
      "https://EXAMPLE.com:443/article?view=one&view=two#heading",
    );
    await nextMicrotask();
    requests[0].resolve({ title: "Article" });
    await result;
    expect(mockedGet.mock.calls[0][0]).toBe(
      "/api/metadata?url=https%3A%2F%2Fexample.com%2Farticle%3Fview%3Done%26view%3Dtwo",
    );
  });

  it.each([
    "invalid",
    "mailto:person@example.com",
    // eslint-disable-next-line no-script-url -- Verify that executable URLs are rejected without making a request.
    "javascript:alert(1)",
    "https://user:secret@example.com",
  ])("does not fetch an unsupported URL: %s", async (url) => {
    await expect(getLinkMetadata(url)).resolves.toEqual({});
    expect(mockedGet).not.toHaveBeenCalled();
  });
});
