type PushCleanup = () => Promise<void>;

let cleanup: PushCleanup = () => Promise.resolve();

export const registerPushCleanup = (next: PushCleanup) => {
  cleanup = next;
};

export const runPushCleanup = () => cleanup();
