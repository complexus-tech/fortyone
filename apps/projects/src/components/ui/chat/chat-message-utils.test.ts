/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { MayaUIMessage } from "@/lib/ai/tools/types";
import {
  getPromptTextSegments,
  getVisibleMessageRenderEntries,
  getVisibleToolPartIndexes,
} from "./chat-message-utils";

describe("getPromptTextSegments", () => {
  it("detects HTTP and www links without including sentence punctuation", () => {
    expect(
      getPromptTextSegments(
        "Review https://fortyone.app/work/PRD-571, then visit www.example.com.",
      ),
    ).toEqual([
      { start: 0, type: "text", value: "Review " },
      {
        type: "link",
        href: "https://fortyone.app/work/PRD-571",
        start: 7,
        value: "https://fortyone.app/work/PRD-571",
      },
      { start: 40, type: "text", value: "," },
      { start: 41, type: "text", value: " then visit " },
      {
        type: "link",
        href: "https://www.example.com",
        start: 53,
        value: "www.example.com",
      },
      { start: 68, type: "text", value: "." },
    ]);
  });

  it("leaves text without links unchanged", () => {
    expect(getPromptTextSegments("Plan the next sprint")).toEqual([
      { start: 0, type: "text", value: "Plan the next sprint" },
    ]);
  });

  it("keeps the exact preview visible for approval, then replaces it with the applied plan", () => {
    const preview = {
      output: { kind: "maya-work-plan", phase: "preview", plan: {} },
      state: "output-available",
      toolCallId: "preview-1",
      type: "tool-mayaWorkPlanTool",
    };
    const approval = {
      approval: { id: "approval-1" },
      input: { runId: "run-1" },
      state: "approval-requested",
      toolCallId: "apply-1",
      type: "tool-applyMayaWorkPlanTool",
    };
    const message = {
      id: "message-1",
      parts: [preview, approval],
      role: "assistant",
    } as unknown as MayaUIMessage;

    expect(Array.from(getVisibleToolPartIndexes(message))).toEqual([0, 1]);

    const appliedMessage = {
      ...message,
      parts: [
        preview,
        {
          ...approval,
          output: { kind: "maya-work-plan", phase: "applied", plan: {} },
          state: "output-available",
        },
      ],
    } as unknown as MayaUIMessage;
    expect(Array.from(getVisibleToolPartIndexes(appliedMessage))).toEqual([1]);
  });
});

describe("getVisibleMessageRenderEntries", () => {
  it("uses stable unique keys when persisted messages have missing or duplicate ids", () => {
    const messages = [
      {
        id: "",
        parts: [{ text: "First", type: "text" }],
        role: "user",
      },
      {
        id: "",
        parts: [{ text: "Second", type: "text" }],
        role: "assistant",
      },
      {
        id: "stable-id",
        parts: [{ text: "Third", type: "text" }],
        role: "user",
      },
      {
        id: "stable-id",
        parts: [{ text: "Fourth", type: "text" }],
        role: "assistant",
      },
    ] satisfies MayaUIMessage[];

    const entries = getVisibleMessageRenderEntries(messages);
    const renderKeys = entries.map(({ renderKey }) => renderKey);

    expect(renderKeys).toEqual([
      "maya-message-user-0",
      "maya-message-assistant-1",
      "stable-id",
      "stable-id-3",
    ]);
    expect(new Set(renderKeys).size).toBe(4);
    expect(entries.map(({ isLast }) => isLast)).toEqual([
      false,
      false,
      false,
      true,
    ]);
  });

  it("marks no visible message as last when the actual final message is private", () => {
    const messages = [
      {
        id: "visible",
        parts: [{ text: "Visible", type: "text" }],
        role: "assistant",
      },
      {
        id: "private",
        parts: [],
        role: "assistant",
      },
    ] satisfies MayaUIMessage[];

    expect(
      getVisibleMessageRenderEntries(messages).map(({ isLast }) => isLast),
    ).toEqual([false]);
  });
});
