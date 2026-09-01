import type { ToolSet } from "ai";
import type { MayaToolName } from "./tool-names";

const EAGER_OPENAI_TOOL_NAMES = new Set<string>(
  ["navigation", "suggestions", "theme"] satisfies readonly MayaToolName[],
);

export const MAYA_TOOL_SEARCH_NAME = "mayaToolSearch" as const;

/**
 * Keep Maya's entire capability catalog discoverable without loading every
 * function schema into the model's working context. OpenAI's hosted tool
 * search matches requests semantically, so discovery does not depend on the
 * user's language, vocabulary, or the previous conversational domain.
 */
export const withOpenAIToolDiscovery = <TOOLS extends ToolSet>(
  toolSet: TOOLS,
  toolSearch: ToolSet[string],
): TOOLS & Record<typeof MAYA_TOOL_SEARCH_NAME, ToolSet[string]> => {
  const deferredTools = Object.entries(toolSet).map(
    ([name, registeredTool]) => {
      if (EAGER_OPENAI_TOOL_NAMES.has(name)) return [name, registeredTool];

      const existingProviderOptions = registeredTool.providerOptions ?? {};
      const existingOpenAIOptions =
        (existingProviderOptions.openai as Record<string, unknown> | undefined) ??
        {};

      return [
        name,
        {
          ...registeredTool,
          providerOptions: {
            ...existingProviderOptions,
            openai: {
              ...existingOpenAIOptions,
              deferLoading: true,
            },
          },
        },
      ];
    },
  );

  return Object.fromEntries([
    ...deferredTools,
    [MAYA_TOOL_SEARCH_NAME, toolSearch],
  ]) as TOOLS & Record<typeof MAYA_TOOL_SEARCH_NAME, ToolSet[string]>;
};
