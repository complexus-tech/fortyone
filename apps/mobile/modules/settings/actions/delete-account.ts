import { getStoredSession } from "@/lib/auth";
import {
  createAccountDeletionOperation,
  type AccountDeletionScope,
} from "@/lib/account-deletion";
import { request } from "@/lib/http";
import { getApiURL } from "@/lib/http/config";
import { useAuthStore } from "@/store/auth";

export const isAccountDeletionSession = (scope: AccountDeletionScope) => {
  const current = useAuthStore.getState();
  return (
    current.isAuthenticated &&
    !current.isLoading &&
    current.userId === scope.userId &&
    current.sessionEpoch === scope.sessionEpoch
  );
};

export const createDeleteAccountOperation = () =>
  createAccountDeletionOperation({
    isCurrent: isAccountDeletionSession,
    getSession: getStoredSession,
    send: (session) => {
      if (session.apiOrigin !== getApiURL().origin)
        throw new Error(
          "Sign in again for this API environment before deleting your account.",
        );
      return request("delete", "users/account", undefined, {
        useWorkspace: false,
        sessionCookie: session.cookie,
        // Only an accepted deletion performs local deletion cleanup. A 401 or
        // 409 remains a visible failure for this particular request.
        handleUnauthorized: false,
      });
    },
    complete: (cookie, result) =>
      useAuthStore.getState().completeAccountDeletion(cookie, result),
  });
