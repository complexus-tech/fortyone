"use client";
import { useLocalStorage } from "@/hooks";
import type { StoriesLayout } from "@/components/ui";
import { useObjectiveStories } from "@/modules/stories/hooks/objective-stories";
import { ObjectiveOptionsProvider } from "./provider";
import { AllStories } from "./all-stories";
import { Header } from "./header";
import { ObjectivePageSkeleton } from "./objective-page-skeleton";

export const ListStories = ({ objectiveId }: { objectiveId: string }) => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "teams:objectives:stories:layout",
    "kanban",
  );
  const { isPending } = useObjectiveStories(objectiveId);

  if (isPending) {
    return <ObjectivePageSkeleton layout={layout} />;
  }

  return (
    <ObjectiveOptionsProvider>
      <Header layout={layout} setLayout={setLayout} />
      <AllStories layout={layout} />
    </ObjectiveOptionsProvider>
  );
};
