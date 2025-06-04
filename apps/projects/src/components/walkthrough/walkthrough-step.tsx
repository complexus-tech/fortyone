"use client";

import { Button, Flex, Text, Box } from "ui";
import { ArrowLeftIcon, CloseIcon } from "icons";
import { cn } from "lib";
import {
  useWalkthrough,
  type WalkthroughStep as WalkthroughStepType,
} from "./walkthrough-provider";

interface ElementPosition {
  top: number;
  left: number;
  width: number;
  height: number;
}

interface WalkthroughStepProps {
  targetPosition: ElementPosition;
  step: WalkthroughStepType;
}

const getPositionStyles = (
  targetPosition: ElementPosition,
  position = "bottom",
) => {
  const padding = 16;
  const arrowSize = 8;

  switch (position) {
    case "top":
      return {
        top: targetPosition.top - padding,
        left: targetPosition.left + targetPosition.width / 2,
        transform: "translate(-50%, -100%)",
        marginBottom: arrowSize,
      };
    case "top-start":
      return {
        top: targetPosition.top - padding,
        left: targetPosition.left,
        transform: "translate(0, -100%)",
        marginBottom: arrowSize,
      };
    case "bottom":
      return {
        top: targetPosition.top + targetPosition.height + padding,
        left: targetPosition.left + targetPosition.width / 2,
        transform: "translate(-50%, 0)",
        marginTop: arrowSize,
      };
    case "bottom-start":
      return {
        top: targetPosition.top + targetPosition.height + padding,
        left: targetPosition.left,
        transform: "translate(0, 0)",
        marginTop: arrowSize,
      };
    case "center":
      return {
        top: targetPosition.top,
        left: targetPosition.left,
        transform: "translate(0, 0)",
      };
    case "left":
      return {
        top: targetPosition.top + targetPosition.height / 2,
        left: targetPosition.left - padding,
        transform: "translate(-100%, -50%)",
        marginRight: arrowSize,
      };
    case "right":
      return {
        top: targetPosition.top + targetPosition.height / 2,
        left: targetPosition.left + targetPosition.width + padding,
        transform: "translate(0, -50%)",
        marginLeft: arrowSize,
      };
    default:
      return {
        top: targetPosition.top + targetPosition.height + padding,
        left: targetPosition.left + targetPosition.width / 2,
        transform: "translate(-50%, 0)",
        marginTop: arrowSize,
      };
  }
};

export const WalkthroughStep = ({
  targetPosition,
  step,
}: WalkthroughStepProps) => {
  const { state, nextStep, prevStep, skipWalkthrough, closeWalkthrough } =
    useWalkthrough();

  const positionStyles = getPositionStyles(targetPosition, step.position);
  const isFirstStep = state.currentStep === 0;
  const isLastStep = state.currentStep === state.totalSteps - 1;

  return (
    <div
      className="pointer-events-auto absolute"
      style={{
        top: positionStyles.top,
        left: positionStyles.left,
        transform: positionStyles.transform,
        zIndex: 60,
      }}
    >
      {/* Arrow pointing to target - hide for center position */}
      {step.position !== "center" && (
        <div
          className={cn("absolute h-0 w-0 border-8", {
            "-top-4 left-1/2 -translate-x-1/2 border-transparent border-b-white dark:border-b-dark-200":
              step.position === "bottom",
            "-top-4 left-8 border-transparent border-b-white dark:border-b-dark-200":
              step.position === "bottom-start",
            "-bottom-4 left-1/2 -translate-x-1/2 border-transparent border-t-white dark:border-t-dark-200":
              step.position === "top",
            "-bottom-4 left-8 border-transparent border-t-white dark:border-t-dark-200":
              step.position === "top-start",
            "-left-4 top-1/2 -translate-y-1/2 border-transparent border-r-white dark:border-r-dark-200":
              step.position === "right",
            "-right-4 top-1/2 -translate-y-1/2 border-transparent border-l-white dark:border-l-dark-200":
              step.position === "left",
          })}
        />
      )}

      {/* Content card */}
      <Box className="w-80 max-w-[90vw] rounded-xl border border-gray-100 bg-white shadow-lg backdrop-blur dark:border-dark-50 dark:bg-dark-200 dark:shadow-dark/20">
        {/* Header */}
        <Flex
          align="center"
          className="border-b border-gray-100 px-6 py-4 dark:border-dark-50"
          justify="between"
        >
          <div>
            <Text className="text-lg" fontWeight="medium">
              {step.title}
            </Text>
            <Text className="text-sm" color="muted">
              Step {state.currentStep + 1} of {state.totalSteps}
            </Text>
          </div>
          <Button
            asIcon
            color="tertiary"
            onClick={closeWalkthrough}
            size="sm"
            variant="naked"
          >
            <CloseIcon className="h-4 w-auto" />
            <span className="sr-only">Close walkthrough</span>
          </Button>
        </Flex>

        {/* Content */}
        <Box className="px-6 py-4">
          {typeof step.content === "string" ? (
            <Text color="muted">{step.content}</Text>
          ) : (
            step.content
          )}
        </Box>

        {/* Footer */}
        <Flex
          align="center"
          className="border-t border-gray-100 px-6 py-4 dark:border-dark-50"
          justify="between"
        >
          <div>
            {step.showSkip !== false && (
              <Button
                color="tertiary"
                onClick={skipWalkthrough}
                size="sm"
                variant="naked"
              >
                Skip tour
              </Button>
            )}
          </div>

          <Flex align="center" gap={2}>
            {!isFirstStep && (
              <Button
                color="tertiary"
                leftIcon={<ArrowLeftIcon className="h-4 w-auto" />}
                onClick={prevStep}
                size="sm"
                variant="outline"
              >
                Back
              </Button>
            )}
            <Button onClick={nextStep} size="sm">
              {isLastStep ? "Finish" : "Next"}
            </Button>
          </Flex>
        </Flex>
      </Box>
    </div>
  );
};
