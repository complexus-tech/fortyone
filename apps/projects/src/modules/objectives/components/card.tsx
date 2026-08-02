"use client";
import {
  Flex,
  Text,
  Box,
  Avatar,
  Button,
  Checkbox,
  DatePicker,
  CircleProgressBar,
} from "ui";
import Link from "next/link";
import { ArrowRight2Icon, ObjectiveIcon, CalendarIcon } from "icons";
import { format, formatISO } from "date-fns";
import { cn } from "lib";
import { RowWrapper } from "@/components/ui/row-wrapper";
import { useTeams } from "@/modules/teams/hooks/teams";
import {
  TeamColor,
  AssigneesMenu,
  PrioritiesMenu,
  PriorityIcon,
  ObjectiveHealthIcon,
} from "@/components/ui";
import { ObjectiveStatusesMenu } from "@/components/ui/objective-statuses-menu";
import { ObjectiveStatusIcon } from "@/components/ui/objective-status-icon";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { hexToRgba } from "@/utils";
import { useTerminology, useWorkspacePath } from "@/hooks";
import { useCanUpdateObjective } from "../hooks/use-can-update-objective";
import { useUpdateObjectiveMutation } from "../hooks/update-mutation";
import type { Objective, ObjectiveUpdate } from "../types";
import { ObjectiveHealthEditor } from "./objective-health-editor";

