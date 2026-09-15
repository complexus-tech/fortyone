/** Keeps a known successful send separate from retryable local draft cleanup. */
export const createCommentFinalizer = (onSent: () => void) => {
  let sent = false;
  let closed = false;
  let cleared = false;
  let pending: Promise<void> | null = null;

  const run = (operation: () => Promise<void>): Promise<void> => {
    if (pending) return pending;
    pending = Promise.resolve()
      .then(operation)
      .finally(() => {
        pending = null;
      });
    return pending;
  };

  const clear = async (removeDraft: () => Promise<void>) => {
    if (cleared) return;
    try {
      await removeDraft();
    } catch (cause) {
      if (!sent) throw cause;
      throw new Error(
        "Your comment was sent, but its draft could not be removed from this device. Try again to finish without sending it again.",
      );
    }
    cleared = true;
  };

  return {
    persistDraft: (write: () => Promise<void>): Promise<void> => {
      if (sent || closed || pending) {
        return Promise.reject(
          new Error("This comment is no longer available for editing."),
        );
      }
      return write();
    },
    send: (create: () => Promise<void>, removeDraft: () => Promise<void>) =>
      run(async () => {
        if (closed) return;
        if (!sent) {
          await create();
          // Close the write gate before publishing state to React or awaiting cleanup.
          sent = true;
          onSent();
        }
        await clear(removeDraft);
        closed = true;
      }),
    close: (removeDraft: () => Promise<void>) =>
      run(async () => {
        if (sent) await clear(removeDraft);
        closed = true;
      }),
    discard: (removeDraft: () => Promise<void>) =>
      run(async () => {
        await clear(removeDraft);
        closed = true;
      }),
  };
};
