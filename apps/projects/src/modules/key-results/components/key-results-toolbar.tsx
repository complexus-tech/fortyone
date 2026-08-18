"use client";

import { useState } from "react";
import { formatISO } from "date-fns";
import { AssigneeIcon, CalendarIcon, CloseIcon, DeleteIcon } from "icons";
import { Button, DatePicker, Flex, Text, Tooltip } from "ui";
import { AssigneesMenu, ConfirmDialog } from "@/components/ui";
import { useTerminology } from "@/hooks";
import { useDeleteKeyResultMutation } from "@/modules/objectives/hooks/use-delete-key-result-mutation";
import { useUpdateKeyResultMutation } from "@/modules/objectives/hooks/use-update-key-result-mutation";
import type { KeyResultWithTeam } from "../types";

export const KeyResultsToolbar = ({
  clearSelection,
  selectedKeyResults,
}: {
  clearSelection: () => void;
  selectedKeyResults: KeyResultWithTeam[];
}) => {
  const { getTermDisplay } = useTerminology();
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const { mutate: updateKeyResult } = useUpdateKeyResultMutation();
  const { mutate: deleteKeyResult } = useDeleteKeyResultMutation();
  const teamIds = new Set(selectedKeyResults.map(({ teamId }) => teamId));
  const sharedTeamId =
    teamIds.size === 1 ? selectedKeyResults[0]?.teamId : undefined;
  const selectedKeyResultLabel = getTermDisplay("keyResultTerm", {
    variant: selectedKeyResults.length === 1 ? "singular" : "plural",
  });

  const updateSelected = (data: { endDate?: string; lead?: string | null }) => {
    for (const keyResult of selectedKeyResults) {
      updateKeyResult({
        keyResultId: keyResult.id,
        objectiveId: keyResult.objectiveId,
        data,
        silent: true,
      });
    }
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
              onClick={clearSelection}
              size="sm"
              variant="outline"
            >
              <span className="sr-only">Clear</span>
            </Button>
          </Tooltip>
          {selectedKeyResults.length} selected
        </Text>
        <DatePicker>
          <DatePicker.Trigger>
            <Button
              color="tertiary"
              leftIcon={
                <CalendarIcon className="text-primary dark:text-primary h-[1.15rem]" />
              }
              variant="naked"
            >
              Deadline
            </Button>
          </DatePicker.Trigger>
          <DatePicker.Calendar
            onDayClick={(day) => {
              updateSelected({ endDate: formatISO(day) });
            }}
          />
        </DatePicker>
        {sharedTeamId ? (
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
              onAssigneeSelected={(lead) => {
                updateSelected({ lead });
              }}
              placeholder="Assign lead..."
              teamId={sharedTeamId}
            />
          </AssigneesMenu>
        ) : null}
        <Button
          leftIcon={<DeleteIcon className="h-[1.15rem] text-current" />}
          onClick={() => {
            setIsDeleteOpen(true);
          }}
        >
          Delete
        </Button>
      </Flex>
      <ConfirmDialog
        confirmText="Yes, Delete"
        description={`Are you sure you want to delete ${selectedKeyResults.length} ${selectedKeyResultLabel}? This action cannot be undone.`}
        isOpen={isDeleteOpen}
        onClose={() => {
          setIsDeleteOpen(false);
        }}
        onConfirm={() => {
          for (const keyResult of selectedKeyResults) {
            deleteKeyResult({
              keyResultId: keyResult.id,
              objectiveId: keyResult.objectiveId,
            });
          }
          clearSelection();
          setIsDeleteOpen(false);
        }}
        title={`Delete ${getTermDisplay("keyResultTerm", {
          variant: "plural",
          capitalize: true,
        })}`}
      />
    </>
  );
};
