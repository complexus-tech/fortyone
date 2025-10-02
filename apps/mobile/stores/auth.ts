import { create } from "zustand";
import {
  getAccessToken,
  saveAccessToken,
  clearAccessToken,
  saveWorkspace,
  getWorkspace,
  clearWorkspace,
} from "@/lib/auth";

interface AuthState {
  // State
  token: string | null;
  workspace: string | null;
  isAuthenticated: boolean;

  // Actions
  setToken: (token: string) => void;
  setWorkspace: (workspace: string) => void;
  setAuthData: (token: string, workspace?: string) => void;
  clearAuth: () => void;
  loadAuthData: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  // Initial state
  token: null,
  workspace: null,
  isAuthenticated: false,

  // Actions
  setToken: (token: string) => {
    set({ token, isAuthenticated: true });
    saveAccessToken(token);
  },

  setWorkspace: (workspace: string) => {
    set({ workspace });
    saveWorkspace(workspace);
  },

  setAuthData: (token: string, workspace?: string) => {
    set({
      token,
      workspace: workspace || null,
      isAuthenticated: true,
    });
    saveAccessToken(token);
    if (workspace) {
      saveWorkspace(workspace);
    }
  },

  clearAuth: () => {
    set({
      token: null,
      workspace: null,
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
