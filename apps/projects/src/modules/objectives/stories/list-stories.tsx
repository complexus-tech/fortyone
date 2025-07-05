"use client";
import { useLocalStorage } from "@/hooks";
import type { StoriesLayout } from "@/components/ui";
import { useObjectiveStoriesGrouped } from "@/modules/stories/hooks/use-objective-stories-grouped";
import { useObjective } from "../hooks";
import { ObjectiveOptionsProvider, useObjectiveOptions } from "./provider";
import { AllStories } from "./all-stories";
import { Header } from "./header";
import { ObjectivePageSkeleton } from "./objective-page-skeleton";

export const ListStories = ({ objectiveId }: { objectiveId: string }) => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "teams:objectives:stories:layout",
    "list",
  );
  const { viewOptions } = useObjectiveOptions();
  const { isPending: isObjectivePending } = useObjective(objectiveId);
  const { isPending: isStoriesPending, data: groupedStories } =
    useObjectiveStoriesGrouped(objectiveId, viewOptions.groupBy, {
      orderBy: viewOptions.orderBy,
    });

  if (isObjectivePending || isStoriesPending) {
    return <ObjectivePageSkeleton layout={layout} />;
  }

  return (
    <ObjectiveOptionsProvider>
      <Header layout={layout} setLayout={setLayout} />
      <AllStories groupedStories={groupedStories} layout={layout} />
    </ObjectiveOptionsProvider>
  );
};
