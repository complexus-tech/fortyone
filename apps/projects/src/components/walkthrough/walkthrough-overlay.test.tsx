/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { act, render, screen, waitFor } from "@testing-library/react";
import {
  getWalkthroughTargetSelector,
  walkthroughTargets,
} from "@/shared/walkthrough/targets";
import { useWalkthrough } from "./walkthrough-provider";
import { WalkthroughOverlay } from "./walkthrough-overlay";

jest.mock("./walkthrough-provider", () => ({
  useWalkthrough: jest.fn(),
}));

jest.mock("./walkthrough-step", () => ({
  WalkthroughStep: ({
    allowsTargetInteraction,
    isFallback,
    step,
    targetPosition,
  }: {
    allowsTargetInteraction?: boolean;
    isFallback?: boolean;
    step: { title: string };
    targetPosition: { left: number; top: number };
  }) => (
    <div
      data-allows-target-interaction={allowsTargetInteraction}
      data-fallback={isFallback}
      data-left={targetPosition.left}
      data-testid="walkthrough-step"
      data-top={targetPosition.top}
    >
      {step.title}
    </div>
  ),
}));

const useWalkthroughMock = jest.mocked(useWalkthrough);
const originalInnerWidth = window.innerWidth;
const originalIntersectionObserver = window.IntersectionObserver;
const originalResizeObserver = window.ResizeObserver;
let targetElement: HTMLButtonElement | null = null;

const createTargetRect = ({
  height,
  left,
  top,
  width,
}: {
  height: number;
  left: number;
  top: number;
  width: number;
}): DOMRect =>
  ({
    bottom: top + height,
    height,
    left,
    right: left + width,
    toJSON: () => ({}),
    top,
    width,
    x: left,
    y: top,
  }) as DOMRect;

