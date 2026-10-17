/* global afterEach, beforeEach, describe, expect, it, jest -- Provided by the projects test runner. */
import type { ReactNode } from "react";
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { get, post } from "@/lib/http";
import {
  getAudioTranscription,
  requestAudioTranscription,
  useAudioTranscription,
} from "./audio-transcription";
import type { ChatTranscription } from "./audio-transcription";

let mockUserId = "member";
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "second" }),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({
    data: mockUserId ? { user: { id: mockUserId } } : null,
  }),
}));
jest.mock("@/lib/http", () => ({ get: jest.fn(), post: jest.fn() }));
const mockGet = jest.mocked(get);
const mockPost = jest.mocked(post);
let client: QueryClient;
const key = ["chat", "second", "transcription", "room", "voice", "member"];
const Wrapper = ({ children }: { children: ReactNode }) => (
  <QueryClientProvider client={client}>{children}</QueryClientProvider>
);

beforeEach(() => {
  jest.useFakeTimers();
  mockUserId = "member";
  client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  mockGet.mockReset().mockResolvedValue({ status: "not_requested" });
  mockPost.mockReset().mockResolvedValue({ status: "pending" });
});
afterEach(() => {
  cleanup();
  client.clear();
  jest.useRealTimers();
});

const hook = () =>
  renderHook(() => useAudioTranscription("room", "voice"), {
    wrapper: Wrapper,
  });

