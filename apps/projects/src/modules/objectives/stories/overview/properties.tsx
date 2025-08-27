import { CalendarIcon } from "icons";
import { Text, Flex, Button, Avatar, DatePicker } from "ui";
import { cn } from "lib";
import { format, formatISO } from "date-fns";
import { useParams } from "next/navigation";
import { useSession } from "next-auth/react";
import { AssigneesMenu, PrioritiesMenu, PriorityIcon } from "@/components/ui";
import { ObjectiveStatusIcon } from "@/components/ui/objective-status-icon";
import { useIsAdminOrOwner } from "@/hooks/owner";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { useMediaQuery } from "@/hooks";
import { hexToRgba } from "@/utils";
import type { ObjectiveUpdate } from "../../types";
import { useObjective, useUpdateObjectiveMutation } from "../../hooks";
import { ObjectiveStatusesMenu } from "../../../../components/ui/objective-statuses-menu";

export const Properties = () => {
  const { data: session } = useSession();
  const isMobile = useMediaQuery("(max-width: 768px)");
  const { objectiveId } = useParams<{ objectiveId: string }>();
  const { data: objective } = useObjective(objectiveId);
  const { data: statuses = [] } = useObjectiveStatuses();
  const { data: members = [] } = useTeamMembers(objective?.teamId);
  const updateMutation = useUpdateObjectiveMutation();
  const leadUser = members.find((m) => m.id === objective?.leadUser);
  const { isAdminOrOwner } = useIsAdminOrOwner(objective?.createdBy);
  const canUpdate = isAdminOrOwner || session?.user?.id === objective?.leadUser;

  const handleUpdate = (data: ObjectiveUpdate) => {
    updateMutation.mutate({
      objectiveId,
      data,
    });
  };

  const status = statuses.find((s) => s.id === objective?.statusId);

  return (
    <Flex align="center" className="mt-6" gap={2} wrap>
      <Text className="hidden md:block" color="muted" fontWeight="semibold">
        Properties:
      </Text>
      <ObjectiveStatusesMenu>
        <ObjectiveStatusesMenu.Trigger>
          <Button
            color="tertiary"
            disabled={!canUpdate}
            leftIcon={<ObjectiveStatusIcon statusId={objective?.statusId} />}
            size="sm"
            style={{
              backgroundColor: hexToRgba(status?.color),
              borderColor: hexToRgba(status?.color),
            }}
            type="button"
            variant={isMobile ? "solid" : "naked"}
          >
            {status?.name}
          </Button>
        </ObjectiveStatusesMenu.Trigger>
        <ObjectiveStatusesMenu.Items
          setStatusId={(statusId) => {
            handleUpdate({ statusId });
          }}
          statusId={objective?.statusId}
        />
      </ObjectiveStatusesMenu>
      <PrioritiesMenu>
        <PrioritiesMenu.Trigger>
          <Button
            color="tertiary"
            disabled={!canUpdate}
            leftIcon={<PriorityIcon priority={objective?.priority} />}
            size="sm"
            type="button"
            variant={isMobile ? "solid" : "naked"}
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
            disabled={!canUpdate}
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
            variant={isMobile ? "solid" : "naked"}
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
          teamId={objective?.teamId}
        />
      </AssigneesMenu>
      <DatePicker>
        <DatePicker.Trigger>
          <Button
            color="tertiary"
            disabled={!canUpdate}
            leftIcon={
              <CalendarIcon
                className={cn("h-[1.15rem] w-auto", {
                  "text-gray/80 dark:text-gray-300/80": !objective?.startDate,
                })}
              />
            }
            size="sm"
            variant={isMobile ? "solid" : "naked"}
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
            handleUpdate({
              startDate: formatISO(day, { representation: "date" }),
            });
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
            disabled={!canUpdate}
            leftIcon={
              <CalendarIcon
                className={cn("h-[1.15rem]", {
                  "text-gray/80 dark:text-gray-300/80": !objective?.endDate,
                })}
              />
            }
            size="sm"
            variant={isMobile ? "solid" : "naked"}
          >
            {objective?.endDate ? (
              format(new Date(objective.endDate), "MMM d, yyyy")
            ) : (
              <Text color="muted">Deadline</Text>
            )}
          </Button>
        </DatePicker.Trigger>
        <DatePicker.Calendar
          onDayClick={(day) => {
            handleUpdate({
              endDate: formatISO(day, { representation: "date" }),
            });
          }}
          selected={
            objective?.endDate ? new Date(objective.endDate) : undefined
          }
        />
      </DatePicker>
    </Flex>
  );
};
