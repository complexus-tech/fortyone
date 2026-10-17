/* global afterEach, beforeEach, describe, expect, it, jest -- Provided by the projects test runner. */
import { StrictMode } from "react";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
} from "@testing-library/react";
import { AudioPlayerControls } from "./audio-player-controls";
import type { AudioTrack } from "./audio-playback-controller";
import {
  createAudioPlaybackController,
  IDLE_AUDIO_PLAYBACK,
} from "./audio-playback-controller";
import { AudioWaveform } from "./audio-waveform";
import {
  AudioWaveformProvider,
  useAudioWaveform,
} from "./audio-waveform-context";
import type { AudioWaveformData } from "./audio-waveform-data";
import { createAudioWaveformStore } from "./audio-waveform-data";

jest.mock("./audio-waveform-data", () => ({
  createAudioWaveformStore: jest.fn(),
}));

class WaveformObserver implements IntersectionObserver {
  static instances: WaveformObserver[] = [];
  readonly root = null;
  readonly rootMargin = "0px";
  readonly scrollMargin = "0px";
  readonly thresholds = [0];
  readonly observe = jest.fn();
  readonly unobserve = jest.fn();
  readonly disconnect = jest.fn();
  readonly takeRecords = jest.fn(() => []);

  constructor(private readonly callback: IntersectionObserverCallback) {
    WaveformObserver.instances.push(this);
  }

  intersect(isIntersecting: boolean) {
    this.callback([{ isIntersecting } as IntersectionObserverEntry], this);
  }
}

const originalObserver = globalThis.IntersectionObserver;
const mockLoad = jest.fn<
  Promise<AudioWaveformData | null>,
  [Pick<AudioTrack, "id" | "url" | "durationSeconds">, AbortSignal]
>();
const mockDestroy = jest.fn();
const track: AudioTrack = {
  id: "recording",
  messageId: "message",
  fileName: "Voice note",
  url: "/api/recording",
  durationSeconds: 60,
};
const waveform: AudioWaveformData = {
  peaks: [0, 0.25, 1, 0.5],
  duration: 60,
};

const WaveformHarness = ({ source = track }: { source?: AudioTrack }) => {
  const { ref, data } = useAudioWaveform(source);
  return (
    <div ref={ref}>
      <output aria-label="Loaded waveform">
        {data ? JSON.stringify(data.peaks) : "Loading"}
      </output>
    </div>
  );
};

const show = (observer = WaveformObserver.instances.at(-1)!) => {
  act(() => {
    observer.intersect(true);
  });
};

const dwell = async () => {
  await act(async () => {
    jest.advanceTimersByTime(200);
  });
};

beforeEach(() => {
  jest.useFakeTimers();
  jest.clearAllMocks();
  WaveformObserver.instances = [];
  globalThis.IntersectionObserver = WaveformObserver;
  mockLoad.mockReset().mockResolvedValue(waveform);
  jest.mocked(createAudioWaveformStore).mockReturnValue({
    load: mockLoad,
    destroy: mockDestroy,
  });
});

afterEach(() => {
  cleanup();
  jest.useRealTimers();
  jest.restoreAllMocks();
  globalThis.IntersectionObserver = originalObserver;
});

