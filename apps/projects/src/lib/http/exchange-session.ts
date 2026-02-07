export const exchangeSessionToken = async () => {
  try {
    await fetch("/api/session/finalize", {
      method: "POST",
      credentials: "include",
    });
  } catch {
    // Exchange is best-effort; keep UI responsive if it fails.
  }
};
