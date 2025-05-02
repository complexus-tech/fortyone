"use client";
import { useLocalStorage } from "@/hooks";
import type { StoriesLayout } from "@/components/ui";
import { useObjectiveStories } from "@/modules/stories/hooks/objective-stories";
import { useObjective } from "../hooks";
import { ObjectiveOptionsProvider } from "./provider";
import { AllStories } from "./all-stories";
import { Header } from "./header";
import { ObjectivePageSkeleton } from "./objective-page-skeleton";

export const ListStories = ({ objectiveId }: { objectiveId: string }) => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "teams:objectives:stories:layout",
    "list",
  );
  const { isPending: isObjectivePending } = useObjective(objectiveId);
  const { isPending: isStoriesPending } = useObjectiveStories(objectiveId);

  if (isObjectivePending || isStoriesPending) {
    return <ObjectivePageSkeleton layout={layout} />;
  }

  return (
    <ObjectiveOptionsProvider>
      <Header layout={layout} setLayout={setLayout} />
      <AllStories layout={layout} />
    </ObjectiveOptionsProvider>
  );
};
