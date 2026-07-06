/* global beforeAll, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type * as ReactTypes from "react";
import { render, screen } from "@testing-library/react";
import { publicPortalFixture } from "./fixtures";
import {
  PublicPortalRequestDetailPage,
  PublicPortalRequestsPage,
  PublicPortalRoadmapPage,
} from ".";

const designSystemOnlyProps = new Set([
  "active",
  "align",
  "asIcon",
  "color",
  "direction",
  "fontSize",
  "fontWeight",
  "fullWidth",
  "gap",
  "justify",
  "rounded",
  "size",
  "variant",
  "wrap",
]);

const getDomProps = <T extends Record<string, unknown>>(props: T) =>
  Object.fromEntries(
    Object.entries(props).filter(([key]) => !designSystemOnlyProps.has(key)),
  );

jest.mock("icons", () => {
  const Icon = () => <svg aria-hidden="true" />;
  return {
    ArrowRightIcon: Icon,
    ArrowRight2Icon: Icon,
    ArrowUpDownIcon: Icon,
    ArrowUpIcon: Icon,
    BellIcon: Icon,
    CommentIcon: Icon,
    CopyIcon: Icon,
    DashboardIcon: Icon,
    GanttIcon: Icon,
    KanbanIcon: Icon,
    ListIcon: Icon,
    LogoutIcon: Icon,
    MoonIcon: Icon,
    PlusIcon: Icon,
    RequestsIcon: Icon,
    RoadmapIcon: Icon,
    SearchIcon: Icon,
    ShareIcon: Icon,
    SettingsIcon: Icon,
    StoryIcon: Icon,
    SunIcon: Icon,
    SystemIcon: Icon,
    UpdatesIcon: Icon,
  };
});

jest.mock("next-themes", () => ({
  useTheme: () => ({
    setTheme: jest.fn(),
    theme: "system",
  }),
}));

jest.mock("@/components/shared/sidebar/actions", () => ({
  logOut: jest.fn(),
}));

jest.mock("@/components/shared/sidebar/utils", () => ({
  clearAllStorage: jest.fn(),
}));

jest.mock("./actions", () => ({
  createFeedbackAction: jest.fn(),
  createFeedbackCommentAction: jest.fn(),
  createStoryFromFeedbackAction: jest.fn(),
  toggleFeedbackVoteAction: jest.fn(),
}));

jest.mock("sonner", () => ({
  toast: {
    error: jest.fn(),
    success: jest.fn(),
  },
}));

jest.mock("ui", () => {
  const React = jest.requireActual("react");

  const Box = ({
    children,
    ...props
  }: ReactTypes.HTMLAttributes<HTMLDivElement> & Record<string, unknown>) => (
    <div {...getDomProps(props)}>{children}</div>
  );
  const Flex = ({
    children,
    ...props
  }: ReactTypes.HTMLAttributes<HTMLDivElement> & Record<string, unknown>) => (
    <div {...getDomProps(props)}>{children}</div>
  );
  const Text = ({
    as: Tag = "p",
    children,
    ...props
  }: ReactTypes.HTMLAttributes<HTMLElement> &
    Record<string, unknown> & {
      as?: ReactTypes.ElementType;
    }) => React.createElement(Tag, getDomProps(props), children);
  const Button = ({
    children,
    href,
    leftIcon,
    rightIcon,
    ...props
  }: ReactTypes.ButtonHTMLAttributes<HTMLButtonElement> &
    Record<string, unknown> & {
      href?: string;
      leftIcon?: ReactTypes.ReactNode;
      rightIcon?: ReactTypes.ReactNode;
    }) =>
    href ? (
      <a {...getDomProps(props)} href={href}>
        {leftIcon}
        {children}
        {rightIcon}
      </a>
    ) : (
      <button {...getDomProps(props)} type="button">
        {leftIcon}
        {children}
        {rightIcon}
      </button>
    );
  const Avatar = ({ name }: { name?: string }) => (
    <div>{name ? name.slice(0, 2).toUpperCase() : "U"}</div>
  );
  const Dialog = ({
    children,
    open,
  }: {
    children: ReactTypes.ReactNode;
    open?: boolean;
  }) => <div>{open ? children : children}</div>;
  function DialogContent({ children }: { children: ReactTypes.ReactNode }) {
    return <div>{children}</div>;
  }
  function DialogHeader({ children }: { children: ReactTypes.ReactNode }) {
    return <div>{children}</div>;
  }
  function DialogTitle({ children }: { children: ReactTypes.ReactNode }) {
    return <div>{children}</div>;
  }
  function DialogBody({ children }: { children: ReactTypes.ReactNode }) {
    return <div>{children}</div>;
  }
  function DialogFooter({ children }: { children: ReactTypes.ReactNode }) {
    return <div>{children}</div>;
  }
  Dialog.Content = DialogContent;
  Dialog.Header = DialogHeader;
  Dialog.Title = DialogTitle;
  Dialog.Body = DialogBody;
  Dialog.Footer = DialogFooter;
  const Input = ({
    leftIcon,
    ...props
  }: ReactTypes.InputHTMLAttributes<HTMLInputElement> &
    Record<string, unknown> & {
      leftIcon?: ReactTypes.ReactNode;
    }) => {
    void leftIcon;
    return <input {...getDomProps(props)} />;
  };
  const TextArea = (
    props: ReactTypes.TextareaHTMLAttributes<HTMLTextAreaElement>,
  ) => <textarea {...props} />;
  const Menu = ({ children }: { children: ReactTypes.ReactNode }) => (
    <div>{children}</div>
  );
  const MenuButton = ({ children }: { children: ReactTypes.ReactNode }) => (
    <>{children}</>
  );
  const MenuItems = ({ children }: { children: ReactTypes.ReactNode }) => (
    <div>{children}</div>
  );
  const MenuGroup = ({ children }: { children: ReactTypes.ReactNode }) => (
    <div>{children}</div>
  );
  const MenuSeparator = () => <hr />;
  const MenuItem = ({ children }: { children: ReactTypes.ReactNode }) => (
    <div>{children}</div>
  );
  const MenuSubMenu = ({ children }: { children: ReactTypes.ReactNode }) => (
    <div>{children}</div>
  );
  const MenuSubTrigger = ({ children }: { children: ReactTypes.ReactNode }) => (
    <div>{children}</div>
  );
  const MenuSubItems = ({ children }: { children: ReactTypes.ReactNode }) => (
    <div>{children}</div>
  );
  Menu.Button = MenuButton;
  Menu.Items = MenuItems;
  Menu.Group = MenuGroup;
  Menu.Separator = MenuSeparator;
  Menu.Item = MenuItem;
  Menu.SubMenu = MenuSubMenu;
  Menu.SubTrigger = MenuSubTrigger;
  Menu.SubItems = MenuSubItems;

  return {
    Avatar,
    Box,
    Button,
    Dialog,
    Flex,
    Input,
    Menu,
    Text,
    TextArea,
  };
});

describe("Public portal UI", () => {
  beforeAll(() => {
    class IntersectionObserverMock {
      disconnect = jest.fn();
      observe = jest.fn();
      unobserve = jest.fn();
    }

    Object.defineProperty(global, "IntersectionObserver", {
      value: IntersectionObserverMock,
      writable: true,
    });
  });

  beforeEach(() => {
    global.fetch = jest.fn(
      async (input: Parameters<typeof fetch>[0]): Promise<Response> => {
        let url: string;
        if (typeof input === "string") {
          url = input;
        } else if (input instanceof URL) {
          url = input.toString();
        } else {
          url = input.url;
        }
        const requestUrl = new URL(url, "https://fortyone.test");
        const status = requestUrl.searchParams.get("status");
        const requests = status
          ? publicPortalFixture.requests.filter(
              (request) => request.status === status,
            )
          : publicPortalFixture.requests;

        return {
          json: async () => ({
            data: {
              ...publicPortalFixture,
              requests,
              requestsHasMore: false,
            },
          }),
          ok: true,
        } as Response;
      },
    );
  });

  it("renders the public feedback page with feedback terminology", () => {
    render(<PublicPortalRequestsPage portal={publicPortalFixture} />);

    expect(
      screen.getAllByRole("link", { name: /^Feedback$/i }).length,
    ).toBeGreaterThan(0);
    expect(
      screen.getByRole("button", { name: /new feedback/i }),
    ).toBeInTheDocument();
    expect(screen.getByText("All Feedback")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Login/signup" })).toHaveAttribute(
      "href",
      "/",
    );
    expect(screen.queryByText("All Requests")).not.toBeInTheDocument();
  });

  it("renders signed-in portal navigation controls", () => {
    render(
      <PublicPortalRequestsPage
        portal={publicPortalFixture}
        viewer={{
          accountHref: "/city-roads/settings/account",
          appHref: "/city-roads/my-work",
          avatarUrl: null,
          email: "ada@example.com",
          name: "Ada Ndlovu",
          notificationsHref: "/city-roads/notifications",
        }}
      />,
    );

    expect(
      screen.queryByRole("link", { name: "Login/signup" }),
    ).not.toBeInTheDocument();
    expect(screen.getByLabelText("Open account menu")).toBeInTheDocument();
    expect(screen.getByText("ada@example.com")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /open app/i })).toHaveAttribute(
      "href",
      "/city-roads/my-work",
    );
    expect(
      screen.getByRole("link", { name: /account settings/i }),
    ).toHaveAttribute("href", "/city-roads/settings/account");
  });

  it("renders roadmap columns from promoted public requests", async () => {
    render(<PublicPortalRoadmapPage portal={publicPortalFixture} />);

    expect(screen.getByText("Planned")).toBeInTheDocument();
    expect(screen.getByText("In Progress")).toBeInTheDocument();
    expect(screen.getByText("Done")).toBeInTheDocument();
    expect(
      await screen.findByText("Resurface Market Road before rainy season"),
    ).toBeInTheDocument();
    expect(
      screen.queryByText("Add pedestrian crossing near East Avenue school"),
    ).not.toBeInTheDocument();
  });

  it("renders a public request detail page with comments and metadata", () => {
    render(
      <PublicPortalRequestDetailPage
        portal={publicPortalFixture}
        request={publicPortalFixture.requests[0]}
      />,
    );

    expect(
      screen.getByRole("heading", {
        name: "Add pedestrian crossing near East Avenue school",
      }),
    ).toBeInTheDocument();
    expect(screen.getByPlaceholderText("Add a comment...")).toBeInTheDocument();
    expect(screen.getByText("Road repairs")).toBeInTheDocument();
    expect(screen.getByText("Copy link")).toBeInTheDocument();
  });
});
