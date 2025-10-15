import React from "react";
import { Row, Text, Col } from "@/components/ui";
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
import { useBottomSheetModal } from "@gorhom/bottom-sheet";
import { formatISO } from "date-fns";

export const Properties = ({ story }: { story: Story }) => {
  const { dismiss } = useBottomSheetModal();
  const sprintsEnabled = useSprintsEnabled(story.teamId);
  const { objectiveEnabled } = useFeatures();

  const updateStoryMutation = useUpdateStoryMutation();
  const updateLabelsMutation = useUpdateLabelsMutation();

  const handleUpdate = (data: Partial<DetailedStory>) => {
    updateStoryMutation.mutate({
      storyId: story.id,
      payload: data,
    });
    dismiss();
  };

  const onLabelsChange = (labelIds: string[]) => {
    updateLabelsMutation.mutate({
      storyId: story.id,
      labels: labelIds,
    });
  };

  return (
    <Row asContainer>
      <Col className="my-4 ">
        <Text className="mb-2.5 opacity-80">Properties</Text>
        <Row wrap gap={2}>
          <StatusBadge
            story={story}
            onStatusChange={(statusId) => handleUpdate({ statusId })}
          />
          <PriorityBadge
            priority={story.priority || "No Priority"}
            onPriorityChange={(priority) => handleUpdate({ priority })}
          />
          <AssigneeBadge
            story={story}
            onAssigneeChange={(assigneeId) => handleUpdate({ assigneeId })}
          />
          {sprintsEnabled && (
            <SprintBadge
              story={story}
              onSprintChange={(sprintId) => handleUpdate({ sprintId })}
            />
          )}
          {objectiveEnabled && (
            <ObjectiveBadge
              story={story}
              onObjectiveChange={(objectiveId) => handleUpdate({ objectiveId })}
            />
          )}
          <LabelsBadge story={story} onLabelsChange={onLabelsChange} />
          <StartDateBadge
            story={story}
            onStartDateChange={(startDate) =>
              handleUpdate({
                startDate: startDate
                  ? formatISO(startDate, { representation: "date" })
                  : null,
              })
            }
          />
          <EndDateBadge
            story={story}
            onEndDateChange={(endDate) =>
              handleUpdate({
                endDate: endDate
                  ? formatISO(endDate, { representation: "date" })
                  : null,
              })
            }
          />
        </Row>
      </Col>
    </Row>
  );
};
