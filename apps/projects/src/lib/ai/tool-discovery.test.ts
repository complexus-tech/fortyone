/** @jest-environment node */ // eslint-disable-line tsdoc/syntax -- Jest requires this docblock.

import { tool, type ToolSet } from "ai";
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
});
