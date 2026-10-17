/* global afterEach, beforeEach, describe, expect, it, jest -- Provided by the projects test runner. */
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { AudioAttachmentPlayer } from "./audio-attachment-player";
import { AudioPlaybackBar } from "./audio-playback-bar";
import { AudioPlaybackProvider } from "./audio-playback-context";
import type { AudioTrack } from "./audio-playback-controller";

const track: AudioTrack = {
  id: "voice",
  messageId: "source-message",
  fileName: "Voice message.webm",
  url: "/api/voice",
  durationSeconds: 60,
};
let audio: HTMLAudioElement;
let paused: boolean;
let play: jest.SpyInstance<Promise<void>, []>;
let pause: jest.SpyInstance<void, []>;
let onIntersection: IntersectionObserverCallback;
const originalIntersectionObserver = global.IntersectionObserver;

beforeEach(() => {
  paused = true;
  audio = document.createElement("audio");
  Object.defineProperties(audio, {
    paused: { get: () => paused },
    duration: { value: 60 },
    readyState: { value: 1 },
  });
  play = jest.spyOn(audio, "play").mockImplementation(async () => {
    paused = false;
  });
  pause = jest.spyOn(audio, "pause").mockImplementation(() => {
    paused = true;
  });
  jest.spyOn(audio, "load").mockImplementation(() => undefined);
  jest.spyOn(window, "Audio").mockImplementation(() => audio);
  const intersections = new Set<IntersectionObserverCallback>();
  onIntersection = (entries, observer) => {
    intersections.forEach((callback) => {
      callback(entries, observer);
    });
  };
  global.IntersectionObserver = jest.fn(
    (callback: IntersectionObserverCallback) => {
      intersections.add(callback);
      return {
        observe: jest.fn(),
        unobserve: jest.fn(),
        disconnect: jest.fn(() => {
          intersections.delete(callback);
        }),
        root: null,
        rootMargin: "",
        scrollMargin: "",
        thresholds: [],
        takeRecords: () => [],
      };
    },
  );
});

afterEach(() => {
  cleanup();
  jest.restoreAllMocks();
  global.IntersectionObserver = originalIntersectionObserver;
});

const Harness = ({
  showRow = true,
  onViewMessage = jest.fn(),
}: {
  showRow?: boolean;
  onViewMessage?: (id: string) => void;
}) => (
  <AudioPlaybackProvider onViewMessage={onViewMessage}>
    {showRow ? <AudioAttachmentPlayer attachment={track} /> : null}
    <AudioPlaybackBar />
  </AudioPlaybackProvider>
);

