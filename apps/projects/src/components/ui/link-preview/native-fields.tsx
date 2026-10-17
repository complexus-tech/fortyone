"use client";

import type { ReactNode } from "react";
import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Avatar, Button, Input } from "ui";
import { EditIcon } from "icons";
import { useSession } from "@/lib/auth/client";
import { useUserRole } from "@/hooks/role";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { memberKeys, statusKeys, teamKeys } from "@/constants/keys";
import { getJoinedTeams } from "@/modules/teams/queries/get-teams";
import { getStatuses } from "@/lib/queries/states/get-states";
import { getTeamMembers } from "@/lib/queries/members/get-members";
import { getObjectiveStatuses } from "@/modules/objectives/queries/statuses";
import { objectiveKeys } from "@/shared/objectives/keys";
import type { StoryPriority, StoryUpdate } from "@/shared/story/types";
import type { ObjectiveUpdate } from "@/modules/objectives/types";
import { useUpdateStoryMutation } from "@/modules/story/hooks/update-mutation";
import { useUpdateObjectiveMutation } from "@/modules/objectives/hooks/update-mutation";
import { StatusesMenu } from "../story/statuses-menu";
import { PrioritiesMenu } from "../story/priorities-menu";
import { AssigneesMenu } from "../story/assignees-menu";
import { ObjectiveStatusesMenu } from "../objective-statuses-menu";
import { PriorityIcon } from "../priority-icon";
import type {
  NativeStorySummary,
  NativeObjectiveSummary,
} from "./native-query";

export const canEditNativeWork = ({
  role,
  sameWorkspace,
  joinedTeam,
  archived = false,
}: {
  role?: string;
  sameWorkspace: boolean;
  joinedTeam: boolean;
  archived?: boolean;
}) =>
  sameWorkspace &&
  joinedTeam &&
  !archived &&
  (role === "admin" || role === "member");

const useFields = (teamId: string, kind: "story" | "objective") => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const { userRole } = useUserRole();
  const ctx = { session, workspaceSlug };
  const teams = useQuery({
    queryKey: teamKeys.joined(workspaceSlug),
    queryFn: () => getJoinedTeams(ctx),
    enabled: Boolean(session),
    staleTime: 60_000,
  });
  const states = useQuery({
    queryKey: statusKeys.lists(workspaceSlug),
    queryFn: () => getStatuses(ctx),
    enabled: Boolean(session) && kind === "story",
    staleTime: 60_000,
  });
  const objectiveStates = useQuery({
    queryKey: objectiveKeys.statuses(workspaceSlug),
    queryFn: () => getObjectiveStatuses(ctx),
    enabled: Boolean(session) && kind === "objective",
    staleTime: 60_000,
  });
  const members = useQuery({
    queryKey: [...memberKeys.team(workspaceSlug, teamId), ""],
    queryFn: () => getTeamMembers(teamId, ctx),
    enabled: Boolean(session && teamId),
    staleTime: 60_000,
  });
  const storyStates = states.isError
    ? []
    : states.data?.filter((state) => state.teamId === teamId) ?? [];
  const objectiveStatusItems = objectiveStates.isError
    ? []
    : objectiveStates.data ?? [];
  return {
    canEdit: canEditNativeWork({
      role: userRole,
      sameWorkspace: true,
      joinedTeam:
        !teams.isError &&
        Boolean(teams.data?.some((team) => team.id === teamId)),
    }),
    states: kind === "story" ? storyStates : objectiveStatusItems,
    members: members.isError ? [] : members.data ?? [],
  };
};

