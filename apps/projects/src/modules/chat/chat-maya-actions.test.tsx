import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { toast } from "sonner";
import { get, post } from "@/lib/http";
import type { ChatMayaRun } from "./use-chat-maya-runs";
import { ChatMayaStatus } from "./chat-maya-status";
import { chatKeys } from "./constants";

jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "requester" } } }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "second" }),
}));
jest.mock("@/lib/http", () => ({ get: jest.fn(), post: jest.fn() }));
jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));
jest.mock("./chat-task-results", () => ({ ChatTaskResults: () => null }));

const run: ChatMayaRun = {
  id: "run",
  sourceId: "source",
  requesterId: "requester",
  status: "awaiting_approval",
  summary: "Review the proposed tasks",
  proposalReview: "Create: Verify onboarding keyboard navigation",
  version: 3,
  updatedAt: "2026-10-16T18:00:00Z",
};
const runsKey = chatKeys.mayaRuns("second", "conversation");
const deferred = <T,>() => {
  let resolve!: (value: T) => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<T>((yes, no) => {
    resolve = yes;
    reject = no;
  });
  return { promise, resolve, reject };
};
const clients: QueryClient[] = [];

const renderCopies = (source: ChatMayaRun) => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  clients.push(client);
  const another = { ...source, id: "another-run", sourceId: "another-source" };
  const cached = { enabled: true, runs: [source, another] };
  client.setQueryData(runsKey, cached);
  jest.mocked(get).mockResolvedValue({ data: cached });
  render(
    <QueryClientProvider client={client}>
      <section aria-label="Conversation">
        <ChatMayaStatus
          conversationId="conversation"
          sourceMessageId={source.sourceId}
        />
      </section>
      <section aria-label="Thread">
        <ChatMayaStatus
          conversationId="conversation"
          sourceMessageId={source.sourceId}
        />
      </section>
      <section aria-label="Another request">
        <ChatMayaStatus
          conversationId="conversation"
          sourceMessageId="another-source"
        />
      </section>
    </QueryClientProvider>,
  );
  return { client, cached };
};

beforeEach(() => jest.clearAllMocks());
afterEach(() => {
  for (const client of clients) client.clear();
  clients.length = 0;
});

it.each([
  {
    action: "approve",
    label: "Approve",
    pending: "Approving…",
    status: "awaiting_approval",
    next: "queued",
  },
  {
    action: "cancel",
    label: "Cancel",
    pending: "Cancelling…",
    status: "awaiting_approval",
    next: "cancelled",
  },
  {
    action: "retry",
    label: "Retry",
    pending: "Retrying…",
    status: "failed",
    next: "queued",
  },
] as const)(
  "shares $pending feedback and prevents duplicate $action requests",
  async ({ action, label, pending, status, next }) => {
    const response = deferred<{ data: ChatMayaRun }>();
    const refresh = deferred<{
      data: { enabled: boolean; runs: ChatMayaRun[] };
    }>();
    jest.mocked(post).mockReturnValue(response.promise);
    const source = { ...run, status };
    const { client, cached } = renderCopies(source);
    jest.mocked(get).mockReturnValue(refresh.promise);
    const conversation = within(
      screen.getByRole("region", { name: "Conversation" }),
    );
    const thread = within(screen.getByRole("region", { name: "Thread" }));
    const first = conversation.getByRole("button", { name: label });
    const second = thread.getByRole("button", { name: label });
    act(() => {
      fireEvent.click(first);
      fireEvent.click(second);
    });
    await waitFor(() => {
      expect(
        conversation.getByRole("button", { name: pending }),
      ).toBeDisabled();
      expect(thread.getByRole("button", { name: pending })).toBeDisabled();
    });
    expect(
      within(screen.getByRole("region", { name: "Another request" })).getByRole(
        "button",
        { name: label },
      ),
    ).toBeEnabled();
    expect(post).toHaveBeenCalledTimes(1);
    expect(post).toHaveBeenCalledWith(
      `chat/conversations/conversation/maya-runs/run/${action}`,
      { version: run.version },
      expect.objectContaining({ workspaceSlug: "second" }),
    );
    expect(client.getQueryData(runsKey)).toEqual(cached);
    const confirmed: ChatMayaRun = { ...source, status: next, version: 4 };
    await act(async () => {
      response.resolve({ data: confirmed });
    });
    await waitFor(() => {
      expect(client.getQueryData(runsKey)).toEqual({
        ...cached,
        runs: [confirmed, cached.runs[1]],
      });
      expect(get).toHaveBeenCalledTimes(1);
    });
    // The response is visible even while the authoritative refresh is pending.
    expect(
      conversation.queryByRole("button", { name: pending }),
    ).not.toBeInTheDocument();
    await act(async () => {
      refresh.resolve({
        data: { ...cached, runs: [confirmed, cached.runs[1]] },
      });
    });
  },
);

it("keeps a failed approval recoverable in every copy and surfaces the error", async () => {
  const response = deferred<never>();
  jest.mocked(post).mockReturnValue(response.promise);
  const { client, cached } = renderCopies(run);
  fireEvent.click(
    within(screen.getByRole("region", { name: "Conversation" })).getByRole(
      "button",
      { name: "Approve" },
    ),
  );
  await waitFor(() => {
    expect(screen.getAllByRole("button", { name: "Approving…" })).toHaveLength(
      2,
    );
  });
  await act(async () => {
    response.reject(new Error("The proposal changed. Please review it again."));
  });
  await waitFor(() => {
    for (const section of ["Conversation", "Thread"]) {
      const copy = within(screen.getByRole("region", { name: section }));
      expect(copy.getByRole("button", { name: "Approve" })).toBeEnabled();
      expect(copy.getByText(run.proposalReview!)).toBeVisible();
    }
  });
  expect(client.getQueryData(runsKey)).toEqual(cached);
  expect(toast.error).toHaveBeenCalledWith("Could not update Maya’s request", {
    description: "The proposal changed. Please review it again.",
  });
});
