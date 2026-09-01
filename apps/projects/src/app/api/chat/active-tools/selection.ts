import type { UIMessage } from "ai";
import {
  MAYA_TOOL_NAMES,
  type MayaToolName,
} from "@/lib/ai/tool-names";

/**
 * Maya's capability contract is deliberately independent of user wording.
 *
 * The model is multilingual and users can change subjects at any point in a
 * conversation. Filtering tool schemas with phrase matching makes capability
 * availability depend on language, inflection, and stale conversational state.
 * Instead, every registered tool is declared on every model turn. Tool schemas,
 * mutation approvals, authentication, and server-side authorization remain the
 * enforcement boundaries for what can actually execute.
 */
const UNIVERSAL_ACTIVE_TOOLS: readonly MayaToolName[] = MAYA_TOOL_NAMES;

export type ActiveToolPlan = {
  activeTools: MayaToolName[];
  source: "universal";
};

export const selectActiveToolPlan = (_input: {
  currentPath?: string;
  messages: UIMessage[];
  storyTerminology?: string;
}): ActiveToolPlan => ({
  // Return a fresh array because the AI SDK accepts a mutable list and route
  // callers must never be able to alter the canonical registry.
  activeTools: [...UNIVERSAL_ACTIVE_TOOLS],
  source: "universal",
});

export const selectActiveTools = (
  input: Parameters<typeof selectActiveToolPlan>[0],
) => selectActiveToolPlan(input).activeTools;
