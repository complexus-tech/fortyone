/** @jest-environment node */ // eslint-disable-line tsdoc/syntax -- Jest requires this docblock.

import type { ToolSet, UIMessage } from "ai";
import { createOpenAI } from "@ai-sdk/openai";
import { convertToModelMessages, generateText, tool } from "ai";
import { z } from "zod";
import {
  MAYA_TOOL_SEARCH_NAME,
  omitMayaToolSearchHistory,
  withOpenAIToolDiscovery,
} from "./tool-discovery";

const registeredTool = () =>
  tool({
    description: "Example capability",
    inputSchema: z.object({}),
    execute: () => ({ success: true }),
  });

describe("withOpenAIToolDiscovery", () => {
  it("defers domain schemas while keeping lightweight UI tools eager", () => {
    const toolSearch = { type: "provider", id: "openai.tool_search" } as never;
    const tools = withOpenAIToolDiscovery(
      {
        createStory: registeredTool(),
        navigation: registeredTool(),
        theme: registeredTool(),
      },
      toolSearch,
    );

    expect(tools.createStory.providerOptions?.openai).toMatchObject({
      deferLoading: true,
    });
    expect(tools.navigation.providerOptions?.openai).toBeUndefined();
    expect(tools.theme.providerOptions?.openai).toBeUndefined();
    expect(tools[MAYA_TOOL_SEARCH_NAME]).toBe(toolSearch);
  });

  it("preserves provider options without mutating the registry", () => {
    const originalTools: ToolSet = {
      createStory: tool({
        description: "Create a story",
        inputSchema: z.object({}),
        providerOptions: { other: { value: 1 } },
        execute: () => ({ success: true }),
      }),
    };
    const discoveredTools = withOpenAIToolDiscovery(
      originalTools,
      { type: "provider", id: "openai.tool_search" } as never,
    );

    expect(discoveredTools.createStory.providerOptions).toMatchObject({
      openai: { deferLoading: true },
      other: { value: 1 },
    });
    expect(originalTools.createStory.providerOptions).toEqual({
      other: { value: 1 },
    });
    expect(originalTools).not.toHaveProperty(MAYA_TOOL_SEARCH_NAME);
  });

  it("removes only hosted search parts from replayed UI history", () => {
    const messages = [
      {
        id: "assistant-1",
        role: "assistant",
        parts: [
          {
            type: "tool-mayaToolSearch",
            toolCallId: "tsc_duplicate",
            state: "output-available",
            input: { arguments: { search_query: "team members" } },
            output: { tools: [] },
          },
          {
            type: "tool-listTeamMembers",
            toolCallId: "call_members",
            state: "output-available",
            input: {},
            output: { members: [] },
          },
          { type: "text", text: "You are the only member." },
        ],
      },
    ] satisfies UIMessage[];

    const sanitized = omitMayaToolSearchHistory(messages);

    expect(sanitized[0].parts).toEqual([
      expect.objectContaining({ type: "tool-listTeamMembers" }),
      { type: "text", text: "You are the only member." },
    ]);
    expect(messages[0].parts).toHaveLength(3);
  });

  it("omits duplicate search items while retaining the real OpenAI tool pair", async () => {
    let requestBody: { input?: Record<string, unknown>[] } | undefined;
    const fetch = jest.fn(
      async (_input: RequestInfo | URL, init?: RequestInit) => {
        requestBody = JSON.parse(String(init?.body)) as {
          input?: Record<string, unknown>[];
        };
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
    const runtimeTools = withOpenAIToolDiscovery(
      { listTeamMembers: registeredTool() },
      openai.tools.toolSearch({ execution: "server" }),
    );
    const messages = [
      {
        id: "assistant-1",
        role: "assistant",
        parts: [
          {
            type: "tool-mayaToolSearch",
            toolCallId: "tsc_duplicate",
            state: "output-available",
            input: { arguments: { search_query: "team members" } },
            output: { tools: [] },
          },
          {
            type: "tool-listTeamMembers",
            toolCallId: "call_members",
            state: "output-available",
            input: {},
            output: { success: true },
          },
          { type: "text", text: "You are the only member." },
        ],
      },
      {
        id: "user-2",
        role: "user",
        parts: [{ type: "text", text: "Continue" }],
      },
    ] satisfies UIMessage[];

    const modelMessages = await convertToModelMessages(
      omitMayaToolSearchHistory(messages),
      { tools: runtimeTools },
    );
    await expect(
      generateText({
        messages: modelMessages,
        model: openai("gpt-5.6-luna"),
        tools: runtimeTools,
      }),
    ).rejects.toThrow("test stop");

    expect(requestBody?.input).not.toEqual(
      expect.arrayContaining([
        expect.objectContaining({ type: "tool_search_call" }),
      ]),
    );
    expect(requestBody?.input).not.toEqual(
      expect.arrayContaining([
        expect.objectContaining({ type: "tool_search_output" }),
      ]),
    );
    expect(requestBody?.input).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          call_id: "call_members",
          name: "listTeamMembers",
          type: "function_call",
        }),
        expect.objectContaining({
          call_id: "call_members",
          type: "function_call_output",
        }),
      ]),
    );
  });
});
