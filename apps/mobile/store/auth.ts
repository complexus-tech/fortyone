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

const sampleToken = process.env.EXPO_PUBLIC_TOKEN ?? "";

export const useAuthStore = create<AuthState>((set) => ({
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
    clearWorkspace();
  },

  loadAuthData: async () => {
    try {
      const [token, workspace] = await Promise.all([
        // getAccessToken(),
        // getWorkspace(),
        "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI4YTc5ODExMi05MGZlLTQ5NWUtOWYxYy1mMzY2NTVlM2Q4YWIiLCJleHAiOjE3NjI4ODMyMDMsIm5iZiI6MTc1OTQyNzIwMywiaWF0IjoxNzU5NDI3MjAzfQ.K05W85tEEWQ5dFqu7bgXjjowkk_zYowwKSJ_VMXR7_o",
        "complexus",
      ]);
      if (token) {
        set({
          token,
          workspace,
          isAuthenticated: true,
        });
      }
    } catch {
      set({
        token: null,
        workspace: null,
        isAuthenticated: false,
      });
    }
  },
}));
