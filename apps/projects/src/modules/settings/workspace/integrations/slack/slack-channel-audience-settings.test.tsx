/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ComponentPropsWithoutRef, ElementType, ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import {
  useSlackChannelAudiences,
  useUpdateSlackChannelAudience,
} from "@/lib/hooks/slack";
import { useTeams } from "@/modules/teams/hooks/teams";
import { SlackChannelAudienceSettings } from "./slack-channel-audience-settings";
import type { SlackChannelAudience } from "./types";

jest.mock("@/lib/hooks/slack", () => ({
  useSlackChannelAudiences: jest.fn(),
  useUpdateSlackChannelAudience: jest.fn(),
}));

jest.mock("@/modules/teams/hooks/teams", () => ({
  useTeams: jest.fn(),
}));

jest.mock("@/components/ui/team-color", () => ({
  TeamColor: () => <span aria-hidden="true" />,
}));

jest.mock("icons", () => {
  const React = jest.requireActual("react");
  const Icon = (props: ComponentPropsWithoutRef<"svg">) =>
    React.createElement("svg", props);

  return {
    ArrowDownIcon: Icon,
    CheckIcon: Icon,
    LockIcon: Icon,
  };
});

jest.mock("ui", () => {
  const React = jest.requireActual("react");
  const Box = ({ children, ...props }: ComponentPropsWithoutRef<"div">) =>
    React.createElement("div", props, children);
  const Flex = ({
    align: _align,
    children,
    gap: _gap,
    justify: _justify,
    wrap: _wrap,
    ...props
  }: ComponentPropsWithoutRef<"div"> & {
    align?: string;
    gap?: number;
    justify?: string;
    wrap?: boolean;
  }) => React.createElement("div", props, children);
  const Text = ({
    as: Component = "span",
    children,
    color: _color,
    ...props
  }: ComponentPropsWithoutRef<"span"> & {
    as?: ElementType;
    color?: string;
  }) => React.createElement(Component, props, children);
  const Button = ({
    children,
    color: _color,
    leftIcon,
    loading,
    loadingText,
    rightIcon,
    variant: _variant,
    ...props
  }: ComponentPropsWithoutRef<"button"> & {
    color?: string;
    leftIcon?: ReactNode;
    loading?: boolean;
    loadingText?: string;
    rightIcon?: ReactNode;
    variant?: string;
  }) =>
    React.createElement(
      "button",
      { ...props, type: "button" },
      leftIcon,
      loading ? loadingText : children,
      rightIcon,
    );
  const Container = ({ children, ...props }: ComponentPropsWithoutRef<"div">) =>
    React.createElement("div", props, children);
  const Item = ({
    children,
    onSelect,
    value: _value,
    ...props
  }: ComponentPropsWithoutRef<"button"> & {
    onSelect?: () => void;
    value?: string;
  }) =>
    React.createElement(
      "button",
      {
        ...props,
        onClick: onSelect,
        type: "button",
      },
      children,
    );
  const Command = Container;
  Object.assign(Command, {
    Empty: Container,
    Group: Container,
    Input: (props: ComponentPropsWithoutRef<"input">) =>
      React.createElement("input", props),
    Item,
    Separator: () => React.createElement("hr"),
  });
  const Popover = Container;
  Object.assign(Popover, {
    Content: Container,
    Trigger: ({ children }: { asChild?: boolean; children: ReactNode }) =>
      children,
  });

  return {
    Badge: ({ children }: { children: ReactNode }) =>
      React.createElement("span", null, children),
    Box,
    Button,
    Command,
    Flex,
    Popover,
    Skeleton: (props: ComponentPropsWithoutRef<"div">) =>
      React.createElement("div", props),
    Text,
  };
});

const mockUseSlackChannelAudiences = jest.mocked(useSlackChannelAudiences);
const mockUseUpdateSlackChannelAudience = jest.mocked(
  useUpdateSlackChannelAudience,
);
const mockUseTeams = jest.mocked(useTeams);

const channelAudience: SlackChannelAudience = {
  channel: {
    id: "channel-record-1",
    slackChannelId: "C123",
    name: "product",
    isPrivate: false,
    isArchived: false,
    isMember: true,
    isActive: true,
    lastSyncedAt: "2026-08-14T09:00:00Z",
    createdAt: "2026-08-14T09:00:00Z",
    updatedAt: "2026-08-14T09:00:00Z",
  },
  teamIds: [],
};

