/** Collect the latest snapshots from active editors before leaving the page. */
export async function flushDesktopEditors() {
  const writes: Promise<unknown>[] = [];
  window.dispatchEvent(
    new CustomEvent("fortyone:desktop-before-close", {
      detail: {
        waitUntil: (write: Promise<unknown>) => {
          writes.push(Promise.resolve(write));
        },
      },
    }),
  );
  let timer: ReturnType<typeof setTimeout> | undefined;
  try {
    await Promise.race([
      Promise.all(writes),
      new Promise<never>((_, reject) => {
        timer = setTimeout(() => {
          reject(
            new Error(
              "Saving took too long. Your current page stayed open; retry saving before continuing.",
            ),
          );
        }, 12_000);
      }),
    ]);
  } finally {
    clearTimeout(timer);
  }
}
