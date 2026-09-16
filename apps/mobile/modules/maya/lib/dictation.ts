export type MayaDictationStatus =
  | "idle"
  | "preparing"
  | "recording"
  | "transcribing";
export const MAYA_DICTATION_SECONDS = 60;

export const isCurrentDictationEvent = (
  event: { url: string | null },
  currentUri: string | null,
  isRecording: boolean,
) =>
  !isRecording &&
  Boolean(currentUri) &&
  (event.url === null || event.url === currentUri);

type DictationSnapshot = {
  status: MayaDictationStatus;
  seconds: number;
  error: string | null;
};
type RecordingRun = {
  scopeKey: string;
  cancelled: boolean;
  request: AbortController;
  operation: Promise<void> | null;
  stop: Promise<string | null> | null;
  cleanup: Promise<void> | null;
  released: boolean;
  uri: string | null;
};

/** Serializes microphone ownership; cancelling never waits for a stale network reply. */
export const createMayaDictation = (options: {
  getScopeKey: () => string;
  isCurrent: () => boolean;
  permission: () => Promise<boolean>;
  prepare: () => Promise<void>;
  record: () => void;
  stop: () => Promise<string | null>;
  release: (uri: string) => void | Promise<void>;
  transcribe: (uri: string, signal: AbortSignal) => Promise<string>;
  onText: (text: string) => void;
}) => {
  let current: RecordingRun | null = null;
  let snapshot: DictationSnapshot = { status: "idle", seconds: 0, error: null };
  const listeners = new Set<() => void>();
  const publish = (change: Partial<DictationSnapshot>) => {
    snapshot = { ...snapshot, ...change };
    listeners.forEach((listener) => listener());
  };
  const assertCurrent = (run: RecordingRun) => {
    if (
      current !== run ||
      run.cancelled ||
      !options.isCurrent() ||
      run.scopeKey !== options.getScopeKey()
    )
      throw new Error("Dictation cancelled.");
  };
  const stop = (run: RecordingRun) => {
    if (!run.stop) {
      run.stop = options
        .stop()
        .then((uri) => {
          run.uri = uri;
          return uri;
        })
        .catch((error: unknown) => {
          run.stop = null;
          throw error;
        });
    }
    return run.stop;
  };
  const release = (run: RecordingRun) => {
    if (!run.cleanup)
      run.cleanup = (async () => {
        await stop(run);
        if (run.uri && !run.released) {
          await options.release(run.uri);
          run.released = true;
        }
      })().catch((error: unknown) => {
        run.cleanup = null;
        throw error;
      });
    return run.cleanup;
  };
  const settle = (run: RecordingRun, error: unknown = null) => {
    if (current !== run) return;
    current = null;
    publish({
      status: "idle",
      seconds: 0,
      error: error instanceof Error ? error.message : null,
    });
  };
  const cleanupFailure = (run: RecordingRun, error: unknown) => {
    if (current !== run) return;
    publish({
      error:
        error instanceof Error
          ? error.message
          : "Could not finish microphone cleanup. Try Cancel again.",
    });
  };
  const finish = async () => {
    const run = current;
    if (!run || snapshot.status !== "recording") return;
    publish({ status: "transcribing", error: null });
    const operation = (async () => {
      let failure: unknown = null;
      try {
        const uri = await stop(run);
        assertCurrent(run);
        if (!uri)
          throw new Error("No recording was captured. Try dictation again.");
        const text = (await options.transcribe(uri, run.request.signal)).trim();
        assertCurrent(run);
        if (!text)
          throw new Error("No speech was detected. Try dictation again.");
        await release(run);
        assertCurrent(run);
        options.onText(text);
      } catch (error) {
        failure = run.cancelled ? null : error;
      }
      try {
        await release(run);
        settle(run, failure);
      } catch (error) {
        cleanupFailure(run, error);
        throw error;
      }
      if (failure) throw failure;
    })();
    run.operation = operation;
    return operation;
  };
  const cancel = async () => {
    const run = current;
    if (!run) return;
    run.cancelled = true;
    run.request.abort();
    // A pending native prepare can acquire the mic after an early stop. Wait
    // for that boundary; transcription itself is safely fenced and aborted.
    if (snapshot.status === "preparing")
      await run.operation?.catch(() => undefined);
    try {
      await release(run);
      settle(run);
    } catch (error) {
      cleanupFailure(run, error);
      throw error;
    }
  };
  return {
    getSnapshot: () => snapshot,
    subscribe: (listener: () => void) => {
      listeners.add(listener);
      return () => listeners.delete(listener);
    },
    clearError: () => publish({ error: null }),
    start: async () => {
      if (current)
        throw new Error("Finish or cancel the current dictation first.");
      if (!options.isCurrent() || !options.getScopeKey())
        throw new Error("Open an active Maya conversation before dictating.");
      const run: RecordingRun = {
        scopeKey: options.getScopeKey(),
        cancelled: false,
        request: new AbortController(),
        operation: null,
        stop: null,
        cleanup: null,
        released: false,
        uri: null,
      };
      current = run;
      publish({ status: "preparing", seconds: 0, error: null });
      const operation = (async () => {
        try {
          if (!(await options.permission()))
            throw new Error(
              "Allow microphone access in Settings to use dictation.",
            );
          assertCurrent(run);
          await options.prepare();
          assertCurrent(run);
          options.record();
          publish({ status: "recording" });
        } catch (error) {
          try {
            await release(run);
            settle(run, run.cancelled ? null : error);
          } catch (cleanupError) {
            cleanupFailure(run, cleanupError);
            throw cleanupError;
          }
          if (!run.cancelled) throw error;
        }
      })();
      run.operation = operation;
      await operation;
    },
    finish,
    cancel,
    interrupt: async () => {
      await cancel();
      publish({
        error: "Dictation was interrupted. Tap the microphone to try again.",
      });
    },
    progress: async (seconds: number) => {
      if (snapshot.status !== "recording") return;
      publish({
        seconds: Math.min(
          MAYA_DICTATION_SECONDS,
          Math.max(0, Math.floor(seconds)),
        ),
      });
      if (seconds >= MAYA_DICTATION_SECONDS) await finish();
    },
  };
};
