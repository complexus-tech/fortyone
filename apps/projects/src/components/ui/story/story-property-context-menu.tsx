"use client";

import { ContextMenu } from "ui";
import { useStatuses } from "@/lib/hooks/statuses";
import { useUpdateStoryMutation } from "@/modules/story/hooks/update-mutation";
import type { Story } from "@/modules/stories/types";
import { PriorityIcon } from "../priority-icon";
import { StoryStatusIcon } from "../story-status-icon";
import { ContextMenuItem } from "./context-menu-item";
import { ObjectiveKeyResultContextSubMenu } from "./objective-key-result-menu";
import { STORY_PRIORITIES } from "./story-action-options";

export const StoryPropertyContextMenu = ({
  disabled,
  story,
}: {
  disabled: boolean;
  story: Story;
}) => {
  const updateStory = useUpdateStoryMutation();
  const { data: statuses = [] } = useStatuses();
  const teamStatuses = statuses.filter(
    (status) => status.teamId === story.teamId,
  );

  return (
    <>
      <ContextMenu.Group>
        <ContextMenuItem
          disabled={disabled}
          icon={<StoryStatusIcon statusId={story.statusId} />}
          label="Status"
          subMenu={teamStatuses.map((status) => ({
            active: status.id === story.statusId,
            icon: <StoryStatusIcon statusId={status.id} />,
            label: status.name,
            onSelect: () => {
              updateStory.mutate({
                payload: { statusId: status.id },
                storyId: story.id,
              });
            },
          }))}
        />
        <ContextMenuItem
          disabled={disabled}
          icon={<PriorityIcon priority={story.priority} />}
          label="Priority"
          subMenu={STORY_PRIORITIES.map((priority) => ({
            active: priority === story.priority,
            icon: <PriorityIcon priority={priority} />,
            label: priority,
            onSelect: () => {
              updateStory.mutate({
                payload: { priority },
                storyId: story.id,
              });
            },
          }))}
        />
        <ObjectiveKeyResultContextSubMenu
          disabled={disabled}
          keyResultId={story.keyResultId}
          objectiveId={story.objectiveId}
          onChange={(payload) => {
            updateStory.mutate({ payload, storyId: story.id });
          }}
          teamId={story.teamId}
        />
      </ContextMenu.Group>
      <ContextMenu.Separator className="my-2" />
    </>
  );
};