export const ObjectiveCard = ({
  id,
  name,
  leadUser,
  teamId,
  endDate,
  isInTeam,
  isInSearch,
  statusId,
  health,
  priority,
  onSelect,
  onSelectionChange,
  selected = false,
  childCount = 0,
  isExpanded = false,
  onToggleExpanded,
  ...rest
}: Objective & {
  childCount?: number;
  isExpanded?: boolean;
  isInTeam?: boolean;
  isInSearch?: boolean;
  onSelect?: () => void;
  onSelectionChange?: (checked: boolean) => void;
  onToggleExpanded?: () => void;
  selected?: boolean;
}) => {
  const canUpdate = useCanUpdateObjective();
  const { data: members = [] } = useTeamMembers(teamId);
  const { data: teams = [] } = useTeams();
  const { data: statuses = [] } = useObjectiveStatuses();
  const updateMutation = useUpdateObjectiveMutation();
  const { withWorkspace } = useWorkspacePath();
  const { getTermDisplay } = useTerminology();

  const lead = members.find((member) => member.id === leadUser);
  const team = teams.find((team) => team.id === teamId);
  const status = statuses.find((s) => s.id === statusId);
  let progress = 0;
  if (rest.stats) {
    progress = Math.round((rest.stats.completed / rest.stats.total) * 100) || 0;
  }

  const handleUpdate = (data: ObjectiveUpdate) => {
    updateMutation.mutate({
      objectiveId: id,
      data,
    });
  };

  return (
    <RowWrapper
      className={cn("@container px-5 py-2.5 md:px-12", {
        "gap-4 md:px-6": isInSearch,
      })}
    >
      <Box
        className={cn(
          "relative flex min-w-10 flex-1 items-center gap-2 @sm:min-w-20",
          {
            "pointer-events-none opacity-40": id === "optimistic",
          },
        )}
      >
        {onSelectionChange ? (
          <Checkbox
            checked={selected}
            className="shrink-0 rounded md:absolute md:-left-[1.6rem]"
            disabled={!canUpdate}
            onCheckedChange={onSelectionChange}
          />
        ) : null}
        {childCount > 0 && onToggleExpanded ? (
          <button
            aria-label={`${isExpanded ? "Collapse" : "Expand"} ${getTermDisplay(
              "keyResultTerm",
              { variant: "plural" },
            )}`}
            className="text-text-muted hover:text-foreground grid h-7 w-7 shrink-0 place-items-center rounded transition-colors"
            onClick={onToggleExpanded}
            type="button"
          >
            <ArrowRight2Icon
              className={cn("h-4 w-4 transition-transform", {
                "rotate-90": isExpanded,
              })}
              strokeWidth={2.5}
            />
          </button>
        ) : null}
        {onSelect ? (
          <button
            className="focus-visible:ring-primary flex min-w-0 flex-1 items-center gap-2 rounded-sm text-left outline-none hover:opacity-90 focus-visible:ring-1"
            onClick={onSelect}
            type="button"
          >
            {onSelectionChange ? null : (
              <Flex
                align="center"
                className="bg-surface-muted size-8 shrink-0 rounded-lg"
                justify="center"
              >
                <ObjectiveIcon className="h-4" />
              </Flex>
            )}
            <Text className="min-w-0 truncate pr-2">{name}</Text>
          </button>
        ) : (
          <Link
            className="flex min-w-0 flex-1 items-center gap-2 hover:opacity-90"
            href={withWorkspace(`/teams/${teamId}/objectives/${id}`)}
            prefetch
          >
            {onSelectionChange ? null : (
              <Flex
                align="center"
                className="bg-surface-muted size-8 shrink-0 rounded-lg"
                justify="center"
              >
                <ObjectiveIcon className="h-4" />
              </Flex>
            )}
            <Text className="min-w-0 truncate pr-2">{name}</Text>
          </Link>
        )}
      </Box>
      <Flex align="center" className="shrink-0 gap-2 md:gap-4">
        {!isInTeam ? (
          <Box className="hidden w-[50px] shrink-0 items-center gap-1.5 md:flex">
            <TeamColor color={team?.color} />
            <Text className="truncate uppercase" color="muted">
              {team?.code}
            </Text>
          </Box>
        ) : null}
        <Box className="hidden w-[40px] shrink-0 items-center md:flex">
          <AssigneesMenu>
            <AssigneesMenu.Trigger>
              <Button
                className={cn({
                  "text-text-secondary": !leadUser,
                })}
                color="tertiary"
                disabled={!canUpdate}
                leftIcon={
                  <Avatar
                    className={cn({
                      "text-foreground/80": !leadUser,
                    })}
                    name={lead?.username}
                    size="xs"
                    src={lead?.avatarUrl}
                  />
                }
                size="sm"
                type="button"
                variant="naked"
              >
                <span className="sr-only">{lead?.username}</span>
              </Button>
            </AssigneesMenu.Trigger>
            <AssigneesMenu.Items
              assigneeId={leadUser}
              onAssigneeSelected={(leadUser) => {
                handleUpdate({ leadUser });
              }}
              teamId={teamId}
            />
          </AssigneesMenu>
        </Box>
        {!isInSearch && (
          <Box className="hidden w-[60px] shrink-0 items-center gap-1.5 pl-0.5 md:flex">
            <CircleProgressBar progress={progress} size={16} strokeWidth={2} />
            {progress}%
          </Box>
        )}
        <Box className="shrink-0 md:w-[96px]">
          <ObjectiveStatusesMenu>
            <ObjectiveStatusesMenu.Trigger>
              <Button
                color="tertiary"
                disabled={!canUpdate}
                leftIcon={<ObjectiveStatusIcon statusId={statusId} />}
                size="sm"
                style={{
                  backgroundColor: hexToRgba(status?.color),
                  borderColor: hexToRgba(status?.color),
                }}
                type="button"
              >
                <span className="hidden max-w-[7ch] truncate md:inline-block">
                  {status?.name ?? "Backlog"}
                </span>
              </Button>
            </ObjectiveStatusesMenu.Trigger>
            <ObjectiveStatusesMenu.Items
              setStatusId={(statusId) => {
                handleUpdate({ statusId });
              }}
              statusId={statusId}
            />
          </ObjectiveStatusesMenu>
        </Box>
        <Box className="shrink-0 md:w-[100px]">
          <PrioritiesMenu>
            <PrioritiesMenu.Trigger>
              <Button
                color="tertiary"
                disabled={!canUpdate}
                leftIcon={<PriorityIcon priority={priority} />}
                size="sm"
                type="button"
                variant="naked"
              >
                <span className="hidden md:inline-block">
                  {priority ?? "No Priority"}
                </span>
              </Button>
            </PrioritiesMenu.Trigger>
            <PrioritiesMenu.Items
              priority={priority}
              setPriority={(priority) => {
                handleUpdate({ priority });
              }}
            />
          </PrioritiesMenu>
        </Box>

        <Box className="hidden w-[100px] shrink-0 md:block">
          <DatePicker>
            <DatePicker.Trigger>
              <Button
                className={cn({
                  "text-text-muted": !endDate,
                })}
                color="tertiary"
                disabled={!canUpdate}
                leftIcon={
                  <CalendarIcon
                    className={cn("h-[1.15rem]", {
                      "text-text-muted": !endDate,
                    })}
                  />
                }
                size="sm"
                variant="naked"
              >
                {endDate ? (
                  format(new Date(endDate), "MMM d, yy")
                ) : (
                  <Text color="muted">Target date</Text>
                )}
              </Button>
            </DatePicker.Trigger>
            <DatePicker.Calendar
              onDayClick={(day) => {
                handleUpdate({
                  endDate: formatISO(day, { representation: "date" }),
                });
              }}
              selected={endDate ? new Date(endDate) : undefined}
            />
          </DatePicker>
        </Box>

        <Box className="shrink-0 md:w-[96px]">
          <ObjectiveHealthEditor health={health} objectiveId={id}>
            <Button
              color="tertiary"
              disabled={!canUpdate}
              leftIcon={<ObjectiveHealthIcon health={health} />}
              size="sm"
              type="button"
              variant="naked"
            >
              <span className="hidden md:inline-block">
                {health ?? "No Health"}
              </span>
            </Button>
          </ObjectiveHealthEditor>
        </Box>
      </Flex>
    </RowWrapper>
  );
};