describe("WalkthroughOverlay", () => {
  beforeEach(() => {
    useWalkthroughMock.mockReturnValue({
      currentStepData: {
        content: "Notifications content",
        id: "my-notifications",
        target: getWalkthroughTargetSelector(walkthroughTargets.notifications),
        title: "Stay Updated",
      },
      state: {
        currentStep: 3,
        hasSeenWalkthrough: false,
        isActive: true,
        totalSteps: 10,
        walkthroughVersion: "1.0.0",
      },
    } as unknown as ReturnType<typeof useWalkthrough>);
  });

  afterEach(() => {
    targetElement?.remove();
    targetElement = null;
    Object.defineProperty(window, "innerWidth", {
      configurable: true,
      value: originalInnerWidth,
      writable: true,
    });
    Object.defineProperty(window, "IntersectionObserver", {
      configurable: true,
      value: originalIntersectionObserver,
      writable: true,
    });
    Object.defineProperty(window, "ResizeObserver", {
      configurable: true,
      value: originalResizeObserver,
      writable: true,
    });
  });

  it("keeps the current step visible when its target is unavailable", () => {
    render(<WalkthroughOverlay />);

    expect(screen.getByText("Stay Updated")).toBeInTheDocument();
    expect(screen.getByTestId("walkthrough-step")).toHaveAttribute(
      "data-fallback",
      "true",
    );
  });

  it("re-resolves the target position when the viewport changes", () => {
    targetElement = document.createElement("button");
    targetElement.dataset.walkthroughTarget = walkthroughTargets.notifications;
    document.body.append(targetElement);

    const getBoundingClientRect = jest
      .spyOn(targetElement, "getBoundingClientRect")
      .mockReturnValue(
        createTargetRect({ height: 56, left: 720, top: 12, width: 56 }),
      );

    render(<WalkthroughOverlay />);

    expect(screen.getByTestId("walkthrough-step")).toHaveAttribute(
      "data-left",
      "720",
    );

    getBoundingClientRect.mockReturnValue(
      createTargetRect({ height: 56, left: 420, top: 12, width: 56 }),
    );
    Object.defineProperty(window, "innerWidth", {
      configurable: true,
      value: 600,
      writable: true,
    });

    act(() => {
      window.dispatchEvent(new Event("resize"));
    });

    expect(screen.getByTestId("walkthrough-step")).toHaveAttribute(
      "data-left",
      "420",
    );
  });

  it("recovers when its target mounts after the walkthrough starts", async () => {
    render(<WalkthroughOverlay />);

    expect(screen.getByTestId("walkthrough-step")).toHaveAttribute(
      "data-fallback",
      "true",
    );

    targetElement = document.createElement("button");
    targetElement.dataset.walkthroughTarget = walkthroughTargets.notifications;
    jest
      .spyOn(targetElement, "getBoundingClientRect")
      .mockReturnValue(
        createTargetRect({ height: 56, left: 720, top: 12, width: 56 }),
      );
    document.body.append(targetElement);

    await waitFor(() => {
      expect(screen.getByTestId("walkthrough-step")).toHaveAttribute(
        "data-fallback",
        "false",
      );
    });
  });

  it("recovers when an existing offscreen target enters the viewport", async () => {
    const intersectionCallbackRef: {
      current?: IntersectionObserverCallback;
    } = {};
    const observe = jest.fn();

    class TestIntersectionObserver implements IntersectionObserver {
      readonly root = null;
      readonly rootMargin = "0px";
      readonly scrollMargin = "0px";
      readonly thresholds = [0];
      readonly disconnect = jest.fn();
      readonly observe = observe;
      readonly takeRecords = jest.fn(() => []);
      readonly unobserve = jest.fn();

      constructor(callback: IntersectionObserverCallback) {
        intersectionCallbackRef.current = callback;
      }
    }

    Object.defineProperty(window, "IntersectionObserver", {
      configurable: true,
      value: TestIntersectionObserver,
      writable: true,
    });

    targetElement = document.createElement("button");
    targetElement.dataset.walkthroughTarget = walkthroughTargets.notifications;
    document.body.append(targetElement);

    const getBoundingClientRect = jest
      .spyOn(targetElement, "getBoundingClientRect")
      .mockReturnValue(
        createTargetRect({
          height: 56,
          left: 720,
          top: window.innerHeight + 40,
          width: 56,
        }),
      );

    render(<WalkthroughOverlay />);

    expect(screen.getByTestId("walkthrough-step")).toHaveAttribute(
      "data-fallback",
      "true",
    );
    expect(observe).toHaveBeenCalledWith(targetElement);

    const visibleRect = createTargetRect({
      height: 56,
      left: 720,
      top: 12,
      width: 56,
    });
    getBoundingClientRect.mockReturnValue(visibleRect);

    const notifyIntersection = intersectionCallbackRef.current;
    if (!notifyIntersection) {
      throw new Error(
        "Expected the target to be observed for visibility changes.",
      );
    }

    act(() => {
      notifyIntersection(
        [
          {
            boundingClientRect: visibleRect,
            intersectionRatio: 1,
            intersectionRect: visibleRect,
            isIntersecting: true,
            rootBounds: null,
            target: targetElement!,
            time: 0,
          },
        ],
        {} as IntersectionObserver,
      );
    });

    await waitFor(() => {
      expect(screen.getByTestId("walkthrough-step")).toHaveAttribute(
        "data-fallback",
        "false",
      );
    });
  });

  it("updates the spotlight when a mounted target changes layout", async () => {
    const resizeCallbackRef: { current?: ResizeObserverCallback } = {};
    const observe = jest.fn();

    class TestResizeObserver implements ResizeObserver {
      readonly disconnect = jest.fn();
      readonly observe = observe;
      readonly unobserve = jest.fn();

      constructor(callback: ResizeObserverCallback) {
        resizeCallbackRef.current = callback;
      }
    }

    Object.defineProperty(window, "ResizeObserver", {
      configurable: true,
      value: TestResizeObserver,
      writable: true,
    });

    targetElement = document.createElement("button");
    targetElement.dataset.walkthroughTarget = walkthroughTargets.notifications;
    document.body.append(targetElement);

    const getBoundingClientRect = jest
      .spyOn(targetElement, "getBoundingClientRect")
      .mockReturnValue(
        createTargetRect({ height: 56, left: 720, top: 12, width: 56 }),
      );

    render(<WalkthroughOverlay />);

    expect(screen.getByTestId("walkthrough-step")).toHaveAttribute(
      "data-left",
      "720",
    );
    expect(observe).toHaveBeenCalledWith(targetElement);

    const movedRect = createTargetRect({
      height: 56,
      left: 520,
      top: 12,
      width: 56,
    });
    getBoundingClientRect.mockReturnValue(movedRect);

    const notifyResize = resizeCallbackRef.current;
    if (!notifyResize) {
      throw new Error("Expected the target layout to be observed.");
    }

    act(() => {
      notifyResize(
        [
          {
            borderBoxSize: [],
            contentBoxSize: [],
            contentRect: movedRect,
            devicePixelContentBoxSize: [],
            target: targetElement!,
          },
        ],
        {} as ResizeObserver,
      );
    });

    await waitFor(() => {
      expect(screen.getByTestId("walkthrough-step")).toHaveAttribute(
        "data-left",
        "520",
      );
    });
  });

  it("leaves the highlighted control interactive while a required action is pending", () => {
    targetElement = document.createElement("button");
    targetElement.dataset.walkthroughTarget = walkthroughTargets.notifications;
    jest
      .spyOn(targetElement, "getBoundingClientRect")
      .mockReturnValue(
        createTargetRect({ height: 56, left: 720, top: 12, width: 56 }),
      );
    document.body.append(targetElement);
    useWalkthroughMock.mockReturnValue({
      currentStepData: {
        content: "Create a task.",
        id: "create-story",
        requiredAction: {
          actionLabel: "Create my first task",
          id: "story-created",
        },
        target: getWalkthroughTargetSelector(walkthroughTargets.notifications),
        title: "Create your first task",
      },
      isWalkthroughActionComplete: () => false,
      state: {
        currentStep: 1,
        hasSeenWalkthrough: false,
        isActive: true,
        totalSteps: 4,
        walkthroughVersion: "1.0.0",
      },
    } as unknown as ReturnType<typeof useWalkthrough>);

    render(<WalkthroughOverlay />);

    expect(screen.getByTestId("walkthrough-step")).toHaveAttribute(
      "data-allows-target-interaction",
      "true",
    );
    expect(
      document.querySelectorAll("[data-walkthrough-interaction-blocker]"),
    ).toHaveLength(4);
  });
});
