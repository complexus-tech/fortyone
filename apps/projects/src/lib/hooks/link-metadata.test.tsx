/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type { ReactNode } from "react";
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { getLinkMetadata } from "../queries/links/get-metadata";
import { useLinkMetadata } from "./link-metadata";

jest.mock("../queries/links/get-metadata", () => ({
  getLinkMetadata: jest.fn(),
}));

let queryClient: QueryClient;
const Wrapper = ({ children }: { children: ReactNode }) => (
  <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
);

beforeEach(() => {
  queryClient = new QueryClient();
  jest
    .mocked(getLinkMetadata)
    .mockReset()
    .mockResolvedValue({ title: "Article" });
});

afterEach(() => {
  cleanup();
  queryClient.clear();
});

describe("link metadata query lifecycle", () => {
  it("shares metadata across repeated links and virtualized remounts", async () => {
    const first = renderHook(
      () => useLinkMetadata("https://example.com/article"),
      { wrapper: Wrapper },
    );
    const second = renderHook(
      () => useLinkMetadata("https://example.com/article"),
      { wrapper: Wrapper },
    );
    await waitFor(() => {
      expect(first.result.current.data?.title).toBe("Article");
    });
    expect(second.result.current.data?.title).toBe("Article");
    first.unmount();
    second.unmount();

    const remounted = renderHook(
      () => useLinkMetadata("https://example.com/article"),
      { wrapper: Wrapper },
    );
    expect(remounted.result.current.data?.title).toBe("Article");
    expect(getLinkMetadata).toHaveBeenCalledTimes(1);
  });

  it("passes query cancellation to the fetcher when the last observer unmounts", () => {
    jest.mocked(getLinkMetadata).mockImplementation(
      () =>
        new Promise(() => {
          // Keep the request pending until the observer is removed.
        }),
    );
    const { unmount } = renderHook(
      () => useLinkMetadata("https://example.com/article"),
      { wrapper: Wrapper },
    );
    const signal = jest.mocked(getLinkMetadata).mock.calls[0][1];
    expect(signal?.aborted).toBe(false);
    unmount();
    expect(signal?.aborted).toBe(true);
  });

  it("does not fetch previews until enabled", () => {
    renderHook(
      () => useLinkMetadata("https://example.com/article", { enabled: false }),
      { wrapper: Wrapper },
    );
    renderHook(() => useLinkMetadata(""), { wrapper: Wrapper });
    expect(getLinkMetadata).not.toHaveBeenCalled();
  });

  it("shares canonical keys across Chat and story URLs but preserves query strings", async () => {
    const first = renderHook(
      () => useLinkMetadata("https://EXAMPLE.com:443/article?view=one#intro"),
      { wrapper: Wrapper },
    );
    await waitFor(() => {
      expect(first.result.current.isSuccess).toBe(true);
    });
    const second = renderHook(
      () => useLinkMetadata("https://example.com/article?view=one#details"),
      { wrapper: Wrapper },
    );
    expect(second.result.current.data?.title).toBe("Article");
    expect(getLinkMetadata).toHaveBeenCalledTimes(1);
    const different = renderHook(
      () => useLinkMetadata("https://example.com/article?view=two"),
      { wrapper: Wrapper },
    );
    await waitFor(() => {
      expect(different.result.current.isSuccess).toBe(true);
    });
    expect(getLinkMetadata).toHaveBeenCalledTimes(2);
  });

  it("retries empty metadata after 30 seconds instead of keeping a failure fresh for an hour", async () => {
    jest.mocked(getLinkMetadata).mockResolvedValueOnce({});
    const first = renderHook(
      () => useLinkMetadata("https://example.com/article"),
      { wrapper: Wrapper },
    );
    await waitFor(() => {
      expect(first.result.current.isSuccess).toBe(true);
    });
    first.unmount();
    act(() => {
      queryClient.setQueryData(
        ["link-metadata", "https://example.com/article"],
        {},
        { updatedAt: Date.now() - 30_001 },
      );
    });
    const second = renderHook(
      () => useLinkMetadata("https://example.com/article"),
      { wrapper: Wrapper },
    );
    await waitFor(() => {
      expect(second.result.current.data?.title).toBe("Article");
    });
    expect(getLinkMetadata).toHaveBeenCalledTimes(2);
  });
});

describe("manual preview refresh", () => {
  it("updates every observer with the new revision while retaining the old card during refresh", async () => {
    const first = renderHook(
      () => useLinkMetadata("https://example.com/article"),
      { wrapper: Wrapper },
    );
    const second = renderHook(
      () => useLinkMetadata("https://example.com/article"),
      { wrapper: Wrapper },
    );
    await waitFor(() => {
      expect(first.result.current.data?.title).toBe("Article");
    });
    let complete!: (value: { title: string; version: string }) => void;
    jest.mocked(getLinkMetadata).mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          complete = resolve;
        }),
    );
    let request!: Promise<void>;
    act(() => {
      request = first.result.current.refreshPreview();
    });
    await waitFor(() => {
      expect(second.result.current.isRefreshing).toBe(true);
    });
    expect(first.result.current.data?.title).toBe("Article");
    await act(async () => {
      complete({ title: "Updated", version: "2" });
      await request;
    });
    await waitFor(() => {
      expect(second.result.current.data?.title).toBe("Updated");
    });
    expect(getLinkMetadata).toHaveBeenLastCalledWith(
      "https://example.com/article",
      undefined,
      { refresh: true },
    );
  });

  it("keeps last-good data and rejects refresh failures for visible error feedback", async () => {
    const { result } = renderHook(
      () => useLinkMetadata("https://example.com/article"),
      { wrapper: Wrapper },
    );
    await waitFor(() => {
      expect(result.current.data?.title).toBe("Article");
    });
    jest
      .mocked(getLinkMetadata)
      .mockRejectedValueOnce(new Error("Unavailable"));
    await act(async () => {
      await expect(result.current.refreshPreview()).rejects.toThrow(
        "Unavailable",
      );
    });
    await waitFor(() => {
      expect(result.current.refreshError).toBeInstanceOf(Error);
    });
    expect(result.current.data?.title).toBe("Article");
  });
});
