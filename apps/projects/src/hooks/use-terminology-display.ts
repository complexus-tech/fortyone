import { useCallback } from "react";
import { useTerminology } from "@/lib/hooks/terminology/terminology";
import type { Terminology } from "@/types";

type TermKey = keyof Terminology;

type DisplayOptions = {
  variant?: "singular" | "plural";
  capitalize?: boolean;
};

type GetTermDisplayFn = (termKey: TermKey, options?: DisplayOptions) => string;

/**
 * Hook for consistent display of terminology throughout the application
 * @returns A function to format terminology terms with options
 */
export const useTerminologyDisplay = () => {
  const {
    data: terminology = {
      storyTerm: "story",
      sprintTerm: "sprint",
      objectiveTerm: "objective",
      keyResultTerm: "key result",
    },
  } = useTerminology();

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
