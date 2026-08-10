/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import { IntegrationRequestBanner } from "./integration-request-banner";

jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/workspace${path}`,
  }),
}));

jest.mock("@/modules/integration-requests/thread-activity", () => ({
  IntegrationRequestThreadActivity: ({ requestId }: { requestId: string }) => (
    <div data-testid="thread-activity">Thread activity for {requestId}</div>
  ),
}));

const link = {
  id: "thread-1",
  integrationRequestId: "request-1",
  teamId: "team-1",
  provider: "slack" as const,
  externalChannelId: "channel-1",
  externalThreadId: "thread-1",
  sourceUrl: "https://slack.com/archives/channel-1/pthread-1",
  requestTitle: "Improve the request flow",
  createdAt: "2026-08-10T10:00:00.000Z",
  updatedAt: "2026-08-10T10:00:00.000Z",
};

describe("IntegrationRequestBanner", () => {
  it("opens the thread by default and toggles from the banner while retaining request actions", () => {
    render(<IntegrationRequestBanner links={[link]} />);

    const trigger = screen.getByRole("button", {
      name: "Toggle Slack conversation for Improve the request flow",
    });
    const activity = screen.getByTestId("thread-activity");
    const content = activity.closest('[data-slot="collapsible-content"]');

    expect(content).not.toBeNull();
    expect(trigger).toHaveAttribute("aria-expanded", "true");
    expect(content).toHaveAttribute("data-state", "open");
    expect(screen.queryByText("Slack thread")).not.toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "Open Slack request" }),
    ).toHaveAttribute("href", "/workspace/teams/team-1/requests/request-1");
    expect(
      screen.getByRole("button", { name: "More Slack request actions" }),
    ).toBeInTheDocument();

    fireEvent.click(trigger);

    expect(trigger).toHaveAttribute("aria-expanded", "false");
    expect(content).toHaveAttribute("data-state", "closed");
  });
});
