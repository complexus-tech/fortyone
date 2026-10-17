import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  within,
} from "@testing-library/react";
import { TypingIndicator } from "./typing-indicator";
import { applyChatTypingEvent, clearChatTyping } from "./typing";
import type { ChatMessage } from "./types";
import type { ChatMayaRun } from "./use-chat-maya-runs";

let mockRuns: ChatMayaRun[] = [];
const mockMutate = jest.fn();
jest.mock("@/lib/hooks/workspaces", () => ({
  useCurrentWorkspace: () => ({ workspace: { id: "workspace" } }),
}));
jest.mock("./chat-context", () => ({
  useChatIdentity: () => ({ userId: "me" }),
}));
jest.mock("./use-chat-maya-runs", () => ({
  useChatMayaRuns: () => ({ data: { enabled: true, runs: mockRuns } }),
  useChatMayaAction: () => ({ mutate: mockMutate, isPending: false }),
}));

const run = (overrides: Partial<ChatMayaRun> = {}): ChatMayaRun => ({
  id: "run",
  sourceId: "root",
  threadRootId: "root",
  requesterId: "me",
  status: "running",
  summary: "",
  version: 1,
  updatedAt: "2026-10-17T10:00:00Z",
  ...overrides,
});

const Indicators = ({
  openThread,
  messages = [],
  replies = [],
}: {
  openThread?: string;
  messages?: ChatMessage[];
  replies?: ChatMessage[];
}) => (
  <>
    <section aria-label="Main composer">
      <TypingIndicator
        activeThreadRootId={openThread}
        conversationId="room"
        messages={messages}
      />
    </section>
    {openThread ? (
      <section aria-label="Thread composer">
        <TypingIndicator
          conversationId="room"
          messages={replies}
          threadRootId={openThread}
        />
      </section>
    ) : null}
  </>
);

beforeEach(() => {
  mockRuns = [];
  mockMutate.mockReset();
});
afterEach(() => {
  cleanup();
  clearChatTyping("workspace");
});

test("human typing stays below its own composer while both panes are open", () => {
  render(<Indicators openThread="root" />);
  act(() => {
    for (const [actorId, fullName, threadRootId] of [
      ["top-level", "Sarah", null],
      ["in-thread", "Joseph", "root"],
      ["other-thread", "Someone else", "other-root"],
      ["me", "My own name", "root"],
    ]) {
      applyChatTypingEvent({
        workspaceId: "workspace",
        conversationId: "room",
        actorId: actorId!,
        fullName: fullName!,
        threadRootId,
        isTyping: true,
        expiresAt: new Date(Date.now() + 6_000).toISOString(),
      });
    }
  });
  expect(
    within(screen.getByRole("region", { name: "Main composer" })).getByRole(
      "status",
    ),
  ).toHaveTextContent("Sarah is typing…");
  expect(
    within(screen.getByRole("region", { name: "Thread composer" })).getByRole(
      "status",
    ),
  ).toHaveTextContent("Joseph is typing…");
  expect(
    screen.queryByText(/Someone else|My own name/),
  ).not.toBeInTheDocument();
});

test("the first Maya ask moves from the main composer to its opened thread exactly once", () => {
  mockRuns = [run()];
  const view = render(<Indicators />);
  expect(screen.getByRole("status")).toHaveTextContent("Maya is thinking…");

  view.rerender(<Indicators openThread="root" />);
  expect(
    within(screen.getByRole("region", { name: "Main composer" })).getByRole(
      "status",
    ),
  ).toBeEmptyDOMElement();
  expect(
    within(screen.getByRole("region", { name: "Thread composer" })).getByRole(
      "status",
    ),
  ).toHaveTextContent("Maya is thinking…");
  expect(screen.getAllByText("Maya is thinking…")).toHaveLength(1);
  fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
  expect(mockMutate).toHaveBeenCalledWith({
    run: mockRuns[0],
    action: "cancel",
  });

  view.rerender(<Indicators openThread="another-thread" />);
  expect(
    within(screen.getByRole("region", { name: "Main composer" })).getByRole(
      "status",
    ),
  ).toHaveTextContent("Maya is thinking…");
  expect(
    within(screen.getByRole("region", { name: "Thread composer" })).getByRole(
      "status",
    ),
  ).toBeEmptyDOMElement();
});

test("Maya follow-ups belong only to their destination thread even when their source is paged out", () => {
  mockRuns = [run({ sourceId: "older-reply" })];
  const view = render(<Indicators openThread="root" />);
  expect(
    within(screen.getByRole("region", { name: "Main composer" })).getByRole(
      "status",
    ),
  ).toBeEmptyDOMElement();
  expect(
    within(screen.getByRole("region", { name: "Thread composer" })).getByRole(
      "status",
    ),
  ).toHaveTextContent("Maya is thinking…");

  view.rerender(<Indicators openThread="another-thread" />);
  for (const indicator of screen.getAllByRole("status"))
    expect(indicator).toBeEmptyDOMElement();
  view.rerender(<Indicators />);
  expect(screen.getByRole("status")).toBeEmptyDOMElement();
});

test("older API responses infer the same scope from loaded messages without showing in both panes", () => {
  mockRuns = [run({ sourceId: "reply", threadRootId: undefined })];
  render(
    <Indicators
      openThread="root"
      replies={[{ id: "reply", threadRootId: "root" } as ChatMessage]}
    />,
  );
  expect(
    within(screen.getByRole("region", { name: "Main composer" })).getByRole(
      "status",
    ),
  ).toBeEmptyDOMElement();
  expect(
    within(screen.getByRole("region", { name: "Thread composer" })).getByRole(
      "status",
    ),
  ).toHaveTextContent("Maya is thinking…");
});

test.each([
  "needs_input",
  "awaiting_approval",
  "succeeded",
  "failed",
  "cancelled",
] as const)(
  "Maya does not appear to be thinking after the run is %s",
  (status) => {
    mockRuns = [run({ status })];
    render(<Indicators openThread="root" />);
    for (const indicator of screen.getAllByRole("status"))
      expect(indicator).toBeEmptyDOMElement();
  },
);
