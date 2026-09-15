import { isRichTextValue, type RichTextValue } from "./content";

/** Expo DOM imperative calls return void; only a matching persisted ACK completes a flush. */
export const createEditorFlushController = (timeoutMs = 15_000) => {
  let nextId = 0;
  let disposed = false;
  let pending: {
    id: number;
    promise: Promise<RichTextValue>;
    resolve: (value: RichTextValue) => void;
    reject: (error: Error) => void;
    timer: ReturnType<typeof setTimeout>;
  } | null = null;

  const fail = (error: Error) => {
    const request = pending;
    if (!request) return;
    pending = null;
    clearTimeout(request.timer);
    request.reject(error);
  };

  return {
    request: (send: (requestId: number) => void): Promise<RichTextValue> => {
      if (disposed) return Promise.reject(new Error("The editor was closed."));
      if (pending) return pending.promise;
      const id = ++nextId;
      let resolve!: (value: RichTextValue) => void;
      let reject!: (error: Error) => void;
      const promise = new Promise<RichTextValue>((onResolve, onReject) => {
        resolve = onResolve;
        reject = onReject;
      });
      const timer = setTimeout(
        () =>
          fail(
            new Error(
              "The editor did not finish saving. Keep this screen open and try again.",
            ),
          ),
        timeoutMs,
      );
      pending = { id, promise, resolve, reject, timer };
      try {
        send(id);
      } catch (cause) {
        fail(
          cause instanceof Error
            ? cause
            : new Error("The editor is not ready."),
        );
      }
      return promise;
    },
    acknowledge: (requestId: number, value: unknown, error: string | null) => {
      if (!pending || pending.id !== requestId) return;
      if (error || !isRichTextValue(value)) {
        fail(
          new Error(
            error || "The editor returned an invalid draft. Please try again.",
          ),
        );
        return;
      }
      const request = pending;
      pending = null;
      clearTimeout(request.timer);
      request.resolve(value);
    },
    fail,
    dispose: () => {
      disposed = true;
      fail(new Error("The editor closed before saving finished."));
    },
  };
};
