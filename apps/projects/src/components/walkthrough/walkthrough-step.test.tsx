/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import {
  getWalkthroughTargetSelector,
  walkthroughTargets,
} from "@/shared/walkthrough/targets";
import { useWalkthrough } from "./walkthrough-provider";
import { WalkthroughStep } from "./walkthrough-step";

jest.mock("./walkthrough-provider", () => ({
  useWalkthrough: jest.fn(),
}));

const useWalkthroughMock = jest.mocked(useWalkthrough);
const closeWalkthroughMock = jest.fn();
const goToStepMock = jest.fn();
const isWalkthroughActionCompleteMock = jest.fn(() => false);
const nextStepMock = jest.fn();
const startActionMock = jest.fn();

describe("WalkthroughStep", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    useWalkthroughMock.mockReturnValue({
      closeWalkthrough: closeWalkthroughMock,
      goToStep: goToStepMock,
      isWalkthroughActionComplete: isWalkthroughActionCompleteMock,
      nextStep: nextStepMock,
      prevStep: jest.fn(),
      skipWalkthrough: jest.fn(),
      steps: [],
      state: {
        currentStep: 3,
        hasSeenWalkthrough: false,
        isActive: true,
        progress: {
          current: 4,
          total: 10,
        },
        totalSteps: 10,
        walkthroughVersion: "1.0.0",
      },
    } as unknown as ReturnType<typeof useWalkthrough>);
  });

  const renderStep = () =>
    render(
      <WalkthroughStep
        step={{
          content: "You have unread notifications.",
          id: "my-notifications",
          target: "body",
          title: "Stay Updated",
        }}
        targetPosition={{ height: 0, left: 0, top: 0, width: 0 }}
      />,
    );

  it("provides a labelled modal with an initial focus target", () => {
    renderStep();

    const dialog = screen.getByRole("dialog", { name: /stay updated/i });
    const nextButton = screen.getByRole("button", { name: "Next" });

    expect(dialog).toHaveAttribute("aria-modal", "true");
    expect(dialog).toHaveAttribute("aria-describedby");
    expect(nextButton).toHaveFocus();
    expect(nextButton).toHaveClass(
      "dark:focus-visible:ring-offset-surface-elevated",
    );
  });

  it("shows progress for the selected journey instead of the global step index", () => {
    useWalkthroughMock.mockReturnValue({
      closeWalkthrough: closeWalkthroughMock,
      goToStep: goToStepMock,
      isWalkthroughActionComplete: isWalkthroughActionCompleteMock,
      nextStep: nextStepMock,
      prevStep: jest.fn(),
      skipWalkthrough: jest.fn(),
      steps: [],
      state: {
        currentStep: 2,
        hasSeenWalkthrough: false,
        isActive: true,
        progress: {
          current: 2,
          total: 3,
        },
        totalSteps: 4,
        walkthroughVersion: "1.0.0",
      },
    } as unknown as ReturnType<typeof useWalkthrough>);

    renderStep();

    expect(screen.getByLabelText("Step 2 of 3")).toBeInTheDocument();
    expect(screen.getByText("[2 / 3]")).toBeInTheDocument();
    expect(screen.queryByText("[3 / 4]")).not.toBeInTheDocument();
  });

  it("closes the walkthrough when Escape is pressed", () => {
    renderStep();

    fireEvent.keyDown(window, { key: "Escape" });

    expect(closeWalkthroughMock).toHaveBeenCalledTimes(1);
  });

  it("wraps keyboard focus within the walkthrough", () => {
    renderStep();

    const closeButton = screen.getByRole("button", {
      name: "Close walkthrough",
    });
    const nextButton = screen.getByRole("button", { name: "Next" });

    fireEvent.keyDown(nextButton, { key: "Tab" });

    expect(closeButton).toHaveFocus();
  });

  it("uses the selected welcome choice to start the relevant compact tip", () => {
    const welcomeChoices = [
      {
        description: "Turn an idea into a visible next step.",
        id: "create-story",
        illustration: <span aria-hidden="true" />,
        targetStepId: "create-story",
        title: "Create a task",
      },
      {
        description: "Connect the work to a bigger outcome.",
        id: "roadmap",
        illustration: <span aria-hidden="true" />,
        startAction: startActionMock,
        targetStepId: "roadmap",
        title: "Create an objective",
      },
      {
        description: "Give the work a real place in your week.",
        id: "calendar",
        illustration: <span aria-hidden="true" />,
        targetStepId: "calendar",
        title: "Plan my time",
      },
    ];
    const welcomeStep = {
      content:
        "Choose a starting point and we’ll guide you to the next action.",
      id: "welcome",
      panelLayout: "welcome" as const,
      position: "center" as const,
      target: "body" as const,
      title: "What would you like to do first?",
      welcomeChoices,
    };

    useWalkthroughMock.mockReturnValue({
      closeWalkthrough: closeWalkthroughMock,
      goToStep: goToStepMock,
      nextStep: jest.fn(),
      prevStep: jest.fn(),
      skipWalkthrough: jest.fn(),
      state: {
        currentStep: 0,
        hasSeenWalkthrough: false,
        isActive: true,
        progress: {
          current: 1,
          total: 4,
        },
        totalSteps: 4,
        walkthroughVersion: "1.0.0",
      },
      steps: [
        welcomeStep,
        { ...welcomeStep, id: "create-story" },
        { ...welcomeStep, id: "roadmap" },
      ],
    } as unknown as ReturnType<typeof useWalkthrough>);

    render(
      <WalkthroughStep
        step={welcomeStep}
        targetPosition={{ height: 0, left: 0, top: 0, width: 0 }}
      />,
    );

    expect(screen.getByLabelText("Step 1 of 4")).toBeInTheDocument();
    expect(screen.getByText("[1 / 4]")).toBeInTheDocument();
    expect(screen.queryByText("< 1 / 4 >")).not.toBeInTheDocument();
    expect(screen.getByText("Choose a starting point")).toBeInTheDocument();

    fireEvent.click(
      screen.getByRole("button", { name: /create an objective/i }),
    );
    fireEvent.click(
      screen.getByRole("button", { name: /start guided setup/i }),
    );

    expect(goToStepMock).toHaveBeenCalledWith(2);
    expect(startActionMock).toHaveBeenCalledTimes(1);
  });

  it("uses the primary button to launch a gated action instead of advancing", () => {
    const target = document.createElement("button");
    const targetClickMock = jest.fn();
    target.dataset.walkthroughTarget = walkthroughTargets.create;
    target.addEventListener("click", targetClickMock);
    document.body.append(target);

    render(
      <WalkthroughStep
        allowsTargetInteraction
        step={{
          content: "Create and save a real task.",
          id: "create-story",
          requiredAction: {
            actionLabel: "Create my first task",
            id: "story-created",
          },
          target: getWalkthroughTargetSelector(walkthroughTargets.create),
          title: "Create your first task",
        }}
        targetPosition={{ height: 40, left: 20, top: 20, width: 40 }}
      />,
    );

    const dialog = screen.getByRole("dialog", {
      name: /create your first task/i,
    });
    fireEvent.click(
      screen.getByRole("button", { name: "Create my first task" }),
    );

    expect(dialog).not.toHaveAttribute("aria-modal");
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    expect(targetClickMock).toHaveBeenCalledTimes(1);
    expect(nextStepMock).not.toHaveBeenCalled();

    target.remove();
  });

  it("keeps keyboard focus within the live target and walkthrough controls", () => {
    const target = document.createElement("button");
    const backgroundButton = document.createElement("button");
    target.dataset.walkthroughTarget = walkthroughTargets.create;
    target.textContent = "Create task";
    backgroundButton.textContent = "Background action";
    document.body.append(target, backgroundButton);

    render(
      <WalkthroughStep
        allowsTargetInteraction
        step={{
          content: "Create and save a real task.",
          id: "create-story",
          requiredAction: {
            actionLabel: "Create my first task",
            id: "story-created",
          },
          target: getWalkthroughTargetSelector(walkthroughTargets.create),
          title: "Create your first task",
        }}
        targetPosition={{ height: 40, left: 20, top: 20, width: 40 }}
      />,
    );

    const closeButton = screen.getByRole("button", {
      name: "Close walkthrough",
    });
    target.focus();

    fireEvent.keyDown(target, { key: "Tab" });
    expect(closeButton).toHaveFocus();

    fireEvent.keyDown(closeButton, { key: "Tab", shiftKey: true });
    expect(target).toHaveFocus();
    expect(backgroundButton).not.toHaveFocus();

    target.remove();
    backgroundButton.remove();
  });

  it("waits for the intended target before enabling a gated action", () => {
    render(
      <WalkthroughStep
        isFallback
        step={{
          content: "Create and save a real task.",
          id: "create-story",
          requiredAction: {
            actionLabel: "Create my first task",
            id: "story-created",
          },
          target: getWalkthroughTargetSelector(walkthroughTargets.create),
          title: "Create your first task",
        }}
        targetPosition={{ height: 0, left: 0, top: 0, width: 0 }}
      />,
    );

    expect(
      screen.getByRole("button", { name: "Getting things ready…" }),
    ).toBeDisabled();
  });
});
