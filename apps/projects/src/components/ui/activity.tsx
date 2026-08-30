import type { ReactNode } from "react";
import { Box, Flex, Text, TimeAgo, Tooltip } from "ui";
import Link from "next/link";
import { cn } from "lib";
import { EstimateIcon, InfoIcon, SprintsIcon, Time02Icon } from "icons";
import { useSession } from "@/lib/auth/client";
import { formatActivityReasonDates } from "@/lib/activity-format";
import { DEFAULT_ESTIMATE_SCHEME, formatEstimate } from "@/lib/estimate";
import { formatTimeNeeded } from "@/lib/time-needed";
import { useTerminology, useWorkspacePath } from "@/hooks";
import { useLabels } from "@/lib/hooks/labels";
import { useMayaAssignee, useMembers } from "@/lib/hooks/members";
import { useProfile } from "@/lib/hooks/profile";
import { useStatuses } from "@/lib/hooks/statuses";
import { useObjective } from "@/modules/objectives/hooks/use-objective";
import { useSprint } from "@/modules/sprints/hooks/sprint-details";
import type { StoryActivity, StoryPriority } from "@/modules/stories/types";
import { useTeamSettings } from "@/modules/teams/hooks/use-team-settings";
import type { Label } from "@/types";
import {
  formatScheduleActivityValue,
  getActivityCopy,
  getActivityValueIds,
  getDisplayActivityReason,
} from "./activity-copy";
import { ActivityActor } from "./activity-actor";
import {
  getActivityFieldMeta,
  getLabelActivityDisplayValue,
} from "./activity-field-renderers";
import { ActivityUpdateSegments } from "./activity-update-segments";
import { PriorityIcon } from "./priority-icon";

const DisplayEstimate = ({
  value,
  teamId,
}: {
  value: string;
  teamId?: string;
}) => {
  const { data: teamSettings } = useTeamSettings(teamId);
  const estimateValue = Number.parseInt(value, 10);
  const estimateScheme =
    teamSettings?.estimationSettings.scheme ?? DEFAULT_ESTIMATE_SCHEME;

  return (
    <span className="flex items-center gap-1">
      <EstimateIcon className="h-5" />
      {Number.isNaN(estimateValue)
        ? "No complexity"
        : formatEstimate(estimateScheme, estimateValue, "full")}
    </span>
  );
};

const DisplayTimeNeeded = ({ value }: { value: string }) => {
  const minutes = Number.parseInt(value, 10);
  return (
    <span className="flex items-center gap-1">
      <Time02Icon className="h-5" />
      {formatTimeNeeded(Number.isNaN(minutes) ? null : minutes, "full")}
    </span>
  );
};

const DisplaySprint = ({
  sprintId,
  teamId,
}: {
  sprintId: string;
  teamId?: string;
}) => {
  const { data: sprint } = useSprint(sprintId, teamId);
  const { withWorkspace } = useWorkspacePath();
  return (
    <>
      {!sprintId || sprintId.includes("nil") ? (
        <span>No sprint</span>
      ) : (
        <Link
          className="flex items-center gap-1"
          href={withWorkspace(
            `/teams/${sprint?.teamId}/sprints/${sprintId}/stories`,
          )}
        >
          <SprintsIcon className="h-5" />
          {sprint?.name}
        </Link>
      )}
    </>
  );
};

const DisplayObjective = ({
  objectiveId,
  teamId,
}: {
  objectiveId: string;
  teamId?: string;
}) => {
  const { data: objective } = useObjective(objectiveId, teamId);
  const { withWorkspace } = useWorkspacePath();
  return (
    <>
      {!objectiveId || objectiveId.includes("nil") ? (
        <span>No objective</span>
      ) : (
        <Link
          href={withWorkspace(
            `/teams/${objective?.teamId}/objectives/${objectiveId}`,
          )}
        >
          {objective?.name}
        </Link>
      )}
    </>
  );
};

const getActivityVerb = (type: StoryActivity["type"], storyTerm: string) => {
  if (type === "create") return `created the ${storyTerm}`;
  if (type === "link") return "linked";
  return "changed";
};