const EditableTitle = ({
  title,
  canEdit,
  pending,
  onSave,
}: {
  title: string;
  canEdit: boolean;
  pending: boolean;
  onSave: (title: string, onSuccess: () => void) => void;
}) => {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(title);
  if (editing && canEdit)
    return (
      <form
        className="space-y-2"
        onSubmit={(event) => {
          event.preventDefault();
          const value = draft.trim();
          if (!value || pending) return;
          if (value === title) setEditing(false);
          else
            onSave(value, () => {
              setEditing(false);
            });
        }}
      >
        <Input
          aria-label="Title"
          autoFocus
          maxLength={500}
          onChange={(event) => {
            setDraft(event.target.value);
          }}
          onKeyDown={(event) => {
            if (event.key === "Escape") {
              event.preventDefault();
              setEditing(false);
            }
          }}
          value={draft}
        />
        <div className="flex gap-2">
          <Button disabled={!draft.trim() || pending} size="sm" type="submit">
            {pending ? "Saving…" : "Save"}
          </Button>
          <Button
            color="tertiary"
            disabled={pending}
            onClick={() => {
              setEditing(false);
            }}
            size="sm"
            type="button"
            variant="naked"
          >
            Cancel
          </Button>
        </div>
      </form>
    );
  return (
    <div className="flex items-start gap-2">
      <p className="min-w-0 flex-1 font-semibold [overflow-wrap:anywhere]">
        {title}
      </p>
      {canEdit ? (
        <Button
          aria-label="Edit title"
          asIcon
          color="tertiary"
          disabled={pending}
          onClick={() => {
            setDraft(title);
            setEditing(true);
          }}
          size="sm"
          variant="naked"
        >
          <EditIcon className="size-4" />
        </Button>
      ) : null}
    </div>
  );
};

const Property = ({
  label,
  children,
}: {
  label: string;
  children: ReactNode;
}) => (
  <div className="min-w-0">
    <p className="text-text-muted mb-1">{label}</p>
    {children}
  </div>
);

const StatusMark = ({ color }: { color?: string }) => (
  <span
    aria-hidden
    className="bg-text-muted size-2.5 shrink-0 rounded-full"
    style={color ? { backgroundColor: color } : undefined}
  />
);

const WorkFields = ({
  states,
  statusId,
  priority,
  assigneeId,
  assigneeName,
  assigneeAvatar,
  teamId,
  kind,
  canEdit,
  pending,
  onStatus,
  onPriority,
  onAssignee,
}: {
  states: { id: string; name: string; color: string }[];
  statusId: string;
  priority: StoryPriority;
  assigneeId: string | null;
  assigneeName?: string;
  assigneeAvatar?: string | null;
  teamId: string;
  kind: "story" | "objective";
  canEdit: boolean;
  pending: boolean;
  onStatus: (id: string) => void;
  onPriority: (priority: StoryPriority) => void;
  onAssignee: (id: string | null) => void;
}) => {
  const state = states.find((item) => item.id === statusId);
  const statusContent = (
    <>
      <StatusMark color={state?.color} />
      <span className="truncate">{state?.name ?? "Status unavailable"}</span>
    </>
  );
  const priorityContent = (
    <>
      <PriorityIcon priority={priority} />
      <span>{priority}</span>
    </>
  );
  const assigneeContent = (
    <>
      <Avatar name={assigneeName} size="sm" src={assigneeAvatar} />
      <span className="truncate">
        {assigneeName ?? (assigneeId ? "Assigned" : "Unassigned")}
      </span>
    </>
  );
  const fieldClass = "flex min-h-8 items-center gap-2";
  return (
    <div className="pointer-events-none mt-3 flex flex-wrap gap-x-4 gap-y-2">
      <Property label="Status">
        {canEdit && states.length > 0 && kind === "story" ? (
          <StatusesMenu>
            <StatusesMenu.Trigger>
              <Button
                aria-label="Change status"
                className="pointer-events-auto"
                color="tertiary"
                disabled={pending}
                size="sm"
                variant="naked"
              >
                {statusContent}
              </Button>
            </StatusesMenu.Trigger>
            <StatusesMenu.Items
              setStatusId={onStatus}
              statusId={statusId}
              teamId={teamId}
            />
          </StatusesMenu>
        ) : null}
        {canEdit && states.length > 0 && kind === "objective" ? (
          <ObjectiveStatusesMenu>
            <ObjectiveStatusesMenu.Trigger>
              <Button
                aria-label="Change status"
                className="pointer-events-auto"
                color="tertiary"
                disabled={pending}
                size="sm"
                variant="naked"
              >
                {statusContent}
              </Button>
            </ObjectiveStatusesMenu.Trigger>
            <ObjectiveStatusesMenu.Items
              setStatusId={onStatus}
              statusId={statusId}
            />
          </ObjectiveStatusesMenu>
        ) : null}
        {!canEdit || states.length === 0 ? (
          <span className={fieldClass}>{statusContent}</span>
        ) : null}
      </Property>
      <Property label="Priority">
        {canEdit ? (
          <PrioritiesMenu>
            <PrioritiesMenu.Trigger>
              <Button
                aria-label="Change priority"
                className="pointer-events-auto"
                color="tertiary"
                disabled={pending}
                size="sm"
                variant="naked"
              >
                {priorityContent}
              </Button>
            </PrioritiesMenu.Trigger>
            <PrioritiesMenu.Items
              priority={priority}
              setPriority={onPriority}
            />
          </PrioritiesMenu>
        ) : (
          <span className={fieldClass}>{priorityContent}</span>
        )}
      </Property>
      <Property label={kind === "story" ? "Assignee" : "Lead"}>
        {canEdit ? (
          <AssigneesMenu>
            <AssigneesMenu.Trigger>
              <Button
                aria-label={
                  kind === "story" ? "Change assignee" : "Change lead"
                }
                className="pointer-events-auto max-w-52"
                color="tertiary"
                disabled={pending}
                size="sm"
                variant="naked"
              >
                {assigneeContent}
              </Button>
            </AssigneesMenu.Trigger>
            <AssigneesMenu.Items
              assigneeId={assigneeId ?? undefined}
              onAssigneeSelected={onAssignee}
              teamId={teamId}
            />
          </AssigneesMenu>
        ) : (
          <span className={fieldClass}>{assigneeContent}</span>
        )}
      </Property>
    </div>
  );
};

