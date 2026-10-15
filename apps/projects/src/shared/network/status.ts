"use client";

import { useSyncExternalStore } from "react";

export const DESKTOP_CONNECTIVITY_EVENT = "fortyone:desktop-connectivity";

declare global {
  interface Window {
    __FORTYONE_OFFLINE_PAGE__?: boolean;
  }
}

// The desktop shell publishes failed service probes here. A network connection
// alone does not mean a restored offline page can reach the application.
export function getNetworkOnline() {
  if (typeof window === "undefined") return true;
  return navigator.onLine && !window.__FORTYONE_OFFLINE_PAGE__;
}

export function subscribeToNetworkStatus(callback: () => void) {
  window.addEventListener("online", callback);
  window.addEventListener("offline", callback);
  window.addEventListener(DESKTOP_CONNECTIVITY_EVENT, callback);
  return () => {
    window.removeEventListener("online", callback);
    window.removeEventListener("offline", callback);
    window.removeEventListener(DESKTOP_CONNECTIVITY_EVENT, callback);
  };
}

export const useNetworkOnline = () =>
  useSyncExternalStore(subscribeToNetworkStatus, getNetworkOnline, () => true);