describe("audio waveforms", () => {
  it("renders the recording's amplitudes and clips the played portion independently", () => {
    const view = render(
      <AudioWaveform peaks={waveform.peaks} progress={0.25} />,
    );
    const paths = Array.from(view.container.querySelectorAll("path"));
    const heights = [
      ...(paths[0].getAttribute("d") ?? "").matchAll(/v[\d.]+/g),
    ].map((match) => Number(match[0].slice(1)));
    expect(heights).toHaveLength(waveform.peaks.length);
    expect(heights[2]).toBeGreaterThan(heights[3]);
    expect(heights[3]).toBeGreaterThan(heights[1]);
    expect(heights[1]).toBeGreaterThan(heights[0]);
    expect(paths[1]).toHaveAttribute("d", paths[0].getAttribute("d"));
    const svg = view.container.querySelector("svg")!;
    const width = Number(svg.getAttribute("viewBox")?.split(" ")[2]);
    expect(
      Number(
        view.container.querySelector("clipPath rect")?.getAttribute("width"),
      ),
    ).toBe(width * 0.25);
    view.rerender(<AudioWaveform peaks={waveform.peaks} progress={0.75} />);
    expect(paths[0]).toHaveAttribute("d", paths[1].getAttribute("d"));
    expect(
      Number(
        view.container.querySelector("clipPath rect")?.getAttribute("width"),
      ),
    ).toBe(width * 0.75);
    expect(svg).toHaveAttribute("aria-hidden", "true");
  });

  it("waits for sustained visibility and skips notes passed during fast scrolling", async () => {
    render(
      <AudioWaveformProvider>
        <WaveformHarness />
      </AudioWaveformProvider>,
    );
    await dwell();
    expect(mockLoad).not.toHaveBeenCalled();
    show();
    act(() => {
      jest.advanceTimersByTime(100);
      WaveformObserver.instances[0].intersect(false);
    });
    await dwell();
    expect(mockLoad).not.toHaveBeenCalled();
    show();
    await dwell();
    expect(mockLoad).toHaveBeenCalledTimes(1);
    expect(
      screen.getByRole("status", { name: "Loaded waveform" }),
    ).toHaveTextContent(JSON.stringify(waveform.peaks));
    show();
    await dwell();
    expect(mockLoad).toHaveBeenCalledTimes(1);
  });

  it("aborts invisible work and ignores its result after a new visible request", async () => {
    let resolveOld!: (data: AudioWaveformData | null) => void;
    mockLoad.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveOld = resolve;
        }),
    );
    render(
      <AudioWaveformProvider>
        <WaveformHarness />
      </AudioWaveformProvider>,
    );
    show();
    await dwell();
    const oldSignal = mockLoad.mock.calls[0][1];
    act(() => {
      WaveformObserver.instances[0].intersect(false);
    });
    expect(oldSignal.aborted).toBe(true);
    show();
    await dwell();
    await act(async () => {
      resolveOld({ peaks: [1, 1, 1], duration: 60 });
    });
    expect(
      screen.getByRole("status", { name: "Loaded waveform" }),
    ).toHaveTextContent(JSON.stringify(waveform.peaks));
    expect(mockLoad).toHaveBeenCalledTimes(2);
  });

  it("does not leak an old URL's peaks while the same attachment resolves a new URL", async () => {
    let resolveNew!: (data: AudioWaveformData | null) => void;
    const view = render(
      <AudioWaveformProvider>
        <WaveformHarness />
      </AudioWaveformProvider>,
    );
    show();
    await dwell();
    mockLoad.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveNew = resolve;
        }),
    );
    view.rerender(
      <AudioWaveformProvider>
        <WaveformHarness
          source={{ ...track, url: "/api/replaced-recording" }}
        />
      </AudioWaveformProvider>,
    );
    expect(
      screen.getByRole("status", { name: "Loaded waveform" }),
    ).toHaveTextContent("Loading");
    expect(WaveformObserver.instances[0].disconnect).toHaveBeenCalledTimes(1);
    show();
    await dwell();
    expect(mockLoad.mock.calls[1][0].url).toBe("/api/replaced-recording");
    await act(async () => {
      resolveNew({ peaks: [1, 0, 1], duration: 60 });
    });
    expect(
      screen.getByRole("status", { name: "Loaded waveform" }),
    ).toHaveTextContent("[1,0,1]");
  });

  it("cancels active subscriptions and destroys the conversation cache on unmount", async () => {
    mockLoad.mockImplementation(() => new Promise(() => {}));
    const view = render(
      <AudioWaveformProvider>
        <WaveformHarness />
      </AudioWaveformProvider>,
    );
    show();
    await dwell();
    const signal = mockLoad.mock.calls[0][1];
    view.unmount();
    expect(signal.aborted).toBe(true);
    expect(mockDestroy).toHaveBeenCalledTimes(1);
    expect(WaveformObserver.instances[0].disconnect).toHaveBeenCalledTimes(1);
  });

  it("subscribes after StrictMode effect replay without duplicate downloads", async () => {
    const view = render(
      <StrictMode>
        <AudioWaveformProvider>
          <WaveformHarness />
        </AudioWaveformProvider>
      </StrictMode>,
    );
    expect(WaveformObserver.instances[0].disconnect).toHaveBeenCalledTimes(1);
    show();
    await dwell();
    expect(mockLoad).toHaveBeenCalledTimes(1);
    expect(
      screen.getByRole("status", { name: "Loaded waveform" }),
    ).toHaveTextContent(JSON.stringify(waveform.peaks));
    view.unmount();
    expect(mockDestroy).toHaveBeenCalledTimes(2);
  });

  it("keeps a focusable native range for keyboard seeking without starting playback", () => {
    const controller = createAudioPlaybackController();
    const seek = jest
      .spyOn(controller, "seek")
      .mockImplementation(() => undefined);
    const play = jest.spyOn(controller, "play").mockResolvedValue();
    render(
      <AudioWaveformProvider>
        <AudioPlayerControls
          controller={controller}
          state={{
            ...IDLE_AUDIO_PLAYBACK,
            track,
            duration: 60,
            currentTime: 15,
            status: "paused",
            canSeek: true,
          }}
          track={track}
        />
      </AudioWaveformProvider>,
    );
    const slider = screen.getByRole("slider", { name: "Seek Voice note" });
    expect(slider).toHaveAttribute("type", "range");
    expect(slider).toBeEnabled();
    expect(slider).toHaveAttribute("aria-valuetext", "0:15 of 1:00");
    slider.focus();
    expect(slider).toHaveFocus();
    // The browser supplies keyboard/pointer range changes; jsdom does not
    // implement the default arrow-key action on native range inputs.
    fireEvent.change(slider, { target: { value: "25" } });
    expect(seek).toHaveBeenCalledWith(25);
    expect(play).not.toHaveBeenCalled();
  });
});
