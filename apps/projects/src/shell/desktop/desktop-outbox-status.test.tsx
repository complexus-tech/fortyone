import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { DesktopOutboxStatus } from "./desktop-outbox-status";
import {
  resolveDesktopOperation,
  useDesktopOutbox,
} from "./desktop-outbox-coordinator";

jest.mock("./desktop-outbox-coordinator", () => ({
  useDesktopOutbox: jest.fn(),
  resolveDesktopOperation: jest.fn(),
  refreshDesktopOutbox: jest.fn(),
}));
const scope = {
  accountId: "account",
  workspaceId: "workspace",
  workspaceSlug: "engineering",
};
const operation = {
  id: "operation",
  kind: "story.update" as const,
  resourceId: "story",
  payload: { title: "My new title" },
  status: "conflict" as const,
  serverValue: { title: "Teammate's title" },
  serverVersion: "v2",
  attempted: true,
};
const show = (client = new QueryClient()) =>
  render(
    <QueryClientProvider client={client}>
      <DesktopOutboxStatus />
    </QueryClientProvider>,
  );

beforeEach(() => {
  jest.clearAllMocks();
  jest.mocked(useDesktopOutbox).mockReturnValue({
    scope,
    snapshot: { operations: [operation], completed: [] },
    busy: false,
    online: true,
  });
  jest.mocked(resolveDesktopOperation).mockResolvedValue(undefined);
});

it("requires explicit review of the current server value before applying a conflict", async () => {
  show();
  fireEvent.click(screen.getByRole("button", { name: "Review saved changes" }));
  expect(screen.getByText("Teammate's title")).toBeVisible();
  const apply = screen.getByRole("button", { name: "Apply reviewed changes" });
  expect(apply).toBeDisabled();
  fireEvent.click(screen.getByRole("checkbox"));
  fireEvent.click(apply);
  await waitFor(() => {
    expect(resolveDesktopOperation).toHaveBeenCalledWith(scope, {
      operationId: "operation",
      resolution: "retry",
      payload: { title: "My new title" },
      baseVersion: "v2",
    });
  });
});

it("requires a second confirmation before discarding a saved change", () => {
  show();
  fireEvent.click(screen.getByRole("button", { name: "Review saved changes" }));
  fireEvent.click(screen.getByRole("button", { name: "Discard local change" }));
  expect(resolveDesktopOperation).not.toHaveBeenCalled();
  fireEvent.click(screen.getByRole("button", { name: "Confirm discard" }));
  expect(resolveDesktopOperation).toHaveBeenCalledWith(scope, {
    operationId: "operation",
    resolution: "discard",
  });
});

it("does not allow an uncertain pending write to be discarded", () => {
  jest.mocked(useDesktopOutbox).mockReturnValue({
    scope,
    snapshot: {
      operations: [{ ...operation, status: "pending" }],
      completed: [],
    },
    busy: false,
    online: true,
  });
  show();
  fireEvent.click(screen.getByRole("button", { name: "Review saved changes" }));
  expect(
    screen.getByRole("button", { name: "Discard local change" }),
  ).toBeDisabled();
  expect(
    screen.getByText(/same operation so it cannot create a duplicate/),
  ).toBeVisible();
});

it("provides a settings review action without a pending count or live sync status", () => {
  show();
  expect(
    screen.getByRole("button", { name: "Review saved changes" }),
  ).toHaveTextContent("Review saved changes");
  expect(
    screen.queryByText(/1 change|Syncing|Sync paused/),
  ).not.toBeInTheDocument();
});

it("shows downloaded status names from this workspace during conflict review", () => {
  const client = new QueryClient();
  client.setQueryData(
    ["statuses", scope.workspaceSlug, "list"],
    [
      { id: "old-status", name: "In progress" },
      { id: "new-status", name: "Complete" },
    ],
  );
  client.setQueryData(
    ["statuses", "another-workspace", "list"],
    [{ id: "new-status", name: "Foreign status" }],
  );
  jest.mocked(useDesktopOutbox).mockReturnValue({
    scope,
    snapshot: {
      operations: [
        {
          ...operation,
          payload: { statusId: "new-status" },
          serverValue: { statusId: "old-status" },
        },
      ],
      completed: [],
    },
    busy: false,
    online: true,
  });
  show(client);
  fireEvent.click(screen.getByRole("button", { name: "Review saved changes" }));
  expect(screen.getByText("In progress")).toBeVisible();
  expect(screen.getByText("Complete")).toBeVisible();
  expect(screen.queryByText("Foreign status")).not.toBeInTheDocument();
});
