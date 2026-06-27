/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type * as ReactTypes from "react";
import { render, screen } from "@testing-library/react";
import {
  PublicPortalRequestDetailPage,
  PublicPortalRequestsPage,
  PublicPortalRoadmapPage,
  publicPortalFixture,
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
    LogoutIcon: Icon,
    MoonIcon: Icon,
    PlusIcon: Icon,
    RequestsIcon: Icon,
    RoadmapIcon: Icon,
    SearchIcon: Icon,
    ShareIcon: Icon,
    SettingsIcon: Icon,
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
    }) => (
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
    )
  );
  const Avatar = ({ name }: { name?: string }) => (
    <div>{name ? name.slice(0, 2).toUpperCase() : "U"}</div>
  );
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
  const MenuSubTrigger = ({
    children,
  }: {
    children: ReactTypes.ReactNode;
  }) => <div>{children}</div>;
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
    Flex,
    Menu,
    Text,
  };
});

describe("Public portal UI", () => {
  it("renders the public requests page with requests terminology", () => {
    render(<PublicPortalRequestsPage portal={publicPortalFixture} />);

    expect(
      screen.getAllByRole("link", { name: /^Requests$/i }).length,
    ).toBeGreaterThan(0);
    expect(
      screen.getByRole("button", { name: /new request/i }),
    ).toBeInTheDocument();
    expect(screen.getByText("All Requests")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Login/signup" })).toHaveAttribute(
      "href",
      "/",
    );
    expect(screen.queryByText("All Feedback")).not.toBeInTheDocument();
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

  it("renders roadmap columns from promoted public requests", () => {
    render(<PublicPortalRoadmapPage portal={publicPortalFixture} />);

    expect(screen.getByText("Planned")).toBeInTheDocument();
    expect(screen.getByText("In Progress")).toBeInTheDocument();
    expect(screen.getByText("Done")).toBeInTheDocument();
    expect(
      screen.getByText("Resurface Market Road before rainy season"),
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
    expect(screen.getByText("Add a comment...")).toBeInTheDocument();
    expect(screen.getByText("Road repairs")).toBeInTheDocument();
    expect(screen.getByText("Copy link")).toBeInTheDocument();
  });
});
