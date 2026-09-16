import assert from "node:assert/strict";
import test from "node:test";
import type { FileUIPart } from "ai";
import type { MayaMessage } from "../types";
import {
  HISTORICAL_ATTACHMENT_PLACEHOLDER,
  prepareMobileMayaMessages,
} from "./chat-protocol.ts";

const file = (filename: string, url: string): FileUIPart => ({
  type: "file",
  mediaType: "application/pdf",
  filename,
  url,
});

test("only the newly submitted final user turn preserves attachment bytes", () => {
  const messages: MayaMessage[] = [
    {
      id: "old-user",
      role: "user",
      metadata: { keep: true },
      parts: [
        { type: "text", text: "Earlier" },
        file("private-old.pdf", "data:application/pdf;base64,b2xk"),
      ],
    },
    {
      id: "assistant",
      role: "assistant",
      parts: [file("assistant-file.pdf", "https://example.test/old.pdf")],
    },
    {
      id: "new-user",
      role: "user",
      parts: [file("current.pdf", "data:application/pdf;base64,bmV3")],
    },
  ];
  const prepared = prepareMobileMayaMessages(messages, {
    preserveLastUserFiles: true,
  });
  assert.equal(prepared[2], messages[2]);
  assert.deepEqual(prepared[0].metadata, { keep: true });
  assert.deepEqual(prepared[0].parts, [
    { type: "text", text: "Earlier" },
    { type: "text", text: HISTORICAL_ATTACHMENT_PLACEHOLDER },
  ]);
  assert.equal(JSON.stringify(prepared).includes("private-old.pdf"), false);
  assert.equal(
    JSON.stringify(prepared).includes("https://example.test"),
    false,
  );
  assert.equal(messages[0].parts[1].type, "file");
});

test("approval continuation omits all files without changing tool receipts", () => {
  const receipt: MayaMessage = {
    id: "approval",
    role: "assistant",
    parts: [
      {
        type: "tool-createStory",
        toolCallId: "call-one",
        state: "approval-responded",
        input: { title: "Task" },
        approval: { id: "approval-one", approved: true },
      },
    ],
  };
  const messages: MayaMessage[] = [
    {
      id: "user",
      role: "user",
      parts: [file("brief.pdf", "data:application/pdf;base64,cGRm")],
    },
    receipt,
  ];
  const prepared = prepareMobileMayaMessages(messages, {
    preserveLastUserFiles: true,
  });
  assert.equal(prepared[1], receipt);
  assert.equal(JSON.stringify(prepared).includes("base64"), false);
  assert.deepEqual(prepared[0].parts, [
    { type: "text", text: HISTORICAL_ATTACHMENT_PLACEHOLDER },
  ]);
});

test("historical files are omitted unless preservation is explicitly requested", () => {
  const messages: MayaMessage[] = [
    {
      id: "user",
      role: "user",
      parts: [file("brief.pdf", "data:application/pdf;base64,cGRm")],
    },
  ];
  assert.deepEqual(prepareMobileMayaMessages(messages)[0].parts, [
    { type: "text", text: HISTORICAL_ATTACHMENT_PLACEHOLDER },
  ]);
});
