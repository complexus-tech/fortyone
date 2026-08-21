/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ComponentPropsWithoutRef, ReactNode } from "react";
import { render, screen, waitFor } from "@testing-library/react";
import { toast } from "sonner";
import {
  useCalendarIntegration,
  useCreateCalendarConnectSession,
  useRevokeCalendarConnection,
  useSyncCalendarConnection,
} from "@/lib/hooks/calendar";
import { CalendarIntegrationSettings } from "./calendar-integration-settings";

let mockSearchParams = new URLSearchParams();

jest.mock("next/navigation", () => ({
  useSearchParams: () => mockSearchParams,
}));

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

jest.mock("@/components/ui", () => ({
  MicrosoftIcon: () => <svg aria-label="Microsoft" />,
}));

jest.mock("icons", () => ({
  CalendarIcon: (props: ComponentPropsWithoutRef<"svg">) => <svg {...props} />,
  CalendarPlusIcon: (props: ComponentPropsWithoutRef<"svg">) => (
    <svg {...props} />
  ),
  ClockIcon: (props: ComponentPropsWithoutRef<"svg">) => <svg {...props} />,
  GoogleCalendarIcon: (props: ComponentPropsWithoutRef<"svg">) => (
    <svg {...props} />
  ),
  MoreHorizontalIcon: (props: ComponentPropsWithoutRef<"svg">) => (
    <svg {...props} />
  ),
  ReloadIcon: (props: ComponentPropsWithoutRef<"svg">) => <svg {...props} />,
  UnlinkIcon: (props: ComponentPropsWithoutRef<"svg">) => <svg {...props} />,
}));

jest.mock("ui", () => {
  const passthrough = ({ children }: { children?: ReactNode }) => (
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
    Item: typeof passthrough;
    Items: typeof passthrough;
  };
  Menu.Button = passthrough;
  Menu.Group = passthrough;
  Menu.Item = passthrough;
  Menu.Items = passthrough;

  return {
    Badge: passthrough,
    Box: passthrough,
    Button: ({ children }: { children?: ReactNode }) => (
      <button type="button">{children}</button>
    ),
    Dialog,
    Flex: passthrough,
    Menu,
    Skeleton: passthrough,
    Text: ({
      as: Tag = "span",
      children,
    }: {
      as?: "h1" | "span";
      children: ReactNode;
    }) => <Tag>{children}</Tag>,
  };
});

jest.mock("@/lib/hooks/calendar", () => ({
  useCalendarIntegration: jest.fn(),
  useCreateCalendarConnectSession: jest.fn(),
  useRevokeCalendarConnection: jest.fn(),
  useSyncCalendarConnection: jest.fn(),
}));

jest.mock("sonner", () => ({
  toast: {
    error: jest.fn(),
    success: jest.fn(),
  },
}));

const mockUseCalendarIntegration = jest.mocked(useCalendarIntegration);
const mockUseCreateCalendarConnectSession = jest.mocked(
  useCreateCalendarConnectSession,
);
const mockUseRevokeCalendarConnection = jest.mocked(
  useRevokeCalendarConnection,
);
const mockUseSyncCalendarConnection = jest.mocked(useSyncCalendarConnection);

beforeEach(() => {
  jest.clearAllMocks();
  window.history.replaceState({}, "", "/acme/settings/account/calendar");
  mockSearchParams = new URLSearchParams();
  mockUseCalendarIntegration.mockReturnValue({
    data: { connections: [] },
    isError: false,
    isFetching: false,
    isPending: false,
    refetch: jest.fn(),
  } as unknown as ReturnType<typeof useCalendarIntegration>);
  mockUseCreateCalendarConnectSession.mockReturnValue({
    isPending: false,
    mutate: jest.fn(),
  } as unknown as ReturnType<typeof useCreateCalendarConnectSession>);
  mockUseRevokeCalendarConnection.mockReturnValue({
    isPending: false,
    mutate: jest.fn(),
  } as unknown as ReturnType<typeof useRevokeCalendarConnection>);
  mockUseSyncCalendarConnection.mockReturnValue({
    isPending: false,
    mutate: jest.fn(),
  } as unknown as ReturnType<typeof useSyncCalendarConnection>);
});

describe("CalendarIntegrationSettings callback feedback", () => {
  it("offers Google and Outlook calendar connections", () => {
    render(<CalendarIntegrationSettings />);

    expect(screen.getByText("Google Calendar")).toBeInTheDocument();
    expect(screen.getByText("Outlook Calendar")).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: "Connect" })).toHaveLength(2);
  });

  it("silently clears a successful connection callback from the URL", async () => {
    window.history.replaceState(
      {},
      "",
      "/acme/settings/account/calendar?connected=1&keep=1",
    );
    mockSearchParams = new URLSearchParams(window.location.search);

    render(<CalendarIntegrationSettings />);

    await waitFor(() => {
      expect(window.location.search).toBe("?keep=1");
    });
    expect(toast.success).not.toHaveBeenCalled();
    expect(toast.error).not.toHaveBeenCalled();
  });

  it("preserves connection errors while clearing callback parameters", async () => {
    window.history.replaceState(
      {},
      "",
      "/acme/settings/account/calendar?calendar_error=access_denied&keep=1",
    );
    mockSearchParams = new URLSearchParams(window.location.search);

    render(<CalendarIntegrationSettings />);

    await waitFor(() => {
      expect(window.location.search).toBe("?keep=1");
    });
    expect(toast.error).toHaveBeenCalledTimes(1);
    expect(toast.error).toHaveBeenCalledWith(
      "Google Calendar connection was cancelled.",
    );
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("attributes a Microsoft callback error to Outlook Calendar", async () => {
    window.history.replaceState(
      {},
      "",
      "/acme/settings/account/calendar?calendar_error=access_denied&calendar_provider=microsoft",
    );
    mockSearchParams = new URLSearchParams(window.location.search);

    render(<CalendarIntegrationSettings />);

    await waitFor(() => {
      expect(window.location.search).toBe("");
    });
    expect(toast.error).toHaveBeenCalledWith(
      "Outlook Calendar connection was cancelled.",
    );
  });
});
