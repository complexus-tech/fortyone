/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type {
  ComponentPropsWithoutRef,
  ElementType,
  MouseEvent as ReactMouseEvent,
  ReactElement,
  ReactNode,
} from "react";
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
    LoadingIcon: Icon,
    MoreHorizontalIcon: Icon,
    PlusIcon: Icon,
    UnlinkIcon: Icon,
  };
});

jest.mock("ui", () => {
  const React = jest.requireActual("react");

  type OverlayState = {
    open: boolean;
    setOpen: (open: boolean) => void;
  };
  type OverlayRootProps = {
    children: ReactNode;
    onOpenChange?: (open: boolean) => void;
    open?: boolean;
  };
  type TriggerProps = {
    asChild?: boolean;
    children: ReactElement<{
      onClick?: ComponentPropsWithoutRef<"button">["onClick"];
    }>;
  };

  const Box = ({ children, ...props }: ComponentPropsWithoutRef<"div">) =>
    React.createElement("div", props, children);
  const Flex = ({
    align: _align,
    children,
    direction: _direction,
    gap: _gap,
    justify: _justify,
    wrap: _wrap,
    ...props
  }: ComponentPropsWithoutRef<"div"> & {
    align?: string;
    direction?: string;
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
    active: _active,
    align: _align,
    asIcon,
    children,
    color: _color,
    fullWidth: _fullWidth,
    leftIcon,
    loading,
    loadingText,
    rightIcon,
    rounded: _rounded,
    size,
    variant,
    ...props
  }: ComponentPropsWithoutRef<"button"> & {
    active?: boolean;
    align?: string;
    asIcon?: boolean;
    color?: string;
    fullWidth?: boolean;
    leftIcon?: ReactNode;
    loading?: boolean;
    loadingText?: string;
    rightIcon?: ReactNode;
    rounded?: string;
    size?: string;
    variant?: string;
  }) =>
    React.createElement(
      "button",
      {
        ...props,
        "data-as-icon": asIcon ? "true" : undefined,
        "data-size": size,
        "data-variant": variant,
        type: "button",
      },
      leftIcon,
      loading ? loadingText || "Loading..." : children,
      rightIcon,
    );
  const Badge = ({
    children,
    color: _color,
    rounded: _rounded,
    size = "md",
    variant: _variant,
    ...props
  }: ComponentPropsWithoutRef<"span"> & {
    color?: string;
    rounded?: string;
    size?: string;
    variant?: string;
  }) => React.createElement("span", { ...props, "data-size": size }, children);
  const Divider = (props: ComponentPropsWithoutRef<"hr">) =>
    React.createElement("hr", props);

  const CommandContext = React.createContext({
    query: "",
    setQuery: (_query: string) => undefined,
  });
  const CommandRoot = ({
    children,
    ...props
  }: ComponentPropsWithoutRef<"div">) => {
    const [query, setQuery] = React.useState("");
    return (
      <CommandContext.Provider value={{ query, setQuery }}>
        <div {...props}>{children}</div>
      </CommandContext.Provider>
    );
  };
  const CommandInput = ({
    onChange,
    ...props
  }: ComponentPropsWithoutRef<"input">) => {
    const { query, setQuery } = React.useContext(CommandContext);
    return (
      <input
        {...props}
        onChange={(event) => {
          setQuery(event.target.value);
          onChange?.(event);
        }}
        value={query}
      />
    );
  };
  const CommandItem = ({
    active: _active,
    children,
    onSelect,
    value = "",
    ...props
  }: ComponentPropsWithoutRef<"button"> & {
    active?: boolean;
    onSelect?: () => void;
    value?: string;
  }) => {
    const { query } = React.useContext(CommandContext) as {
      query: string;
    };
    if (
      query.trim() &&
      !value.toLocaleLowerCase().includes(query.trim().toLocaleLowerCase())
    ) {
      return null;
    }
    return (
      <button {...props} onClick={onSelect} type="button">
        {children}
      </button>
    );
  };
  const Command = Object.assign(CommandRoot, {
    Empty: () => null,
    Group: Box,
    Input: CommandInput,
    Item: CommandItem,
    List: (props: ComponentPropsWithoutRef<"div">) => (
      <div {...props} role="listbox" />
    ),
    Separator: Divider,
  });

  const createOverlay = () => {
    const OverlayContext = React.createContext(null as OverlayState | null);
    const Root = ({ children, onOpenChange, open }: OverlayRootProps) => {
      const [internalOpen, setInternalOpen] = React.useState(false);
      const effectiveOpen = open ?? internalOpen;
      const setOpen = (nextOpen: boolean) => {
        if (open === undefined) setInternalOpen(nextOpen);
        onOpenChange?.(nextOpen);
      };
      return (
        <OverlayContext.Provider value={{ open: effectiveOpen, setOpen }}>
          <div>{children}</div>
        </OverlayContext.Provider>
      );
    };
    const Trigger = ({ children }: TriggerProps) => {
      const overlay = React.useContext(OverlayContext);
      return React.cloneElement(children, {
        onClick: (event: ReactMouseEvent<HTMLButtonElement>) => {
          children.props.onClick?.(event);
          overlay?.setOpen(!overlay.open);
        },
      });
    };
    const Content = ({
      align: _align,
      children,
      ...props
    }: ComponentPropsWithoutRef<"div"> & { align?: string }) => {
      const overlay = React.useContext(OverlayContext);
      return overlay?.open ? <div {...props}>{children}</div> : null;
    };
    return {
      Content,
      Root,
      Trigger,
      useOverlay: () => React.useContext(OverlayContext),
    };
  };

  const popover = createOverlay();
  const Popover = Object.assign(popover.Root, {
    Content: popover.Content,
    Trigger: popover.Trigger,
  });

  const menu = createOverlay();
  const MenuItem = ({
    children,
    onSelect,
    ...props
  }: ComponentPropsWithoutRef<"button"> & { onSelect?: () => void }) => {
    const overlay = menu.useOverlay();
    return (
      <button
        {...props}
        onClick={() => {
          onSelect?.();
          overlay?.setOpen(false);
        }}
        type="button"
      >
        {children}
      </button>
    );
  };
  const MenuItems = ({
    align: _align,
    children,
    ...props
  }: ComponentPropsWithoutRef<"div"> & { align?: string }) => {
    const overlay = menu.useOverlay();
    return overlay?.open ? <div {...props}>{children}</div> : null;
  };
  const Menu = Object.assign(menu.Root, {
    Button: menu.Trigger,
    Group: Box,
    Item: MenuItem,
    Items: MenuItems,
  });

  return {
    Badge,
    Box,
    Button,
    Command,
    Divider,
    Flex,
    Menu,
    Popover,
    Skeleton: (props: ComponentPropsWithoutRef<"div">) => <div {...props} />,
    Text,
  };
});

