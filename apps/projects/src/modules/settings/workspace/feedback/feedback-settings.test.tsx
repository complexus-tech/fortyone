/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type {
  ChangeEvent,
  ComponentPropsWithoutRef,
  ElementType,
  ReactNode,
} from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { useTeams } from "@/modules/teams/hooks/teams";
import {
  useCreateFeedbackBoardMutation,
  useFeedbackBoardReviewers,
  useFeedbackPortals,
  useUpdateFeedbackBoardReviewerMutation,
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
    ArrowDownIcon: Icon,
    LinkIcon: Icon,
    PlusIcon: Icon,
    TeamIcon: Icon,
  };
});

jest.mock("ui", () => {
  const React = jest.requireActual("react");
  const DialogContext = React.createContext(true);
  const SelectContext = React.createContext({
    disabled: false,
    onValueChange: (_value: string) => undefined,
    value: "",
  });

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
  const Dialog = ({
    children,
    open = true,
  }: {
    children: ReactNode;
    open?: boolean;
  }) => React.createElement(DialogContext.Provider, { value: open }, children);
  const DialogPart = ({
    children,
    ...props
  }: ComponentPropsWithoutRef<"div">) =>
    React.createElement("div", props, children);
  const DialogContent = ({
    children,
    ...props
  }: ComponentPropsWithoutRef<"div">) => {
    const open = React.useContext(DialogContext);
    return open ? React.createElement("div", props, children) : null;
  };

  Object.assign(Dialog, {
    Body: DialogPart,
    Content: DialogContent,
    Description: DialogPart,
    Footer: DialogPart,
    Header: DialogPart,
    Title: DialogPart,
  });

  const Select = ({
    children,
    disabled = false,
    onValueChange,
    value,
  }: {
    children: ReactNode;
    disabled?: boolean;
    onValueChange: (value: string) => void;
    value: string;
  }) =>
    React.createElement(
      SelectContext.Provider,
      { value: { disabled, onValueChange, value } },
      children,
    );
  const SelectTrigger = ({
    children: _children,
    ...props
  }: ComponentPropsWithoutRef<"select">) => {
    const context = React.useContext(SelectContext);
    return React.createElement(
      "select",
      {
        ...props,
        disabled: context.disabled,
        onChange: (event: ChangeEvent<HTMLSelectElement>) => {
          context.onValueChange(event.target.value);
        },
        value: context.value,
      },
      React.createElement("option", { value: "off" }, "Off"),
      React.createElement("option", { value: "daily" }, "Daily"),
      React.createElement("option", { value: "weekly" }, "Weekly"),
    );
  };
  Object.assign(Select, {
    Content: () => null,
    Input: () => null,
    Option: () => null,
    Trigger: SelectTrigger,
  });

  return {
    Avatar: ({ name, ...props }: { name: string }) =>
      React.createElement("div", { ...props, "aria-label": name }),
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
    Select,
    Skeleton: (props: ComponentPropsWithoutRef<"div">) =>
      React.createElement("div", props),
    Text,
  };
});

jest.mock("./hooks", () => ({
  useCreateFeedbackBoardMutation: jest.fn(),
  useFeedbackBoardReviewers: jest.fn(),
  useFeedbackPortals: jest.fn(),
  useUpdateFeedbackBoardReviewerMutation: jest.fn(),
  useUpdateFeedbackPortalMutation: jest.fn(),
}));

const mockUseTeams = jest.mocked(useTeams);
const mockUseFeedbackPortals = jest.mocked(useFeedbackPortals);
const mockUseFeedbackBoardReviewers = jest.mocked(useFeedbackBoardReviewers);
const mockUseCreateFeedbackBoardMutation = jest.mocked(
  useCreateFeedbackBoardMutation,
);
const mockUseUpdateFeedbackBoardReviewerMutation = jest.mocked(
  useUpdateFeedbackBoardReviewerMutation,
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
  const updateReviewer = jest.fn();

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
    mockUseFeedbackBoardReviewers.mockReturnValue({
      data: [
        {
          avatarUrl: "https://cdn.example.com/amina.jpg",
          email: "amina@example.com",
          emailFrequency: "daily",
          name: "Amina Moyo",
          role: "admin",
          userId: "user-1",
        },
        {
          email: "tariro@example.com",
          emailFrequency: "off",
          name: "Tariro Ncube",
          role: "member",
          userId: "user-2",
        },
      ],
      isError: false,
      isLoading: false,
      refetch: jest.fn(),
    } as never);
    mockUseUpdateFeedbackBoardReviewerMutation.mockReturnValue({
      isPending: false,
      mutate: updateReviewer,
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

  it("shows the board color without a wrapper or slug", () => {
    const { container } = render(<FeedbackSettings />);

    const boardSwatch = container.querySelector<HTMLElement>(
      '[style*="background-color: red"]',
    );

    expect(boardSwatch).toBeInTheDocument();
    expect(boardSwatch?.parentElement).not.toHaveClass("bg-surface-muted/70");
    expect(screen.queryByText("road-safety")).not.toBeInTheDocument();
  });

  it("disables board creation when every team already has a board", () => {
    render(<FeedbackSettings />);

    expect(screen.getByRole("button", { name: "Create Board" })).toBeDisabled();
  });

  it("offers only teams that do not already have a board", () => {
    mockUseTeams.mockReturnValue({
      data: [
        { id: "team-1", name: "Operations" },
        { id: "team-2", name: "Customer success" },
      ],
    } as ReturnType<typeof useTeams>);
    render(<FeedbackSettings />);

    fireEvent.click(screen.getByRole("button", { name: "Create Board" }));

    const teamSelect = screen.getByRole("combobox", { name: "Owning team" });
    expect(teamSelect).toHaveValue("team-2");
    expect(teamSelect).toHaveClass("appearance-none", "pr-10", "pl-3");
    expect(teamSelect.nextElementSibling).toHaveClass("right-3.5");
    expect(
      screen.queryByRole("option", { name: "Operations" }),
    ).not.toBeInTheDocument();
    expect(
      screen.getByRole("option", { name: "Customer success" }),
    ).toBeInTheDocument();
  });

  it("loads reviewers only when opened and auto-saves their frequency", async () => {
    render(<FeedbackSettings />);

    expect(mockUseFeedbackBoardReviewers).toHaveBeenCalledWith(
      "board-1",
      false,
    );
    expect(screen.queryByText("1 subscribed")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Reviewers" }));

    await waitFor(() => {
      expect(mockUseFeedbackBoardReviewers).toHaveBeenLastCalledWith(
        "board-1",
        true,
      );
    });

    expect(screen.getByText("Feedback stays immediate")).toBeInTheDocument();
    expect(
      screen.getByText("Feedback stays immediate").parentElement,
    ).toHaveClass("dark:bg-white/[0.04]");
    expect(screen.getByText("Amina Moyo")).toBeInTheDocument();
    expect(screen.getByLabelText("Amina Moyo")).toHaveAttribute(
      "src",
      "https://cdn.example.com/amina.jpg",
    );
    expect(screen.getByText("amina@example.com")).toBeInTheDocument();
    expect(screen.getByText("Admin")).toBeInTheDocument();
    expect(screen.getByText("1 subscribed")).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /save/i }),
    ).not.toBeInTheDocument();

    fireEvent.change(
      screen.getByRole("combobox", {
        name: "Email summary for Tariro Ncube",
      }),
      { target: { value: "weekly" } },
    );

    expect(updateReviewer).toHaveBeenCalledWith({
      input: { emailFrequency: "weekly" },
      userId: "user-2",
    });
  });
});
