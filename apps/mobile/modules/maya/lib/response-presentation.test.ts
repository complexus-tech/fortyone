import assert from "node:assert/strict";
import test from "node:test";
import type { UIMessage } from "ai";
import {
  captureMayaResponse,
  selectMayaResponseRows,
  type MayaResponseRow,
} from "./response-presentation.ts";

const text = (
  id: string,
  role: UIMessage["role"],
  value: string,
): UIMessage => ({
  id,
  role,
  parts: [{ type: "text", text: value }],
});
const previous = [
  text("old-user", "user", "Earlier"),
  text("old-reply", "assistant", "Earlier reply"),
];
const user = text("new-user", "user", "Create a task");
const snapshot = () =>
  captureMayaResponse({
    chatId: "chat",
    requestId: "request",
    mode: "send",
    messages: previous,
  });
const approval = (): UIMessage => ({
  id: "proposal",
  role: "assistant",
  parts: [
    { type: "text", text: "I can create that task." },
    {
      type: "tool-createStory",
      toolCallId: "call",
      state: "approval-requested",
      input: { title: "Task" },
      approval: { id: "approve" },
    },
  ],
});
const continuation = () => {
  const message = approval();
  const response = captureMayaResponse({
    chatId: "chat",
    requestId: "approve-request",
    mode: "approval",
    messages: [...previous, user, message],
  });
  return { message, response };
};
const rowText = (row: MayaResponseRow) =>
  row.kind === "thinking"
    ? "thinking"
    : row.message.parts
        .flatMap((part) => (part.type === "text" ? part.text : []))
        .join("");

test("preparation preserves history without putting progress above the later user row", () => {
  const rows = selectMayaResponseRows({
    chatId: "chat",
    messages: previous,
    response: snapshot(),
    active: true,
  });
  assert.deepEqual(
    rows.map((row) => row.id),
    ["old-user", "old-reply"],
  );
});

test("empty starts, whitespace, and tool deltas share exactly one response slot", () => {
  const response = snapshot();
  const starts: UIMessage[] = [
    { id: "reply", role: "assistant", parts: [] },
    text("reply", "assistant", ""),
    text("reply", "assistant", " \n"),
    {
      id: "reply",
      role: "assistant",
      parts: [{ type: "reasoning", text: "Internal work", state: "streaming" }],
    },
    {
      id: "reply",
      role: "assistant",
      parts: [
        {
          type: "tool-searchStories",
          toolCallId: "search",
          state: "input-available",
          input: { query: "Task" },
        },
      ],
    },
    {
      id: "reply",
      role: "assistant",
      parts: [
        {
          type: "tool-searchStories",
          toolCallId: "search",
          state: "output-available",
          input: { query: "Task" },
          output: { message: "Found a task" },
        },
      ],
    },
  ];
  for (const reply of [null, ...starts]) {
    const rows = selectMayaResponseRows({
      chatId: "chat",
      messages: [...previous, user, ...(reply ? [reply] : [])],
      response,
      active: true,
    });
    assert.deepEqual(
      rows.map((row) => row.id),
      ["old-user", "old-reply", "new-user", "maya-response:request"],
    );
    assert.equal(rows.at(-1)?.kind, "thinking");
  }
});

test("first readable assistant text replaces thinking with the same row identity", () => {
  const response = snapshot();
  const reply = text("reply", "assistant", "Hello");
  for (const active of [true, false]) {
    const rows = selectMayaResponseRows({
      chatId: "chat",
      messages: [...previous, user, reply],
      response,
      active,
    });
    assert.equal(rows.at(-1)?.id, "maya-response:request");
    assert.equal(rowText(rows.at(-1)!), "Hello");
    assert.equal(rows.filter((row) => row.kind === "thinking").length, 0);
  }
});

test("a files-only user turn anchors progress without requiring user text", () => {
  const file: UIMessage = {
    id: "file-user",
    role: "user",
    parts: [
      {
        type: "file",
        mediaType: "application/pdf",
        url: "data:application/pdf;base64,JVBERg==",
        filename: "brief.pdf",
      },
    ],
  };
  const rows = selectMayaResponseRows({
    chatId: "chat",
    messages: [...previous, file],
    response: snapshot(),
    active: true,
  });
  assert.deepEqual(
    rows.slice(-2).map((row) => row.id),
    ["file-user", "maya-response:request"],
  );
});

test("approval continuation freezes existing text and cards until new text arrives", () => {
  const { message, response } = continuation();
  message.parts[0] = { type: "text", text: "I can create that task." };
  message.parts[1] = {
    type: "tool-createStory",
    toolCallId: "call",
    state: "output-available",
    input: { title: "Task" },
    output: { message: "Created" },
  };
  message.parts.push({ type: "text", text: "  " });
  const rows = selectMayaResponseRows({
    chatId: "chat",
    messages: [...previous, user, message],
    response,
    active: true,
  });
  assert.equal(rows.at(-1)?.kind, "thinking");
  const history = rows.at(-2);
  assert.equal(history?.kind, "message");
  if (history?.kind === "message")
    assert.equal(
      (history.message.parts[1] as { state: string }).state,
      "approval-requested",
    );
  assert.equal(rowText(rows.at(-2)!), "I can create that task.");
});

