/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ComponentPropsWithoutRef, ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import {
  useCreateSlackAccountLinkSession,
  useDisconnectSlackAccount,
  useSlackAccountLinkToken,
  useSlackIntegration,
} from "@/lib/hooks/slack";
import { SlackAccountLinkSettings } from "./slack-account-link-settings";

jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/acme${path}`,
  }),
}));

jest.mock("@/modules/settings/components", () => ({
  SettingsBackButton: ({ href, label }: { href: string; label: string }) => (
    <a href={href}>{label}</a>
  ),
}));

jest.mock("icons", () => ({
  SlackIcon: (props: ComponentPropsWithoutRef<"svg">) => <svg {...props} />,
  UnlinkIcon: (props: ComponentPropsWithoutRef<"svg">) => <svg {...props} />,
}));

jest.mock("ui", () => {
  const passthrough = ({ children }: { children: ReactNode }) => (
    <div>{children}</div>
  );
  const Dialog = passthrough as typeof passthrough & {
    Body: typeof passthrough;
    Content: typeof passthrough;
    Footer: typeof passthrough;
    Header: typeof passthrough;
    Title: typeof passthrough;
  };
  Dialog.Body = passthrough;
  Dialog.Content = passthrough;
  Dialog.Footer = passthrough;
  Dialog.Header = passthrough;
  Dialog.Title = passthrough;

  return {
    Box: passthrough,
    Button: ({
      children,
      loading,
      onClick,
    }: {
      children: ReactNode;
      loading?: boolean;
      onClick?: () => void;
    }) => (
      <button onClick={onClick} type="button">
        {loading ? "Loading..." : children}
      </button>
    ),
    Dialog,
    Flex: passthrough,
    Text: ({ children }: { children: ReactNode }) => <span>{children}</span>,
  };
});

jest.mock("@/lib/hooks/slack", () => ({
  useCreateSlackAccountLinkSession: jest.fn(),
  useDisconnectSlackAccount: jest.fn(),
  useSlackAccountLinkToken: jest.fn(),
  useSlackIntegration: jest.fn(),
}));

const mockUseCreateSlackAccountLinkSession = jest.mocked(
  useCreateSlackAccountLinkSession,
);
const mockUseDisconnectSlackAccount = jest.mocked(useDisconnectSlackAccount);
const mockUseSlackAccountLinkToken = jest.mocked(useSlackAccountLinkToken);
const mockUseSlackIntegration = jest.mocked(useSlackIntegration);

describe("SlackAccountLinkSettings", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockUseCreateSlackAccountLinkSession.mockReturnValue({
      isPending: false,
      mutate: jest.fn(),
    } as unknown as ReturnType<typeof useCreateSlackAccountLinkSession>);
    mockUseDisconnectSlackAccount.mockReturnValue({
      isPending: false,
      mutate: jest.fn(),
    } as unknown as ReturnType<typeof useDisconnectSlackAccount>);
    mockUseSlackAccountLinkToken.mockReturnValue({
      errorMessage: null,
      hasToken: false,
      retry: undefined,
      status: "idle",
    });
  });

  it("shows the current member's connected Slack account and disconnect action", () => {
    mockUseSlackIntegration.mockReturnValue({
      data: {
        accountLink: {
          linkedAt: "2026-08-26T08:00:00Z",
          linkedVia: "dashboard_oauth",
          slackUserId: "U123",
        },
        channels: [],
        slackWorkspace: {
          createdAt: "2026-08-14T09:00:00Z",
          id: "installation-1",
          isActive: true,
          slackTeamDomain: "acme",
          slackTeamId: "T123",
          slackTeamName: "Acme",
          updatedAt: "2026-08-14T09:00:00Z",
        },
      },
    } as unknown as ReturnType<typeof useSlackIntegration>);

    render(<SlackAccountLinkSettings />);

    expect(screen.getByText("Slack account connected")).toBeInTheDocument();
    expect(
      screen.getByText("Your account is already connected with Slack."),
    ).toBeInTheDocument();
    expect(screen.getByText("Acme")).toBeInTheDocument();
    expect(screen.getByText("U123")).toBeInTheDocument();
    expect(screen.getByText("Slack authorization")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Disconnect account" }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Connect Slack account" }),
    ).not.toBeInTheDocument();
  });

  it("does not offer a second connection while an already-linked response refreshes", () => {
    mockUseSlackIntegration.mockReturnValue({
      data: { accountLink: null, channels: [], slackWorkspace: null },
    } as unknown as ReturnType<typeof useSlackIntegration>);
    mockUseSlackAccountLinkToken.mockReturnValue({
      errorMessage: null,
      hasToken: true,
      retry: undefined,
      status: "already_connected",
    });

    render(<SlackAccountLinkSettings />);

    expect(
      screen.getByText("Your account is already connected with Slack."),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Connect Slack account" }),
    ).not.toBeInTheDocument();
  });
});
