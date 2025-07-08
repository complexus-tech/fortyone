"use client";

import { ObjectiveOptionsProvider } from "./provider";
import { AllStories } from "./all-stories";

export const ListStories = ({ objectiveId }: { objectiveId: string }) => {
  return (
    <ObjectiveOptionsProvider>
      <AllStories objectiveId={objectiveId} />
    </ObjectiveOptionsProvider>
  );
};
