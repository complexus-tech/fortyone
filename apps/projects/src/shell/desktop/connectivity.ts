"use client";

import { useSyncExternalStore } from "react";
import { getApiUrl } from "@/lib/api-url";
import { getDesktopBridge } from "./bridge";

export const DESKTOP_CONNECTIVITY_EVENT = "fortyone:desktop-connectivity";
const RETRY_INTERVAL = 20_000;
const PROBE_TIMEOUT = 5_000;
let probe: Promise<boolean> | null = null;

declare global {
  interface Window {
    __FORTYONE_OFFLINE_PAGE__?: boolean;
  }
}

export function getDesktopOnline() {
  if (typeof window === "undefined") return true;
  return navigator.onLine && !window.__FORTYONE_OFFLINE_PAGE__;
}

function publish(online: boolean) {
  const previous = getDesktopOnline();
  window.__FORTYONE_OFFLINE_PAGE__ = !online;
  if (previous !== getDesktopOnline())
    window.dispatchEvent(
      new CustomEvent(DESKTOP_CONNECTIVITY_EVENT, { detail: { online } }),
    );
}

export function probeDesktopConnectivity(): Promise<boolean> {
  if (!getDesktopBridge()) return Promise.resolve(navigator.onLine);
  if (probe) return probe;
  if (!navigator.onLine) {
    publish(false);
    return Promise.resolve(false);
  }
  const controller = new AbortController();
  const timeout = setTimeout(() => {
    controller.abort();
  }, PROBE_TIMEOUT);
  const task = (async () => {
    try {
      const web = new URL("/desktop-worker.js", location.origin);
      web.searchParams.set("__fortyone_probe", String(Date.now()));
      const api = new URL("/readiness", getApiUrl());
      api.searchParams.set("__fortyone_probe", String(Date.now()));
      const options = {
        cache: "no-store" as const,
        credentials: "omit" as const,
        signal: controller.signal,
      };
      const results = await Promise.all([
        fetch(web, { ...options, method: "HEAD" }),
        fetch(api, { ...options, method: "GET" }),
      ]);
      const online = results.every((result) => result.ok);
      publish(online);
      return online;
    } catch {
      publish(false);
      return false;
    } finally {
      clearTimeout(timeout);
    }
  })();
  probe = task;
  void task.finally(() => {
    if (probe === task) probe = null;
  });
  return task;
}

export function monitorDesktopConnectivity() {
  if (!getDesktopBridge()) return () => undefined;
  let timer: ReturnType<typeof setTimeout> | undefined;
  let active = true;
  const schedule = () => {
    clearTimeout(timer);
    if (active && !getDesktopOnline())
      timer = setTimeout(check, RETRY_INTERVAL);
  };
  const check = () => {
    void probeDesktopConnectivity().finally(schedule);
  };
  const offline = () => {
    publish(false);
    schedule();
  };
  window.addEventListener("online", check);
  window.addEventListener("offline", offline);
  window.addEventListener("focus", check);
  window.addEventListener(DESKTOP_CONNECTIVITY_EVENT, schedule);
  if (!getDesktopOnline()) check();
  return () => {
    active = false;
    clearTimeout(timer);
    window.removeEventListener("online", check);
    window.removeEventListener("offline", offline);
    window.removeEventListener("focus", check);
    window.removeEventListener(DESKTOP_CONNECTIVITY_EVENT, schedule);
  };
}

const subscribe = (callback: () => void) => {
  window.addEventListener(DESKTOP_CONNECTIVITY_EVENT, callback);
  window.addEventListener("offline", callback);
  return () => {
    window.removeEventListener(DESKTOP_CONNECTIVITY_EVENT, callback);
    window.removeEventListener("offline", callback);
  };
};
export const useDesktopOnline = () =>
  useSyncExternalStore(subscribe, getDesktopOnline, () => true);