describe("inline audio messages", () => {
  it("plays explicitly, exposes accessible seeking, and cycles playback speed", async () => {
    render(<Harness />);
    expect(window.Audio).not.toHaveBeenCalled();
    expect(
      screen.getByRole("slider", { name: "Seek Voice message.webm" }),
    ).toBeDisabled();
    fireEvent.click(
      screen.getByRole("button", { name: "Play Voice message.webm" }),
    );
    await waitFor(() => {
      expect(play).toHaveBeenCalledTimes(1);
    });
    act(() => {
      onIntersection(
        [{ isIntersecting: true } as IntersectionObserverEntry],
        {} as IntersectionObserver,
      );
      audio.dispatchEvent(new Event("loadedmetadata"));
    });
    const seek = screen.getByRole("slider", {
      name: "Seek Voice message.webm",
    });
    expect(seek).toBeEnabled();
    fireEvent.change(seek, { target: { value: "25" } });
    expect(audio.currentTime).toBe(25);
    expect(seek).toHaveAttribute("aria-valuetext", "0:25 of 1:00");
    fireEvent.click(
      screen.getByRole("button", { name: "Playback speed 1 times" }),
    );
    expect(audio.playbackRate).toBe(1.5);
    fireEvent.click(
      screen.getByRole("button", { name: "Pause Voice message.webm" }),
    );
    expect(paused).toBe(true);
  });

  it("keeps playback alive after row virtualization and offers return and stop controls", async () => {
    const onViewMessage = jest.fn();
    const view = render(<Harness onViewMessage={onViewMessage} />);
    act(() => {
      onIntersection(
        [{ isIntersecting: true } as IntersectionObserverEntry],
        {} as IntersectionObserver,
      );
    });
    fireEvent.click(
      screen.getByRole("button", { name: "Play Voice message.webm" }),
    );
    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: "Pause Voice message.webm" }),
      ).toBeInTheDocument();
    });
    expect(
      screen.queryByRole("region", { name: "Audio playing outside this view" }),
    ).not.toBeInTheDocument();
    const pauseCount = pause.mock.calls.length;
    view.rerender(<Harness onViewMessage={onViewMessage} showRow={false} />);
    expect(pause).toHaveBeenCalledTimes(pauseCount);
    expect(
      screen.getByRole("region", { name: "Audio playing outside this view" }),
    ).toBeInTheDocument();
    fireEvent.click(
      screen.getByRole("button", { name: "Return to audio message" }),
    );
    expect(onViewMessage).toHaveBeenCalledWith("source-message");
    fireEvent.click(
      screen.getByRole("button", { name: "Stop audio playback" }),
    );
    expect(
      screen.queryByRole("region", { name: "Audio playing outside this view" }),
    ).not.toBeInTheDocument();
    expect(audio).not.toHaveAttribute("src");
  });

  it("requests transcription only on demand and expands ready text without starting audio", () => {
    const request = jest.fn();
    const view = render(
      <AudioPlaybackProvider>
        <AudioAttachmentPlayer
          attachment={track}
          onRequestTranscript={request}
        />
      </AudioPlaybackProvider>,
    );
    expect(request).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Transcribe" }));
    expect(request).toHaveBeenCalledTimes(1);
    view.rerender(
      <AudioPlaybackProvider>
        <AudioAttachmentPlayer
          attachment={track}
          onRequestTranscript={request}
          transcript={{ status: "pending" }}
        />
      </AudioPlaybackProvider>,
    );
    expect(
      screen.getByRole("button", { name: "Transcribing…" }),
    ).toBeDisabled();
    view.rerender(
      <AudioPlaybackProvider>
        <AudioAttachmentPlayer
          attachment={track}
          onRequestTranscript={request}
          transcript={{
            status: "ready",
            text: "Let’s review the welcome flow tomorrow.",
          }}
        />
      </AudioPlaybackProvider>,
    );
    expect(
      screen.getByText("Let’s review the welcome flow tomorrow."),
    ).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Hide transcript" }));
    expect(
      screen.queryByText("Let’s review the welcome flow tomorrow."),
    ).not.toBeInTheDocument();
    expect(window.Audio).not.toHaveBeenCalled();
  });

  it("shows a retryable transcription error and stops playback on provider teardown", async () => {
    const request = jest.fn();
    const view = render(
      <AudioPlaybackProvider>
        <AudioAttachmentPlayer
          attachment={track}
          onRequestTranscript={request}
          transcript={{ status: "error", error: "Please try again." }}
        />
      </AudioPlaybackProvider>,
    );
    expect(screen.getByText("Please try again.")).toBeInTheDocument();
    fireEvent.click(
      screen.getByRole("button", { name: "Retry transcription" }),
    );
    expect(request).toHaveBeenCalledTimes(1);
    fireEvent.click(
      screen.getByRole("button", { name: "Play Voice message.webm" }),
    );
    await waitFor(() => {
      expect(play).toHaveBeenCalledTimes(1);
    });
    view.unmount();
    expect(paused).toBe(true);
    expect(audio).not.toHaveAttribute("src");
  });

  it("explains a ready transcript with no detected speech", () => {
    render(
      <AudioPlaybackProvider>
        <AudioAttachmentPlayer
          attachment={track}
          transcript={{ status: "ready", text: "" }}
        />
      </AudioPlaybackProvider>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Show transcript" }));
    expect(
      screen.getByText("No speech was detected in this recording."),
    ).toBeInTheDocument();
  });
});
