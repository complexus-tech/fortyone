"use client";
import {
  Flex,
  Text,
  Box,
  Avatar,
  Button,
  DatePicker,
  CircleProgressBar,
} from "ui";
import Link from "next/link";
import { ObjectiveIcon, CalendarIcon } from "icons";
import { format } from "date-fns";
import { cn } from "lib";
import { useSession } from "next-auth/react";
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
import { HealthMenu } from "@/components/ui/health-menu";
import { ObjectiveStatusIcon } from "@/components/ui/objective-status-icon";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { useIsAdminOrOwner } from "@/hooks/owner";
import { hexToRgba } from "@/utils";
import { useUpdateObjectiveMutation } from "../hooks";
import type { Objective, ObjectiveUpdate } from "../types";

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
  createdBy,
  ...rest
}: Objective & { isInTeam?: boolean; isInSearch?: boolean }) => {
  const { data: session } = useSession();
  const { data: members = [] } = useTeamMembers(teamId);
  const { data: teams = [] } = useTeams();
  const { data: statuses = [] } = useObjectiveStatuses();
  const updateMutation = useUpdateObjectiveMutation();
  const { isAdminOrOwner } = useIsAdminOrOwner(createdBy);
  const canUpdate = isAdminOrOwner || session?.user?.id === leadUser;

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
      className={cn("px-5 py-3 md:px-12", {
        "gap-4 py-2 md:px-6": isInSearch,
      })}
    >
      <Box
        className={cn("flex shrink-0 items-center gap-2 md:w-[300px]", {
          "pointer-events-none opacity-40": id === "optimistic",
        })}
      >
        <Link
          className="flex w-full items-center gap-2 hover:opacity-90"
          href={`/teams/${teamId}/objectives/${id}`}
          prefetch
        >
          <Flex
            align="center"
            className="size-8 shrink-0 rounded-[0.6rem] bg-gray-100/50 dark:bg-dark-200"
            justify="center"
          >
            <ObjectiveIcon className="h-4" />
          </Flex>
          <Text className="truncate font-medium">{name}</Text>
        </Link>
      </Box>
      <Flex align="center" className="gap-2 md:gap-4">
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
                className={cn("font-medium", {
                  "text-gray-200 dark:text-gray-300": !leadUser,
                })}
                color="tertiary"
                disabled={!canUpdate}
                leftIcon={
                  <Avatar
                    className={cn({
                      "text-dark/80 dark:text-gray-200": !leadUser,
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
        <Box className="shrink-0 md:w-[120px]">
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
                  "text-gray/80 dark:text-gray-300/80": !endDate,
                })}
                color="tertiary"
                disabled={!canUpdate}
                leftIcon={
                  <CalendarIcon
                    className={cn("h-[1.15rem]", {
                      "text-gray/80 dark:text-gray-300/80": !endDate,
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
                handleUpdate({ endDate: day.toISOString() });
              }}
              selected={endDate ? new Date(endDate) : undefined}
            />
          </DatePicker>
        </Box>

        <Box className="shrink-0 md:w-[120px]">
          <HealthMenu>
            <HealthMenu.Trigger>
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
            </HealthMenu.Trigger>
            <HealthMenu.Items
              health={health}
              setHealth={(health) => {
                handleUpdate({ health });
              }}
            />
          </HealthMenu>
        </Box>
      </Flex>
    </RowWrapper>
  );
};
