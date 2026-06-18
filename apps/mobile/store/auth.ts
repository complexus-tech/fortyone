import { create } from "zustand";
import {
  saveWorkspace,
  getWorkspace,
  clearWorkspace,
  saveSessionFlag,
  getSessionFlag,
  clearSessionFlag,
} from "@/lib/auth";
import { getCurrentUser, clearSession } from "@/lib/actions/session";

interface AuthState {
  // State
  workspace: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;

  // Actions
  setWorkspace: (workspace: string) => void;
  setAuthData: (workspace: string) => void;
  clearAuth: () => Promise<void>;
  loadAuthData: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set) => ({
  // Initial state
  workspace: null,
  isAuthenticated: false,
  isLoading: true,

  setWorkspace: (workspace: string) => {
    set({ workspace });
    saveWorkspace(workspace);
  },

  setAuthData: (workspace: string) => {
    set({
      workspace,
      isAuthenticated: true,
      isLoading: false,
    });
    saveSessionFlag();
    saveWorkspace(workspace);
  },

  clearAuth: async () => {
    set({
      workspace: null,
      isAuthenticated: false,
      isLoading: false,
    });
    await Promise.all([clearSessionFlag(), clearWorkspace()]);

    try {
      await clearSession();
    } catch {
      // Local auth state should stay cleared if the remote session is already gone.
    }
  },

  loadAuthData: async () => {
    try {
      const [hasSession, workspace] = await Promise.all([
        getSessionFlag(),
        getWorkspace(),
      ]);

      if (!hasSession) {
        set({
          workspace: null,
          isAuthenticated: false,
          isLoading: false,
        });
        return;
      }

      await getCurrentUser();
      set({
        workspace,
        isAuthenticated: true,
        isLoading: false,
      });
    } catch {
      await Promise.all([clearSessionFlag(), clearWorkspace()]);
      set({
        workspace: null,
        isAuthenticated: false,
        isLoading: false,
      });
    }
  },
}));
