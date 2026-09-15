import type { StoredSession } from "./auth-contract";

export type AccountDeletionResult = "deleted" | "cleanup_pending";
export type AccountDeletionScope = { userId: string; sessionEpoch: number };
type DeletionResponse = { status: number; json: () => Promise<unknown> };
type AccountDeletionDependencies = {
  isCurrent: (scope: AccountDeletionScope) => boolean;
  getSession: () => Promise<StoredSession | null>;
  send: (session: StoredSession) => Promise<DeletionResponse>;
  complete: (cookie: string, result: AccountDeletionResult) => Promise<boolean>;
};

export class AccountDeletionCleanupError extends Error {
  constructor(result: AccountDeletionResult) {
    super(
      `${accountDeletionMessage(result)} Could not finish signing out on this device. Choose Finish signing out to retry local cleanup.`,
    );
    this.name = "AccountDeletionCleanupError";
  }
}

export const accountDeletionMessage = (result: AccountDeletionResult) =>
  "Your account has been deleted. Shared workspace contributions remain attributed to Former user." +
  (result === "cleanup_pending"
    ? " Cleanup of connected-service data is continuing."
    : "");

export async function readAccountDeletionResult(
  response: DeletionResponse,
): Promise<AccountDeletionResult> {
  if (response.status === 204) return "deleted";
  if (response.status === 202) {
    const body: unknown = await response.json();
    if (
      body &&
      typeof body === "object" &&
      "data" in body &&
      body.data &&
      typeof body.data === "object" &&
      "status" in body.data &&
      body.data.status === "cleanup_pending"
    ) {
      return "cleanup_pending";
    }
  }
  throw new Error("Account deletion was not confirmed. Please try again.");
}

/** A confirmation can outlive its screen; always send the original credential. */
export async function deleteConfirmedAccount(
  scope: AccountDeletionScope,
  dependencies: AccountDeletionDependencies,
) {
  const assertCurrent = () => {
    if (!dependencies.isCurrent(scope))
      throw new Error("Your session changed. Open Settings and try again.");
  };
  assertCurrent();
  const session = await dependencies.getSession();
  assertCurrent();
  if (!session || session.userId !== scope.userId)
    throw new Error(
      "Your session expired. Sign in again before deleting your account.",
    );
  const result = await readAccountDeletionResult(
    await dependencies.send(session),
  );
  const signedOut = await dependencies.complete(session.cookie, result);
  return { result, signedOut };
}

/** Once accepted, retries can only finish cleanup for that original session. */
export function createAccountDeletionOperation(
  dependencies: AccountDeletionDependencies,
) {
  let accepted: {
    scope: AccountDeletionScope;
    cookie: string;
    result: AccountDeletionResult;
  } | null = null;

  const complete = async (cookie: string, result: AccountDeletionResult) => {
    try {
      return await dependencies.complete(cookie, result);
    } catch {
      // Keep the credential private and preserve the accepted checkpoint.
      throw new AccountDeletionCleanupError(result);
    }
  };

  return async (scope: AccountDeletionScope) => {
    if (accepted) {
      if (
        accepted.scope.userId !== scope.userId ||
        accepted.scope.sessionEpoch !== scope.sessionEpoch ||
        !dependencies.isCurrent(scope)
      ) {
        throw new Error("Your session changed. Open Settings and try again.");
      }
      const signedOut = await complete(accepted.cookie, accepted.result);
      return { result: accepted.result, signedOut };
    }
    return deleteConfirmedAccount(scope, {
      ...dependencies,
      complete: (cookie, result) => {
        accepted = { scope: { ...scope }, cookie, result };
        return complete(cookie, result);
      },
    });
  };
}

/** Check both persisted identity and transitions that can start during the read. */
export async function completeDeletedAccount(
  cookie: string,
  result: AccountDeletionResult,
  dependencies: {
    getVersion: () => number;
    getSession: () => Promise<StoredSession | null>;
    clearLocalSession: (message: string) => Promise<void>;
    reportCleanupFailure: (message: string) => void;
  },
) {
  const version = dependencies.getVersion();
  let session: StoredSession | null;
  try {
    session = await dependencies.getSession();
  } catch {
    // Without a verified identity we cannot erase a potentially newer session.
    throw new AccountDeletionCleanupError(result);
  }
  if (version !== dependencies.getVersion() || session?.cookie !== cookie)
    return false;
  const message = accountDeletionMessage(result);
  try {
    await dependencies.clearLocalSession(message);
  } catch {
    // The server already accepted deletion. Do not invite another DELETE or
    // falsely describe this as an account deletion failure.
    dependencies.reportCleanupFailure(
      `${message} Some saved data could not be removed from this device. Close and reopen the app to retry cleanup.`,
    );
  }
  return true;
}
