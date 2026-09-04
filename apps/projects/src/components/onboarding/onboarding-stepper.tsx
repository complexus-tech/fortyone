"use client";

import { CheckIcon } from "icons";
import { cn } from "lib";

export const WORKSPACE_ONBOARDING_STEPS = [
  "Workspace",
  "Your work",
  "Get started",
] as const;
export const JOIN_ONBOARDING_STEPS = [
  "Join workspace",
  "Your profile",
  "Get started",
] as const;

export type OnboardingStep = number;

type OnboardingStepperProps = {
  currentStep: OnboardingStep;
  steps?: readonly string[];
  furthestStep?: OnboardingStep;
  onStepChange?: (step: OnboardingStep) => void;
};

export const OnboardingStepper = ({
  currentStep,
  steps = WORKSPACE_ONBOARDING_STEPS,
  furthestStep = currentStep,
  onStepChange,
}: OnboardingStepperProps) => (
  <nav aria-label="Setup progress" className="mb-8 w-full">
    <p aria-live="polite" className="sr-only">
      Step {currentStep + 1} of {steps.length}: {steps[currentStep]}
    </p>
    <div className="relative">
      <div
        aria-hidden="true"
        className="bg-border-strong absolute top-3 right-3 left-3 h-px"
      >
        <span
          className="bg-foreground block h-full"
          style={{
            width: `${(furthestStep / Math.max(steps.length - 1, 1)) * 100}%`,
          }}
        />
      </div>
      <ol
        className="relative grid w-full"
        style={{
          gridTemplateColumns: `repeat(${steps.length}, minmax(0, 1fr))`,
        }}
      >
        {steps.map((label, step) => {
          const isCurrent = step === currentStep;
          const isCompleted = step < furthestStep;
          const canNavigate = Boolean(onStepChange) && step <= furthestStep;
          const position = (step / Math.max(steps.length - 1, 1)) * 100;
          const markerPosition = {
            left: `${position}%`,
            transform: `translateX(-${position}%)`,
          };
          const alignment = cn(
            "relative flex w-6 flex-col",
            step === 0 && "items-start",
            step > 0 && step < steps.length - 1 && "items-center",
            step === steps.length - 1 && "items-end",
          );
          const content = (
            <>
              <span
                aria-hidden="true"
                className={cn(
                  "relative z-1 flex aspect-square size-6 shrink-0 items-center justify-center rounded-full border text-xs leading-none font-medium transition-colors duration-150",
                  "border-border-strong bg-background text-text-muted",
                  (isCompleted || isCurrent) &&
                    "border-foreground bg-foreground text-background",
                )}
              >
                {isCompleted && !isCurrent ? (
                  <CheckIcon
                    className="size-3.5 text-current"
                    strokeWidth={2}
                  />
                ) : (
                  step + 1
                )}
              </span>
              <span
                className={cn(
                  "mt-2 text-sm whitespace-nowrap",
                  isCurrent ? "text-foreground font-medium" : "text-text-muted",
                )}
              >
                {label}
              </span>
            </>
          );

          return (
            <li
              aria-current={isCurrent ? "step" : undefined}
              className="relative min-w-0"
              key={label}
            >
              {canNavigate ? (
                <button
                  aria-label={`Step ${step + 1}: ${label}`}
                  className={cn(
                    alignment,
                    "focus-visible:ring-foreground focus-visible:ring-offset-background rounded-md outline-none focus-visible:ring-2 focus-visible:ring-offset-4",
                  )}
                  onClick={() => onStepChange?.(step)}
                  style={markerPosition}
                  type="button"
                >
                  {content}
                </button>
              ) : (
                <span className={alignment} style={markerPosition}>
                  {content}
                </span>
              )}
            </li>
          );
        })}
      </ol>
    </div>
  </nav>
);
