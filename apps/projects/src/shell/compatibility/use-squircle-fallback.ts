import type { HyperellipseController } from "hyperellipse";
import { useEffect } from "react";

const loadHyperellipse = () => import("hyperellipse");

/** Mount from the hydrated workspace or settings shell, not a streaming ancestor. */
export const useSquircleFallback = () => {
  useEffect(() => {
    if (
      typeof CSS === "undefined" ||
      typeof CSS.supports !== "function" ||
      CSS.supports("corner-shape", "squircle")
    ) {
      return;
    }

    let cancelled = false;
    let controller: HyperellipseController | undefined;
    const root = document.documentElement;

    void loadHyperellipse()
      .then(({ registerHyperellipse }) => {
        if (cancelled) return;

        // CSS carriers are scoped to this marker, including portaled menus.
        root.dataset.squircleFallback = "true";
        controller = registerHyperellipse({ pendingRadiusScale: 0.75 / 2.2 });
      })
      .catch((error: unknown) => {
        if (cancelled) return;
        delete root.dataset.squircleFallback;
        // eslint-disable-next-line no-console -- A cosmetic fallback must not interrupt the app; retain diagnostics.
        console.error("Could not initialize squircle fallback", error);
      });

    return () => {
      cancelled = true;
      controller?.destroy();
      delete root.dataset.squircleFallback;
    };
  }, []);
};
