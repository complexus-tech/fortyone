"use client";

import { useEffect, useId, useLayoutEffect, useRef, useState } from "react";
import { Box, Button, Flex, Text, Tooltip } from "ui";
import { ArrowLeft2Icon, ArrowRight2Icon, CheckIcon, CloseIcon } from "icons";
import { cn } from "lib";
import {
  getWalkthroughPanelPosition,
  type WalkthroughPanelSize,
  type WalkthroughTargetPosition,
} from "./walkthrough-position";
import {
  useWalkthrough,
  type WalkthroughStep as WalkthroughStepType,
  type WalkthroughWelcomeChoice,
} from "./walkthrough-provider";

const DEFAULT_PANEL_SIZE: WalkthroughPanelSize = {
  width: 528,
  height: 360,
};

const DEFAULT_PANEL_SIZES = {
  standard: DEFAULT_PANEL_SIZE,
  welcome: {
    width: 896,
    height: 536,
  },
} satisfies Record<
  NonNullable<WalkthroughStepType["panelLayout"]>,
  WalkthroughPanelSize
>;

const WALKTHROUGH_SUBTLE_BUTTON_CLASS =
  "bg-surface/45 hover:bg-surface/65 dark:bg-surface/45 dark:hover:bg-surface/65";
const WALKTHROUGH_PRIMARY_BUTTON_CLASS =
  "bg-primary-solid/90 hover:bg-primary-solid focus-visible:ring-primary/45 focus-visible:ring-2 focus-visible:ring-offset-2 dark:focus-visible:ring-offset-surface-elevated";

const getViewportSize = () => ({
  height: typeof window === "undefined" ? 0 : window.innerHeight,
  width: typeof window === "undefined" ? 0 : window.innerWidth,
});

const focusableSelector = [
  "a[href]",
  "button:not([disabled])",
  "input:not([disabled])",
  "select:not([disabled])",
  "textarea:not([disabled])",
  '[tabindex]:not([tabindex="-1"])',
].join(",");

const getFocusableElements = (element: HTMLElement | null) => {
  if (!element) {
    return [];
  }

  const descendants = Array.from(
    element.querySelectorAll<HTMLElement>(focusableSelector),
  );

  return element.matches(focusableSelector)
    ? [element, ...descendants]
    : descendants;
};

const WelcomeChoiceCard = ({
  choice,
  isSelected,
  onSelect,
}: {
  choice: WalkthroughWelcomeChoice;
  isSelected: boolean;
  onSelect: () => void;
}) => (
  <button
    aria-pressed={isSelected}
    className={cn(
      "border-border bg-surface/45 focus-visible:ring-primary/40 flex min-h-44 flex-col rounded-xl border p-4 text-center transition-[border-color,background-color,box-shadow] outline-none focus-visible:ring-2",
      isSelected
        ? "border-primary bg-primary/5 shadow-sm"
        : "hover:border-primary/35 hover:bg-surface/65",
    )}
    onClick={onSelect}
    type="button"
  >
    <Box
      as="span"
      className="text-primary flex h-20 items-center justify-center"
    >
      {choice.illustration}
    </Box>
    <Text
      as="span"
      className="mt-2 text-base whitespace-nowrap"
      fontWeight="semibold"
    >
      {choice.title}
    </Text>
    <Text as="span" className="mt-1 leading-5" color="muted">
      {choice.description}
    </Text>
  </button>
);

const WalkthroughTitle = ({
  currentStepNumber,
  id,
  isWelcome,
  title,
  totalSteps,
}: {
  currentStepNumber: number;
  id: string;
  isWelcome: boolean;
  title: string;
  totalSteps: number;
}) => (
  <Text
    as="h2"
    className={cn(
      "tracking-tight",
      isWelcome ? "text-3xl leading-tight sm:text-[2.1rem]" : "text-[1.1rem]",
    )}
    fontWeight="medium"
    id={id}
  >
    {title}{" "}
    <span
      aria-label={`Step ${currentStepNumber} of ${totalSteps}`}
      className="whitespace-nowrap"
    >
      {`[${currentStepNumber} / ${totalSteps}]`}
    </span>
  </Text>
);

