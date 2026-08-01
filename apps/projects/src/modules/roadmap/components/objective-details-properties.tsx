"use client";

import type { ReactNode } from "react";
import { useState } from "react";
import { format, formatISO } from "date-fns";
import { CalendarIcon } from "icons";
import { cn } from "lib";
import {
  Avatar,
  Box,
  Button,
  DatePicker,
  Dialog,
  Flex,
  Text,
  TextArea,
} from "ui";
import { toast } from "sonner";
import { useSession } from "@/lib/auth/client";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";
import { useIsAdminOrOwner } from "@/hooks/owner";
import { useTerminology, useUserRole } from "@/hooks";
import { useUpdateObjectiveMutation } from "@/modules/objectives/hooks/update-mutation";
import type {
  Objective,
  ObjectiveHealth,
  ObjectiveUpdate,
} from "@/modules/objectives/types";
import { AssigneesMenu } from "@/components/ui/story/assignees-menu";
import { HealthMenu } from "@/components/ui/health-menu";
import { ObjectiveHealthIcon } from "@/components/ui/objective-health-icon";
import { ObjectiveStatusesMenu } from "@/components/ui/objective-statuses-menu";
import { ObjectiveStatusIcon } from "@/components/ui/objective-status-icon";
import { PrioritiesMenu } from "@/components/ui/story/priorities-menu";
import { PriorityIcon } from "@/components/ui/priority-icon";

const DetailRow = ({ label, value }: { label: string; value: ReactNode }) => (
  <Flex align="center" className="min-h-9" gap={4}>
    <Text className="w-28 shrink-0" color="muted">
      {label}
    </Text>
    <Box className="min-w-0 flex-1">{value}</Box>
  </Flex>
);

const propertyButtonClassName =
  "-ml-2 min-w-0 max-w-full justify-start px-2 font-normal";

