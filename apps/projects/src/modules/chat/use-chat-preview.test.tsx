/* global afterEach, expect, it -- Jest globals are provided by the projects test runner. */

import { act, cleanup, renderHook } from "@testing-library/react";
import type { PreviewDraft } from "./chat-rich-text";
import type { PreviewMessage } from "./preview-data";
import { PREVIEW_CONVERSATIONS } from "./preview-data";
import { useChatPreview } from "./use-chat-preview";

const getMessage = (messages: PreviewMessage[], id: string) => {
  const message = messages.find((item) => item.id === id);
  if (!message) throw new Error(`Expected message ${id} in preview`);
  return message;
};

afterEach(cleanup);

it("preserves message data and row callbacks while conversation and thread drafts change", () => {
  const { result } = renderHook(useChatPreview);
  const root = getMessage(result.current.conversation.messages, "product-4");
  act(() => {
    result.current.openThread(root, document.createElement("button"));
  });
  const initial = result.current;
  const draft: PreviewDraft = { text: "An unfinished message", content: null };

  act(() => {
    result.current.updateDraft(initial.activeId, draft);
  });
  act(() => {
    result.current.updateDraft(`${initial.activeId}:thread:${root.id}`, {
      text: "An unfinished reply",
      content: null,
    });
  });

  expect(result.current.draft).toBe(draft);
  expect(result.current.conversation).toBe(initial.conversation);
  expect(result.current.conversation.messages).toBe(
    initial.conversation.messages,
  );
  expect(result.current.thread).toBe(initial.thread);
  const currentRoot = getMessage(result.current.conversation.messages, root.id);
  expect(currentRoot).toBe(root);
  expect(currentRoot.replies).toBe(root.replies);
  for (const callback of [
    "editMessage",
    "deleteMessage",
    "toggleReaction",
    "openThread",
    "closeThread",
    "viewMessage",
    "selectTab",
  ] as const) {
    expect(result.current[callback]).toBe(initial[callback]);
  }
});

it("applies edits and reactions, then retains replies when their root becomes a tombstone", () => {
  const { result } = renderHook(useChatPreview);
  const original = getMessage(PREVIEW_CONVERSATIONS[0].messages, "product-4");
  const originalText = original.text;
  const draft: PreviewDraft = {
    text: "Updated welcome plan",
    content: {
      type: "doc",
      content: [
        {
          type: "paragraph",
          content: [
            {
              type: "text",
              text: "Updated welcome plan",
              marks: [{ type: "bold" }],
            },
          ],
        },
      ],
    },
  };

  act(() => {
    result.current.editMessage(original.id, draft);
  });
  expect(
    getMessage(result.current.conversation.messages, original.id),
  ).toMatchObject({
    text: draft.text,
    richContent: draft.content,
    edited: true,
  });

  act(() => {
    result.current.toggleReaction(original.id, "✅");
  });
  expect(
    getMessage(result.current.conversation.messages, original.id).reactions,
  ).toContainEqual({ emoji: "✅", count: 1, reacted: true });

  act(() => {
    result.current.deleteMessage(original.id);
  });
  const tombstone = getMessage(
    result.current.conversation.messages,
    original.id,
  );
  expect(tombstone).toMatchObject({
    deleted: true,
    text: "",
    richContent: null,
    reactions: [],
  });
  expect(tombstone.replies?.map((reply) => reply.id)).toEqual(
    original.replies?.map((reply) => reply.id),
  );
  expect(original.text).toBe(originalText);
  expect(original.deleted).toBeUndefined();
});

it("removes an own leaf message while retaining another person's message", () => {
  const { result } = renderHook(useChatPreview);
  const initialCount = result.current.conversation.messages.length;

  act(() => {
    result.current.deleteMessage("product-today-03");
    result.current.deleteMessage("product-1");
  });

  expect(result.current.conversation.messages).toHaveLength(initialCount - 1);
  expect(
    result.current.conversation.messages.some(
      (message) => message.id === "product-today-03",
    ),
  ).toBe(false);
  expect(
    getMessage(result.current.conversation.messages, "product-1").deleted,
  ).toBeUndefined();
});

it("requests the original root when closing a thread whose trigger was virtualized away", () => {
  const { result } = renderHook(useChatPreview);
  const root = getMessage(result.current.conversation.messages, "product-4");
  const unmountedTrigger = document.createElement("button");

  act(() => {
    result.current.openThread(root, unmountedTrigger);
  });
  expect(result.current.thread?.id).toBe(root.id);

  act(() => {
    result.current.closeThread();
  });
  expect(result.current.thread).toBeUndefined();
  expect(result.current.targetMessageId).toBe(root.id);
});
