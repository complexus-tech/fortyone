import { CalendarIcon } from "icons";
import { Text, Flex, Button, Avatar, DatePicker } from "ui";
import { cn } from "lib";
import { format } from "date-fns";
import { useParams } from "next/navigation";
import {
  AssigneesMenu,
  PrioritiesMenu,
  PriorityIcon,
  StoryStatusIcon,
} from "@/components/ui";
import { useStatuses } from "@/lib/hooks/statuses";
import { useMembers } from "@/lib/hooks/members";
import type { ObjectiveUpdate } from "../../types";
import { useObjective, useUpdateObjectiveMutation } from "../../hooks";
import { ObjectiveStatusesMenu } from "../../../../components/ui/objective-statuses-menu";

export const Properties = () => {
  const { objectiveId } = useParams<{ objectiveId: string }>();
  const { data: objective } = useObjective(objectiveId);
  const { data: statuses = [] } = useStatuses();
  const { data: members = [] } = useMembers();
  const updateMutation = useUpdateObjectiveMutation();
  const leadUser = members.find((m) => m.id === objective?.leadUser);

  const handleUpdate = (data: ObjectiveUpdate) => {
    updateMutation.mutate({
      objectiveId,
      data,
    });
  };

  const status = statuses.find((s) => s.id === objective?.statusId);

  return (
    <Flex align="center" className="mt-6" gap={2} wrap>
      <Text color="muted" fontWeight="semibold">
        Properties:
      </Text>
      <ObjectiveStatusesMenu>
        <ObjectiveStatusesMenu.Trigger>
          <Button
            color="tertiary"
            leftIcon={<StoryStatusIcon statusId={objective?.statusId} />}
            size="sm"
            type="button"
            variant="naked"
          >
            {status?.name ?? "Backlog"}
          </Button>
        </ObjectiveStatusesMenu.Trigger>
        <ObjectiveStatusesMenu.Items
          setStatusId={(statusId) => {
            handleUpdate({ statusId });
          }}
          statusId={objective?.statusId}
          teamId={objective!.teamId}
        />
      </ObjectiveStatusesMenu>
      <PrioritiesMenu>
        <PrioritiesMenu.Trigger>
          <Button
            color="tertiary"
            leftIcon={<PriorityIcon priority={objective?.priority} />}
            size="sm"
            type="button"
            variant="naked"
          >
            {objective?.priority ?? "No Priority"}
          </Button>
        </PrioritiesMenu.Trigger>
        <PrioritiesMenu.Items
          priority={objective?.priority}
          setPriority={(priority) => {
            handleUpdate({ priority });
          }}
        />
      </PrioritiesMenu>
      <AssigneesMenu>
        <AssigneesMenu.Trigger>
          <Button
            className={cn("font-medium", {
              "text-gray-200 dark:text-gray-300": !objective?.leadUser,
            })}
            color="tertiary"
            leftIcon={
              <Avatar
                className={cn({
                  "text-dark/80 dark:text-gray-200": !objective?.leadUser,
                })}
                name={leadUser?.username}
                size="xs"
                src={leadUser?.avatarUrl}
              />
            }
            size="sm"
            type="button"
            variant="naked"
          >
            {leadUser ? (
              leadUser.username
            ) : (
              <Text as="span" color="muted">
                Assign lead
              </Text>
            )}
          </Button>
        </AssigneesMenu.Trigger>
        <AssigneesMenu.Items
          assigneeId={objective?.leadUser}
          onAssigneeSelected={(leadUser) => {
            handleUpdate({ leadUser: leadUser ?? undefined });
          }}
        />
      </AssigneesMenu>
      <DatePicker>
        <DatePicker.Trigger>
          <Button
            color="tertiary"
            leftIcon={
              <CalendarIcon
                className={cn("h-[1.15rem] w-auto", {
                  "text-gray/80 dark:text-gray-300/80": !objective?.startDate,
                })}
              />
            }
            size="sm"
            variant="naked"
          >
            {objective?.startDate ? (
              format(new Date(objective.startDate), "MMM d, yyyy")
            ) : (
              <Text as="span" color="muted">
                Start date
              </Text>
            )}
          </Button>
        </DatePicker.Trigger>
        <DatePicker.Calendar
          onDayClick={(day) => {
            handleUpdate({ startDate: day.toISOString() });
          }}
          selected={
            objective?.startDate ? new Date(objective.startDate) : undefined
          }
        />
      </DatePicker>
      <DatePicker>
        <DatePicker.Trigger>
          <Button
            className={cn({
              "text-gray/80 dark:text-gray-300/80": !objective?.endDate,
            })}
            color="tertiary"
            leftIcon={
              <CalendarIcon
                className={cn("h-[1.15rem]", {
                  "text-gray/80 dark:text-gray-300/80": !objective?.endDate,
                })}
              />
            }
            size="sm"
            variant="naked"
          >
            {objective?.endDate ? (
              format(new Date(objective.endDate), "MMM d, yy")
            ) : (
              <Text color="muted">Target date</Text>
            )}
          </Button>
        </DatePicker.Trigger>
        <DatePicker.Calendar
          onDayClick={(day) => {
            handleUpdate({ endDate: day.toISOString() });
          }}
          selected={
            objective?.endDate ? new Date(objective.endDate) : undefined
          }
        />
      </DatePicker>
    </Flex>
  );
};
