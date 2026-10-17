/** One request in flight and one replaceable hint; slow networks never build a backlog. */
export const createTypingActivity = (
  send: (isTyping: boolean) => Promise<unknown>,
) => {
  let active = false;
  let disposed = false;
  let inFlight = false;
  let pending: boolean | undefined;
  let lastSent = 0;
  let idle: ReturnType<typeof setTimeout> | undefined;

  const drain = () => {
    if (inFlight || pending === undefined) return;
    const value = pending;
    pending = undefined;
    inFlight = true;
    void send(value)
      .catch(() => {
        // A lost hint expires on the server and must never fail the message.
      })
      .finally(() => {
        inFlight = false;
        drain();
      });
  };
  const enqueue = (value: boolean) => {
    pending = value;
    drain();
  };
  const stop = () => {
    clearTimeout(idle);
    if (!active) return;
    active = false;
    enqueue(false);
  };

  return {
    changed() {
      if (disposed) return;
      const now = Date.now();
      if (!active || now - lastSent >= 3000) {
        active = true;
        lastSent = now;
        enqueue(true);
      }
      clearTimeout(idle);
      idle = setTimeout(stop, 4000);
    },
    stop,
    dispose() {
      disposed = true;
      stop();
      // Discard queued refreshes from this scope, retaining only its final stop.
      if (pending === true) pending = false;
    },
  };
};
