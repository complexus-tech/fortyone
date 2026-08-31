/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { render, screen } from "@testing-library/react";
import { useWalkthrough } from "./walkthrough-provider";
import { WalkthroughIntegration } from "./walkthrough-integration";

let pathname = "/acme/calendar";

jest.mock("next/navigation", () => ({
  usePathname: () => pathname,
}));

jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));

jest.mock("./use-walkthrough-steps", () => ({
  useWalkthroughSteps: () => [],
}));

jest.mock("./walkthrough-overlay", () => ({
  WalkthroughOverlay: () => null,
}));

jest.mock("./walkthrough-provider", () => {
  const React = jest.requireActual("react");

  return {
    WalkthroughProvider: ({
      autoStart,
      children,
      dismissOnClose,
      tourKey,
      version,
    }: {
      autoStart?: boolean;
      children: React.ReactNode;
      dismissOnClose?: boolean;
      tourKey?: string;
      version?: string;
    }) =>
      React.createElement(
        "div",
        {
          "data-auto-start": autoStart ? "true" : "false",
          "data-dismiss-on-close": dismissOnClose ? "true" : "false",
          "data-testid": "module-walkthrough-provider",
          "data-tour-key": tourKey,
          "data-tour-version": version,
        },
        children,
      ),
    useWalkthrough: jest.fn(),
  };
});

const useWalkthroughMock = jest.mocked(useWalkthrough);
const setStepsMock = jest.fn();

const mockWalkthroughState = ({
  hasSeenWalkthrough,
  isActive,
}: {
  hasSeenWalkthrough: boolean;
  isActive: boolean;
}) => {
  useWalkthroughMock.mockReturnValue({
    setSteps: setStepsMock,
    state: {
      currentStep: 0,
      hasSeenWalkthrough,
      isActive,
      progress: {
        current: 0,
        total: 0,
      },
      totalSteps: 0,
      walkthroughVersion: "1.0.0",
    },
  } as unknown as ReturnType<typeof useWalkthrough>);
};

describe("WalkthroughIntegration", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    pathname = "/acme/calendar";
  });

  it("mounts the persisted tour for the current module after setup", () => {
    mockWalkthroughState({ hasSeenWalkthrough: true, isActive: false });

    render(<WalkthroughIntegration />);

    expect(screen.getByTestId("module-walkthrough-provider")).toHaveAttribute(
      "data-tour-key",
      "workspace-module-calendar",
    );
    expect(screen.getByTestId("module-walkthrough-provider")).toHaveAttribute(
      "data-tour-version",
      "1.0.0",
    );
    expect(screen.getByTestId("module-walkthrough-provider")).toHaveAttribute(
      "data-auto-start",
      "true",
    );
    expect(screen.getByTestId("module-walkthrough-provider")).toHaveAttribute(
      "data-dismiss-on-close",
      "true",
    );
  });

  it("does not overlap a running getting-started tour", () => {
    mockWalkthroughState({ hasSeenWalkthrough: true, isActive: true });

    render(<WalkthroughIntegration />);

    expect(
      screen.queryByTestId("module-walkthrough-provider"),
    ).not.toBeInTheDocument();
  });

  it("does not attach route tours to unsupported screens", () => {
    pathname = "/acme/settings";
    mockWalkthroughState({ hasSeenWalkthrough: true, isActive: false });

    render(<WalkthroughIntegration />);

    expect(
      screen.queryByTestId("module-walkthrough-provider"),
    ).not.toBeInTheDocument();
  });
});
