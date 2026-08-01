"use client";

import { useState } from "react";
import { formatISO } from "date-fns";
import { AssigneeIcon, CalendarIcon, CloseIcon, DeleteIcon } from "icons";
import { Button, DatePicker, Dialog, Flex, Text, TextArea, Tooltip } from "ui";
import {
  AssigneesMenu,
  ObjectiveHealthIcon,
  PrioritiesMenu,
  PriorityIcon,
} from "@/components/ui";
import { HealthMenu } from "@/components/ui/health-menu";
import { ObjectiveStatusIcon } from "@/components/ui/objective-status-icon";
import { ObjectiveStatusesMenu } from "@/components/ui/objective-statuses-menu";
import {
  useBulkDeleteObjectivesMutation,
  useBulkUpdateObjectivesMutation,
} from "@/modules/objectives/hooks";
import type {
  Objective,
  ObjectiveHealth,
  ObjectiveUpdate,
} from "@/modules/objectives/types";
import { openDialogAfterMenuClose } from "@/utils/menu-dialog-state";

export const ObjectivesToolbar = ({
  objectives,
  onClear,
  selectedObjectiveIds,
}: {
  objectives: Objective[];
  onClear: () => void;
  selectedObjectiveIds: string[];
}) => {
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [isHealthOpen, setIsHealthOpen] = useState(false);
  const [healthComment, setHealthComment] = useState("");
  const [pendingHealth, setPendingHealth] = useState<ObjectiveHealth>(null);
  const bulkUpdate = useBulkUpdateObjectivesMutation();
  const bulkDelete = useBulkDeleteObjectivesMutation();
  const selectedObjectiveIdSet = new Set(selectedObjectiveIds);
  const selectedObjectives = objectives.filter(({ id }) =>
    selectedObjectiveIdSet.has(id),
  );
  const selectedTeamIds = new Set(
    selectedObjectives.map(({ teamId }) => teamId),
  );
  const teamId =
    selectedTeamIds.size === 1 ? selectedObjectives[0]?.teamId : undefined;

  const handleBulkUpdate = (data: ObjectiveUpdate) => {
    bulkUpdate.mutate({ objectiveIds: selectedObjectiveIds, data });
  };

  const closeHealthDialog = () => {
    setIsHealthOpen(false);
    setPendingHealth(null);
    setHealthComment("");
  };

  const handleHealthUpdate = () => {
    if (!pendingHealth || !healthComment.trim()) return;
    bulkUpdate.mutate(
      {
        objectiveIds: selectedObjectiveIds,
        data: { health: pendingHealth, comment: healthComment.trim() },
      },
      { onSuccess: closeHealthDialog },
    );
  };

  return (
    <>
      <Flex
        align="center"
        className="border-border bg-surface/90 shadow-shadow fixed right-1/2 bottom-8 left-1/2 z-50 w-max -translate-x-1/2 rounded-2xl border-[0.5px] px-2.5 py-2 shadow-lg backdrop-blur"
        gap={1}
      >
        <Text
          as="span"
          className="mr-4 flex items-center gap-1.5 px-1 opacity-80"
        >
          <Tooltip title="Clear selection">
            <Button
              color="tertiary"
              leftIcon={<CloseIcon className="h-4" strokeWidth={3} />}
              onClick={onClear}
              size="sm"
              variant="outline"
            >
              <span className="sr-only">Clear</span>
            </Button>
          </Tooltip>
          {selectedObjectiveIds.length} selected
        </Text>

        <ObjectiveStatusesMenu>
          <ObjectiveStatusesMenu.Trigger>
            <Button
              color="tertiary"
              leftIcon={<ObjectiveStatusIcon />}
              variant="naked"
            >
              Status
            </Button>
          </ObjectiveStatusesMenu.Trigger>
          <ObjectiveStatusesMenu.Items
            setStatusId={(statusId) => {
              handleBulkUpdate({ statusId });
            }}
          />
        </ObjectiveStatusesMenu>

        <PrioritiesMenu>
          <PrioritiesMenu.Trigger>
            <Button
              color="tertiary"
              leftIcon={<PriorityIcon />}
              variant="naked"
            >
              Priority
            </Button>
          </PrioritiesMenu.Trigger>
          <PrioritiesMenu.Items
            setPriority={(priority) => {
              handleBulkUpdate({ priority });
            }}
          />
        </PrioritiesMenu>

        <AssigneesMenu>
          <AssigneesMenu.Trigger>
            <Button
              color="tertiary"
              leftIcon={<AssigneeIcon className="h-[1.15rem]" />}
              variant="naked"
            >
              Lead
            </Button>
          </AssigneesMenu.Trigger>
          <AssigneesMenu.Items
            onAssigneeSelected={(leadUser) => {
              handleBulkUpdate({ leadUser });
            }}
            showMaya={false}
            teamId={teamId}
          />
        </AssigneesMenu>

        <DatePicker>
          <DatePicker.Trigger>
            <Button
              color="tertiary"
              leftIcon={<CalendarIcon className="h-[1.15rem]" />}
              variant="naked"
            >
              Target
            </Button>
          </DatePicker.Trigger>
          <DatePicker.Calendar
            onDayClick={(day) => {
              handleBulkUpdate({
                endDate: formatISO(day, { representation: "date" }),
              });
            }}
          />
        </DatePicker>

        <HealthMenu>
          <HealthMenu.Trigger>
            <Button
              color="tertiary"
              leftIcon={<ObjectiveHealthIcon health={null} />}
              variant="naked"
            >
              Health
            </Button>
          </HealthMenu.Trigger>
          <HealthMenu.Items
            setHealth={(health) => {
              setPendingHealth(health);
              openDialogAfterMenuClose(setIsHealthOpen);
            }}
          />
        </HealthMenu>

        <Button
          leftIcon={<DeleteIcon className="h-[1.15rem] text-current" />}
          onClick={() => {
            setIsDeleteOpen(true);
          }}
        >
          Delete
        </Button>
      </Flex>

      <Dialog onOpenChange={setIsDeleteOpen} open={isDeleteOpen}>
        <Dialog.Content>
          <Dialog.Header className="px-6 pt-6">
            <Dialog.Title className="text-lg">
              Delete {selectedObjectiveIds.length} objective
              {selectedObjectiveIds.length === 1 ? "" : "s"}?
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="pt-0">
            <Text color="muted">
              This will permanently delete the selected objectives and their
              associated key results.
            </Text>
            <Flex align="center" className="mt-4" gap={2} justify="end">
              <Button
                color="tertiary"
                disabled={bulkDelete.isPending}
                onClick={() => {
                  setIsDeleteOpen(false);
                }}
              >
                Cancel
              </Button>
              <Button
                leftIcon={<DeleteIcon className="text-current" />}
                loading={bulkDelete.isPending}
                loadingText="Deleting..."
                onClick={() => {
                  bulkDelete.mutate(selectedObjectiveIds, {
                    onSuccess: () => {
                      setIsDeleteOpen(false);
                      onClear();
                    },
                  });
                }}
              >
                Delete
              </Button>
            </Flex>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>

      <Dialog
        onOpenChange={(open) => {
          if (!open) closeHealthDialog();
        }}
        open={isHealthOpen}
      >
        <Dialog.Content>
          <Dialog.Header className="px-6 pt-6">
            <Dialog.Title className="flex items-center gap-2 text-lg">
              Change health to
              <ObjectiveHealthIcon health={pendingHealth} />
              {pendingHealth}
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Description>
            Add one comment explaining this health change for the selected
            objectives.
          </Dialog.Description>
          <Dialog.Body>
            <Text className="mb-1.5" color="muted">
              Comment*
            </Text>
            <TextArea
              className="border-border/80 resize-none rounded-2xl border bg-transparent py-4 leading-normal"
              onChange={(event) => {
                setHealthComment(event.target.value);
              }}
              placeholder="Explain why the health status is changing..."
              rows={4}
              value={healthComment}
            />
          </Dialog.Body>
          <Dialog.Footer className="justify-end gap-2">
            <Button color="tertiary" onClick={closeHealthDialog}>
              Cancel
            </Button>
            <Button
              disabled={!healthComment.trim() || bulkUpdate.isPending}
              loading={bulkUpdate.isPending}
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
