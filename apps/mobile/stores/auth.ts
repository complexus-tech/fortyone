import { create } from "zustand";
import { getAccessToken, saveAccessToken, clearAccessToken } from "@/lib/auth";

interface AuthState {
  // State
  token: string | null;
  workspaceId: string | null;
  isAuthenticated: boolean;

  // Actions
  setToken: (token: string) => void;
  setWorkspaceId: (workspaceId: string) => void;
  setAuthData: (token: string, workspaceId?: string) => void;
  clearAuth: () => void;
  loadAuthData: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  // Initial state
  token: null,
  workspaceId: null,
  isAuthenticated: false,

  // Actions
  setToken: (token: string) => {
    set({ token, isAuthenticated: true });
    saveAccessToken(token);
  },

  setWorkspaceId: (workspaceId: string) => {
    set({ workspaceId });
    // Also save to SecureStore if needed
  },

  setAuthData: (token: string, workspaceId?: string) => {
    set({
      token,
      workspaceId: workspaceId || null,
      isAuthenticated: true,
    });
    saveAccessToken(token);
    // Save workspaceId to SecureStore if needed
  },

  clearAuth: () => {
    set({
      token: null,
      workspaceId: null,
      isAuthenticated: false,
    });
    clearAccessToken();
  },

  loadAuthData: async () => {
    try {
      const token = await getAccessToken();
      if (token) {
        set({
          token,
          isAuthenticated: true,
        });
      }
    } catch (error) {
      console.error("Failed to load auth data:", error);
    }
  },
}));
