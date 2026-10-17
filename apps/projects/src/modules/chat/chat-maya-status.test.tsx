import { fireEvent, render, screen } from "@testing-library/react";
import { ChatMayaStatus } from "./chat-maya-status";
import type { ChatMayaRun } from "./use-chat-maya-runs";

const mockMutate = jest.fn();
let mockUserID = "requester";
let mockRun: ChatMayaRun;

jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: mockUserID } } }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "second" }),
}));
jest.mock("@/hooks", () => ({
  useTerminology: () => ({ getTermDisplay: () => "Tasks" }),
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/second${path}`,
  }),
}));
jest.mock("@/lib/hooks/statuses", () => ({
  useStatuses: () => ({ data: [] }),
}));
jest.mock("./use-chat-maya-runs", () => ({
  useChatMayaRuns: () => ({ data: { enabled: true, runs: [mockRun] } }),
  useChatMayaAction: () => ({ mutate: mockMutate, isPending: false }),
}));

beforeEach(() => {
  mockMutate.mockReset();
  mockUserID = "requester";
  mockRun = {
    id: "run",
    sourceId: "source",
    requesterId: "requester",
    status: "awaiting_approval",
    summary: "Review the proposed tasks",
    proposalReview: "Create: Verify onboarding keyboard navigation",
    version: 3,
    updatedAt: "2026-10-16T18:00:00Z",
  };
});

test("the requester reviews the proposal before approving its current version", () => {
  render(
    <ChatMayaStatus conversationId="conversation" sourceMessageId="source" />,
  );
  expect(screen.getByText(mockRun.proposalReview!)).toBeInTheDocument();
  fireEvent.click(screen.getByRole("button", { name: "Approve" }));
  expect(mockMutate).toHaveBeenCalledWith({ run: mockRun, action: "approve" });
});

test("other members can see the proposal but cannot approve or cancel it", () => {
  mockUserID = "another-member";
  render(
    <ChatMayaStatus conversationId="conversation" sourceMessageId="source" />,
  );
  expect(screen.getByText(mockRun.proposalReview!)).toBeInTheDocument();
  expect(
    screen.queryByRole("button", { name: "Approve" }),
  ).not.toBeInTheDocument();
  expect(
    screen.queryByRole("button", { name: "Cancel" }),
  ).not.toBeInTheDocument();
});

test("failed requests expose retry only to their requester", () => {
  mockRun = { ...mockRun, status: "failed", summary: "Please try again." };
  render(
    <ChatMayaStatus conversationId="conversation" sourceMessageId="source" />,
  );
  fireEvent.click(screen.getByRole("button", { name: "Retry" }));
  expect(mockMutate).toHaveBeenCalledWith({ run: mockRun, action: "retry" });
  expect(
    screen.queryByRole("button", { name: "Approve" }),
  ).not.toBeInTheDocument();
});

test("published task cards link to the created task in the current workspace", () => {
  mockRun = {
    ...mockRun,
    status: "succeeded",
    replyMessageId: "reply",
    tasks: [
      {
        id: "created-task",
        reference: "PRO-42",
        title: "Verify keyboard navigation",
        teamId: "team",
        priority: "High",
      },
    ],
  };
  render(
    <ChatMayaStatus conversationId="conversation" sourceMessageId="reply" />,
  );
  expect(
    screen.getByRole("link", { name: /Open Verify keyboard navigation/ }),
  ).toHaveAttribute("href", "/second/work/PRO-42");
  expect(
    screen.queryByRole("button", { name: "Approve" }),
  ).not.toBeInTheDocument();
});
