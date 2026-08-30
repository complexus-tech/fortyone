"use client";

import { useEffect, useReducer } from "react";
import { createPortal } from "react-dom";
import { useWalkthrough } from "./walkthrough-provider";
import { WalkthroughStep } from "./walkthrough-step";

interface ElementPosition {
  top: number;
  left: number;
  width: number;
  height: number;
}

const getTargetPosition = (
  isActive: boolean,
  currentTarget: string | undefined,
): ElementPosition | null => {
  if (!isActive || !currentTarget || typeof window === "undefined") {
    return null;
  }

  if (currentTarget === "body") {
    return {
      top: window.scrollY + window.innerHeight / 2 - 100,
      left: window.innerWidth / 2 - 160,
      width: 320,
      height: 200,
    };
  }

  const targetElement = document.querySelector(currentTarget);
  if (!targetElement) {
    return null;
  }

  const rect = targetElement.getBoundingClientRect();
  return {
    top: rect.top + window.scrollY,
    left: rect.left + window.scrollX,
    width: rect.width,
    height: rect.height,
  };
};

export const WalkthroughOverlay = () => {
  const { state, currentStepData } = useWalkthrough();
  const [, refreshLayout] = useReducer(
    (currentValue: number) => currentValue + 1,
    0,
  );
  const currentTarget = currentStepData?.target;

  useEffect(() => {
    if (!state.isActive || !currentTarget) {
      return;
    }

    const handleResize = () => {
      refreshLayout();
    };
    const handleScroll = () => {
      refreshLayout();
    };
    const scrollListenerOptions: AddEventListenerOptions = { passive: true };

    window.addEventListener("resize", handleResize);
    window.addEventListener("scroll", handleScroll, scrollListenerOptions);

    const timeout = window.setTimeout(() => {
      refreshLayout();
    }, 100);

    return () => {
      window.removeEventListener("resize", handleResize);
      window.removeEventListener("scroll", handleScroll, scrollListenerOptions);
      clearTimeout(timeout);
    };
  }, [state.isActive, currentTarget]);

  const targetPosition = getTargetPosition(state.isActive, currentTarget);

  if (
    typeof document === "undefined" ||
    !state.isActive ||
    !currentStepData ||
    !targetPosition
  ) {
    return null;
  }

  return createPortal(
    <div className="pointer-events-none fixed inset-0 z-50">
      {/* Dark overlay - always present */}
      <div className="absolute inset-0 bg-black/40">
        {/* Spotlight cutout - only for non-body targets */}
        {currentStepData.target !== "body" && (
          <div
            className="border-primary/50 absolute rounded-lg border-2 bg-transparent shadow-xl"
            style={{
              top: targetPosition.top - 8,
              left: targetPosition.left - 8,
              width: targetPosition.width + 16,
              height: targetPosition.height + 16,
              boxShadow: `
                0 0 0 4px rgba(0, 0, 0, 0.1),
                0 0 0 9999px rgba(0, 0, 0, 0.6)
              `,
            }}
          />
        )}
      </div>

      {/* Walkthrough step content */}
      <WalkthroughStep step={currentStepData} targetPosition={targetPosition} />
    </div>,
    document.body,
  );
};
