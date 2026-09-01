/** @jest-environment node */ // eslint-disable-line tsdoc/syntax -- Jest requires this docblock.

import { createOpenAI } from "@ai-sdk/openai";
import { generateText, tool, type ToolSet } from "ai";
import { z } from "zod";
import {
  MAYA_TOOL_SEARCH_NAME,
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

  it("preserves existing provider options and tool behavior", () => {
    const execute = jest.fn(() => ({ success: true }));
    const original = tool({
      description: "Create a story",
      inputSchema: z.object({ title: z.string() }),
      execute,
      providerOptions: {
        openai: { customFlag: "preserved" },
        other: { value: 1 },
      },
    });
    const tools = withOpenAIToolDiscovery(
      { createStory: original },
      { type: "provider", id: "openai.tool_search" } as never,
    );

    expect(tools.createStory).toMatchObject({
      description: "Create a story",
      execute,
      providerOptions: {
        openai: { customFlag: "preserved", deferLoading: true },
        other: { value: 1 },
      },
    });
    expect(tools.createStory.inputSchema).toBe(original.inputSchema);
  });

  it("does not mutate the registered tool set", () => {
    const originalTools: ToolSet = { createStory: registeredTool() };
    withOpenAIToolDiscovery(
      originalTools,
      { type: "provider", id: "openai.tool_search" } as never,
    );

    expect(originalTools.createStory.providerOptions).toBeUndefined();
    expect(originalTools).not.toHaveProperty(MAYA_TOOL_SEARCH_NAME);
  });

  it("serializes hosted search and deferred functions to OpenAI", async () => {
    let requestBody: { tools?: Record<string, unknown>[] } | undefined;
    const fetch = jest.fn(
      async (_input: RequestInfo | URL, init?: RequestInit) => {
        requestBody = JSON.parse(String(init?.body)) as {
          tools?: Record<string, unknown>[];
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
    const tools = withOpenAIToolDiscovery(
      {
        createStory: registeredTool(),
        navigation: registeredTool(),
      },
      openai.tools.toolSearch({ execution: "server" }),
    );

    await expect(
      generateText({
        model: openai("gpt-5.6-luna"),
        prompt: "Gadzira story",
        tools,
      }),
    ).rejects.toThrow("test stop");

    expect(requestBody?.tools).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          defer_loading: true,
          name: "createStory",
          type: "function",
        }),
        expect.objectContaining({
          name: "navigation",
          type: "function",
        }),
        expect.objectContaining({
          execution: "server",
          type: "tool_search",
        }),
      ]),
    );
    expect(
      requestBody?.tools?.find(({ name }) => name === "navigation"),
    ).not.toHaveProperty("defer_loading");
  });
});