test("continued new parts do not repeat old text, and remain stable after completion", () => {
  const { message, response } = continuation();
  message.parts.push({ type: "text", text: "The task is ready." });
  for (const active of [true, false]) {
    const rows = selectMayaResponseRows({
      chatId: "chat",
      messages: [...previous, user, message],
      response,
      active,
    });
    assert.equal(rows.at(-2)?.id, "proposal");
    assert.equal(rowText(rows.at(-2)!), "I can create that task.");
    assert.equal(rows.at(-1)?.id, "maya-response:approve-request");
    assert.equal(rowText(rows.at(-1)!), "The task is ready.");
    if (rows.at(-1)?.kind === "message")
      assert.equal(
        (rows.at(-1) as Extract<MayaResponseRow, { kind: "message" }>).message
          .id,
        "proposal",
      );
  }
});

test("new text appended to an existing text part is detected without duplication", () => {
  const { message, response } = continuation();
  message.parts[0] = { type: "text", text: "I can create that task. Done." };
  const rows = selectMayaResponseRows({
    chatId: "chat",
    messages: [...previous, user, message],
    response,
    active: true,
  });
  assert.equal(rowText(rows.at(-2)!), "I can create that task.");
  assert.equal(rowText(rows.at(-1)!), " Done.");
});

test("idle cancellation/errors remove progress and expose actual actionable tools", () => {
  for (const reply of [
    approval(),
    {
      id: "failure",
      role: "assistant",
      parts: [
        {
          type: "tool-createStory",
          toolCallId: "call",
          state: "output-error",
          input: {},
          errorText: "Task creation failed",
        },
      ],
    } as UIMessage,
  ]) {
    const rows = selectMayaResponseRows({
      chatId: "chat",
      messages: [...previous, user, reply],
      response: snapshot(),
      active: false,
    });
    assert.equal(rows.at(-1)?.kind, "message");
    assert.equal(
      rows.some((row) => row.kind === "thinking"),
      false,
    );
  }
  const rows = selectMayaResponseRows({
    chatId: "chat",
    messages: [...previous, user, text("empty", "assistant", "")],
    response: snapshot(),
    active: false,
  });
  assert.equal(rows.at(-1)?.id, "new-user");
});

test("conversation changes ignore stale response snapshots", () => {
  const rows = selectMayaResponseRows({
    chatId: "another-chat",
    messages: [user],
    response: snapshot(),
    active: true,
  });
  assert.deepEqual(
    rows.map((row) => row.id),
    ["new-user"],
  );
});

test("inactive pending tools never leave a padded blank assistant row", () => {
  for (const state of ["input-available", "approval-responded"] as const) {
    const message = {
      id: "pending",
      role: "assistant",
      parts: [
        {
          type: "tool-createStory",
          toolCallId: "call",
          state,
          input: {},
          approval: { id: "approve", approved: true },
        },
      ],
    } as UIMessage;
    const rows = selectMayaResponseRows({
      chatId: "chat",
      messages: [user, message],
      response: null,
      active: false,
    });
    assert.deepEqual(
      rows.map((row) => row.id),
      ["new-user"],
    );
  }
});

test("a later request does not treat the previous reply as its response", () => {
  const messages = [...previous, user, text("reply", "assistant", "Done")];
  const response = captureMayaResponse({
    chatId: "chat",
    requestId: "next",
    mode: "send",
    messages,
  });
  assert.equal(
    selectMayaResponseRows({
      chatId: "chat",
      messages,
      response,
      active: true,
    }).some((row) => row.kind === "thinking"),
    false,
  );
  const rows = selectMayaResponseRows({
    chatId: "chat",
    messages: [...messages, text("next-user", "user", "Another task")],
    response,
    active: true,
  });
  assert.equal(rows.at(-1)?.id, "maya-response:next");
});

test("successful tool results leave no empty history rows while errors and replies remain", () => {
  const message = (id: string, output: unknown): UIMessage => ({
    id,
    role: "assistant",
    parts: [
      {
        type: "tool-searchStories",
        toolCallId: id,
        state: "output-available",
        input: {},
        output,
      },
    ],
  });
  const rows = selectMayaResponseRows({
    chatId: "chat",
    active: false,
    response: null,
    messages: [
      message("empty", {
        message: "Found 0 stories in this team.",
        stories: [],
      }),
      message("card", {
        message: "Found 1 story in this team.",
        stories: [
          { id: "01234567-1234-4234-8234-0123456789ab", title: "Task" },
        ],
      }),
      message("error", { success: false, message: "Access denied" }),
      text("reply", "assistant", "Here is today's focus."),
    ],
  });
  assert.deepEqual(
    rows.map((row) => row.id),
    ["error", "reply"],
  );
});
