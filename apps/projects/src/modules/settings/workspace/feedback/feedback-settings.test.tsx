/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ComponentPropsWithoutRef, ElementType, ReactNode } from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { useTeams } from "@/modules/teams/hooks/teams";
import {
  useCreateFeedbackBoardMutation,
  useFeedbackPortals,
  useUpdateFeedbackPortalMutation,
} from "./hooks";
import { FeedbackSettings } from ".";

jest.mock("@/modules/teams/hooks/teams", () => ({
  useTeams: jest.fn(),
}));

jest.mock("icons", () => {
  const React = jest.requireActual("react");
  const Icon = (props: ComponentPropsWithoutRef<"svg">) =>
    React.createElement("svg", props);

  return {
    LinkIcon: Icon,
    PlusIcon: Icon,
    TeamIcon: Icon,
  };
});

jest.mock("ui", () => {
  const React = jest.requireActual("react");

  const Box = ({ children, ...props }: ComponentPropsWithoutRef<"div">) =>
    React.createElement("div", props, children);
  const Flex = Box;
  const Text = ({
    as = "span",
    children,
    color: _color,
    fontWeight: _fontWeight,
    ...props
  }: ComponentPropsWithoutRef<"span"> & {
    as?: ElementType;
    color?: string;
    fontWeight?: string;
  }) => React.createElement(as, props, children);
  const Button = ({
    children,
    color: _color,
    href,
    leftIcon,
    size: _size,
    ...props
  }: ComponentPropsWithoutRef<"button"> & {
    color?: string;
    href?: string;
    leftIcon?: ReactNode;
    size?: string;
  }) =>
    React.createElement(
      href ? "a" : "button",
      href ? { ...props, href } : props,
      leftIcon,
      children,
    );
  const Dialog = ({ children }: { children: ReactNode }) =>
    React.createElement(React.Fragment, null, children);
  const DialogPart = ({
    children,
    ...props
  }: ComponentPropsWithoutRef<"div">) =>
    React.createElement("div", props, children);

  Object.assign(Dialog, {
    Body: DialogPart,
    Content: DialogPart,
    Footer: DialogPart,
    Header: DialogPart,
    Title: DialogPart,
  });

  return {
    Box,
    Button,
    Dialog,
    Flex,
    Input: (props: ComponentPropsWithoutRef<"input">) =>
      React.createElement("input", props),
    Switch: ({
      checked,
      onCheckedChange,
      ...props
    }: ComponentPropsWithoutRef<"button"> & {
      checked: boolean;
      onCheckedChange: (checked: boolean) => void;
    }) =>
      React.createElement("button", {
        ...props,
        "aria-checked": checked,
        onClick: () => {
          onCheckedChange(!checked);
        },
        role: "switch",
        type: "button",
      }),
    Text,
  };
});

jest.mock("./hooks", () => ({
  useCreateFeedbackBoardMutation: jest.fn(),
  useFeedbackPortals: jest.fn(),
  useUpdateFeedbackPortalMutation: jest.fn(),
}));

const mockUseTeams = jest.mocked(useTeams);
const mockUseFeedbackPortals = jest.mocked(useFeedbackPortals);
const mockUseCreateFeedbackBoardMutation = jest.mocked(
  useCreateFeedbackBoardMutation,
);
const mockUseUpdateFeedbackPortalMutation = jest.mocked(
  useUpdateFeedbackPortalMutation,
);

const portal = {
  id: "portal-1",
  workspaceId: "workspace-1",
  name: "City Roads",
  slug: "city-roads",
  isPublic: true,
  createdAt: "2026-07-19T00:00:00.000Z",
  updatedAt: "2026-07-19T00:00:00.000Z",
  boards: [
    {
      id: "board-1",
      workspaceId: "workspace-1",
      portalId: "portal-1",
      teamId: "team-1",
      name: "Road safety",
      slug: "road-safety",
      color: "red",
      orderIndex: 0,
      createdAt: "2026-07-19T00:00:00.000Z",
      updatedAt: "2026-07-19T00:00:00.000Z",
    },
  ],
};

describe("FeedbackSettings", () => {
  const updatePortal = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    updatePortal.mockResolvedValue({ data: portal });
    mockUseTeams.mockReturnValue({
      data: [{ id: "team-1", name: "Operations" }],
    } as ReturnType<typeof useTeams>);
    mockUseFeedbackPortals.mockReturnValue({
      data: [portal],
      isLoading: false,
    } as ReturnType<typeof useFeedbackPortals>);
    mockUseCreateFeedbackBoardMutation.mockReturnValue({
      isPending: false,
      mutateAsync: jest.fn(),
    } as never);
    mockUseUpdateFeedbackPortalMutation.mockReturnValue({
      isPending: false,
      mutateAsync: updatePortal,
    } as never);
  });

  it("saves portal availability as soon as the switch changes", async () => {
    render(<FeedbackSettings />);

    expect(
      screen.getByText(/customers can submit requests, vote on ideas/i),
    ).toBeInTheDocument();
    expect(
      screen.queryByPlaceholderText("Describe what people should submit here"),
    ).not.toBeInTheDocument();
    expect(screen.queryByText("/feedback")).not.toBeInTheDocument();

    const enabledSwitch = screen.getByRole("switch", {
      name: "Enable public feedback portal",
    });

    expect(
      screen.queryByRole("button", { name: "Save Changes" }),
    ).not.toBeInTheDocument();
    fireEvent.click(enabledSwitch);

    await waitFor(() => {
      expect(updatePortal).toHaveBeenCalledWith({
        input: { isPublic: false },
        portalId: portal.id,
      });
    });
  });

  it.each([
    ["an API error response", { error: { message: "Could not update" } }],
    ["a rejected request", new Error("Network unavailable")],
  ])("rolls the switch back after %s", async (_label, result) => {
    if (result instanceof Error) {
      updatePortal.mockRejectedValueOnce(result);
    } else {
      updatePortal.mockResolvedValueOnce(result);
    }
    render(<FeedbackSettings />);

    const enabledSwitch = screen.getByRole("switch", {
      name: "Enable public feedback portal",
    });
    fireEvent.click(enabledSwitch);

    await waitFor(() => {
      expect(enabledSwitch).toHaveAttribute("aria-checked", "true");
    });
  });

  it("places the board color inside a muted surface", () => {
    const { container } = render(<FeedbackSettings />);

    const swatches = container.querySelectorAll<HTMLElement>(
      '[style*="background-color: red"]',
    );
    const boardSwatch = Array.from(swatches).find((swatch) =>
      swatch.parentElement?.classList.contains("bg-surface-muted/70"),
    );

    expect(boardSwatch).toBeInTheDocument();
  });
});
