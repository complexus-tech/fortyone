/** Attempt every cleanup step and keep authentication blocked until drafts settle. */
export async function clearLocalAccountData(steps: {
  resetCache: () => Promise<void>;
  clearCredentials: () => Promise<void>;
  clearDrafts: () => Promise<void>;
  finish: () => void;
}) {
  try {
    await steps.resetCache();
  } finally {
    try {
      await steps.clearCredentials();
    } finally {
      try {
        await steps.clearDrafts();
      } finally {
        steps.finish();
      }
    }
  }
}