export const NativeStoryFields = ({ story }: { story: NativeStorySummary }) => {
  const {
    states,
    members,
    canEdit: canEditTeam,
  } = useFields(story.teamId, "story");
  const mutation = useUpdateStoryMutation();
  const canEdit = canEditTeam && !story.archivedAt && !story.deletedAt;
  const assignee =
    members.find((member) => member.id === story.assigneeId) ?? story.assignee;
  const update = (payload: StoryUpdate, onSuccess?: () => void) => {
    if (canEdit) mutation.mutate({ storyId: story.id, payload }, { onSuccess });
  };
  return (
    <>
      <p className="font-semibold [overflow-wrap:anywhere]">{story.title}</p>
      <WorkFields
        assigneeAvatar={assignee?.avatarUrl}
        assigneeId={story.assigneeId}
        assigneeName={assignee?.fullName || assignee?.username}
        canEdit={canEdit}
        kind="story"
        onAssignee={(assigneeId) => {
          update({ assigneeId });
        }}
        onPriority={(priority) => {
          update({ priority });
        }}
        onStatus={(statusId) => {
          update({ statusId });
        }}
        pending={mutation.isPending}
        priority={story.priority}
        states={states}
        statusId={story.statusId}
        teamId={story.teamId}
      />
      {mutation.isError ? (
        <p className="text-danger mt-3" role="alert">
          The change could not be saved. Try again.
        </p>
      ) : null}
    </>
  );
};

export const NativeObjectiveFields = ({
  objective,
}: {
  objective: NativeObjectiveSummary;
}) => {
  const { states, members, canEdit } = useFields(objective.teamId, "objective");
  const mutation = useUpdateObjectiveMutation();
  const lead = members.find((member) => member.id === objective.leadUser);
  const update = (data: ObjectiveUpdate, onSuccess?: () => void) => {
    if (canEdit)
      mutation.mutate({ objectiveId: objective.id, data }, { onSuccess });
  };
  return (
    <>
      <EditableTitle
        canEdit={canEdit}
        onSave={(name, done) => {
          update({ name }, done);
        }}
        pending={mutation.isPending}
        title={objective.name}
      />
      <WorkFields
        assigneeAvatar={lead?.avatarUrl}
        assigneeId={objective.leadUser}
        assigneeName={lead?.fullName || lead?.username}
        canEdit={canEdit}
        kind="objective"
        onAssignee={(leadUser) => {
          update({ leadUser });
        }}
        onPriority={(priority) => {
          update({ priority });
        }}
        onStatus={(statusId) => {
          update({ statusId });
        }}
        pending={mutation.isPending}
        priority={objective.priority ?? "No Priority"}
        states={states}
        statusId={objective.statusId}
        teamId={objective.teamId}
      />
      {mutation.isError ? (
        <p className="text-danger mt-3" role="alert">
          The change could not be saved. Try again.
        </p>
      ) : null}
    </>
  );
};
