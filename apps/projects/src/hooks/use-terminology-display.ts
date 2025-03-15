import { useTerminology } from "@/lib/hooks/terminology/terminology";
import type { Terminology } from "@/types";

type TermKey = keyof Terminology;

type DisplayOptions = {
  variant?: "singular" | "plural";
};

/**
 * Hook for consistent display of terminology throughout the application
 * @param termKey - The database key for the term (e.g., "storyTerm")
 * @param options - Display options for term variant (singular/plural)
 * @returns Formatted terminology string
 */
export const useTerminologyDisplay = (
  termKey: TermKey,
  options: DisplayOptions = {},
): string => {
  const {
    data: terminology = {
      storyTerm: "story",
      sprintTerm: "sprint",
      objectiveTerm: "objective",
      keyResultTerm: "key result",
    },
  } = useTerminology();

  const { variant = "singular" } = options;

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

  return result;
};
