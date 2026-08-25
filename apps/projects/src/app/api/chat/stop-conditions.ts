import type { StopCondition } from "ai";
import type { tools } from "@/lib/ai/tools";

const STORY_CREATION_TOOL_NAMES = new Set(["createStory", "bulkCreateStories"]);

const isTerminalOutput = (output: unknown) =>
  Boolean(
    output &&
      typeof output === "object" &&
      ("success" in output || "needsConfirmation" in output),
  );

export const hasTerminalStoryCreationResult: StopCondition<typeof tools> = ({
  steps,
}) =>
  steps
    .at(-1)
    ?.toolResults.some(
      (result) =>
        STORY_CREATION_TOOL_NAMES.has(result.toolName) &&
        isTerminalOutput(result.output),
    ) ?? false;
