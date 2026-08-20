import { useCallback } from "react";
import { useWorkspaceSettings } from "@/lib/hooks/workspace/settings";
import type { WorkspaceSettings } from "@/types";

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
export const useTerminology = () => {
  const { data: terminology = DEFAULT_WORKSPACE_SETTINGS } =
    useWorkspaceSettings();

  const getTermDisplay = useCallback<GetTermDisplayFn>(
    (termKey, options = {}) => {
      const { variant = "singular", capitalize = false } = options;

      // Get the current term value directly using the key
      const currentValue = terminology[termKey];
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
    [terminology],
  );

  return { getTermDisplay };
};