interface WalkthroughStepProps {
  allowsTargetInteraction?: boolean;
  isFallback?: boolean;
  targetPosition: WalkthroughTargetPosition;
  step: WalkthroughStepType;
}

export const WalkthroughStep = ({
  allowsTargetInteraction = false,
  isFallback = false,
  targetPosition,
  step,
}: WalkthroughStepProps) => {
  const {
    closeWalkthrough,
    goToStep,
    isWalkthroughActionComplete,
    nextStep,
    prevStep,
    skipWalkthrough,
    state,
    steps,
  } = useWalkthrough();
  const panelLayout = step.panelLayout ?? "standard";
  const isWelcome = panelLayout === "welcome";
  const requiredAction = step.requiredAction;
  const isRequiredActionComplete = requiredAction
    ? isWalkthroughActionComplete(requiredAction.id)
    : true;
  const isRequiredActionPending =
    Boolean(requiredAction) && !isRequiredActionComplete;
  const isRequiredActionReady = isRequiredActionPending && !isFallback;
  const panelRef = useRef<HTMLDivElement>(null);
  const nextStepButtonRef = useRef<HTMLButtonElement>(null);
  const previouslyFocusedElementRef = useRef<HTMLElement | null>(null);
  const [panelSize, setPanelSize] = useState<WalkthroughPanelSize | null>(null);
  const firstWelcomeChoiceId = step.welcomeChoices?.[0]?.id ?? null;
  const [selectedWelcomeChoiceId, setSelectedWelcomeChoiceId] =
    useState(firstWelcomeChoiceId);
  const activeWelcomeChoiceId = step.welcomeChoices?.some(
    (choice) => choice.id === selectedWelcomeChoiceId,
  )
    ? selectedWelcomeChoiceId
    : firstWelcomeChoiceId;
  const titleId = useId();
  const descriptionId = useId();

  useEffect(() => {
    const activeElement = document.activeElement;
    previouslyFocusedElementRef.current =
      activeElement instanceof HTMLElement ? activeElement : null;

    return () => {
      previouslyFocusedElementRef.current?.focus();
    };
  }, []);

  useEffect(() => {
    if (!allowsTargetInteraction) {
      nextStepButtonRef.current?.focus();
    }
  }, [allowsTargetInteraction, step.id]);

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        const hasOpenModalDialog =
          allowsTargetInteraction &&
          Array.from(
            document.querySelectorAll<HTMLElement>(
              '[role="dialog"][aria-modal="true"]',
            ),
          ).some((dialog) => dialog !== panelRef.current);

        if (hasOpenModalDialog) {
          return;
        }

        event.preventDefault();
        closeWalkthrough();
        return;
      }

      if (event.key !== "Tab") {
        return;
      }

      const hasOpenModalDialog =
        allowsTargetInteraction &&
        Array.from(
          document.querySelectorAll<HTMLElement>(
            '[role="dialog"][aria-modal="true"]',
          ),
        ).some((dialog) => dialog !== panelRef.current);

      if (hasOpenModalDialog) {
        return;
      }

      const panel = panelRef.current;
      const targetElement = allowsTargetInteraction
        ? document.querySelector<HTMLElement>(step.target)
        : null;
      const focusableElements = Array.from(
        new Set([
          ...getFocusableElements(targetElement),
          ...getFocusableElements(panel),
        ]),
      );

      if (!focusableElements.length) {
        event.preventDefault();
        panel?.focus();
        return;
      }

      const activeElement = document.activeElement;
      const activeElementIndex = focusableElements.indexOf(
        activeElement as HTMLElement,
      );

      if (activeElementIndex === -1) {
        event.preventDefault();
        (event.shiftKey
          ? focusableElements.at(-1)
          : focusableElements[0]
        )?.focus();
        return;
      }

      event.preventDefault();
      const nextIndex = event.shiftKey
        ? (activeElementIndex - 1 + focusableElements.length) %
          focusableElements.length
        : (activeElementIndex + 1) % focusableElements.length;
      focusableElements[nextIndex]?.focus();
    };

    window.addEventListener("keydown", handleKeyDown);

    return () => {
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [allowsTargetInteraction, closeWalkthrough, step.target]);

  useLayoutEffect(() => {
    const panel = panelRef.current;

    if (!panel) {
      return;
    }

    const updatePanelSize = () => {
      const { height, width } = panel.getBoundingClientRect();

      if (!height || !width) {
        return;
      }

      setPanelSize((currentSize) => {
        if (
          currentSize &&
          currentSize.height === height &&
          currentSize.width === width
        ) {
          return currentSize;
        }

        return { height, width };
      });
    };

    updatePanelSize();

    if (typeof ResizeObserver === "undefined") {
      return;
    }

    const observer = new ResizeObserver(updatePanelSize);
    observer.observe(panel);

    return () => {
      observer.disconnect();
    };
  }, [step.id]);

  const startGuidedSetup = () => {
    const selectedChoice = step.welcomeChoices?.find(
      (choice) => choice.id === activeWelcomeChoiceId,
    );
    selectedChoice?.startAction?.();
    const targetStepIndex = selectedChoice
      ? steps.findIndex(
          (candidate) => candidate.id === selectedChoice.targetStepId,
        )
      : -1;

    if (targetStepIndex >= 0) {
      goToStep(targetStepIndex);
      return;
    }

    nextStep();
  };
  const triggerRequiredAction = () => {
    const targetElement = document.querySelector<HTMLElement>(step.target);

    if (!targetElement) {
      return;
    }

    const focusTarget = targetElement.matches(focusableSelector)
      ? targetElement
      : targetElement.querySelector<HTMLElement>(focusableSelector);
    focusTarget?.focus();
    targetElement.click();
  };
  const handlePrimaryAction = () => {
    if (isRequiredActionPending) {
      triggerRequiredAction();
      return;
    }

    if (isWelcome) {
      startGuidedSetup();
      return;
    }

    nextStep();
  };
  const panelPosition = getWalkthroughPanelPosition({
    panelSize: panelSize ?? DEFAULT_PANEL_SIZES[panelLayout],
    position: isFallback ? "center" : step.position,
    targetPosition,
    viewport: getViewportSize(),
  });
  const isFirstStep = state.progress.current <= 1;
  const isLastStep = state.progress.current === state.progress.total;
  const currentStepNumber = state.progress.current;
  const primaryActionLabel = (() => {
    if (isRequiredActionPending && isFallback) {
      return "Getting things ready…";
    }

    if (isRequiredActionPending) {
      return requiredAction?.actionLabel ?? "Continue";
    }

    if (isLastStep) {
      return "Finish";
    }

    if (isWelcome) {
      return "Start guided setup";
    }

    return step.nextActionLabel ?? "Next";
  })();

  return (
    <div
      aria-describedby={descriptionId}
      aria-labelledby={titleId}
      aria-modal={allowsTargetInteraction ? undefined : "true"}
      className="pointer-events-auto absolute"
      ref={panelRef}
      role="dialog"
      style={{
        top: panelPosition.top,
        left: panelPosition.left,
        zIndex: 60,
      }}
      tabIndex={-1}
    >
      <Box
        className={cn(
          "bg-surface-elevated shadow-shadow flex max-h-[calc(100dvh-2rem)] flex-col overflow-hidden rounded-2xl shadow-lg",
          isWelcome
            ? "w-[calc(100vw-2rem)] max-w-[56rem]"
            : "w-[33rem] max-w-[calc(100vw-2rem)]",
          step.highlight
            ? "from-primary to-warning border border-transparent bg-linear-to-br p-px"
            : "border-border/60 border-[0.5px]",
        )}
      >
        <Box
          className={cn(
            "bg-surface-elevated flex min-h-0 flex-1 flex-col rounded-2xl",
            step.highlight ? "text-foreground" : "",
          )}
        >
          {isWelcome ? (
            <Box className="absolute top-5 right-5 z-10 sm:top-6 sm:right-6">
              <Tooltip className="w-56" title="Close walkthrough">
                <Button
                  asIcon
                  color="tertiary"
                  onClick={closeWalkthrough}
                  size="sm"
                >
                  <CloseIcon className="h-5 w-auto" strokeWidth={3} />
                  <span className="sr-only">Close walkthrough</span>
                </Button>
              </Tooltip>
            </Box>
          ) : (
            <Flex
              align="start"
              className="shrink-0 px-5 pt-5"
              justify="between"
            >
              <Box className="min-w-0 pr-4">
                <WalkthroughTitle
                  currentStepNumber={currentStepNumber}
                  id={titleId}
                  isWelcome={false}
                  title={step.title}
                  totalSteps={state.progress.total}
                />
              </Box>
              <Tooltip className="w-56" title="Close walkthrough">
                <Button
                  asIcon
                  color="tertiary"
                  onClick={closeWalkthrough}
                  size="sm"
                >
                  <CloseIcon className="h-5 w-auto" strokeWidth={3} />
                  <span className="sr-only">Close walkthrough</span>
                </Button>
              </Tooltip>
            </Flex>
          )}

          <Box
            aria-live="polite"
            className="min-h-0 flex-1 overflow-y-auto"
            id={descriptionId}
          >
            {isWelcome ? (
              <Box className="px-7 pt-7 pb-7 sm:px-9 sm:pt-8 sm:pb-9">
                <Box className="max-w-2xl pr-10">
                  <WalkthroughTitle
                    currentStepNumber={currentStepNumber}
                    id={titleId}
                    isWelcome
                    title={step.title}
                    totalSteps={state.progress.total}
                  />
                  <Box className="mt-4 text-lg">{step.content}</Box>
                </Box>

                {step.welcomeChoices?.length ? (
                  <Box className="mt-8">
                    <Text className="text-base" fontWeight="semibold">
                      Choose a starting point
                    </Text>
                    <Box
                      aria-label="Choose a walkthrough starting point"
                      className="mt-4 grid grid-cols-1 gap-3 lg:grid-cols-3"
                      role="group"
                    >
                      {step.welcomeChoices.map((choice) => (
                        <WelcomeChoiceCard
                          choice={choice}
                          isSelected={choice.id === activeWelcomeChoiceId}
                          key={choice.id}
                          onSelect={() => {
                            setSelectedWelcomeChoiceId(choice.id);
                          }}
                        />
                      ))}
                    </Box>
                  </Box>
                ) : null}
              </Box>
            ) : (
              <Box className="p-5 pt-4 text-[1.06rem]">{step.content}</Box>
            )}
          </Box>

          <Flex
            align="center"
            className={cn(
              "shrink-0 px-6 py-4",
              isWelcome
                ? "px-7 pb-7 sm:px-9 sm:pb-8"
                : "border-border border-t",
            )}
            justify="between"
          >
            <div>
              {step.showSkip !== false && (
                <Button
                  className={WALKTHROUGH_SUBTLE_BUTTON_CLASS}
                  color="tertiary"
                  onClick={skipWalkthrough}
                >
                  {isWelcome ? "Explore on my own" : "Skip tour"}
                </Button>
              )}
            </div>

            <Flex align="center" gap={3}>
              {!isFirstStep && (
                <Button
                  className={cn("pl-2", WALKTHROUGH_SUBTLE_BUTTON_CLASS)}
                  color="tertiary"
                  leftIcon={<ArrowLeft2Icon />}
                  onClick={prevStep}
                >
                  Back
                </Button>
              )}
              <Button
                className={cn(
                  isRequiredActionPending
                    ? "min-w-[13.5rem] justify-center"
                    : "pl-6",
                  WALKTHROUGH_PRIMARY_BUTTON_CLASS,
                )}
                disabled={Boolean(
                  isRequiredActionPending && !isRequiredActionReady,
                )}
                onClick={handlePrimaryAction}
                ref={nextStepButtonRef}
                rightIcon={
                  isLastStep ? (
                    <CheckIcon className="text-current" />
                  ) : (
                    <ArrowRight2Icon className="text-current" />
                  )
                }
              >
                {primaryActionLabel}
              </Button>
            </Flex>
          </Flex>
        </Box>
      </Box>
    </div>
  );
};
