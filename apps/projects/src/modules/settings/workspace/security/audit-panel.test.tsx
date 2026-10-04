import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import type { AuditEvent } from "./types";
import { listAudit, downloadAudit } from "./api";
import { SecurityAudit } from "./audit-panel";

const ACTOR_ID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa";
const RESOURCE_ID = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb";
const FIRST_EVENT: AuditEvent = {
  id: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
  source: "security",
  actorId: ACTOR_ID,
  actorType: "human_user",
  resourceType: "workspace_policy",
  resourceId: RESOURCE_ID,
  operation: "workspace.security_policy_updated",
  metadata: { count: 1 },
  createdAt: "2026-10-04T09:00:00Z",
};
const NEXT_EVENT: AuditEvent = {
  ...FIRST_EVENT,
  id: "dddddddd-dddd-4ddd-8ddd-dddddddddddd",
  resourceType: "browser_session",
  operation: "workspace.session_revoked",
  metadata: { reason: "Lost device" },
};

jest.mock("@/hooks/role", () => ({
  useUserRole: () => ({ userRole: "admin" }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: ACTOR_ID } } }),
}));
jest.mock("@/lib/hooks/members", () => ({
  useMembers: () => ({
    data: [
      {
        id: ACTOR_ID,
        username: "ada",
        fullName: "Ada Lovelace",
        email: "ada@example.com",
      },
    ],
  }),
}));
jest.mock("./api", () => ({ listAudit: jest.fn(), downloadAudit: jest.fn() }));

const list = jest.mocked(listAudit);
const download = jest.mocked(downloadAudit);
const renderAudit = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: Infinity } },
  });
  return render(
    <QueryClientProvider client={client}>
      <SecurityAudit />
    </QueryClientProvider>,
  );
};

const resizeObserverDescriptor = Object.getOwnPropertyDescriptor(
  globalThis,
  "ResizeObserver",
);
beforeAll(() => {
  Object.defineProperty(globalThis, "ResizeObserver", {
    configurable: true,
    value: class {
      observe() {}
      unobserve() {}
      disconnect() {}
    },
  });
});
afterAll(() => {
  if (resizeObserverDescriptor)
    Object.defineProperty(
      globalThis,
      "ResizeObserver",
      resizeObserverDescriptor,
    );
  else Reflect.deleteProperty(globalThis, "ResizeObserver");
});

const openFilters = async () => {
  const trigger = screen.getByRole("button", { name: /^Filters/ });
  fireEvent.click(trigger);
  return screen.findByRole("dialog", { name: "Audit log filters" });
};

const openDetails = async (label: string, event: AuditEvent) => {
  const trigger = screen.getByRole("button", {
    name: `Event actions for ${label} (${event.id.slice(0, 8)})`,
  });
  trigger.focus();
  fireEvent.keyDown(trigger, { key: "Enter", code: "Enter" });
  const item = await screen.findByRole("menuitem", { name: "Field Details" });
  item.focus();
  fireEvent.keyDown(item, { key: "Enter", code: "Enter" });
  await screen.findByRole("dialog", { name: "Audit event details" });
  return trigger;
};

beforeEach(() => {
  jest.clearAllMocks();
  list.mockImplementation(async (_ctx, filters, cursor) => ({
    items: cursor || filters.resourceType ? [NEXT_EVENT] : [FIRST_EVENT],
    nextCursor: cursor || filters.resourceType ? "" : "signed-next-cursor",
  }));
  download.mockResolvedValue(undefined);
});

it("uses cursor pages, preserves earlier pages, and exposes complete event details", async () => {
  renderAudit();
  expect(
    await screen.findByText("Security policy updated"),
  ).toBeInTheDocument();
  expect(screen.getByRole("button", { name: "Previous page" })).toBeDisabled();
  fireEvent.click(screen.getByRole("button", { name: "Next page" }));
  expect(await screen.findByText("Session revoked")).toBeInTheDocument();
  expect(screen.queryByText("Security policy updated")).not.toBeInTheDocument();
  expect(screen.getByText("Page 2 · 1 event")).toBeInTheDocument();
  expect(screen.getByRole("button", { name: "Next page" })).toBeDisabled();
  expect(list).toHaveBeenLastCalledWith(
    expect.anything(),
    {},
    "signed-next-cursor",
  );

  const trigger = await openDetails("Session revoked", NEXT_EVENT);
  const dialog = screen.getByRole("dialog", { name: "Audit event details" });
  expect(within(dialog).getByText(ACTOR_ID)).toBeInTheDocument();
  expect(within(dialog).getByText(RESOURCE_ID)).toBeInTheDocument();
  expect(within(dialog).getByText(NEXT_EVENT.id)).toBeInTheDocument();
  expect(
    within(dialog).getByText(/"reason": "Lost device"/),
  ).toBeInTheDocument();
  fireEvent.click(within(dialog).getByRole("button", { name: "Done" }));
  await waitFor(() => {
    expect(trigger).toHaveFocus();
  });
  fireEvent.click(screen.getByRole("button", { name: "Previous page" }));
  expect(
    await screen.findByText("Security policy updated"),
  ).toBeInTheDocument();
  expect(screen.queryByText("Session revoked")).not.toBeInTheDocument();
});

