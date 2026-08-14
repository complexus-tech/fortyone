/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ComponentPropsWithoutRef, ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import {
  useCreateSlackInstallSession,
  useDisconnectSlackWorkspace,
  useResyncSlackChannels,
  useSlackAccountLinkToken,
  useSlackAgentSettings,
  useSlackIntegration,
  useUpdateSlackAgentSettings,
} from "@/lib/hooks/slack";
import { SlackIntegrationSettings } from "./slack-integration-settings";

jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/acme${path}`,
  }),
}));

jest.mock("@/modules/settings/components", () => ({
  SectionHeader: ({
    action,
    description,
    title,
  }: {
    action?: ReactNode;
    description: string;
    title: string;
  }) => (
    <header>
      <h2>{title}</h2>
      <p>{description}</p>
      {action}
    </header>
  ),
  SettingsBackButton: ({ href, label }: { href: string; label: string }) => (
    <a href={href}>{label}</a>
  ),
}));

jest.mock("./slack-channel-audience-settings", () => ({
  SlackChannelAudienceSettings: () => <div>Channel audience settings</div>,
}));

jest.mock("icons", () => ({
  MoreHorizontalIcon: (props: ComponentPropsWithoutRef<"svg">) => (
    <svg {...props} />
  ),
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

  const Menu = passthrough as typeof passthrough & {
    Button: typeof passthrough;
    Group: typeof passthrough;
    Item: (props: { children: ReactNode; onSelect?: () => void }) => ReactNode;
    Items: typeof passthrough;
  };
  Menu.Button = passthrough;
  Menu.Group = passthrough;
  Menu.Item = function MenuItem({ children }: { children: ReactNode }) {
    return <div>{children}</div>;
  };
  Menu.Items = passthrough;

  return {
    Badge: passthrough,
    Box: passthrough,
    Button: ({
      children,
      disabled,
      href,
      loading,
      onClick,
    }: {
      children?: ReactNode;
      disabled?: boolean;
      href?: string;
      loading?: boolean;
      onClick?: () => void;
    }) =>
      href ? (
        <a href={href}>{children}</a>
      ) : (
        <button disabled={disabled} onClick={onClick} type="button">
          {loading ? "Loading..." : children}
        </button>
      ),
    Dialog,
    Flex: passthrough,
    Menu,
    Text: ({
      as: Tag = "span",
      children,
    }: {
      as?: "h1" | "span";
      children: ReactNode;
    }) => <Tag>{children}</Tag>,
    TextArea: (props: ComponentPropsWithoutRef<"textarea">) => (
      <textarea {...props} />
    ),
  };
});

jest.mock("@/lib/hooks/slack", () => ({
  useCreateSlackInstallSession: jest.fn(),
  useDisconnectSlackWorkspace: jest.fn(),
  useResyncSlackChannels: jest.fn(),
  useSlackAccountLinkToken: jest.fn(),
  useSlackAgentSettings: jest.fn(),
  useSlackIntegration: jest.fn(),
  useUpdateSlackAgentSettings: jest.fn(),
}));

const mockUseCreateSlackInstallSession = jest.mocked(
  useCreateSlackInstallSession,
);
const mockUseDisconnectSlackWorkspace = jest.mocked(
  useDisconnectSlackWorkspace,
);
const mockUseResyncSlackChannels = jest.mocked(useResyncSlackChannels);
const mockUseSlackAccountLinkToken = jest.mocked(useSlackAccountLinkToken);
const mockUseSlackAgentSettings = jest.mocked(useSlackAgentSettings);
const mockUseSlackIntegration = jest.mocked(useSlackIntegration);
const mockUseUpdateSlackAgentSettings = jest.mocked(
  useUpdateSlackAgentSettings,
);

describe("SlackIntegrationSettings", () => {
  const createInstallSession = jest.fn();
  const resyncChannels = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    mockUseSlackIntegration.mockReturnValue({
      data: {
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
    mockUseCreateSlackInstallSession.mockReturnValue({
      isPending: false,
      mutate: createInstallSession,
    } as unknown as ReturnType<typeof useCreateSlackInstallSession>);
    mockUseResyncSlackChannels.mockReturnValue({
      isPending: false,
      mutate: resyncChannels,
    } as unknown as ReturnType<typeof useResyncSlackChannels>);
    mockUseDisconnectSlackWorkspace.mockReturnValue({
      isPending: false,
      mutate: jest.fn(),
    } as unknown as ReturnType<typeof useDisconnectSlackWorkspace>);
    mockUseSlackAccountLinkToken.mockReturnValue({
      errorMessage: null,
      hasToken: false,
      retry: undefined,
      status: "idle",
    });
    mockUseSlackAgentSettings.mockReturnValue({
      data: { guidance: "" },
      isLoading: false,
    } as unknown as ReturnType<typeof useSlackAgentSettings>);
    mockUseUpdateSlackAgentSettings.mockReturnValue({
      isPending: false,
      mutate: jest.fn(),
    } as unknown as ReturnType<typeof useUpdateSlackAgentSettings>);
  });

  it("resyncs channels instead of restarting OAuth for a connected workspace", () => {
    render(<SlackIntegrationSettings />);

    fireEvent.click(screen.getByRole("button", { name: "Resync channels" }));

    expect(resyncChannels).toHaveBeenCalledTimes(1);
    expect(createInstallSession).not.toHaveBeenCalled();
  });

  it("prevents a second channel sync while the first request is pending", () => {
    mockUseResyncSlackChannels.mockReturnValue({
      isPending: true,
      mutate: resyncChannels,
    } as unknown as ReturnType<typeof useResyncSlackChannels>);

    render(<SlackIntegrationSettings />);

    expect(screen.getByRole("button", { name: "Loading..." })).toBeDisabled();
  });

  it("keeps OAuth as the action for a workspace that is not connected", () => {
    mockUseSlackIntegration.mockReturnValue({
      data: { channels: [], slackWorkspace: null },
    } as unknown as ReturnType<typeof useSlackIntegration>);

    render(<SlackIntegrationSettings />);

    fireEvent.click(screen.getByRole("button", { name: "Connect Slack" }));

    expect(createInstallSession).toHaveBeenCalledTimes(1);
    expect(resyncChannels).not.toHaveBeenCalled();
  });
});
