import type { ToolExecutionOptions, ToolSet } from "ai";
import type { MayaToolName } from "@/lib/ai/tool-names";

// Native renders canonical approval cards for these task operations. Desktop
// navigation, billing and unrelated management tools are intentionally absent.
export const MOBILE_MAYA_TOOL_NAMES = new Set<string>([
  "members",
  "resolveMember",
  "search",
  "listTeams",
  "listTeamMembers",
  "listTeamStories",
  "searchStories",
  "focusBrief",
  "getStoryDetails",
  "createStory",
  "updateStory",
  "deleteStory",
  "statuses",
  "labels",
  "storyLabels",
  "comments",
  "storyActivities",
  "listSprints",
  "listRunningSprints",
  "listObjectivesTool",
  "listTeamObjectivesTool",
  "listMemories",
  "suggestions",
] satisfies MayaToolName[]);

const READ_ACTIONS: Readonly<Partial<Record<string, readonly string[]>>> = {
  statuses: ["list-all-statuses", "list-team-statuses", "get-status-details"],
  labels: ["list-labels"],
};

export const canUseMobileMayaTool = (name: string, input: unknown) => {
  if (!MOBILE_MAYA_TOOL_NAMES.has(name)) return false;
  const actions = READ_ACTIONS[name];
  if (!actions) return true;
  return Boolean(
    input &&
      typeof input === "object" &&
      actions.includes(Reflect.get(input, "action") as string),
  );
};

export const mobileMayaTools = <T extends ToolSet>(tools: T): T =>
  Object.fromEntries(
    Object.entries(tools)
      .filter(([name]) => MOBILE_MAYA_TOOL_NAMES.has(name))
      .map(([name, registered]) => {
        const actions = READ_ACTIONS[name];
        if (!actions) return [name, registered];
        const execute = registered.execute as
          | NonNullable<ToolSet[string]["execute"]>
          | undefined;
        return [
          name,
          {
            ...registered,
            description: `Mobile supports only these read operations: ${actions.join(", ")}. Workflow or label management is available in the web app.`,
            execute: (input: unknown, options: ToolExecutionOptions) => {
              if (!canUseMobileMayaTool(name, input))
                return {
                  success: false,
                  error: "This operation is available in the web app.",
                };
              return execute?.(input, options);
            },
          },
        ];
      }),
  ) as T;
