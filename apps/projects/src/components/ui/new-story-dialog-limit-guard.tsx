import type { ReactNode } from "react";
import { useTerminology } from "@/hooks/use-terminology-display";
import { FeatureGuard } from "./feature-guard";
import { NewStoryDialogLimit } from "./new-story-dialog-limit";

export const NewStoryDialogLimitGuard = ({
  billingHref,
  canUpgrade,
  children,
  isOpen,
  maxStories,
  onClose,
  planName,
  totalStories,
}: {
  billingHref: string;
  canUpgrade: boolean;
  children: ReactNode;
  isOpen: boolean;
  maxStories: number;
  onClose: () => void;
  planName: string;
  totalStories: number;
}) => {
  const { getTermDisplay } = useTerminology();

  return (
    <FeatureGuard
      count={totalStories}
      fallback={
        <NewStoryDialogLimit
          billingHref={billingHref}
          canUpgrade={canUpgrade}
          isOpen={isOpen}
          maxStories={maxStories}
          onClose={onClose}
          planName={planName}
          storyTerm={getTermDisplay("storyTerm", { capitalize: true })}
          storyTermPlural={getTermDisplay("storyTerm", {
            variant: "plural",
          })}
          storyTermPluralCapitalized={getTermDisplay("storyTerm", {
            capitalize: true,
            variant: "plural",
          })}
          totalStories={totalStories}
        />
      }
      feature="maxStories"
    >
      {children}
    </FeatureGuard>
  );
};
