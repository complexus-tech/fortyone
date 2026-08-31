import { useCallback, useEffect, useState } from "react";
import type { WalkthroughTargetPosition } from "./walkthrough-position";

export interface ResolvedWalkthroughTarget {
  element: Element | null;
  position: WalkthroughTargetPosition;
  isFallback: boolean;
  isSpotlightTarget: boolean;
}

interface WalkthroughTargetSnapshot {
  key: string;
  target: ResolvedWalkthroughTarget | null;
}

interface UseWalkthroughTargetOptions {
  currentTarget: string | undefined;
  isActive: boolean;
}

const fallbackTargetPosition: WalkthroughTargetPosition = {
  top: 0,
  left: 0,
  width: 0,
  height: 0,
};

const TARGET_LAYOUT_CHECK_INTERVAL_MS = 100;
const TARGET_LAYOUT_CHECK_TIMEOUT_MS = 3000;
const TARGET_LAYOUT_STABLE_CHECKS = 2;

const isVisibleTarget = (rect: DOMRect) =>
  rect.height > 0 &&
  rect.width > 0 &&
  rect.bottom > 0 &&
  rect.left < window.innerWidth &&
  rect.right > 0 &&
  rect.top < window.innerHeight;

const resolveWalkthroughTarget = (
  isActive: boolean,
  currentTarget: string | undefined,
): ResolvedWalkthroughTarget | null => {
  if (!isActive || !currentTarget || typeof window === "undefined") {
    return null;
  }

  if (currentTarget === "body") {
    return {
      element: null,
      position: fallbackTargetPosition,
      isFallback: false,
      isSpotlightTarget: false,
    };
  }

  const targetElement = document.querySelector(currentTarget);
  if (!targetElement) {
    return {
      element: null,
      position: fallbackTargetPosition,
      isFallback: true,
      isSpotlightTarget: false,
    };
  }

  const rect = targetElement.getBoundingClientRect();
  if (!isVisibleTarget(rect)) {
    return {
      element: targetElement,
      position: fallbackTargetPosition,
      isFallback: true,
      isSpotlightTarget: false,
    };
  }

  return {
    element: targetElement,
    position: {
      top: rect.top,
      left: rect.left,
      width: rect.width,
      height: rect.height,
    },
    isFallback: false,
    isSpotlightTarget: true,
  };
};

const areTargetPositionsEqual = (
  first: WalkthroughTargetPosition,
  second: WalkthroughTargetPosition,
) =>
  first.top === second.top &&
  first.left === second.left &&
  first.width === second.width &&
  first.height === second.height;

const areResolvedTargetsEqual = (
  first: ResolvedWalkthroughTarget | null,
  second: ResolvedWalkthroughTarget | null,
) => {
  if (first === second) {
    return true;
  }

  if (!first || !second) {
    return false;
  }

  return (
    first.element === second.element &&
    first.isFallback === second.isFallback &&
    first.isSpotlightTarget === second.isSpotlightTarget &&
    areTargetPositionsEqual(first.position, second.position)
  );
};

