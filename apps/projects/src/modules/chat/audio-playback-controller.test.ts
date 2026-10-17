/* global describe, expect, it, jest -- Provided by the projects test runner. */
import { createAudioPlaybackController } from "./audio-playback-controller";
import type { AudioTrack } from "./audio-playback-controller";

const first: AudioTrack = {
  id: "first",
  messageId: "message-one",
  fileName: "First.webm",
  url: "/api/attachments/first",
  durationSeconds: 30,
};
const second: AudioTrack = {
  id: "second",
  messageId: "message-two",
  fileName: "Second.webm",
  url: "/api/attachments/second",
  durationSeconds: 45,
};

const setup = () => {
  const audio = Object.assign(new EventTarget(), {
    preload: "",
    src: "",
    currentTime: 0,
    duration: 0,
    readyState: 0,
    paused: true,
    ended: false,
    playbackRate: 1,
    error: null as { code: number } | null,
    load: jest.fn(),
    removeAttribute: jest.fn((name: string) => {
      if (name === "src") audio.src = "";
    }),
    pause: jest.fn(() => {
      audio.paused = true;
      audio.dispatchEvent(new Event("pause"));
    }),
    play: jest.fn(async () => {
      audio.paused = false;
    }),
  });
  const createAudio = jest.fn(() => audio as unknown as HTMLAudioElement);
  return {
    audio,
    createAudio,
    controller: createAudioPlaybackController(createAudio),
  };
};

describe("conversation audio playback", () => {
  it("creates no media element until an explicit play and reuses one element for subsequent messages", async () => {
    const { audio, createAudio, controller } = setup();
    expect(createAudio).not.toHaveBeenCalled();
    await controller.play(first);
    expect(audio.preload).toBe("none");
    expect(audio.src).toBe(first.url);
    expect(controller.getSnapshot().status).toBe("playing");
    await controller.play(second);
    expect(createAudio).toHaveBeenCalledTimes(1);
    expect(audio.pause).toHaveBeenCalledTimes(2);
    expect(audio.src).toBe(second.url);
    expect(controller.getSnapshot().track).toBe(second);
  });

  it("ignores an old play rejection after another recording is selected", async () => {
    const { audio, controller } = setup();
    let rejectOld: (reason: Error) => void = () => undefined;
    audio.play.mockImplementationOnce(
      () =>
        new Promise<void>((_resolve, reject) => {
          rejectOld = reject;
        }),
    );
    const pending = controller.play(first);
    await controller.play(second);
    rejectOld(new Error("Source changed"));
    await pending;
    expect(controller.getSnapshot()).toMatchObject({
      track: second,
      status: "playing",
      error: null,
    });
  });

  it("keeps a pending play paused when its promise resolves after the user pauses", async () => {
    const { audio, controller } = setup();
    let finish: () => void = () => undefined;
    audio.play.mockImplementationOnce(
      () =>
        new Promise<void>((resolve) => {
          finish = resolve;
        }),
    );
    const pending = controller.play(first);
    controller.pause();
    finish();
    await pending;
    expect(controller.getSnapshot().status).toBe("paused");
  });

  it("uses the recorded duration when WebM metadata is infinite and bounds seeking and speed", async () => {
    const { audio, controller } = setup();
    await controller.play(first);
    audio.duration = Infinity;
    audio.readyState = 1;
    audio.dispatchEvent(new Event("loadedmetadata"));
    expect(controller.getSnapshot()).toMatchObject({
      duration: 30,
      canSeek: true,
    });
    controller.seek(100);
    expect(audio.currentTime).toBe(30);
    controller.seek(-1);
    expect(audio.currentTime).toBe(0);
    controller.setRate(1.5);
    expect(audio.playbackRate).toBe(1.5);
    controller.setRate(9);
    expect(audio.playbackRate).toBe(1.5);
  });

  it("keeps playing when its last visible row disappears and stops only for its deleted message", async () => {
    const { audio, controller } = setup();
    const main = Symbol("main");
    const thread = Symbol("thread");
    controller.setRowVisible(first.id, main, true);
    controller.setRowVisible(first.id, thread, true);
    await controller.play(first);
    expect(controller.getSnapshot().sourceVisible).toBe(true);
    controller.setRowVisible(first.id, main, false);
    expect(controller.getSnapshot().sourceVisible).toBe(true);
    controller.setRowVisible(first.id, thread, false);
    expect(controller.getSnapshot()).toMatchObject({
      sourceVisible: false,
      status: "playing",
    });
    controller.invalidateMessage(second.messageId);
    expect(audio.src).toBe(first.url);
    controller.invalidateMessage(first.messageId);
    expect(audio.src).toBe("");
    expect(controller.getSnapshot().track).toBeNull();
  });

  it("surfaces media errors and reloads the source on explicit retry", async () => {
    const { audio, controller } = setup();
    await controller.play(first);
    audio.error = { code: 4 };
    audio.dispatchEvent(new Event("error"));
    expect(controller.getSnapshot().error).toContain("cannot play");
    expect(controller.getSnapshot().status).toBe("error");
    audio.error = null;
    await controller.play(first);
    expect(audio.load).toHaveBeenCalledTimes(2);
    expect(controller.getSnapshot()).toMatchObject({
      status: "playing",
      error: null,
    });
  });

  it("coordinates draft review audio and releases media/listeners on teardown", async () => {
    const { audio, controller } = setup();
    const pauseDraft = jest.fn();
    const unregister = controller.registerExternalPause(pauseDraft);
    await controller.play(first);
    expect(pauseDraft).toHaveBeenCalledTimes(1);
    unregister();
    await controller.play(second);
    expect(pauseDraft).toHaveBeenCalledTimes(1);
    controller.destroy();
    expect(audio.src).toBe("");
    expect(audio.paused).toBe(true);
    audio.currentTime = 25;
    audio.dispatchEvent(new Event("timeupdate"));
    expect(controller.getSnapshot()).toMatchObject({
      track: null,
      status: "idle",
      currentTime: 0,
    });
  });
});
