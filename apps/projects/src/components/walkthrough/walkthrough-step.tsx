"use client";

import { Button, Flex, Text, Box, Tooltip } from "ui";
import { ArrowLeft2Icon, ArrowRight2Icon, CheckIcon, CloseIcon } from "icons";
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
      <Box className="w-108 max-w-[95vw] rounded-2xl border border-border bg-surface-elevated shadow-lg shadow-shadow">
        <Flex
          align="center"
          className="border-b border-border px-4 py-3.5"
          justify="between"
        >
          <Text className="text-[1.1rem]" fontWeight="medium">
            {step.title} [{state.currentStep + 1}/{state.totalSteps}]
          </Text>
          <Tooltip
            className="w-56"
            title="This will temporarily dismiss the walkthrough"
          >
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
        <Box className="p-4">
          <Box className="text-[1.06rem]">{step.content}</Box>
        </Box>

        {/* Footer */}
        <Flex
          align="center"
          className="border-t border-border px-6 py-4"
          justify="between"
        >
          <div>
            {step.showSkip !== false && (
              <Button color="tertiary" onClick={skipWalkthrough}>
                Skip tour
              </Button>
            )}
          </div>

          <Flex align="center" gap={3}>
            {!isFirstStep && (
              <Button
                className="pl-2"
                color="tertiary"
                leftIcon={<ArrowLeft2Icon />}
                onClick={prevStep}
              >
                Back
              </Button>
            )}
            <Button
              className="pl-6"
              onClick={nextStep}
              rightIcon={
                isLastStep ? (
                  <CheckIcon className="text-white" />
                ) : (
                  <ArrowRight2Icon className="text-white" />
                )
              }
            >
              {isLastStep ? "Finish" : "Next"}
            </Button>
          </Flex>
        </Flex>
      </Box>
    </div>
  );
};
