/** @jest-environment node */ // eslint-disable-line tsdoc/syntax -- Jest requires this docblock.
/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ModelMessage } from "ai";
import { generateText } from "ai";
import { createOpenAI } from "@ai-sdk/openai";
import { sanitizeOpenAIHistoryItemReferences } from "./openai-history";

describe("OpenAI model history", () => {
  it("reconstructs an orphaned output message instead of referencing it", async () => {
    let requestBody: { input?: unknown } | undefined;
    const fetch = jest.fn(
      async (_input: RequestInfo | URL, init?: RequestInit) => {
        requestBody = JSON.parse(String(init?.body)) as { input?: unknown };

        return new Response(
          JSON.stringify({ error: { message: "test stop" } }),
          {
            headers: { "content-type": "application/json" },
            status: 400,
          },
        );
      },
    );
    const openai = createOpenAI({ apiKey: "test", fetch });
    const messages = [
      {
        content: [
          {
            providerOptions: {
              openai: {
                itemId: "msg_orphaned",
                phase: "final_answer",
              },
              testProvider: { preserved: true },
            },
            text: "Earlier answer",
            type: "text",
          },
        ],
        role: "assistant",
      },
      { content: "Continue", role: "user" },
    ] satisfies ModelMessage[];

    const sanitizedMessages = sanitizeOpenAIHistoryItemReferences(messages);
    await expect(
      generateText({
        messages: sanitizedMessages,
        model: openai("gpt-5.6-luna"),
      }),
    ).rejects.toThrow("test stop");

    expect(fetch).toHaveBeenCalledTimes(1);
    expect(requestBody?.input).toEqual([
      {
        content: [{ text: "Earlier answer", type: "output_text" }],
        phase: "final_answer",
        role: "assistant",
      },
      {
        content: [{ text: "Continue", type: "input_text" }],
        role: "user",
      },
    ]);
    expect(sanitizedMessages[0]).toHaveProperty("content.0.providerOptions", {
      openai: { phase: "final_answer" },
      testProvider: { preserved: true },
    });
    expect(messages[0]).toHaveProperty(
      "content.0.providerOptions.openai.itemId",
      "msg_orphaned",
    );
  });

  it("keeps output item references when their reasoning item is present", () => {
    const messages = [
      {
        content: [
          {
            providerOptions: {
              openai: { itemId: "rs_linked" },
            },
            text: "",
            type: "reasoning",
          },
          {
            providerOptions: {
              openai: {
                itemId: "msg_linked",
                phase: "final_answer",
              },
            },
            text: "Earlier answer",
            type: "text",
          },
        ],
        role: "assistant",
      },
    ] satisfies ModelMessage[];

    const sanitizedMessages = sanitizeOpenAIHistoryItemReferences(messages);

    expect(sanitizedMessages).toEqual(messages);
    expect(sanitizedMessages[0]).toBe(messages[0]);
  });
});
