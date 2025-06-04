"use client";

import { useEffect, useState, useCallback } from "react";
import { createPortal } from "react-dom";
import { useWalkthrough } from "./walkthrough-provider";
import { WalkthroughStep } from "./walkthrough-step";

interface ElementPosition {
  top: number;
  left: number;
  width: number;
  height: number;
}

export const WalkthroughOverlay = () => {
  const { state, currentStepData } = useWalkthrough();
  const [targetPosition, setTargetPosition] = useState<ElementPosition | null>(
    null,
  );
  const [mounted, setMounted] = useState(false);

  const updateTargetPosition = useCallback(() => {
    if (!currentStepData?.target) return;

    // Special handling for body target - position in center of viewport
    if (currentStepData.target === "body") {
      setTargetPosition({
        top: window.scrollY + window.innerHeight / 2 - 100,
        left: window.innerWidth / 2 - 160,
        width: 320,
        height: 200,
      });
      return;
    }

    const targetElement = document.querySelector(currentStepData.target);
    if (targetElement) {
      const rect = targetElement.getBoundingClientRect();
      setTargetPosition({
        top: rect.top + window.scrollY,
        left: rect.left + window.scrollX,
        width: rect.width,
        height: rect.height,
      });
    }
  }, [currentStepData?.target]);

  // Update position when step changes or window resizes
  useEffect(() => {
    if (!state.isActive || !currentStepData) {
      setTargetPosition(null);
      return;
    }

    updateTargetPosition();

    const handleResize = () => {
      updateTargetPosition();
    };
    const handleScroll = () => {
      updateTargetPosition();
    };

    window.addEventListener("resize", handleResize);
    window.addEventListener("scroll", handleScroll);

    // Also update position after a short delay to handle dynamic content
    const timeout = setTimeout(updateTargetPosition, 100);

    return () => {
      window.removeEventListener("resize", handleResize);
      window.removeEventListener("scroll", handleScroll);
      clearTimeout(timeout);
    };
  }, [state.isActive, currentStepData, updateTargetPosition]);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted || !state.isActive || !currentStepData || !targetPosition) {
    return null;
  }

  return createPortal(
    <div className="pointer-events-none fixed inset-0 z-50">
      {/* Dark overlay with spotlight effect - only for non-body targets */}
      {currentStepData.target !== "body" && (
        <div className="absolute inset-0 bg-black/60 dark:bg-black/80">
          {/* Spotlight cutout */}
          <div
            className="absolute rounded-lg border-2 border-primary/50 bg-transparent shadow-xl"
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
        </div>
      )}

      {/* Walkthrough step content */}
      <WalkthroughStep step={currentStepData} targetPosition={targetPosition} />
    </div>,
    document.body,
  );
};
