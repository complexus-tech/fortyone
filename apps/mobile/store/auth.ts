import { create } from "zustand";
import {
  clearStoredSession,
  getStoredSession,
  saveSession,
  clearSignInTransaction,
} from "@/lib/auth";
import { getCurrentUser, clearSession } from "@/lib/actions/session";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { HttpError } from "@/lib/http";
import { resetSessionCache } from "@/lib/session-cache";
import { clearMobileDrafts } from "@/components/rich-text/draft-store";
import {
  completeDeletedAccount,
  type AccountDeletionResult,
} from "@/lib/account-deletion";
import { clearLocalAccountData } from "@/lib/session-cleanup";

type AuthData = { workspace: string; userId: string };
interface AuthState {
  workspace: string | null;
  userId: string | null;
  sessionEpoch: number;
  sessionError: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  setWorkspace: (workspace: string) => Promise<void>;
  setAuthData: (data: AuthData) => Promise<void>;
  clearAuth: () => Promise<void>;
  completeAccountDeletion: (
    cookie: string,
    result: AccountDeletionResult,
  ) => Promise<boolean>;
  expireSession: (cookie: string) => Promise<void>;
  loadAuthData: () => Promise<void>;
}

let operationVersion = 0;
let clearingSession: Promise<void> | null = null;

export const useAuthStore = create<AuthState>((set, get) => {
  const clearLocalSession = async (
    message: string | null = null,
    clearTransaction = true,
  ) => {
    if (clearingSession) return clearingSession;
    operationVersion++;
    set({ isLoading: true });
    clearingSession = clearLocalAccountData({
      resetCache: resetSessionCache,
      clearCredentials: async () => {
        await Promise.all([
          clearStoredSession(),
          ...(clearTransaction ? [clearSignInTransaction()] : []),
        ]);
      },
      clearDrafts: clearMobileDrafts,
      finish: () => {
        set((state) => ({
          workspace: null,
          userId: null,
          isAuthenticated: false,
          isLoading: false,
          sessionEpoch: state.sessionEpoch + 1,
          sessionError: message,
        }));
      },
    });
    try {
      await clearingSession;
    } finally {
      clearingSession = null;
    }
  };

  return {
    workspace: null,
    userId: null,
    sessionEpoch: 0,
    sessionError: null,
    isAuthenticated: false,
    isLoading: true,

    setWorkspace: async (workspace) => {
      const session = await getStoredSession();
      if (!session)
        throw new Error("Your session expired. Please sign in again.");
      const version = ++operationVersion;
      await saveSession(
        { ...session, workspace },
        () => version === operationVersion,
      );
      if (version !== operationVersion) return;
      try {
        await resetSessionCache();
      } finally {
        if (version === operationVersion) {
          set((state) => ({
            workspace,
            sessionEpoch: state.sessionEpoch + 1,
            sessionError: null,
          }));
        }
      }
    },

    setAuthData: async ({ workspace, userId }) => {
      const version = ++operationVersion;
      const session = await getStoredSession();
      if (!session || session.userId !== userId)
        throw new Error("Sign-in could not be restored. Please try again.");
      await saveSession(
        { ...session, workspace },
        () => version === operationVersion,
      );
      if (version !== operationVersion) return;
      try {
        await resetSessionCache();
      } finally {
        if (version === operationVersion) {
          set((state) => ({
            workspace,
            userId,
            isAuthenticated: true,
            isLoading: false,
            sessionEpoch: state.sessionEpoch + 1,
            sessionError: null,
          }));
        }
      }
    },

    clearAuth: async () => {
      const session = await getStoredSession();
      await clearLocalSession();
      if (!session) return;
      try {
        await clearSession(session.cookie, session.apiOrigin);
      } catch {
        if (!get().isAuthenticated) {
          set({
            sessionError:
              "Signed out on this device. The server could not confirm session revocation.",
          });
        }
      }
    },

    completeAccountDeletion: (cookie, result) =>
      completeDeletedAccount(cookie, result, {
        getVersion: () => operationVersion,
        getSession: getStoredSession,
        clearLocalSession,
        reportCleanupFailure: (message) => {
          if (!get().isAuthenticated) set({ sessionError: message });
        },
      }),

    expireSession: async (cookie) => {
      const session = await getStoredSession();
      // A late 401 from an old account must never clear a newer sign-in.
      if (session ? session.cookie !== cookie : cookie !== "") return;
      await clearLocalSession("Your session expired. Please sign in again.");
    },

    loadAuthData: async () => {
      const version = ++operationVersion;
      set({ isLoading: true, sessionError: null });
      let session: Awaited<ReturnType<typeof getStoredSession>> = null;
      try {
        session = await getStoredSession();
        if (version !== operationVersion) return;
        if (!session) {
          await clearLocalSession(null, false);
          return;
        }
        const user = await getCurrentUser();
        if (version !== operationVersion) return;
        if (user.id !== session.userId)
          throw new HttpError("The session account changed.", 401);
        const workspaces = await getWorkspaces();
        if (version !== operationVersion) return;
        const workspace =
          workspaces.find((item) => item.slug === session?.workspace) ??
          workspaces.find((item) => item.id === user.lastUsedWorkspaceId) ??
          workspaces[0];
        if (!workspace)
          throw new Error(
            "This account has no workspace. Complete workspace setup to continue.",
          );
        await saveSession(
          { ...session, workspace: workspace.slug },
          () => version === operationVersion,
        );
        if (version !== operationVersion) return;
        set({
          workspace: workspace.slug,
          userId: user.id,
          isAuthenticated: true,
          isLoading: false,
          sessionError: null,
        });
      } catch (error) {
        if (version !== operationVersion) return;
        if (error instanceof HttpError && error.status === 401) {
          await clearLocalSession(
            "Your session expired. Please sign in again.",
          );
          return;
        }
        // Connectivity and service failures do not revoke an existing session.
        set({
          workspace: session?.workspace ?? null,
          userId: session?.userId ?? null,
          isAuthenticated: Boolean(session?.workspace),
          isLoading: false,
          sessionError: session?.workspace
            ? "Unable to reconnect. Your saved work is available; try again when you are online."
            : error instanceof Error
              ? error.message
              : "Unable to restore sign-in. Please try again.",
        });
      }
    },
  };
});
