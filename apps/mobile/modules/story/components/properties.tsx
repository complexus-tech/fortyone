import React, { useState } from "react";
import { GlassSurface } from "@/components/ui/glass-surface";
import { PropertyExpandButton } from "./properties/property-expand-button";
import { Row, Col } from "@/components/ui";
import { DetailedStory, Story } from "@/modules/stories/types";
import { useFeatures, useSprintsEnabled } from "@/hooks";
import { PriorityBadge } from "./properties/priority";
import { StatusBadge } from "./properties/status";
import { AssigneeBadge } from "./properties/assignee";
import { ObjectiveBadge } from "./properties/objective";
import { SprintBadge } from "./properties/sprint";
import { LabelsBadge } from "./properties/labels";
import { StartDateBadge } from "./properties/start-date";
import { EndDateBadge } from "./properties/end-date";
import { useUpdateStoryMutation } from "../hooks/use-update-story-mutation";
import { useUpdateLabelsMutation } from "../hooks/use-update-labels-mutation";
import { formatISO } from "date-fns";

export const Properties = ({ story }: { story: Story }) => {
  const [expanded, setExpanded] = useState(false);
  const sprintsEnabled = useSprintsEnabled(story.teamId);
  const { objectiveEnabled } = useFeatures();

  const updateStoryMutation = useUpdateStoryMutation();
  const updateLabelsMutation = useUpdateLabelsMutation();

  const disabled =
    updateStoryMutation.isPending ||
    updateLabelsMutation.isPending ||
    Boolean(story.deletedAt);

  const handleUpdate = async (data: Partial<DetailedStory>) => {
    await updateStoryMutation.mutateAsync({
      storyId: story.id,
      payload: data,
    });
  };

  const onLabelsChange = async (labelIds: string[]) => {
    await updateLabelsMutation.mutateAsync({
      storyId: story.id,
      labels: labelIds,
    });
  };

  return (
    <Col asContainer align="stretch" className="my-[8px]">
      <GlassSurface cornerRadius={20} style={{ padding: 8 }}>
        <Row
          wrap
          align="center"
          style={{ columnGap: 4, rowGap: 0, minWidth: 0, maxWidth: "100%" }}
        >
          <StatusBadge
            disabled={disabled}
            story={story}
            onStatusChange={(statusId) => handleUpdate({ statusId })}
          />
          <PriorityBadge
            disabled={disabled}
            priority={story.priority || "No Priority"}
            onPriorityChange={(priority) => handleUpdate({ priority })}
          />
          <AssigneeBadge
            disabled={disabled}
            story={story}
            onAssigneeChange={(assigneeId) => handleUpdate({ assigneeId })}
          />
          {(expanded || Boolean(story.labels?.length)) && (
            <LabelsBadge
              disabled={disabled}
              story={story}
              onLabelsChange={onLabelsChange}
            />
          )}
          {sprintsEnabled && (expanded || story.sprintId) && (
            <SprintBadge
              disabled={disabled}
              story={story}
              onSprintChange={(sprintId) => handleUpdate({ sprintId })}
            />
          )}
          {objectiveEnabled && (expanded || story.objectiveId) && (
            <ObjectiveBadge
              disabled={disabled}
              story={story}
              onObjectiveChange={(objectiveId) => handleUpdate({ objectiveId })}
            />
          )}
          {(expanded || story.startDate) && (
            <StartDateBadge
              disabled={disabled}
              story={story}
              onStartDateChange={(startDate) =>
                handleUpdate({
                  startDate: startDate
                    ? formatISO(startDate, { representation: "date" })
                    : null,
                })
              }
            />
          )}
          {(expanded || story.endDate) && (
            <EndDateBadge
              disabled={disabled}
              story={story}
              onEndDateChange={(endDate) =>
                handleUpdate({
                  endDate: endDate
                    ? formatISO(endDate, { representation: "date" })
                    : null,
                })
              }
            />
          )}
          <PropertyExpandButton
            expanded={expanded}
            onPress={() => setExpanded((value) => !value)}
          />
        </Row>
      </GlassSurface>
    </Col>
  );
};
