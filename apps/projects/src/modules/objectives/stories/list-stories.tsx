"use client";

import type { StoriesLayout } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { ObjectiveOptionsProvider } from "./provider";
import { AllStories } from "./all-stories";

export const ListStories = ({ objectiveId }: { objectiveId: string }) => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "teams:objectives:stories:layout",
    "list",
  );
  return (
    <ObjectiveOptionsProvider layout={layout}>
      <AllStories
        layout={layout}
        objectiveId={objectiveId}
        setLayout={setLayout}
      />
    </ObjectiveOptionsProvider>
  );
};
