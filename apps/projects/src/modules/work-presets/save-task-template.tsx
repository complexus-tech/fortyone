"use client";

import type { DetailedStory } from "@/shared/story/types";
import { useStoryCustomFields } from "@/modules/custom-fields/public/client";
import { buildTaskTemplate } from "@/shared/story/task-template";
import { usePresetMutations } from "./hooks";
import { PresetNameDialog } from "./name-dialog";

export const SaveTaskTemplate = ({
  story,
  onOpenChange,
}: {
  story: DetailedStory;
  onOpenChange: (open: boolean) => void;
}) => {
  const { create } = usePresetMutations(story.teamId, "template");
  const fields = useStoryCustomFields(story.id);
  return (
    <PresetNameDialog
      initialName={story.title}
      onOpenChange={onOpenChange}
      onSave={(name, visibility) => {
        if (fields.isPending)
          throw new Error(
            "The task's custom fields are still loading. Try again shortly.",
          );
        if (fields.isError)
          throw new Error(
            "The task's custom fields could not be loaded. Close this dialog and retry.",
          );
        return create.mutateAsync({
          teamId: story.teamId,
          kind: "template",
          name,
          visibility,
          configuration: buildTaskTemplate(story, fields.data.values),
        });
      }}
      open
      shared
      title="Save as task template"
    />
  );
};
