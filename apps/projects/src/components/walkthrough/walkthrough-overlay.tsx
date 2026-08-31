"use client";

import { useEffect, useReducer } from "react";
import { createPortal } from "react-dom";
import { cn } from "lib";
import type { WalkthroughTargetPosition } from "./walkthrough-position";
import { useWalkthrough } from "./walkthrough-provider";
import { WalkthroughStep } from "./walkthrough-step";

interface ResolvedWalkthroughTarget {
  element: Element | null;
  position: WalkthroughTargetPosition;
  isFallback: boolean;
  isSpotlightTarget: boolean;
}

const fallbackTargetPosition: WalkthroughTargetPosition = {
  top: 0,
  left: 0,
  width: 0,
  height: 0,
};

const SPOTLIGHT_PADDING = 8;

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

export const WalkthroughOverlay = () => {
  const { currentStepData, isWalkthroughActionComplete, state } =
    useWalkthrough();
  const [, refreshLayout] = useReducer(
    (currentValue: number) => currentValue + 1,
    0,
  );
  const currentTarget = currentStepData?.target;
  const target = resolveWalkthroughTarget(state.isActive, currentTarget);
  const targetElement = target?.element ?? null;
  const isActionPending = Boolean(
    currentStepData?.requiredAction &&
      !isWalkthroughActionComplete(currentStepData.requiredAction.id),
  );
  const allowsTargetInteraction =
    isActionPending && Boolean(target?.isSpotlightTarget);

  useEffect(() => {
    if (!state.isActive || !currentTarget) {
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

    const scheduleTargetLayoutRefresh = () => {
      if (animationFrameId !== null) {
        return;
      }

      animationFrameId = window.requestAnimationFrame(() => {
        animationFrameId = null;
        refreshTargetLayout();
      });
    };
    const timeoutId = window.setTimeout(refreshTargetLayout, 100);

    window.addEventListener("resize", refreshTargetLayout);
    window.addEventListener(
      "scroll",
      scheduleTargetLayoutRefresh,
      scrollListenerOptions,
    );

    return () => {
      window.removeEventListener("resize", refreshTargetLayout);
      window.removeEventListener("scroll", scheduleTargetLayoutRefresh, true);
      window.clearTimeout(timeoutId);
      if (animationFrameId !== null) {
        window.cancelAnimationFrame(animationFrameId);
      }
    };
  }, [state.isActive, currentTarget]);

  useEffect(() => {
    if (
      !state.isActive ||
      !targetElement ||
      typeof ResizeObserver === "undefined"
    ) {
      return;
    }

    const observer = new ResizeObserver(refreshLayout);
    observer.observe(targetElement);

    return () => {
      observer.disconnect();
    };
  }, [state.isActive, targetElement]);

  useEffect(() => {
    if (
      !state.isActive ||
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
  }, [state.isActive, currentTarget, targetElement]);

  useEffect(() => {
    if (!allowsTargetInteraction || !(targetElement instanceof HTMLElement)) {
      return;
    }

    const animationFrameId = window.requestAnimationFrame(() => {
      const focusTarget = targetElement.matches(
        'button:not([disabled]), input:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
      )
        ? targetElement
        : targetElement.querySelector<HTMLElement>(
            'button:not([disabled]), input:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
          );

      focusTarget?.focus();
    });

    return () => {
      window.cancelAnimationFrame(animationFrameId);
    };
  }, [allowsTargetInteraction, targetElement]);

  if (
    typeof document === "undefined" ||
    !state.isActive ||
    !currentStepData ||
    !target
  ) {
    return null;
  }

  const spotlightTop = Math.max(0, target.position.top - SPOTLIGHT_PADDING);
  const spotlightLeft = Math.max(0, target.position.left - SPOTLIGHT_PADDING);
  const spotlightBottom = Math.min(
    window.innerHeight,
    target.position.top + target.position.height + SPOTLIGHT_PADDING,
  );
  const spotlightRight = Math.min(
    window.innerWidth,
    target.position.left + target.position.width + SPOTLIGHT_PADDING,
  );
  const spotlightHeight = Math.max(0, spotlightBottom - spotlightTop);

  return createPortal(
    <div className="pointer-events-none fixed inset-0 z-50">
      {/* Dark overlay */}
      <div
        aria-hidden="true"
        className={cn(
          "absolute inset-0",
          allowsTargetInteraction
            ? "pointer-events-none"
            : "pointer-events-auto",
          target.isSpotlightTarget ? "bg-transparent" : "bg-black/40",
        )}
      >
        {/* Spotlight cutout for resolved element targets */}
        {target.isSpotlightTarget ? (
          <div
            className="border-primary/50 absolute rounded-lg border-2 bg-transparent shadow-xl"
            style={{
              top: spotlightTop,
              left: spotlightLeft,
              width: spotlightRight - spotlightLeft,
              height: spotlightHeight,
              boxShadow: `
                0 0 0 4px rgba(0, 0, 0, 0.1),
                0 0 0 9999px rgba(0, 0, 0, 0.6)
              `,
            }}
          />
        ) : null}
      </div>

      {allowsTargetInteraction ? (
        <>
          <div
            aria-hidden="true"
            className="pointer-events-auto absolute top-0 right-0 left-0"
            data-walkthrough-interaction-blocker
            style={{ height: spotlightTop }}
          />
          <div
            aria-hidden="true"
            className="pointer-events-auto absolute right-0 bottom-0 left-0"
            data-walkthrough-interaction-blocker
            style={{ top: spotlightBottom }}
          />
          <div
            aria-hidden="true"
            className="pointer-events-auto absolute left-0"
            data-walkthrough-interaction-blocker
            style={{
              top: spotlightTop,
              width: spotlightLeft,
              height: spotlightHeight,
            }}
          />
          <div
            aria-hidden="true"
            className="pointer-events-auto absolute right-0"
            data-walkthrough-interaction-blocker
            style={{
              top: spotlightTop,
              left: spotlightRight,
              height: spotlightHeight,
            }}
          />
        </>
      ) : null}

      {/* Walkthrough step content */}
      <WalkthroughStep
        allowsTargetInteraction={allowsTargetInteraction}
        isFallback={target.isFallback}
        step={currentStepData}
        targetPosition={target.position}
      />
    </div>,
    document.body,
  );
};
