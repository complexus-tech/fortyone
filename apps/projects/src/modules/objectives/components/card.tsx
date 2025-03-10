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
import { useUpdateObjectiveMutation } from "../hooks";
import type { Objective, ObjectiveUpdate } from "../types";

export const ObjectiveCard = ({
  id,
  name,
  leadUser,
  teamId,
  endDate,
  stats: { completed, total },
  isInTeam,
  statusId,
  health,
  priority,
  createdBy,
}: Objective & { isInTeam?: boolean }) => {
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
  const progress = Math.round((completed / total) * 100) || 0;

  const handleUpdate = (data: ObjectiveUpdate) => {
    updateMutation.mutate({
      objectiveId: id,
      data,
    });
  };

  return (
    <RowWrapper className="px-5 py-3 md:px-12">
      <Box className="flex w-[300px] shrink-0 items-center gap-2">
        <Link
          className="flex w-full items-center gap-2 hover:opacity-90"
          href={`/teams/${teamId}/objectives/${id}`}
        >
          <Flex
            align="center"
            className="size-8 shrink-0 rounded-lg bg-gray-100/50 dark:bg-dark-200"
            justify="center"
          >
            <ObjectiveIcon className="h-4" />
          </Flex>
          <Text className="truncate font-medium">{name}</Text>
        </Link>
      </Box>
      <Flex align="center" gap={4}>
        {!isInTeam ? (
          <Box className="flex w-[45px] shrink-0 items-center gap-1.5">
            <TeamColor color={team?.color} />
            <Text className="truncate uppercase" color="muted">
              {team?.code}
            </Text>
          </Box>
        ) : null}

        <Box className="flex w-[40px] shrink-0 items-center">
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

        <Box className="flex w-[60px] shrink-0 items-center gap-1.5 pl-0.5">
          <CircleProgressBar progress={progress} size={16} strokeWidth={2} />
          {progress}%
        </Box>
        <Box className="w-[120px] shrink-0">
          <ObjectiveStatusesMenu>
            <ObjectiveStatusesMenu.Trigger>
              <Button
                color="tertiary"
                disabled={!canUpdate}
                leftIcon={<ObjectiveStatusIcon statusId={statusId} />}
                size="sm"
                type="button"
                variant="naked"
              >
                <span className="inline-block max-w-[7ch] truncate">
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
        <Box className="w-[100px] shrink-0">
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
                {priority ?? "No Priority"}
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

        <Box className="w-[100px] shrink-0">
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

        <Box className="w-[120px] shrink-0">
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
                {health ?? "No Health"}
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
