"use client";
import { Badge, Box, Container, Divider, Text } from "ui";
import { useRef, useState } from "react";
import { addDays, differenceInDays } from "date-fns";
import { cn } from "lib";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { useStatuses } from "@/lib/hooks/statuses";
import { useStoryById } from "@/modules/story/hooks/story";
import { useIsAdminOrOwner } from "@/hooks/owner";
import {
  useFeatures,
  useMediaQuery,
  useTerminology,
  useUserRole,
  useSprintsEnabled,
} from "@/hooks";
import { useMayaAssignee, useMembers } from "@/lib/hooks/members";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { useSprint } from "@/modules/sprints/hooks/sprint-details";
import { useObjective } from "@/modules/objectives/hooks/use-objective";
import { useKeyResults } from "@/modules/objectives/hooks";
import { useUpdateStoryMutation } from "../hooks/update-mutation";
import type { DetailedStory } from "../types";
import { AddLinks } from "./add-links";
import { OptionsHeader } from "./options-header";
import { CoreOptions } from "./story-options/core-options";
import { DateOptions } from "./story-options/date-options";
import { LabelsOption } from "./story-options/labels-option";
import { PlanningOptions } from "./story-options/planning-options";

export const Options = ({
  storyId,
  isNotifications,
  isDialog,
  variant = "sidebar",
}: {
  storyId: string;
  isNotifications: boolean;
  isDialog?: boolean;
  variant?: "sidebar" | "inline";
}) => {
  const { data } = useStoryById(storyId);
  const {
    statusId,
    startDate,
    endDate,
    objectiveId,
    keyResultId,
    assigneeId,
    reporterId,
    teamId,
    labels: storyLabels,
    sprintId,
    deletedAt,
    subStories,
  } = data!;
  const { getTermDisplay } = useTerminology();
  const features = useFeatures();
  const sprintsEnabled = useSprintsEnabled(teamId);
  const isMobile = useMediaQuery("(max-width: 768px)");
  const isCompact = isMobile || variant === "inline";
  const isInline = variant === "inline";
  const [showChildrenDialog, setShowChildrenDialog] = useState(false);
  const pendingStatusIdRef = useRef<string | null>(null);
  const { data: statuses = [] } = useStatuses();
  const { data: members = [] } = useMembers();
  const { hasFeature } = useSubscriptionFeatures();
  const canUseBackgroundMaya = hasFeature("backgroundMaya");
  const { data: mayaAssignee } = useMayaAssignee(canUseBackgroundMaya);
  const { data: sprint } = useSprint(sprintId, teamId);
  const { data: objective } = useObjective(objectiveId, teamId);
  const { data: keyResults = [] } = useKeyResults(
    objectiveId ?? "",
    Boolean(objectiveId && keyResultId),
  );
  const keyResult = keyResults.find(({ id }) => id === keyResultId);
  const objectiveName =
    typeof objective?.name === "string" ? objective.name.trim() || null : null;
  const keyResultName =
    typeof keyResult?.name === "string" ? keyResult.name.trim() || null : null;
  const status =
    statuses.find((state) => state.id === statusId) || statuses.at(0);
  const name = status?.name;
  const isDeleted = Boolean(deletedAt);
  const assignee = data?.assignee ?? members.find((m) => m.id === assigneeId);
  const { mutate } = useUpdateStoryMutation();
  const { isAdminOrOwner } = useIsAdminOrOwner(reporterId);
  const { userRole } = useUserRole();
  const isGuest = userRole === "guest";
  const isEditingDisabled = isDeleted || isGuest;

  const incompleteStatusIds = new Set<string>();
  const completedStatusIds = new Set<string>();
  for (const statusOption of statuses) {
    if (statusOption.category === "completed") {
      completedStatusIds.add(statusOption.id);
    } else if (
      statusOption.category === "started" ||
      statusOption.category === "unstarted" ||
      statusOption.category === "backlog"
    ) {
      incompleteStatusIds.add(statusOption.id);
    }
  }

  const getUndoneChildren = () => {
    const undoneChildrenIds: string[] = [];
    for (const subStory of subStories) {
      if (incompleteStatusIds.has(subStory.statusId)) {
        undoneChildrenIds.push(subStory.id);
      }
    }
    return undoneChildrenIds;
  };

  const handleUpdate = (data: Partial<DetailedStory>) => {
    // If updating status to a "done" state and has undone children
    if (data.statusId && completedStatusIds.has(data.statusId)) {
      const undoneChildrenIds = getUndoneChildren();
      if (undoneChildrenIds.length > 0) {
        pendingStatusIdRef.current = data.statusId;
        setShowChildrenDialog(true);
        return; // Don't update yet, wait for user confirmation
      }
    }

    // Normal update if no confirmation needed
    mutate({ storyId, payload: data });
  };

  const handleConfirmStatusChange = (markChildrenAsDone: boolean) => {
    const pendingStatusId = pendingStatusIdRef.current;
    if (!pendingStatusId) return;

    // Update the main story
    mutate({ storyId, payload: { statusId: pendingStatusId } });

    if (markChildrenAsDone) {
      const undoneChildrenIds = getUndoneChildren();
      // Update all undone children to the same status
      for (const childId of undoneChildrenIds) {
        mutate({ storyId: childId, payload: { statusId: pendingStatusId } });
      }
    }
    // Reset dialog state
    setShowChildrenDialog(false);
    pendingStatusIdRef.current = null;
  };

  return (
    <Box
      className={cn(
        isInline
          ? "h-auto bg-transparent bg-none p-0 md:h-auto md:overflow-visible md:pb-0"
          : "md:bg-surface-muted/50 dark:md:bg-surface-elevated/40 bg-transparent pb-2 md:h-full md:min-h-0 md:overflow-y-auto md:pb-6",
        {
          "dark:md:bg-surface-elevated/80 h-[85dvh]": isDialog,
        },
      )}
    >
      {!isInline ? (
        <Box className="hidden md:block">
          <OptionsHeader
            isAdminOrOwner={isAdminOrOwner}
            isDialog={isDialog}
            storyId={storyId}
          />
        </Box>
      ) : null}
      <Container
        className={cn("text-text-muted px-0.5 pt-4 md:px-6", {
          "px-0 pt-0 md:px-0": isInline,
        })}
      >
        <Box
          className={cn(
            "mb-0 grid grid-cols-[9rem_auto] items-center gap-3 md:mb-6",
            {
              "hidden md:mb-0": isInline && !isDeleted,
            },
          )}
        >
          {!isNotifications && (
            <Text className="hidden md:block" fontWeight="semibold">
              Properties
            </Text>
          )}
          {isDeleted ? (
            <Badge
              className="text-foreground border-opacity-30 bg-opacity-30 px-2"
              color="tertiary"
              size="lg"
            >
              {differenceInDays(
                addDays(new Date(deletedAt!), 30),
                new Date(deletedAt!),
              )}{" "}
              days left in bin
            </Badge>
          ) : null}
        </Box>
        <Box
          className={cn("flex flex-wrap gap-2", {
            "md:block": !isCompact,
          })}
        >
          <CoreOptions
            assignee={assignee}
            canUseBackgroundMaya={canUseBackgroundMaya}
            disabled={isEditingDisabled}
            isCompact={isCompact}
            isNotifications={isNotifications}
            mayaAssigneeId={mayaAssignee?.id}
            members={members}
            onUpdate={handleUpdate}
            statusName={name}
            story={data!}
            storyId={storyId}
          />
          <DateOptions
            disabled={isEditingDisabled}
            endDate={endDate}
            isCompact={isCompact}
            isNotifications={isNotifications}
            onUpdate={handleUpdate}
            startDate={startDate}
            storyTerm={getTermDisplay("storyTerm")}
          />
          <PlanningOptions
            disabled={isEditingDisabled}
            getTermDisplay={getTermDisplay}
            isCompact={isCompact}
            isNotifications={isNotifications}
            keyResultId={keyResultId}
            keyResultName={keyResultName}
            objectiveEnabled={features.objectiveEnabled}
            objectiveId={objectiveId}
            objectiveName={objectiveName}
            onUpdate={handleUpdate}
            sprintId={sprintId}
            sprintName={sprint?.name}
            sprintsEnabled={sprintsEnabled}
            teamId={teamId}
          />
          <LabelsOption
            disabled={isEditingDisabled}
            isCompact={isCompact}
            isNotifications={isNotifications}
            storyId={storyId}
            storyLabels={storyLabels}
            teamId={teamId}
          />
        </Box>

        <Divider className={cn("my-4", { "mt-6 mb-4": isInline })} />
        <Box className={cn({ "flex justify-end": isInline })}>
          <AddLinks storyId={storyId} />
        </Box>
      </Container>

      <ConfirmDialog
        cancelText="No, leave as is"
        confirmText="Yes, mark as done"
        description={`You're about to mark this ${getTermDisplay(
          "storyTerm",
        )} as done. This ${getTermDisplay(
          "storyTerm",
        )} has sub-${getTermDisplay("storyTerm", {
          variant: subStories.length > 1 ? "plural" : "singular",
        })} that are still in progress. Would you like to mark all sub-${getTermDisplay(
          "storyTerm",
          { variant: subStories.length > 1 ? "plural" : "singular" },
        )} as done as well?`}
        hideClose
        isOpen={showChildrenDialog}
        onCancel={() => {
          handleConfirmStatusChange(false);
        }}
        onConfirm={() => {
          handleConfirmStatusChange(true);
        }}
        title={`Mark sub-${getTermDisplay("storyTerm", {
          variant: subStories.length > 1 ? "plural" : "singular",
        })} as done too?`}
      />
    </Box>
  );
};