export const ObjectiveDetailsProperties = ({
  objective,
}: {
  objective: Objective;
}) => {
  const { data: session } = useSession();
  const { userRole } = useUserRole();
  const { getTermDisplay } = useTerminology();
  const { data: statuses = [] } = useObjectiveStatuses();
  const { data: members = [] } = useTeamMembers(objective.teamId);
  const { isAdminOrOwner } = useIsAdminOrOwner(objective.createdBy);
  const updateMutation = useUpdateObjectiveMutation();
  const [comment, setComment] = useState("");
  const [isCommentOpen, setIsCommentOpen] = useState(false);
  const [pendingHealth, setPendingHealth] = useState<ObjectiveHealth>(null);
  const canUpdate = isAdminOrOwner || session?.user.id === objective.leadUser;
  const status = statuses.find((item) => item.id === objective.statusId);
  const lead = members.find((member) => member.id === objective.leadUser);

  const handleUpdate = (data: ObjectiveUpdate) => {
    updateMutation.mutate({
      objectiveId: objective.id,
      data,
    });
  };

  const closeHealthDialog = () => {
    setIsCommentOpen(false);
    setComment("");
    setPendingHealth(null);
  };

  const handleHealthUpdate = () => {
    if (!pendingHealth) {
      toast.warning("Validation error", {
        description: "Please select a health status",
      });
      return;
    }

    const trimmedComment = comment.trim();
    if (!trimmedComment) {
      toast.warning("Validation error", {
        description: "Please provide a comment",
      });
      return;
    }

    handleUpdate({ health: pendingHealth, comment: trimmedComment });
    closeHealthDialog();
  };

  return (
    <>
      <Text className="mb-3">Properties</Text>
      <Flex direction="column" gap={2}>
        <DetailRow
          label="Status"
          value={
            <ObjectiveStatusesMenu>
              <ObjectiveStatusesMenu.Trigger>
                <Button
                  align="left"
                  className={propertyButtonClassName}
                  color="tertiary"
                  disabled={!canUpdate}
                  leftIcon={
                    <ObjectiveStatusIcon statusId={objective.statusId} />
                  }
                  size="sm"
                  type="button"
                  variant="naked"
                >
                  <span className="truncate">
                    {status?.name ?? "No status"}
                  </span>
                </Button>
              </ObjectiveStatusesMenu.Trigger>
              <ObjectiveStatusesMenu.Items
                setStatusId={(statusId) => {
                  handleUpdate({ statusId });
                }}
                statusId={objective.statusId}
              />
            </ObjectiveStatusesMenu>
          }
        />
        <DetailRow
          label="Health"
          value={
            <HealthMenu>
              <HealthMenu.Trigger>
                <Button
                  align="left"
                  className={propertyButtonClassName}
                  color="tertiary"
                  disabled={!canUpdate}
                  leftIcon={<ObjectiveHealthIcon health={objective.health} />}
                  size="sm"
                  type="button"
                  variant="naked"
                >
                  {objective.health ?? "No health"}
                </Button>
              </HealthMenu.Trigger>
              <HealthMenu.Items
                health={objective.health}
                setHealth={(health) => {
                  setPendingHealth(health);
                  setIsCommentOpen(true);
                }}
              />
            </HealthMenu>
          }
        />
        <DetailRow
          label="Priority"
          value={
            <PrioritiesMenu>
              <PrioritiesMenu.Trigger>
                <Button
                  align="left"
                  className={propertyButtonClassName}
                  color="tertiary"
                  disabled={!canUpdate}
                  leftIcon={<PriorityIcon priority={objective.priority} />}
                  size="sm"
                  type="button"
                  variant="naked"
                >
                  {objective.priority ?? "No priority"}
                </Button>
              </PrioritiesMenu.Trigger>
              <PrioritiesMenu.Items
                priority={objective.priority}
                setPriority={(priority) => {
                  handleUpdate({ priority });
                }}
              />
            </PrioritiesMenu>
          }
        />
        <DetailRow
          label="Lead"
          value={
            <AssigneesMenu>
              <AssigneesMenu.Trigger>
                <Button
                  align="left"
                  className={propertyButtonClassName}
                  color="tertiary"
                  disabled={userRole === "guest"}
                  leftIcon={
                    <Avatar
                      className={cn({
                        "text-foreground/80": !objective.leadUser,
                      })}
                      name={lead?.fullName || lead?.username}
                      size="xs"
                      src={lead?.avatarUrl}
                    />
                  }
                  size="sm"
                  type="button"
                  variant="naked"
                >
                  <span className="truncate">
                    {lead?.username ?? "Assign lead"}
                  </span>
                </Button>
              </AssigneesMenu.Trigger>
              <AssigneesMenu.Items
                assigneeId={objective.leadUser}
                onAssigneeSelected={(leadUser) => {
                  handleUpdate({ leadUser: leadUser ?? undefined });
                }}
                placeholder="Assign lead..."
                teamId={objective.teamId}
              />
            </AssigneesMenu>
          }
        />
        <DetailRow
          label="Start date"
          value={
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  align="left"
                  className={propertyButtonClassName}
                  color="tertiary"
                  disabled={!canUpdate}
                  leftIcon={
                    <CalendarIcon
                      className={cn("h-4 w-auto", {
                        "text-text-muted": !objective.startDate,
                      })}
                    />
                  }
                  size="sm"
                  type="button"
                  variant="naked"
                >
                  {objective.startDate
                    ? format(new Date(objective.startDate), "MMM d, yyyy")
                    : "Start date"}
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar
                onDayClick={(day) => {
                  handleUpdate({
                    startDate: formatISO(day, { representation: "date" }),
                  });
                }}
                selected={
                  objective.startDate
                    ? new Date(objective.startDate)
                    : undefined
                }
              />
            </DatePicker>
          }
        />
        <DetailRow
          label="Target date"
          value={
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  align="left"
                  className={propertyButtonClassName}
                  color="tertiary"
                  disabled={!canUpdate}
                  leftIcon={
                    <CalendarIcon
                      className={cn("h-4 w-auto", {
                        "text-text-muted": !objective.endDate,
                      })}
                    />
                  }
                  size="sm"
                  type="button"
                  variant="naked"
                >
                  {objective.endDate
                    ? format(new Date(objective.endDate), "MMM d, yyyy")
                    : "Target date"}
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar
                onDayClick={(day) => {
                  handleUpdate({
                    endDate: formatISO(day, { representation: "date" }),
                  });
                }}
                selected={
                  objective.endDate ? new Date(objective.endDate) : undefined
                }
              />
            </DatePicker>
          }
        />
      </Flex>

      <Dialog
        onOpenChange={(open) => {
          if (open) {
            setIsCommentOpen(true);
            return;
          }
          closeHealthDialog();
        }}
        open={isCommentOpen}
      >
        <Dialog.Content>
          <Dialog.Header>
            <Dialog.Title className="flex items-center gap-2 px-6 pt-0.5 text-lg">
              Change {getTermDisplay("objectiveTerm")} health to{" "}
              <ObjectiveHealthIcon health={pendingHealth} />
              {pendingHealth}
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Description>
            Please provide a brief comment explaining why you&apos;re changing
            the objective health status.
          </Dialog.Description>
          <Dialog.Body>
            <Text className="mt-3 mb-1.5" color="muted">
              Comment*
            </Text>
            <TextArea
              className="border-border/80 resize-none rounded-2xl border bg-transparent py-4 leading-normal"
              onChange={(event) => {
                setComment(event.target.value);
              }}
              placeholder={`e.g, We're on track to complete the ${getTermDisplay("objectiveTerm")} by the end of the quarter.`}
              rows={4}
              value={comment}
            />
          </Dialog.Body>
          <Dialog.Footer className="justify-end gap-2">
            <Button
              color="tertiary"
              onClick={closeHealthDialog}
              variant="outline"
            >
              Cancel
            </Button>
            <Button
              disabled={!comment.trim() || updateMutation.isPending}
              loading={updateMutation.isPending}
              onClick={handleHealthUpdate}
            >
              Update health
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
