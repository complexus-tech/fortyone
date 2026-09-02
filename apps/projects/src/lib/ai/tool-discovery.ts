import type { ToolSet, UIMessage } from "ai";
import type { MayaToolName } from "./tool-names";

const EAGER_OPENAI_TOOL_NAMES = new Set<string>(
  ["navigation", "suggestions", "theme"] satisfies readonly MayaToolName[],
);

export const MAYA_TOOL_SEARCH_NAME = "mayaToolSearch" as const;
const MAYA_TOOL_SEARCH_UI_PART_TYPE = `tool-${MAYA_TOOL_SEARCH_NAME}`;

/**
 * Hosted tool-search calls are orchestration details, not conversational
 * evidence. AI SDK currently folds the hosted call and output into one UI part
 * and loses their distinct provider item IDs; replaying that part reconstructs
 * two OpenAI items with the same `tsc_` ID. Remove only this historical part
 * before model conversion while retaining the discovered tool calls, their
 * outputs, and the user-visible answer.
 */
export const omitMayaToolSearchHistory = (
  messages: UIMessage[],
): UIMessage[] =>
  messages.map((message) => {
    const parts = message.parts.filter(
      (part) => part.type !== MAYA_TOOL_SEARCH_UI_PART_TYPE,
    );

    return parts.length === message.parts.length
      ? message
      : { ...message, parts };
  });

/**
 * Keep Maya's entire capability catalog discoverable without loading every
 * function schema into the model's working context. Hosted tool search is
 * semantic, so availability does not depend on language, vocabulary, route,
 * or whichever domain was discussed previously.
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
