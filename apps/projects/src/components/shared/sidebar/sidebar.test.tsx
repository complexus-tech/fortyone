/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type { HTMLAttributes, ReactNode } from "react";
import { Sidebar } from "./sidebar";

let mockUserRole = "admin";
let mockHasMeeting = true;
let mockTier = "free";
let mockSidebarCollapsed = false;
let mockAssistantCollapsed = false;

jest.mock("ui", () => {
  const MockBox = ({ children, ...props }: HTMLAttributes<HTMLDivElement>) => (
    <div {...props}>{children}</div>
  );
  const MockButton = ({
    children,
    href,
    onClick,
  }: {
    children: ReactNode;
    href?: string;
    onClick?: () => void;
  }) =>
    href ? (
      <a href={href}>{children}</a>
    ) : (
      <button onClick={onClick} type="button">
        {children}
      </button>
    );
  const MockFlex = ({ children }: { children: ReactNode }) => (
    <div>{children}</div>
  );
  const MockMenu = ({ children }: { children: ReactNode }) => (
    <div>{children}</div>
  );
  const MockMenuChild = ({ children }: { children: ReactNode }) => (
    <div>{children}</div>
  );
  MockMenu.Button = MockMenuChild;
  MockMenu.Group = MockMenuChild;
  MockMenu.Item = MockMenuChild;
  MockMenu.Items = MockMenuChild;
  const MockText = ({ children }: { children: ReactNode }) => (
    <span>{children}</span>
  );
  const MockTooltip = ({ children }: { children: ReactNode }) => (
    <div>{children}</div>
  );

  return {
    Box: MockBox,
    Button: MockButton,
    Flex: MockFlex,
    Menu: MockMenu,
    Text: MockText,
    Tooltip: MockTooltip,
  };
});

jest.mock("@/lib/hooks/subscription-features", () => ({
  useSubscriptionFeatures: () => ({
    tier: mockTier,
    trialDaysRemaining: 14,
  }),
}));

jest.mock("@/hooks", () => ({
  useLocalStorage: <T,>(_key: string, initialValue: T) => [
    initialValue,
    jest.fn(),
  ],
  useUserRole: () => ({ userRole: mockUserRole }),
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/acme${path}`,
  }),
}));

jest.mock("@/lib/hooks/workspaces", () => ({
  useCurrentWorkspace: () => ({ workspace: { deletedAt: null } }),
}));

jest.mock("@/components/ui", () => ({
  InviteMembersDialog: ({ isOpen }: { isOpen: boolean }) =>
    isOpen ? <div>Invite members dialog</div> : null,
}));

jest.mock("@/components/shared/keyboard-shortcuts", () => ({
  KeyboardShortcuts: () => null,
}));

jest.mock("./sidebar-context", () => ({
  useSidebar: () => ({
    isCollapsed: mockSidebarCollapsed,
    setIsCollapsed: jest.fn(),
    toggleSidebar: jest.fn(),
  }),
}));
jest.mock("./navigation", () => ({ Navigation: () => <div>Navigation</div> }));
jest.mock("./teams", () => ({ Teams: () => <div>Teams</div> }));
jest.mock("./upcoming-meeting-card", () => ({
  SidebarAssistantCards: ({
    fallback,
    isCollapsed,
  }: {
    fallback: ReactNode;
    isCollapsed?: boolean;
  }) => {
    mockAssistantCollapsed = Boolean(isCollapsed);
    return mockHasMeeting ? <div>Upcoming meeting</div> : fallback;
  },
}));

describe("Sidebar", () => {
  beforeEach(() => {
    mockHasMeeting = true;
    mockTier = "free";
    mockUserRole = "admin";
    mockSidebarCollapsed = false;
    mockAssistantCollapsed = false;
  });

  it("keeps navigation scrollable and the footer fixed", () => {
    render(<Sidebar />);

    const content = screen
      .getByText("Navigation")
      .closest("[data-sidebar-content]");
    const footer = screen
      .getByText("Upcoming meeting")
      .closest("[data-sidebar-footer]");

    expect(content).toHaveClass("min-h-0", "flex-1", "overflow-y-auto");
    expect(footer).toHaveClass("shrink-0");
    expect(content?.parentElement).toHaveClass("h-full", "overflow-hidden");
    expect(content?.parentElement).not.toHaveClass("border-r-[0.5px]");
  });

  it("replaces the footer actions with an upcoming meeting", () => {
    render(<Sidebar />);

    expect(screen.getByText("Upcoming meeting")).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Upgrade" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Upgrade" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Invite members" })).toBeNull();
    expect(screen.queryByText("You're on the free plan")).toBeNull();
    expect(screen.queryByText("Upgrade plan")).toBeNull();
  });

  it("passes the collapsed state to the sidebar assistant", () => {
    mockSidebarCollapsed = true;

    render(<Sidebar />);

    expect(mockAssistantCollapsed).toBe(true);
  });

  it("keeps the Upgrade action non-navigational for members", () => {
    mockHasMeeting = false;
    mockUserRole = "member";
    render(<Sidebar />);

    expect(screen.getByRole("button", { name: "Upgrade" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Upgrade" })).toBeNull();
  });

  it("keeps Upgrade in the footer action row when there is no meeting", () => {
    mockHasMeeting = false;
    render(<Sidebar />);

    expect(screen.queryByRole("button", { name: "Invite members" })).toBeNull();
    expect(screen.getByRole("link", { name: "Upgrade" })).toBeInTheDocument();
  });

  it("shows Invite members instead of Upgrade for paid-plan admins", () => {
    mockHasMeeting = false;
    mockTier = "pro";
    render(<Sidebar />);

    const inviteMembers = screen.getByRole("button", {
      name: "Invite members",
    });

    expect(inviteMembers).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Upgrade" })).toBeNull();

    fireEvent.click(inviteMembers);

    expect(screen.getByText("Invite members dialog")).toBeInTheDocument();
  });

  it("does not show Invite members to paid-plan non-admins", () => {
    mockHasMeeting = false;
    mockTier = "business";
    mockUserRole = "member";
    render(<Sidebar />);

    expect(screen.queryByRole("button", { name: "Invite members" })).toBeNull();
    expect(screen.queryByRole("link", { name: "Upgrade" })).toBeNull();
  });
});
