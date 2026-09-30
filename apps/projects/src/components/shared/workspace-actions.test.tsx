/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { HTMLAttributes, ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { WorkspaceActions } from "./workspace-actions";

let mockUserRole: "admin" | "member" | "guest" | undefined = "admin";
let mockTier = "trial";
let mockTrialDaysRemaining = 14;
let mockSubscriptionPending = false;
let mockSubscriptionError = false;
let mockSidebarCollapsed = true;
let mockHasAssistantCards = false;
let mockWorkspace: { deletedAt: string | null } | undefined = {
  deletedAt: null,
};

jest.mock("ui", () => {
  const MockButton = ({
    children,
    className,
    href,
    leftIcon,
    onClick,
    fullWidth,
    "aria-label": ariaLabel,
    "data-invite-button": inviteButton,
  }: {
    children: ReactNode;
    className?: string;
    href?: string;
    leftIcon?: ReactNode;
    onClick?: () => void;
    fullWidth?: boolean;
    "aria-label"?: string;
    "data-invite-button"?: boolean;
  }) =>
    href ? (
      <a className={className} href={href}>
        {children}
      </a>
    ) : (
      <button
        aria-label={ariaLabel}
        className={className}
        data-full-width={fullWidth ? "true" : undefined}
        data-invite-button={inviteButton ? "true" : undefined}
        onClick={onClick}
        type="button"
      >
        {leftIcon}
        {children}
      </button>
    );
  const MockFlex = ({
    children,
    direction,
    align,
    gap,
    ...props
  }: HTMLAttributes<HTMLDivElement> & {
    direction?: string;
    align?: string;
    gap?: number;
  }) => (
    <div
      data-align={align}
      data-direction={direction}
      data-gap={gap}
      {...props}
    >
      {children}
    </div>
  );
  const MockTooltip = ({
    children,
    title,
  }: {
    children: ReactNode;
    title?: string | null;
  }) => <div title={title ?? undefined}>{children}</div>;

  return { Button: MockButton, Flex: MockFlex, Tooltip: MockTooltip };
});

jest.mock("icons", () => ({
  InviteMembersIcon: () => <span aria-hidden />,
}));
jest.mock("@/hooks/role", () => ({
  useUserRole: () => ({ userRole: mockUserRole }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/acme${path}`,
  }),
}));
jest.mock("@/lib/hooks/workspaces", () => ({
  useCurrentWorkspace: () => ({ workspace: mockWorkspace }),
}));
jest.mock("@/lib/hooks/subscription-features", () => ({
  useSubscriptionFeatures: () => ({
    tier: mockTier,
    trialDaysRemaining: mockTrialDaysRemaining,
  }),
}));
jest.mock("@/lib/hooks/subscriptions/subscription", () => ({
  useSubscription: () => ({
    isPending: mockSubscriptionPending,
    isError: mockSubscriptionError,
  }),
}));
jest.mock("@/components/ui/invite-members", () => ({
  InviteMembersDialog: ({
    isOpen,
    setIsOpen,
  }: {
    isOpen: boolean;
    setIsOpen: (open: boolean) => void;
  }) =>
    isOpen ? (
      <div aria-label="Invite people" role="dialog">
        <button
          onClick={() => {
            setIsOpen(false);
          }}
          type="button"
        >
          Close invitation dialog
        </button>
      </div>
    ) : null,
}));

const TopbarActions = () => (
  <WorkspaceActions
    sidebarHasActions={!mockSidebarCollapsed && !mockHasAssistantCards}
  />
);

describe("WorkspaceActions", () => {
  beforeEach(() => {
    mockUserRole = "admin";
    mockTier = "trial";
    mockTrialDaysRemaining = 14;
    mockSubscriptionPending = false;
    mockSubscriptionError = false;
    mockSidebarCollapsed = true;
    mockHasAssistantCards = false;
    mockWorkspace = { deletedAt: null };
  });

  it.each(["free", "trial", "pro", "business", "enterprise"])(
    "keeps Invite people available to admins on the %s tier",
    (tier) => {
      mockTier = tier;
      render(<TopbarActions />);

      expect(
        screen.getByRole("button", { name: "Invite people" }),
      ).toHaveAttribute("data-invite-button");
    },
  );

  it.each(["member", "guest", undefined] as const)(
    "does not offer invitation controls to a %s user",
    (role) => {
      mockUserRole = role;
      render(<TopbarActions />);

      expect(
        screen.queryByRole("button", { name: "Invite people" }),
      ).not.toBeInTheDocument();
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    },
  );

  it("opens and closes the existing invitation dialog", () => {
    render(<TopbarActions />);

    fireEvent.click(screen.getByRole("button", { name: "Invite people" }));
    expect(
      screen.getByRole("dialog", { name: "Invite people" }),
    ).toBeInTheDocument();

    fireEvent.click(
      screen.getByRole("button", { name: "Close invitation dialog" }),
    );
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it.each([
    [1, "1 day left in trial"],
    [14, "14 days left in trial"],
  ])(
    "shows %s remaining trial days as a mobile sidebar billing link",
    (days, label) => {
      mockTrialDaysRemaining = days;
      render(<WorkspaceActions variant="mobile" />);

      expect(screen.getByRole("link", { name: label })).toHaveAttribute(
        "href",
        "/acme/settings/workspace/billing",
      );
    },
  );

  it("offers Upgrade to free-plan admins independently of invitations", () => {
    mockTier = "free";
    render(<TopbarActions />);

    expect(screen.getByRole("link", { name: "Upgrade" })).toHaveAttribute(
      "href",
      "/acme/settings/workspace/billing",
    );
    expect(
      screen.getByRole("button", { name: "Invite people" }),
    ).toBeInTheDocument();
  });

  it.each(["member", "guest"] as const)(
    "shows informational trial status in the mobile sidebar to a %s without a billing control",
    (role) => {
      mockUserRole = role;
      render(<WorkspaceActions variant="mobile" />);

      expect(screen.getByText("14 days left in trial")).toBeInTheDocument();
      expect(screen.queryByRole("link")).not.toBeInTheDocument();
      expect(screen.queryByRole("button")).not.toBeInTheDocument();
    },
  );

  it("shows Free plan as information for non-admins", () => {
    mockTier = "free";
    mockUserRole = "member";
    render(<TopbarActions />);

    expect(screen.getByText("Free plan")).toBeInTheDocument();
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
    expect(screen.queryByRole("link")).not.toBeInTheDocument();
  });

  it.each(["pro", "business", "enterprise"])(
    "hides subscription status on the %s tier",
    (tier) => {
      mockTier = tier;
      render(<TopbarActions />);

      expect(screen.queryByRole("link")).not.toBeInTheDocument();
      expect(screen.queryByText(/left in trial/)).not.toBeInTheDocument();
      expect(screen.queryByText("Free plan")).not.toBeInTheDocument();
    },
  );

  it("renders no workspace actions for a paid-plan non-admin", () => {
    mockTier = "pro";
    mockUserRole = "member";
    const { container } = render(<TopbarActions />);

    expect(container).toBeEmptyDOMElement();
  });

  it.each(["loading", "error"])(
    "hides potentially incorrect plan status during subscription %s",
    (state) => {
      mockTier = "free";
      mockSubscriptionPending = state === "loading";
      mockSubscriptionError = state === "error";
      render(<TopbarActions />);

      expect(screen.queryByText("Upgrade")).not.toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: "Invite people" }),
      ).toBeInTheDocument();
    },
  );

  it.each([undefined, { deletedAt: "2026-09-30T00:00:00Z" }])(
    "hides actions for a missing or deleted workspace",
    (workspace) => {
      mockWorkspace = workspace;
      const { container } = render(<TopbarActions />);

      expect(container).toBeEmptyDOMElement();
    },
  );

  it("provides a full-width invitation action in the mobile menu", () => {
    mockSidebarCollapsed = false;
    render(<WorkspaceActions variant="mobile" />);

    const invite = screen.getByRole("button", { name: "Invite people" });
    expect(invite).toHaveAttribute("data-full-width", "true");
    expect(invite.closest("[data-workspace-actions]")).toHaveClass("flex-col");
    fireEvent.click(invite);
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "14 days left in trial" }),
    ).toBeInTheDocument();
  });

  it.each(["free", "trial"])(
    "keeps invitations in the topbar while %s plan status is in the expanded sidebar",
    (tier) => {
      mockSidebarCollapsed = false;
      mockTier = tier;
      render(<TopbarActions />);

      expect(
        screen.getByRole("button", { name: "Invite people" }),
      ).toBeInTheDocument();
      expect(screen.queryByRole("link")).not.toBeInTheDocument();
    },
  );

  it("keeps paid-plan invitations in the expanded sidebar without duplicating them", () => {
    mockSidebarCollapsed = false;
    mockTier = "pro";
    const { container } = render(<TopbarActions />);

    expect(container).toBeEmptyDOMElement();
  });

  it.each(["sidebar expands", "last assistant card disappears"])(
    "keeps an open invitation dialog when the %s",
    (change) => {
      mockTier = "pro";
      mockSidebarCollapsed = change === "sidebar expands";
      mockHasAssistantCards = change === "last assistant card disappears";
      const { rerender } = render(<TopbarActions />);
      fireEvent.click(screen.getByRole("button", { name: "Invite people" }));

      mockSidebarCollapsed = false;
      mockHasAssistantCards = false;
      rerender(<TopbarActions />);

      expect(
        screen.getByRole("dialog", { name: "Invite people" }),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: "Invite people" }),
      ).not.toBeInTheDocument();
      fireEvent.click(
        screen.getByRole("button", { name: "Close invitation dialog" }),
      );
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    },
  );

  it("shows paid-plan invitations in the topbar when assistant cards occupy the sidebar", () => {
    mockSidebarCollapsed = false;
    mockHasAssistantCards = true;
    mockTier = "pro";
    render(<TopbarActions />);

    expect(
      screen.getByRole("button", { name: "Invite people" }),
    ).toBeInTheDocument();
  });

  it("shows free-plan status in the topbar when assistant cards occupy the sidebar", () => {
    mockTier = "free";
    mockSidebarCollapsed = false;
    mockHasAssistantCards = true;
    render(<TopbarActions />);

    expect(screen.getByRole("link", { name: "Upgrade" })).toBeInTheDocument();
  });

  it.each([
    [false, false],
    [false, true],
    [true, false],
    [true, true],
  ])(
    "keeps trial status off the topbar for collapsed=%s and assistant cards=%s",
    (collapsed, hasCards) => {
      mockSidebarCollapsed = collapsed;
      mockHasAssistantCards = hasCards;
      render(<TopbarActions />);

      expect(screen.queryByText(/left in trial/)).not.toBeInTheDocument();
      expect(screen.queryByRole("link")).not.toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: "Invite people" }),
      ).toBeInTheDocument();
    },
  );

  it("does not duplicate paid-plan sidebar invitations when subscription refetching fails", () => {
    mockSidebarCollapsed = false;
    mockTier = "pro";
    mockSubscriptionError = true;
    const { container } = render(<TopbarActions />);

    expect(container).toBeEmptyDOMElement();
  });

  it("keeps paid-plan invitations available in the mobile menu regardless of desktop sidebar state", () => {
    mockSidebarCollapsed = false;
    mockTier = "pro";
    render(<WorkspaceActions variant="mobile" />);

    expect(
      screen.getByRole("button", { name: "Invite people" }),
    ).toBeInTheDocument();
  });
});
