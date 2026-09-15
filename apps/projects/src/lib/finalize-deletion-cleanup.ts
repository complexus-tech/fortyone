/** Once deletion succeeds, local storage failure must not leave its UI active. */
export const finalizeDeletionCleanup = (
  cleanup: () => void | Promise<void>,
  navigate: () => void,
): Promise<void> => Promise.resolve().then(cleanup).finally(navigate);