describe("audio transcription requests", () => {
  it("uses workspace-relative transport paths without a leading slash", async () => {
    await getAudioTranscription({ workspaceSlug: "second" }, "room", "voice");
    await requestAudioTranscription(
      { workspaceSlug: "second" },
      "room",
      "voice",
      false,
    );
    expect(mockGet.mock.calls[0][0]).toBe(
      "chat/conversations/room/attachments/voice/transcription",
    );
    expect(mockPost.mock.calls[0][0]).toBe(
      "chat/conversations/room/attachments/voice/transcription",
    );
  });
  it("does no network work until explicit request and reuses a saved ready transcript without POST", async () => {
    mockGet.mockResolvedValue({ status: "ready", text: "Saved transcript" });
    const view = hook();
    expect(view.result.current.transcript.status).toBe("idle");
    await act(async () => {
      await jest.advanceTimersByTimeAsync(10_000);
    });
    expect(mockGet).not.toHaveBeenCalled();
    expect(mockPost).not.toHaveBeenCalled();
    act(() => {
      view.result.current.request();
    });
    await waitFor(() => {
      expect(view.result.current.transcript).toEqual({
        status: "ready",
        text: "Saved transcript",
      });
    });
    expect(mockGet).toHaveBeenCalledTimes(1);
    expect(mockPost).not.toHaveBeenCalled();
    view.unmount();
    const remounted = hook();
    expect(remounted.result.current.transcript.status).toBe("ready");
    expect(mockGet).toHaveBeenCalledTimes(1);
  });

  it("coalesces same-tick and duplicate-row requests before asking for inference", async () => {
    let resolve!: (value: ChatTranscription) => void;
    mockGet.mockReturnValue(
      new Promise((accept) => {
        resolve = accept;
      }),
    );
    const first = hook();
    const second = hook();
    act(() => {
      first.result.current.request();
      first.result.current.request();
      second.result.current.request();
    });
    await waitFor(() => {
      expect(mockGet).toHaveBeenCalledTimes(1);
    });
    await act(async () => {
      resolve({ status: "not_requested" });
    });
    await waitFor(() => {
      expect(first.result.current.transcript.status).toBe("pending");
    });
    expect(mockPost).toHaveBeenCalledTimes(1);
    expect(mockPost).toHaveBeenCalledWith(
      "chat/conversations/room/attachments/voice/transcription",
      { retry: false },
      { workspaceSlug: "second" },
      { retry: 0 },
      expect.any(Function),
    );
  });

  it("retries a failed job only after an explicit action", async () => {
    client.setQueryData(key, { status: "failed", error: "interrupted" });
    mockGet.mockResolvedValue({ status: "failed", error: "interrupted" });
    const view = hook();
    expect(view.result.current.transcript).toEqual({
      status: "error",
      error: "Transcription was interrupted. You can try again.",
    });
    expect(mockGet).not.toHaveBeenCalled();
    expect(mockPost).not.toHaveBeenCalled();
    act(() => {
      view.result.current.request();
    });
    await waitFor(() => {
      expect(mockPost).toHaveBeenCalledTimes(1);
    });
    expect(mockPost.mock.calls[0][1]).toEqual({ retry: true });
  });

  it("resumes polling a cached pending request after virtualization without another POST", async () => {
    client.setQueryData(key, { status: "processing" });
    mockGet.mockResolvedValue({ status: "ready", text: "Finished" });
    const first = hook();
    expect(first.result.current.transcript.status).toBe("pending");
    first.unmount();
    const remounted = hook();
    await act(async () => {
      await jest.advanceTimersByTimeAsync(2_500);
    });
    await waitFor(() => {
      expect(remounted.result.current.transcript).toEqual({
        status: "ready",
        text: "Finished",
      });
    });
    expect(mockGet).toHaveBeenCalledTimes(1);
    expect(mockPost).not.toHaveBeenCalled();
    await act(async () => {
      await jest.advanceTimersByTimeAsync(10_000);
    });
    expect(mockGet).toHaveBeenCalledTimes(1);
  });

  it("surfaces a polling error instead of leaving stale pending data spinning and allows an explicit recovery", async () => {
    client.setQueryData(key, { status: "pending" });
    mockGet.mockRejectedValueOnce(new Error("Connection lost"));
    const view = hook();
    await act(async () => {
      await jest.advanceTimersByTimeAsync(2_500);
    });
    await waitFor(() => {
      expect(view.result.current.transcript).toEqual({
        status: "error",
        error: "Could not load the transcript. Please try again.",
      });
    });
    await act(async () => {
      await jest.advanceTimersByTimeAsync(10_000);
    });
    expect(mockGet).toHaveBeenCalledTimes(1);
    mockGet.mockResolvedValue({ status: "ready", text: "Recovered" });
    act(() => {
      view.result.current.request();
    });
    await waitFor(() => {
      expect(view.result.current.transcript).toEqual({
        status: "ready",
        text: "Recovered",
      });
    });
    expect(mockPost).not.toHaveBeenCalled();
  });

  it("surfaces request errors and releases the duplicate-action guard for retry", async () => {
    mockPost.mockRejectedValueOnce(new Error("Service unavailable"));
    const view = hook();
    act(() => {
      view.result.current.request();
    });
    await waitFor(() => {
      expect(view.result.current.transcript).toEqual({
        status: "error",
        error: "Could not load the transcript. Please try again.",
      });
    });
    act(() => {
      view.result.current.request();
    });
    await waitFor(() => {
      expect(view.result.current.transcript.status).toBe("pending");
    });
    expect(mockPost).toHaveBeenCalledTimes(2);
  });

  it("cancels polling when its last observer unmounts", async () => {
    client.setQueryData(key, { status: "pending" });
    mockGet.mockReturnValue(
      new Promise(() => {
        // Keep polling in flight until the observer is removed.
      }),
    );
    const view = hook();
    await act(async () => {
      await jest.advanceTimersByTimeAsync(2_500);
    });
    const signal = mockGet.mock.calls[0][2]?.signal;
    expect(signal?.aborted).toBe(false);
    view.unmount();
    expect(signal?.aborted).toBe(true);
  });

  it("cancels the saved-result check before starting inference if its row unmounts", async () => {
    let resolve!: (value: ChatTranscription) => void;
    mockGet.mockReturnValue(
      new Promise((accept) => {
        resolve = accept;
      }),
    );
    const view = hook();
    act(() => {
      view.result.current.request();
    });
    await waitFor(() => {
      expect(mockGet).toHaveBeenCalledTimes(1);
    });
    const signal = mockGet.mock.calls[0][2]?.signal;
    view.unmount();
    expect(signal?.aborted).toBe(true);
    await act(async () => {
      resolve({ status: "not_requested" });
    });
    expect(mockPost).not.toHaveBeenCalled();
  });

  it("does not carry explicit-request state or cached transcript into another account", async () => {
    mockGet.mockResolvedValue({ status: "ready", text: "Member transcript" });
    const view = hook();
    act(() => {
      view.result.current.request();
    });
    await waitFor(() => {
      expect(view.result.current.transcript.status).toBe("ready");
    });
    mockUserId = "other-member";
    view.rerender();
    expect(view.result.current.transcript.status).toBe("idle");
    expect(mockGet).toHaveBeenCalledTimes(1);
    mockUserId = "";
    view.rerender();
    act(() => {
      view.result.current.request();
    });
    expect(mockGet).toHaveBeenCalledTimes(1);
  });
});