it("resets the cursor on filter changes and exports the complete applied filter", async () => {
  renderAudit();
  await screen.findByText("Security policy updated");
  fireEvent.click(screen.getByRole("button", { name: "Next page" }));
  await screen.findByText("Session revoked");
  expect(screen.queryByLabelText("Resource type")).not.toBeInTheDocument();
  await openFilters();
  fireEvent.change(screen.getByLabelText("Resource type"), {
    target: { value: "browser_session" },
  });
  fireEvent.change(screen.getByLabelText("Through date (UTC)"), {
    target: { value: "2026-10-04" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Apply filters" }));
  await waitFor(() => {
    expect(list).toHaveBeenLastCalledWith(
      expect.anything(),
      { resourceType: "browser_session", to: "2026-10-05T00:00:00.000Z" },
      "",
    );
  });
  await screen.findByText("Page 1 · 1 event");
  expect(
    screen.queryByRole("dialog", { name: "Audit log filters" }),
  ).not.toBeInTheDocument();
  fireEvent.click(screen.getByRole("button", { name: "Export filtered CSV" }));
  await waitFor(() => {
    expect(download).toHaveBeenCalledWith(expect.anything(), {
      resourceType: "browser_session",
      to: "2026-10-05T00:00:00.000Z",
    });
  });
  await openFilters();
  fireEvent.click(screen.getByRole("button", { name: "Clear filters" }));
  expect(
    await screen.findByText("Security policy updated"),
  ).toBeInTheDocument();
  await openFilters();
  expect(screen.getByLabelText("Resource type")).toHaveValue("");
  expect(screen.getByLabelText("Through date (UTC)")).toHaveValue("");
});

it("keeps invalid filter drafts open and leaves the applied query unchanged", async () => {
  renderAudit();
  await screen.findByText("Security policy updated");
  await openFilters();
  fireEvent.change(screen.getByLabelText("Resource ID"), {
    target: { value: "invalid-id" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Apply filters" }));
  expect(screen.getByRole("alert")).toHaveTextContent(
    "Use a valid resource ID",
  );
  expect(
    screen.getByRole("dialog", { name: "Audit log filters" }),
  ).toBeInTheDocument();
  expect(list).toHaveBeenCalledTimes(1);
  fireEvent.change(screen.getByLabelText("Resource ID"), {
    target: { value: RESOURCE_ID },
  });
  fireEvent.change(screen.getByLabelText("From date (UTC)"), {
    target: { value: "2026-10-05" },
  });
  fireEvent.change(screen.getByLabelText("Through date (UTC)"), {
    target: { value: "2026-10-04" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Apply filters" }));
  expect(list).toHaveBeenCalledTimes(1);
  fireEvent.change(screen.getByLabelText("From date (UTC)"), {
    target: { value: "2026-10-03" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Apply filters" }));
  await waitFor(() => {
    expect(list).toHaveBeenLastCalledWith(
      expect.anything(),
      {
        resourceId: RESOURCE_ID,
        from: "2026-10-03T00:00:00.000Z",
        to: "2026-10-05T00:00:00.000Z",
      },
      "",
    );
  });
});

it("returns keyboard focus to the event menu trigger after Escape from details", async () => {
  renderAudit();
  await screen.findByText("Security policy updated");
  const table = screen.getByRole("table", { name: /Workspace audit events/ });
  expect(within(table).getByText("ada")).toBeInTheDocument();
  expect(within(table).queryByText("Security")).not.toBeInTheDocument();
  const trigger = await openDetails("Security policy updated", FIRST_EVENT);
  const dialog = screen.getByRole("dialog", { name: "Audit event details" });
  expect(within(dialog).getByText("Human user")).toBeInTheDocument();
  fireEvent.keyDown(dialog, { key: "Escape", code: "Escape" });
  await waitFor(() => {
    expect(trigger).toHaveFocus();
  });
  expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
});

it("shows export failures outside the closed filters popover", async () => {
  download.mockRejectedValue(new Error("Export is unavailable"));
  renderAudit();
  await screen.findByText("Security policy updated");
  fireEvent.click(screen.getByRole("button", { name: "Export filtered CSV" }));
  expect(await screen.findByRole("alert")).toHaveTextContent(
    "Export is unavailable",
  );
  expect(
    screen.queryByRole("dialog", { name: "Audit log filters" }),
  ).not.toBeInTheDocument();
});

it("retains recovery navigation after a failed cursor page", async () => {
  list.mockImplementation(async (_ctx, _filters, cursor) => {
    if (cursor) throw new Error("Expired cursor");
    return { items: [FIRST_EVENT], nextCursor: "expired-cursor" };
  });
  renderAudit();
  await screen.findByText("Security policy updated");
  fireEvent.click(screen.getByRole("button", { name: "Next page" }));
  expect(await screen.findByRole("alert")).toHaveTextContent(
    "Audit events could not be loaded",
  );
  expect(screen.getByRole("button", { name: "Previous page" })).toBeEnabled();
  fireEvent.click(
    screen.getByRole("button", { name: "Back to newest events" }),
  );
  expect(
    await screen.findByText("Security policy updated"),
  ).toBeInTheDocument();
  expect(screen.getByText(/Page 1/)).toBeInTheDocument();
});
