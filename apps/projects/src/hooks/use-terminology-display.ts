import { useCallback } from "react";
import { useParams } from "next/navigation";
import { useWorkspaceSettings } from "@/lib/hooks/workspace/settings";
import { teamKeys } from "@/constants/keys";
import { useCachedQueryData } from "@/shared/query-cache/use-cached-query-data";
import type { TeamStoryTerm } from "@/shared/story/terminology";
import type { WorkspaceSettings } from "@/types";
import { useWorkspacePath } from "./use-workspace-path";

type TermKey = "storyTerm" | "sprintTerm" | "objectiveTerm" | "keyResultTerm";

type DisplayOptions = {
  variant?: "singular" | "plural";
  capitalize?: boolean;
};

type GetTermDisplayFn = (termKey: TermKey, options?: DisplayOptions) => string;

const DEFAULT_WORKSPACE_SETTINGS: Pick<WorkspaceSettings, TermKey> = {
  storyTerm: "story",
  sprintTerm: "sprint",
  objectiveTerm: "objective",
  keyResultTerm: "key result",
};

/**
 * Hook for consistent display of terminology throughout the application
 * @returns A function to format terminology terms with options
 */
export const useTerminology = (explicitTeamId?: string | null) => {
  const params = useParams<{ teamId?: string }>();
  const teamId = explicitTeamId === undefined ? params.teamId : explicitTeamId;
  const { workspaceSlug } = useWorkspacePath();
  // Workspace hydration and the Teams owner populate this canonical resource.
  const teams =
    useCachedQueryData<{ id: string; storyTerm?: TeamStoryTerm | null }[]>(
      teamKeys.lists(workspaceSlug),
    ) ?? [];
  const teamStoryTerm = teamId
    ? teams.find((team) => team.id === teamId)?.storyTerm
    : null;
  const { data: terminology = DEFAULT_WORKSPACE_SETTINGS } =
    useWorkspaceSettings();

  const getTermDisplay = useCallback<GetTermDisplayFn>(
    (termKey, options = {}) => {
      const { variant = "singular", capitalize = false } = options;

      // Get the current term value directly using the key
      const currentValue =
        termKey === "storyTerm" && teamStoryTerm
          ? teamStoryTerm
          : terminology[termKey];
      // Handle singular/plural variants
      let result: string = currentValue;

      if (variant === "plural") {
        if (currentValue.endsWith("y")) {
          result = `${currentValue.slice(0, -1)}ies`;
        } else if (currentValue === "focus area") {
          result = "focus areas";
        } else {
          result = `${currentValue}s`;
        }
      }

      // Apply capitalization if requested
      if (capitalize) {
        result = result.charAt(0).toUpperCase() + result.slice(1);
      }

      return result;
    },
    [terminology, teamStoryTerm],
  );

  return { getTermDisplay };
};

export const TEAM_STORY_TERMS: { value: TeamStoryTerm; label: string }[] = [
  { value: "story", label: "Story" },
  { value: "task", label: "Task" },
  { value: "issue", label: "Issue" },
  { value: "ticket", label: "Ticket" },
  { value: "work item", label: "Work item" },
  { value: "deal", label: "Deal" },
];