const mockUseSlackChannelAudiences = jest.mocked(useSlackChannelAudiences);
const mockUseUpdateSlackChannelAudience = jest.mocked(
  useUpdateSlackChannelAudience,
);
const mockUseTeams = jest.mocked(useTeams);

const makeChannelAudience = ({
  id = "channel-record-1",
  isArchived = false,
  isConfigured = false,
  isPrivate = false,
  name = "product",
  slackChannelId = "C123",
  teamIds = [],
}: {
  id?: string;
  isArchived?: boolean;
  isConfigured?: boolean;
  isPrivate?: boolean;
  name?: string;
  slackChannelId?: string;
  teamIds?: string[];
} = {}): SlackChannelAudience => ({
  channel: {
    id,
    slackChannelId,
    name,
    isPrivate,
    isArchived,
    isMember: true,
    isActive: true,
    lastSyncedAt: "2026-08-14T09:00:00Z",
    createdAt: "2026-08-14T09:00:00Z",
    updatedAt: "2026-08-14T09:00:00Z",
  },
  isConfigured,
  teamIds,
});

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

  const setAudiences = (data: SlackChannelAudience[]) => {
    mockUseSlackChannelAudiences.mockReturnValue({
      data,
      isError: false,
      isFetching: false,
      isPending: false,
      refetch: refetchAudiences,
    } as unknown as ReturnType<typeof useSlackChannelAudiences>);
  };

  beforeEach(() => {
    jest.clearAllMocks();
    setAudiences([makeChannelAudience()]);
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

  it("renders an accessible loading state without channel actions", () => {
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
    expect(
      screen.queryByRole("button", { name: "Add channel" }),
    ).not.toBeInTheDocument();
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

  it("renders only configured channels and adds a searched channel", () => {
    setAudiences([
      makeChannelAudience({
        id: "configured-channel",
        isConfigured: true,
        name: "design",
        slackChannelId: "C-DESIGN",
      }),
      makeChannelAudience(),
      makeChannelAudience({
        id: "private-channel",
        isPrivate: true,
        name: "leadership",
        slackChannelId: "C-LEADERSHIP",
        teamIds: [privateTeam.id],
      }),
      makeChannelAudience({
        id: "archived-channel",
        isArchived: true,
        name: "archive",
        slackChannelId: "C-ARCHIVE",
      }),
    ]);

    render(<SlackChannelAudienceSettings />);

    expect(screen.getByText("#design")).toBeInTheDocument();
    expect(screen.queryByText("#product")).not.toBeInTheDocument();
    expect(screen.queryByText("#leadership")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Add channel" }));
    expect(screen.getByRole("listbox")).toBeInTheDocument();
    const search = screen.getByPlaceholderText("Search Slack channels...");
    fireEvent.change(search, { target: { value: "leadership" } });

    expect(screen.queryByText("#product")).not.toBeInTheDocument();
    expect(screen.queryByText("#archive")).not.toBeInTheDocument();
    fireEvent.click(
      screen.getByRole("button", { name: /#leadership private/i }),
    );

    expect(mutate).toHaveBeenCalledWith({
      channelId: "C-LEADERSHIP",
      isConfigured: true,
      teamIds: [privateTeam.id],
    });
  });

  it("autosaves a public-team selection without a Save button", () => {
    setAudiences([makeChannelAudience({ isConfigured: true })]);

    render(<SlackChannelAudienceSettings />);

    fireEvent.click(
      screen.getByRole("button", {
        name: "Choose work access for #product",
      }),
    );
    expect(screen.getByRole("listbox")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Product" }));

    expect(mutate).toHaveBeenCalledWith({
      channelId: "C123",
      isConfigured: true,
      teamIds: [publicTeam.id],
    });
    expect(
      screen.queryByRole("button", { name: /^Save$/ }),
    ).not.toBeInTheDocument();
    expect(mockUseUpdateSlackChannelAudience).toHaveBeenCalledWith("C123");
  });

  it("autosaves Personal only while keeping the channel configured", () => {
    setAudiences([
      makeChannelAudience({
        isConfigured: true,
        teamIds: [publicTeam.id],
      }),
    ]);

    render(<SlackChannelAudienceSettings />);

    fireEvent.click(
      screen.getByRole("button", {
        name: "Choose work access for #product",
      }),
    );
    fireEvent.click(screen.getByRole("button", { name: "Personal only" }));

    expect(mutate).toHaveBeenCalledWith({
      channelId: "C123",
      isConfigured: true,
      teamIds: [],
    });
  });

  it("removes a configured channel through the standard options menu", () => {
    setAudiences([
      makeChannelAudience({
        isConfigured: true,
        teamIds: [publicTeam.id],
      }),
    ]);

    render(<SlackChannelAudienceSettings />);

    fireEvent.click(
      screen.getByRole("button", {
        name: "Channel options for #product",
      }),
    );
    fireEvent.click(screen.getByRole("button", { name: /Remove channel/i }));

    expect(mutate).toHaveBeenCalledWith({
      channelId: "C123",
      isConfigured: false,
      teamIds: [],
    });
  });

  it("keeps private teams available only as taller disabled menu items", () => {
    setAudiences([makeChannelAudience({ isConfigured: true })]);

    render(<SlackChannelAudienceSettings />);

    expect(screen.getByText("Public")).toHaveAttribute("data-size", "md");
    expect(screen.queryByText("Public Slack channel")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Add channel" })).toHaveAttribute(
      "data-size",
      "sm",
    );
    expect(
      screen.getByRole("button", {
        name: "Choose work access for #product",
      }),
    ).toHaveAttribute("data-size", "sm");
    expect(
      screen.getByRole("button", {
        name: "Channel options for #product",
      }),
    ).toHaveAttribute("data-as-icon", "true");
    expect(
      screen.getByRole("button", {
        name: "Channel options for #product",
      }),
    ).toHaveAttribute("data-size", "sm");
    expect(
      screen.getByRole("group", { name: "Actions for #product" }),
    ).toHaveClass("gap-1");
    fireEvent.click(
      screen.getByRole("button", {
        name: "Choose work access for #product",
      }),
    );

    expect(screen.getByText("Private teams")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Payroll" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Payroll" })).toHaveClass(
      "min-h-11",
    );
    expect(screen.queryByText(/^Private$/)).not.toBeInTheDocument();
    expect(
      screen.queryByText(/private teams cannot be included/i),
    ).not.toBeInTheDocument();
  });

  it("shows a compact row-level saving state without blocking other rows", () => {
    setAudiences([
      makeChannelAudience({ isConfigured: true }),
      makeChannelAudience({
        id: "general-channel",
        isConfigured: true,
        name: "general",
        slackChannelId: "C-GENERAL",
      }),
    ]);
    mockUseUpdateSlackChannelAudience.mockImplementation(
      (channelId) =>
        ({
          isPending: channelId === "C123",
          mutate,
        }) as unknown as ReturnType<typeof useUpdateSlackChannelAudience>,
    );

    render(<SlackChannelAudienceSettings />);

    expect(screen.getByLabelText("Saving #product")).toBeInTheDocument();
    expect(
      screen.getByRole("button", {
        name: "Choose work access for #product",
      }),
    ).toBeDisabled();
    expect(
      screen.getByRole("button", {
        name: "Choose work access for #general",
      }),
    ).toBeEnabled();
  });

  it("distinguishes no synced channels from no configured channels", () => {
    const { rerender } = render(<SlackChannelAudienceSettings />);

    expect(screen.getByText("No channels configured")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Add channel" })).toBeEnabled();

    setAudiences([]);
    rerender(<SlackChannelAudienceSettings />);

    expect(screen.getByText("No Slack channels synced")).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Add channel" }),
    ).not.toBeInTheDocument();
  });
});
