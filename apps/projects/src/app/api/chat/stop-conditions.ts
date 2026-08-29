import type { StopCondition } from "ai";
import type { tools } from "@/lib/ai/tools";
import {
  isMutationToolCall,
  isTerminalMutationToolCall,
} from "@/lib/ai/tool-policy";

const getOutputState = (output: unknown) => {
  if (!output || typeof output !== "object") return undefined;
  if ("needsConfirmation" in output && output.needsConfirmation === true) {
    return "confirmation";
  }
  if ("success" in output) return "complete";

  return undefined;
};

export const hasTerminalMutationResult: StopCondition<typeof tools> = ({
  steps,
}) =>
  steps.at(-1)?.toolResults.some((result) => {
    const outputState = getOutputState(result.output);
    if (!outputState) return false;
    if (outputState === "confirmation") {
      return isMutationToolCall(result.toolName, result.input);
    }

    return isTerminalMutationToolCall(result.toolName, result.input);
  }) ?? false;
