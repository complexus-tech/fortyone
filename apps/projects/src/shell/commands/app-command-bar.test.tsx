/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { AppCommandBar } from "./app-command-bar";

jest.mock("next/navigation", () => ({
  useParams: () => ({}),
  usePathname: () => "/acme/my-work",
}));

jest.mock("ui", () => {
  const MockButton = ({
    children,
    disabled,
    href,
    leftIcon,
    onClick,
    "aria-label": ariaLabel,
  }: {
    children?: ReactNode;
    disabled?: boolean;
    href?: string;
    leftIcon?: ReactNode;
    onClick?: () => void;
    "aria-label"?: string;
  }) =>
    href ? (
      <a aria-label={ariaLabel} href={href}>
        {leftIcon}
        {children}
      </a>
    ) : (
      <button
        aria-label={ariaLabel}
        disabled={disabled}
        onClick={onClick}
        type="button"
      >
        {leftIcon}
        {children}
      </button>
    );
  const MockContainer = ({ children }: { children: ReactNode }) => (
    <div>{children}</div>
  );
  const MockMenuItem = ({
    children,
    onSelect,
  }: {
    children: ReactNode;
    onSelect?: () => void;
  }) => (
    <button onClick={onSelect} type="button">
      {children}
    </button>
  );

  const MockMenu = Object.assign(MockContainer, {
    Button: MockContainer,
    Group: MockContainer,
    Item: MockMenuItem,
    Items: MockContainer,
  });

  return {
    Badge: MockContainer,
    Button: MockButton,
    Flex: MockContainer,
    Menu: MockMenu,
    Tooltip: MockContainer,
  };
});

jest.mock("@/hooks/use-terminology-display", () => ({
  useTerminology: () => ({
    getTermDisplay: () => "story",
  }),
}));

jest.mock("@/hooks/role", () => ({
  useUserRole: () => ({ userRole: "admin" }),
}));

jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/acme${path}`,
  }),
}));

jest.mock("@/modules/notifications/hooks/unread", () => ({
  useUnreadNotifications: () => ({ data: 3 }),
}));

jest.mock("@/components/ui/new-objective", () => ({
  NewObjectiveDialog: () => null,
}));

jest.mock("@/components/ui/new-story-dialog", () => ({
  NewStoryDialog: () => null,
}));

jest.mock("@/components/shared/keyboard-shortcuts", () => ({
  KeyboardShortcuts: ({ isOpen }: { isOpen: boolean }) =>
    isOpen ? <div>Keyboard shortcuts dialog</div> : null,
}));

jest.mock("./commands", () => ({
  Commands: () => <div>Search</div>,
}));

jest.mock("@/components/shared/sidebar/profile-menu", () => ({
  ProfileMenu: () => <div>Profile</div>,
}));

jest.mock("@/components/shared/sidebar/workspaces-menu", () => ({
  WorkspacesMenu: () => <div>Workspace</div>,
}));

jest.mock("@/components/shared/sidebar/sidebar-context", () => ({
  useSidebar: () => ({ isCollapsed: false }),
}));

jest.mock("@/components/shared/app-command-action-context", () => ({
  useCurrentAppCommandAction: () => null,
}));

describe("AppCommandBar", () => {
  it("shows Help with the existing command controls", () => {
    render(<AppCommandBar />);

    expect(screen.getByRole("button", { name: "Help" })).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "Notifications" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Create story" }),
    ).toBeInTheDocument();
    expect(screen.getByText("Profile")).toBeInTheDocument();
  });

  it("opens keyboard shortcuts from Help", () => {
    render(<AppCommandBar />);

    fireEvent.click(screen.getByRole("button", { name: "Keyboard shortcuts" }));

    expect(screen.getByText("Keyboard shortcuts dialog")).toBeInTheDocument();
  });
});
