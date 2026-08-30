"use client";

import type { RefObject } from "react";
import { useEffect } from "react";
import type { FeedbackWidgetMode, FeedbackWidgetTheme } from "../protocol";
import { postFeedbackWidgetMessage } from "../protocol";

export const useWidgetTheme = (theme: FeedbackWidgetTheme) => {
  useEffect(() => {
    const documentRoot = document.documentElement;
    const initiallyDark = documentRoot.classList.contains("dark");
    const applyTheme = (dark: boolean) => {
      documentRoot.classList.toggle("dark", dark);
    };

    if (theme !== "auto") {
      applyTheme(theme === "dark");
      return () => {
        documentRoot.classList.toggle("dark", initiallyDark);
      };
    }
    const media = window.matchMedia("(prefers-color-scheme: dark)");
    const update = () => {
      applyTheme(media.matches);
    };
    update();
    media.addEventListener("change", update);
    return () => {
      media.removeEventListener("change", update);
      documentRoot.classList.toggle("dark", initiallyDark);
    };
  }, [theme]);
};

export const useWidgetFrameEmbedEvents = (
  instanceId: string,
  mode: FeedbackWidgetMode,
  rootRef: RefObject<HTMLDivElement | null>,
  trustedParentOrigin: string | null,
) => {
  useEffect(() => {
    if (!trustedParentOrigin || mode === "inline") return;
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        postFeedbackWidgetMessage("escape", instanceId, trustedParentOrigin);
      }
    };
    window.addEventListener("keydown", handleEscape);
    return () => {
      window.removeEventListener("keydown", handleEscape);
    };
  }, [instanceId, mode, trustedParentOrigin]);

  useEffect(() => {
    if (!trustedParentOrigin || mode !== "inline" || !rootRef.current) return;
    const root = rootRef.current;
    const observer = new ResizeObserver(() => {
      postFeedbackWidgetMessage("resize", instanceId, trustedParentOrigin, {
        height: Math.ceil(root.scrollHeight),
      });
    });
    observer.observe(root);
    return () => {
      observer.disconnect();
    };
  }, [instanceId, mode, rootRef, trustedParentOrigin]);
};
