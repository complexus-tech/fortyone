import { useMemo } from "react";
import { useTerminology } from "@/hooks/use-terminology-display";

export const useImportTerms = () => {
  const { getTermDisplay } = useTerminology();
  const storyTerm = getTermDisplay("storyTerm");
  const storyTermPlural = getTermDisplay("storyTerm", { variant: "plural" });
  const storyTermCapitalized = getTermDisplay("storyTerm", {
    capitalize: true,
  });
  const objectiveTerm = getTermDisplay("objectiveTerm");
  const objectiveTermPlural = getTermDisplay("objectiveTerm", {
    variant: "plural",
  });
  const objectiveTermPluralCapitalized = getTermDisplay("objectiveTerm", {
    capitalize: true,
    variant: "plural",
  });
  const keyResultTerm = getTermDisplay("keyResultTerm");
  const keyResultTermPlural = getTermDisplay("keyResultTerm", {
    variant: "plural",
  });
  const sprintTerm = getTermDisplay("sprintTerm");
  const sprintTermPlural = getTermDisplay("sprintTerm", { variant: "plural" });
  return useMemo(
    () => ({
      storyTerm,
      storyTermPlural,
      storyTermCapitalized,
      objectiveTerm,
      objectiveTermPlural,
      objectiveTermPluralCapitalized,
      keyResultTerm,
      keyResultTermPlural,
      sprintTerm,
      sprintTermPlural,
    }),
    [
      storyTerm,
      storyTermPlural,
      storyTermCapitalized,
      objectiveTerm,
      objectiveTermPlural,
      objectiveTermPluralCapitalized,
      keyResultTerm,
      keyResultTermPlural,
      sprintTerm,
      sprintTermPlural,
    ],
  );
};
export type ImportTerms = ReturnType<typeof useImportTerms>;
