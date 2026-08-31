"use client";

import { useEffect } from "react";
import { createPortal } from "react-dom";
import { cn } from "lib";
import { useWalkthroughTarget } from "./use-walkthrough-target";
import { useWalkthrough } from "./walkthrough-provider";
import { WalkthroughStep } from "./walkthrough-step";

const SPOTLIGHT_PADDING = 8;

export const WalkthroughOverlay = () => {
  const { currentStepData, isWalkthroughActionComplete, state } =
    useWalkthrough();
  const currentTarget = currentStepData?.target;
  const target = useWalkthroughTarget({
    currentTarget,
    isActive: state.isActive,
  });
  const targetElement = target?.element ?? null;
  const isActionPending = Boolean(
    currentStepData?.requiredAction &&
      !isWalkthroughActionComplete(currentStepData.requiredAction.id),
  );
  const allowsTargetInteraction =
    isActionPending && Boolean(target?.isSpotlightTarget);

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