export const useWalkthroughTarget = ({
  currentTarget,
  isActive,
}: UseWalkthroughTargetOptions) => {
  const targetSnapshotKey = `${isActive}:${currentTarget ?? ""}`;
  const [targetSnapshot, setTargetSnapshot] =
    useState<WalkthroughTargetSnapshot | null>(null);
  const refreshLayout = useCallback(() => {
    const nextTarget = resolveWalkthroughTarget(isActive, currentTarget);

    setTargetSnapshot((currentSnapshot) => {
      if (
        currentSnapshot?.key === targetSnapshotKey &&
        areResolvedTargetsEqual(currentSnapshot.target, nextTarget)
      ) {
        return currentSnapshot;
      }

      return {
        key: targetSnapshotKey,
        target: nextTarget,
      };
    });
  }, [currentTarget, isActive, targetSnapshotKey]);
  const target =
    targetSnapshot?.key === targetSnapshotKey ? targetSnapshot.target : null;
  const targetElement = target?.element ?? null;

  useEffect(() => {
    if (!isActive || !currentTarget) {
      return;
    }

    const refreshTargetLayout = () => {
      refreshLayout();
    };
    const scrollListenerOptions: AddEventListenerOptions = {
      capture: true,
      passive: true,
    };
    let animationFrameId: number | null = null;
    let targetLayoutIntervalId: number | null = null;
    let targetLayoutTimeoutId: number | null = null;
    let previousTargetRectKey: string | null = null;
    let stableTargetLayoutChecks = 0;

    const clearTargetLayoutStabilityCheck = () => {
      if (targetLayoutIntervalId !== null) {
        window.clearInterval(targetLayoutIntervalId);
        targetLayoutIntervalId = null;
      }
      if (targetLayoutTimeoutId !== null) {
        window.clearTimeout(targetLayoutTimeoutId);
        targetLayoutTimeoutId = null;
      }
    };
    const checkTargetLayoutStability = () => {
      const currentTargetElement = document.querySelector(currentTarget);

      refreshTargetLayout();

      if (!currentTargetElement) {
        previousTargetRectKey = null;
        stableTargetLayoutChecks = 0;
        return;
      }

      const rect = currentTargetElement.getBoundingClientRect();
      const targetRectKey = [rect.top, rect.left, rect.width, rect.height]
        .map((value) => value.toFixed(2))
        .join(":");

      stableTargetLayoutChecks =
        isVisibleTarget(rect) && targetRectKey === previousTargetRectKey
          ? stableTargetLayoutChecks + 1
          : 0;
      previousTargetRectKey = targetRectKey;

      if (stableTargetLayoutChecks >= TARGET_LAYOUT_STABLE_CHECKS) {
        clearTargetLayoutStabilityCheck();
      }
    };
    const scheduleTargetLayoutStabilityCheck = () => {
      clearTargetLayoutStabilityCheck();
      previousTargetRectKey = null;
      stableTargetLayoutChecks = 0;
      checkTargetLayoutStability();
      targetLayoutIntervalId = window.setInterval(
        checkTargetLayoutStability,
        TARGET_LAYOUT_CHECK_INTERVAL_MS,
      );
      targetLayoutTimeoutId = window.setTimeout(
        clearTargetLayoutStabilityCheck,
        TARGET_LAYOUT_CHECK_TIMEOUT_MS,
      );
    };
    const scheduleTargetLayoutRefresh = () => {
      if (animationFrameId !== null) {
        return;
      }

      animationFrameId = window.requestAnimationFrame(() => {
        animationFrameId = null;
        refreshTargetLayout();
      });
    };
    const handleViewportResize = () => {
      refreshTargetLayout();
      scheduleTargetLayoutStabilityCheck();
    };

    scheduleTargetLayoutStabilityCheck();

    window.addEventListener("resize", handleViewportResize);
    window.addEventListener(
      "scroll",
      scheduleTargetLayoutRefresh,
      scrollListenerOptions,
    );

    return () => {
      window.removeEventListener("resize", handleViewportResize);
      window.removeEventListener("scroll", scheduleTargetLayoutRefresh, true);
      clearTargetLayoutStabilityCheck();
      if (animationFrameId !== null) {
        window.cancelAnimationFrame(animationFrameId);
      }
    };
  }, [isActive, currentTarget, refreshLayout]);

  useEffect(() => {
    if (!isActive || !targetElement || typeof ResizeObserver === "undefined") {
      return;
    }

    const observer = new ResizeObserver(refreshLayout);
    observer.observe(targetElement);
    observer.observe(document.documentElement);

    return () => {
      observer.disconnect();
    };
  }, [isActive, targetElement, refreshLayout]);

  useEffect(() => {
    if (
      !isActive ||
      !targetElement ||
      typeof IntersectionObserver === "undefined"
    ) {
      return;
    }

    const observer = new IntersectionObserver(refreshLayout, {
      threshold: [0, 0.01, 1],
    });
    observer.observe(targetElement);

    return () => {
      observer.disconnect();
    };
  }, [isActive, targetElement, refreshLayout]);

  useEffect(() => {
    if (
      !isActive ||
      !currentTarget ||
      currentTarget === "body" ||
      targetElement ||
      typeof MutationObserver === "undefined"
    ) {
      return;
    }

    const observer = new MutationObserver(() => {
      if (document.querySelector(currentTarget)) {
        refreshLayout();
      }
    });
    observer.observe(document.body, {
      attributeFilter: ["data-walkthrough-target"],
      attributes: true,
      childList: true,
      subtree: true,
    });

    return () => {
      observer.disconnect();
    };
  }, [isActive, currentTarget, targetElement, refreshLayout]);

  return target;
};
