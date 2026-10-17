/* global afterEach, beforeEach, describe, expect, it, jest -- Provided by the projects test runner. */
import { act, cleanup, render } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { getMessage } from "./api";
import { AudioAccessGuard } from "./audio-access-guard";
import { createAudioPlaybackController } from "./audio-playback-controller";
import type {
  AudioPlaybackController,
  AudioTrack,
} from "./audio-playback-controller";
import type { ChatMessage } from "./types";

jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "Second" }),
}));
jest.mock("./api", () => ({ getMessage: jest.fn() }));
jest.mock("./audio-playback-context", () => ({
  useAudioPlaybackController: () => mockPlayback,
}));
const mockGetMessage = jest.mocked(getMessage);
let mockPlayback: AudioPlaybackController;
let client: QueryClient;
let audio: ReturnType<typeof createFakeAudio>;
const track: AudioTrack = {
  id: "voice",
  messageId: "message",
  fileName: "Voice note",
  url: "/voice",
  durationSeconds: 10,
};
const key = ["chat", "second", "messages", "room"];
const createFakeAudio = () => {
  const element = Object.assign(new EventTarget(), {
    preload: "none",
    src: "",
    currentTime: 0,
    duration: 10,
    readyState: 1,
    paused: true,
    ended: false,
    playbackRate: 1,
    error: null,
    load: jest.fn(),
    removeAttribute: jest.fn((name: string) => {
      if (name === "src") element.src = "";
    }),
    pause: jest.fn(() => {
      element.paused = true;
    }),
    play: jest.fn(async () => {
      element.paused = false;
    }),
  });
  return element;
};
const validMessage = {
  id: "message",
  deletedAt: null,
  attachments: [{ id: "voice" }],
} as ChatMessage;

beforeEach(() => {
  jest.useFakeTimers();
  client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  audio = createFakeAudio();
  mockPlayback = createAudioPlaybackController(
    () => audio as unknown as HTMLAudioElement,
  );
  mockGetMessage.mockReset().mockResolvedValue(validMessage);
});
afterEach(() => {
  cleanup();
  mockPlayback.destroy();
  client.clear();
  jest.useRealTimers();
});
const mount = () =>
  render(
    <QueryClientProvider client={client}>
      <AudioAccessGuard />
    </QueryClientProvider>,
  );
const invalidate = (queryKey = key) => {
  client.setQueryData(queryKey, {});
  void client.invalidateQueries({ queryKey, exact: true, refetchType: "none" });
};
const advance = async (ms = 200) => {
  await act(async () => {
    await jest.advanceTimersByTimeAsync(ms);
  });
};

describe("active audio access guard", () => {
  it("does no access checks without playback and ignores unrelated workspace invalidation", async () => {
    mount();
    act(() => {
      invalidate();
    });
    await advance();
    expect(mockGetMessage).not.toHaveBeenCalled();
    await mockPlayback.play(track);
    act(() => {
      invalidate(["chat", "another-workspace", "messages", "room"]);
    });
    await advance();
    expect(mockGetMessage).not.toHaveBeenCalled();
  });

  it("coalesces frequent invalidations without waiting for a quiet period", async () => {
    await mockPlayback.play(track);
    mount();
    act(() => {
      invalidate();
    });
    await advance(100);
    act(() => {
      invalidate();
    });
    await advance(100);
    expect(mockGetMessage).toHaveBeenCalledTimes(1);
    expect(mockGetMessage).toHaveBeenCalledWith(
      { workspaceSlug: "Second" },
      "message",
      expect.any(AbortSignal),
    );
    expect(mockPlayback.getSnapshot().status).toBe("playing");
  });

  it.each([
    { ...validMessage, deletedAt: "2026-10-17T00:00:00Z" },
    { ...validMessage, attachments: [] },
  ])(
    "stops and releases audio when its message or attachment is removed",
    async (message) => {
      mockGetMessage.mockResolvedValue(message);
      await mockPlayback.play(track);
      mount();
      act(() => {
        invalidate();
      });
      await advance();
      expect(mockPlayback.getSnapshot().track).toBeNull();
      expect(audio.paused).toBe(true);
      expect(audio.src).toBe("");
    },
  );

  it.each([401, 403, 404])(
    "stops active audio when access returns %s",
    async (status) => {
      mockGetMessage.mockRejectedValue(
        Object.assign(new Error("Unavailable"), { status }),
      );
      await mockPlayback.play(track);
      mount();
      act(() => {
        invalidate();
      });
      await advance();
      expect(mockPlayback.getSnapshot().track).toBeNull();
      expect(audio.src).toBe("");
    },
  );

  it("does not revoke playback for a temporary network failure", async () => {
    mockGetMessage.mockRejectedValue(
      Object.assign(new Error("Unavailable"), { status: 503 }),
    );
    await mockPlayback.play(track);
    mount();
    act(() => {
      invalidate();
    });
    await advance();
    expect(mockPlayback.getSnapshot().status).toBe("playing");
  });

  it("ignores an old response after another attachment in the same message begins playing", async () => {
    let resolve!: (message: ChatMessage) => void;
    mockGetMessage.mockReturnValue(
      new Promise((accept) => {
        resolve = accept;
      }),
    );
    await mockPlayback.play(track);
    mount();
    act(() => {
      invalidate();
    });
    await advance();
    const next = { ...track, id: "another-voice", url: "/another-voice" };
    await mockPlayback.play(next);
    await act(async () => {
      resolve({ ...validMessage, attachments: [] });
    });
    expect(mockPlayback.getSnapshot()).toMatchObject({
      status: "playing",
      track: next,
    });
  });

  it("keeps one request in flight and catches up once more when hints arrive during it", async () => {
    let resolve!: (message: ChatMessage) => void;
    mockGetMessage.mockReturnValueOnce(
      new Promise((accept) => {
        resolve = accept;
      }),
    );
    await mockPlayback.play(track);
    mount();
    act(() => {
      invalidate();
    });
    await advance();
    const signal = mockGetMessage.mock.calls[0][2];
    act(() => {
      invalidate();
      invalidate();
    });
    await advance(1000);
    expect(mockGetMessage).toHaveBeenCalledTimes(1);
    expect(signal?.aborted).toBe(false);
    await act(async () => {
      resolve(validMessage);
    });
    await advance();
    expect(mockGetMessage).toHaveBeenCalledTimes(2);
  });

  it("aborts an outstanding check and ignores its late result on unmount", async () => {
    let resolve!: (message: ChatMessage) => void;
    mockGetMessage.mockReturnValue(
      new Promise((accept) => {
        resolve = accept;
      }),
    );
    await mockPlayback.play(track);
    const view = mount();
    act(() => {
      invalidate();
    });
    await advance();
    const signal = mockGetMessage.mock.calls[0][2];
    view.unmount();
    expect(signal?.aborted).toBe(true);
    await act(async () => {
      resolve({ ...validMessage, attachments: [] });
    });
    expect(mockPlayback.getSnapshot().status).toBe("playing");
  });
});
