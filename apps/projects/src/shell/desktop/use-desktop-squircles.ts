import type { HyperellipseController } from "hyperellipse";
import { useEffect } from "react";

const loadHyperellipse = () => import("hyperellipse");

/** Load the compatibility renderer only in desktop webviews that need it. */
export const useDesktopSquircles = (isDesktop: boolean) => {
  useEffect(() => {
    if (
      !isDesktop ||
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
        root.dataset.desktopSquircles = "true";
        controller = registerHyperellipse({ pendingRadiusScale: 0.75 / 2.2 });
      })
      .catch((error: unknown) => {
        if (cancelled) return;
        delete root.dataset.desktopSquircles;
        // eslint-disable-next-line no-console -- A cosmetic fallback must not interrupt the app; retain diagnostics.
        console.error("Could not initialize desktop squircles", error);
      });

    return () => {
      cancelled = true;
      controller?.destroy();
      delete root.dataset.desktopSquircles;
    };
  }, [isDesktop]);
};