export const Activity = ({
  avatarSurfaceClassName,
  teamId,
  field,
  currentValue,
  type,
  createdAt,
  user,
  newValue,
  oldValue,
  reason,
}: StoryActivity & { avatarSurfaceClassName?: string; teamId?: string }) => {
  const { data: members = [] } = useMembers();
  const { data: mayaAssignee } = useMayaAssignee();
  const { data: statuses = [] } = useStatuses();
  const { data: allLabels = [] } = useLabels();
  const { data: profile } = useProfile();
  const { withWorkspace } = useWorkspacePath();
  const { getTermDisplay } = useTerminology();
  const { data: session } = useSession();
  const storyTerm = getTermDisplay("storyTerm");
  const member = user;
  const isSelfActivity = session?.user.id === member.id;
  const actorDisplayName = isSelfActivity ? "You" : member.fullName;
  const actorDisplayUsername = isSelfActivity ? "you" : member.username;
  const activityAssignees = mayaAssignee
    ? [...members.filter(({ id }) => id !== mayaAssignee.id), mayaAssignee]
    : members;
  const activityVerb = getActivityVerb(type, storyTerm);
  const activityReason = formatActivityReasonDates(
    getDisplayActivityReason(reason),
  );
  const isLinkedUrl =
    type === "link" &&
    currentValue &&
    typeof newValue === "string" &&
    newValue.startsWith("http");
  let linkedValue: ReactNode = null;

  if (type === "link" && currentValue) {
    if (isLinkedUrl && typeof newValue === "string") {
      linkedValue = (
        <a
          className="inline-block min-w-0 truncate text-sm text-black underline md:text-[0.95rem] dark:text-white"
          href={newValue}
          rel="noopener noreferrer"
          target="_blank"
        >
          {currentValue}
        </a>
      );
    } else {
      linkedValue = (
        <Text
          as="span"
          className="inline-block min-w-0 truncate text-sm text-black md:text-[0.95rem] dark:text-white"
          fontWeight="medium"
        >
          {currentValue}
        </Text>
      );
    }
  }

  if (field === "completed_at") {
    return null;
  }

  const fieldMeta = getActivityFieldMeta(field, {
    activityAssignees,
    renderEstimate: (value) => (
      <DisplayEstimate teamId={teamId} value={value} />
    ),
    renderObjective: (value) => (
      <DisplayObjective objectiveId={value} teamId={teamId} />
    ),
    renderPriority: (value) => (
      <span className="flex items-center gap-1">
        <PriorityIcon className="h-5" priority={value as StoryPriority} />
        {value}
      </span>
    ),
    renderSprint: (value) => <DisplaySprint sprintId={value} teamId={teamId} />,
    renderTimeNeeded: (value) => <DisplayTimeNeeded value={value} />,
    statuses,
    withWorkspace,
  });
  const activityLabels =
    field === "labels"
      ? getActivityValueIds(newValue)
          .map((labelId) => allLabels.find((label) => label.id === labelId))
          .filter((label): label is Label => Boolean(label))
      : [];
  const memberById = new Map(members.map((member) => [member.id, member]));
  const collaboratorIdsFromNewValue = getActivityValueIds(newValue);
  let activityCollaboratorIds: string[] = [];
  if (field === "collaborator_ids") {
    activityCollaboratorIds =
      collaboratorIdsFromNewValue.length > 0
        ? collaboratorIdsFromNewValue
        : getActivityValueIds(currentValue);
  }
  const activityCollaborators =
    field === "collaborator_ids"
      ? activityCollaboratorIds.flatMap((userId) => {
          const collaborator = memberById.get(userId);
          return collaborator ? [collaborator] : [];
        })
      : [];
  let collaboratorDisplayValue = "No collaborators";
  if (activityCollaboratorIds.length === 1) {
    collaboratorDisplayValue =
      activityCollaborators[0]?.fullName ||
      activityCollaborators[0]?.username ||
      "1 collaborator";
  } else if (activityCollaboratorIds.length > 1) {
    collaboratorDisplayValue = `${activityCollaboratorIds.length} collaborators`;
  }

  let displayCurrentValue = currentValue;
  if (field === "labels" && activityLabels.length > 0) {
    displayCurrentValue = getLabelActivityDisplayValue(activityLabels);
  } else if (field === "collaborator_ids") {
    displayCurrentValue = collaboratorDisplayValue;
  } else if (field === "auto_scheduling_time") {
    displayCurrentValue = formatScheduleActivityValue(
      currentValue,
      newValue,
      profile?.timezone,
    );
  }
  const activityCopy = getActivityCopy({
    currentValue: displayCurrentValue,
    field,
    fieldLabel: fieldMeta.label,
    oldValue,
    reason,
    storyTerm,
    type,
  });
  return (
    <Box className="relative pb-2 last-of-type:pb-0 md:pb-4">
      <Box
        className={cn(
          "border-border pointer-events-none absolute top-0 left-4 z-0 h-full border-l border-dashed",
        )}
      />
      <Flex align="center" className="z-1 min-w-0" gap={1}>
        <ActivityActor
          avatarSurfaceClassName={avatarSurfaceClassName}
          displayName={actorDisplayName}
          displayUsername={actorDisplayUsername}
          isSelfActivity={isSelfActivity}
          member={member}
          withWorkspace={withWorkspace}
        />
        <Box className="flex min-w-0 flex-1 items-center gap-1 overflow-hidden text-sm md:text-[0.95rem]">
          {type === "update" ? (
            <ActivityUpdateSegments
              activityLabels={activityLabels}
              currentValue={currentValue}
              field={field}
              fieldMeta={fieldMeta}
              segments={activityCopy.segments}
            />
          ) : (
            <Text as="span" className="text-sm md:text-[0.95rem]" color="muted">
              {activityVerb}
            </Text>
          )}
          {linkedValue}
          {activityReason ? (
            <Tooltip title={activityReason}>
              <span className="inline-flex shrink-0 cursor-help items-center">
                <InfoIcon className="text-icon-muted h-4" />
              </span>
            </Tooltip>
          ) : null}
          <Text
            as="span"
            className="mx-0.5 shrink-0 text-sm md:text-[0.95rem]"
            color="muted"
          >
            ·
          </Text>
          <Text
            as="span"
            className="shrink-0 text-sm md:text-[0.95rem]"
            color="muted"
          >
            <TimeAgo timestamp={createdAt} />
          </Text>
        </Box>
      </Flex>
    </Box>
  );
};
