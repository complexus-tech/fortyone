const MAX_ACTIVE_THUMBNAILS = 2;
const waiting: (() => void)[] = [];
let active = 0;

const drain = () => {
  while (active < MAX_ACTIVE_THUMBNAILS && waiting.length) {
    waiting.shift()?.();
  }
};

const cancellationError = () => {
  const error = new Error("PDF thumbnail canceled");
  error.name = "AbortError";
  return error;
};

/** A slot includes downloading, painting, and releasing the document's worker. */
export const queuePdfThumbnail = (
  render: () => Promise<void>,
  signal: AbortSignal,
): Promise<void> =>
  new Promise((resolve, reject) => {
    const cancelQueued = () => {
      const index = waiting.indexOf(start);
      if (index >= 0) waiting.splice(index, 1);
      signal.removeEventListener("abort", cancelQueued);
      reject(cancellationError());
    };
    const start = () => {
      signal.removeEventListener("abort", cancelQueued);
      if (signal.aborted) {
        reject(cancellationError());
        return;
      }
      active++;
      void Promise.resolve()
        .then(render)
        .then(resolve, reject)
        .finally(() => {
          active--;
          drain();
        });
    };
    if (signal.aborted) {
      reject(cancellationError());
      return;
    }
    signal.addEventListener("abort", cancelQueued, { once: true });
    waiting.push(start);
    drain();
  });
