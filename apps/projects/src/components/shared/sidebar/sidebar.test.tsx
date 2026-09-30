/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type { HTMLAttributes, ReactNode } from "react";
import { Sidebar } from "./sidebar";

let mockUserRole = "admin";
let mockHasMeeting = true;
let mockTier = "free";
let mockTrialDaysRemaining = 14;
let mockWorkspaceDeletedAt: string | null = null;
let mockSidebarCollapsed = false;
let mockAssistantCollapsed = false;

jest.mock("ui", () => {
  const MockBox = ({ children, ...props }: HTMLAttributes<HTMLDivElement>) => (
    <div {...props}>{children}</div>
  );
  const MockButton = ({
    children,
    className,
    href,
    onClick,
  }: {
    children: ReactNode;
    className?: string;
    href?: string;
    onClick?: () => void;
  }) =>
    href ? (
      <a className={className} href={href}>
        {children}
      </a>
    ) : (
      <button className={className} onClick={onClick} type="button">
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
  const MockTooltip = ({
    children,
    title,
  }: {
    children: ReactNode;
    title: string;
  }) => <div title={title}>{children}</div>;

  return {
    Box: MockBox,
    Button: MockButton,
    Flex: MockFlex,
    Menu: MockMenu,
    Text: MockText,
    Tooltip: MockTooltip,
  };
});

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
  useCurrentWorkspace: () => ({
    workspace: { deletedAt: mockWorkspaceDeletedAt },
  }),
}));

jest.mock("@/lib/hooks/subscription-features", () => ({
  useSubscriptionFeatures: () => ({
    tier: mockTier,
    trialDaysRemaining: mockTrialDaysRemaining,
  }),
}));

jest.mock("@/components/ui", () => ({
  InviteMembersDialog: ({ isOpen }: { isOpen: boolean }) =>
    isOpen ? <div>Invite members dialog</div> : null,
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
    mockTrialDaysRemaining = 14;
    mockWorkspaceDeletedAt = null;
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

  it("keeps labelled workspace actions out of the collapsed sidebar", () => {
    mockHasMeeting = false;
    mockSidebarCollapsed = true;

    render(<Sidebar />);

    expect(screen.queryByRole("link", { name: "Upgrade" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Invite members" })).toBeNull();
  });

  it("keeps the Upgrade action non-navigational for members", () => {
    mockHasMeeting = false;
    mockUserRole = "member";
    render(<Sidebar />);

    expect(screen.getByRole("button", { name: "Upgrade" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Upgrade" })).toBeNull();
  });

  it("keeps Upgrade in the sidebar when there is no assistant card", () => {
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

  it.each([
    [false, false],
    [false, true],
    [true, false],
    [true, true],
  ])(
    "keeps the trial countdown visible when collapsed is %s and assistant cards are %s",
    (isCollapsed, hasMeeting) => {
      mockTier = "trial";
      mockTrialDaysRemaining = 32;
      mockSidebarCollapsed = isCollapsed;
      mockHasMeeting = hasMeeting;

      render(<Sidebar />);

      const trialLink = screen.getByRole("link", {
        name: "32 days left in trial",
      });
      expect(trialLink).toHaveAttribute(
        "href",
        "/acme/settings/workspace/billing",
      );
      expect(trialLink.closest("[data-sidebar-trial]")).not.toBeNull();
      expect(trialLink.closest("[data-sidebar-footer]")).toHaveClass(
        "shrink-0",
      );
      expect(
        screen.getByText(isCollapsed ? "32d left" : "32 days left in trial"),
      ).toBeInTheDocument();
      expect(trialLink.closest("[title]")).toHaveAttribute(
        "title",
        "32 days left in your trial. Upgrade to a paid plan to get more premium features.",
      );
      if (isCollapsed) {
        expect(trialLink).toHaveClass(
          "max-w-full",
          "min-w-0",
          "truncate",
          "px-1",
        );
        expect(screen.getByText("32d left")).toHaveClass("min-w-0", "truncate");
      }
      expect(screen.queryByText("Upcoming meeting") !== null).toBe(hasMeeting);
    },
  );

  it("uses the singular trial label for the last day", () => {
    mockTier = "trial";
    mockTrialDaysRemaining = 1;

    render(<Sidebar />);

    const trialLink = screen.getByRole("link", { name: "1 day left in trial" });
    expect(trialLink).toHaveAttribute(
      "href",
      "/acme/settings/workspace/billing",
    );
    expect(trialLink.closest("[title]")).toHaveAttribute(
      "title",
      "1 day left in your trial. Upgrade to a paid plan to get more premium features.",
    );
  });

  it("keeps the trial countdown informational for non-admins", () => {
    mockTier = "trial";
    mockUserRole = "member";
    mockSidebarCollapsed = true;

    render(<Sidebar />);

    const trialButton = screen.getByRole("button", {
      name: "14 days left in trial",
    });
    expect(trialButton.closest("[title]")).toHaveAttribute(
      "title",
      "14 days left in your trial. Ask your admin to upgrade to a paid plan to get more premium features.",
    );
    expect(
      screen.queryByRole("link", { name: "14 days left in trial" }),
    ).toBeNull();
  });

  it("keeps the deletion warning and restore action instead of trial billing", () => {
    mockTier = "trial";
    mockWorkspaceDeletedAt = new Date().toISOString();

    render(<Sidebar />);

    expect(
      screen.getByText("Workspace scheduled for deletion"),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "Restore workspace" }),
    ).toHaveAttribute("href", "/acme/settings");
    expect(screen.queryByText("Upcoming meeting")).toBeNull();
    expect(
      screen.queryByRole("link", { name: "14 days left in trial" }),
    ).toBeNull();
  });
});