const publicTeam = {
  id: "team-public",
  name: "Product",
  code: "PROD",
  color: "blue",
  isPrivate: false,
  workspaceId: "workspace-1",
  createdAt: "2026-08-14T09:00:00Z",
  updatedAt: "2026-08-14T09:00:00Z",
  memberCount: 4,
  sprintsEnabled: true,
};

const privateTeam = {
  ...publicTeam,
  id: "team-private",
  name: "Payroll",
  code: "PAY",
  isPrivate: true,
};

describe("SlackChannelAudienceSettings", () => {
  const mutate = jest.fn();
  const refetchAudiences = jest.fn();
  const refetchTeams = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    mockUseSlackChannelAudiences.mockReturnValue({
      data: [channelAudience],
      isError: false,
      isFetching: false,
      isPending: false,
      refetch: refetchAudiences,
    } as unknown as ReturnType<typeof useSlackChannelAudiences>);
    mockUseTeams.mockReturnValue({
      data: [publicTeam, privateTeam],
      isError: false,
      isFetching: false,
      isPending: false,
      refetch: refetchTeams,
    } as unknown as ReturnType<typeof useTeams>);
    mockUseUpdateSlackChannelAudience.mockReturnValue({
      isPending: false,
      mutate,
    } as unknown as ReturnType<typeof useUpdateSlackChannelAudience>);
  });

  it("renders an accessible loading state", () => {
    mockUseSlackChannelAudiences.mockReturnValue({
      isError: false,
      isFetching: true,
      isPending: true,
      refetch: refetchAudiences,
    } as unknown as ReturnType<typeof useSlackChannelAudiences>);

    render(<SlackChannelAudienceSettings />);

    expect(
      screen.getByRole("status", { name: "Loading Slack channel access" }),
    ).toBeInTheDocument();
  });

  it("offers a retry when either admin query fails", () => {
    mockUseSlackChannelAudiences.mockReturnValue({
      isError: true,
      isFetching: false,
      isPending: false,
      refetch: refetchAudiences,
    } as unknown as ReturnType<typeof useSlackChannelAudiences>);

    render(<SlackChannelAudienceSettings />);

    expect(screen.getByRole("alert")).toHaveTextContent(
      "Couldn't load Slack channel access",
    );
    fireEvent.click(screen.getByRole("button", { name: "Try again" }));
    expect(refetchAudiences).toHaveBeenCalledTimes(1);
    expect(refetchTeams).toHaveBeenCalledTimes(1);
  });

  it("saves selected public teams and keeps private teams unavailable", () => {
    render(<SlackChannelAudienceSettings />);

    expect(screen.getByText("#product")).toBeInTheDocument();
    expect(screen.getByText("Public Slack channel")).toBeInTheDocument();
    expect(screen.getByText("Personal only")).toBeInTheDocument();
    expect(
      screen.getByText(/cross-person reports require at least one/i),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Payroll/ })).toBeDisabled();

    fireEvent.click(screen.getByRole("button", { name: "Product" }));
    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(mutate).toHaveBeenCalledWith({
      channelId: "C123",
      teamIds: ["team-public"],
    });
  });

  it("preserves private-team mappings used by other Slack features", () => {
    mockUseSlackChannelAudiences.mockReturnValue({
      data: [{ ...channelAudience, teamIds: [privateTeam.id] }],
      isError: false,
      isFetching: false,
      isPending: false,
      refetch: refetchAudiences,
    } as unknown as ReturnType<typeof useSlackChannelAudiences>);

    render(<SlackChannelAudienceSettings />);

    expect(screen.getByRole("status")).toHaveTextContent(
      "1 existing private team mapping is preserved for other Slack features, but Maya will not use it for shared reports.",
    );
    expect(screen.getByText("Personal only")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Save" })).toBeDisabled();

    fireEvent.click(screen.getByRole("button", { name: "Product" }));
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(mutate).toHaveBeenCalledWith({
      channelId: "C123",
      teamIds: [privateTeam.id, publicTeam.id],
    });
  });

  it("shows the successful empty state when no channels are synced", () => {
    mockUseSlackChannelAudiences.mockReturnValue({
      data: [],
      isError: false,
      isFetching: false,
      isPending: false,
      refetch: refetchAudiences,
    } as unknown as ReturnType<typeof useSlackChannelAudiences>);

    render(<SlackChannelAudienceSettings />);

    expect(screen.getByText("No Slack channels synced")).toBeInTheDocument();
  });
});
