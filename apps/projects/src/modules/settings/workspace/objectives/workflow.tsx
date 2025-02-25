"use client";

import { useState } from "react";
import { Box, Text, Button, Flex, Tooltip } from "ui";
import { PlusIcon } from "icons";
import { toast } from "sonner";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";
import { SectionHeader } from "@/modules/settings/components/section-header";
import { ConfirmDialog } from "@/components/ui";
import type { State, StateCategory } from "@/types/states";
import {
  useCreateObjectiveStatusMutation,
  useDeleteObjectiveStatusMutation,
} from "@/modules/objectives/hooks/statuses";
import { StateRow } from "./components/state-row";

const categories: { label: string; value: StateCategory }[] = [
  { label: "Unstarted", value: "unstarted" },
  { label: "Started", value: "started" },
  { label: "Paused", value: "paused" },
  { label: "Completed", value: "completed" },
  { label: "Cancelled", value: "cancelled" },
];

export const WorkflowSettings = () => {
  const { data: states = [] } = useObjectiveStatuses();
  const deleteMutation = useDeleteObjectiveStatusMutation();
  const createMutation = useCreateObjectiveStatusMutation();
  const [selectedCategory, setSelectedCategory] =
    useState<StateCategory | null>(null);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [stateToDelete, setStateToDelete] = useState<State | null>(null);

  const handleDeleteState = (state: State) => {
    const categoryStates = states.filter((s) => s.category === state.category);
    if (categoryStates.length <= 1) {
      toast.warning("Status cannot be removed", {
        description:
          "Create another status in this category before deleting this one. You can also update the status instead.",
      });
      return;
    }
    setStateToDelete(state);
  };

  const confirmDelete = () => {
    if (stateToDelete) {
      deleteMutation.mutate(stateToDelete.id);
      setStateToDelete(null);
    }
  };

  const handleCreateState = (state: State) => {
    if (selectedCategory) {
      createMutation.mutate(
        {
          name: state.name,
          category: selectedCategory,
          color: state.color,
        },
        {
          onSuccess: () => {
            setIsCreateOpen(false);
          },
        },
      );
    }
  };

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Objectives
      </Text>
      <Box className="mb-6 rounded-lg border border-gray-100 bg-white pb-6 dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          className="mb-4"
          description="Configure custom workflow states to track the progress of objectives across your workspace. Each category represents a different phase in your workflow process."
        />
        <Flex direction="column" gap={4}>
          {categories.map(({ label, value }) => {
            const categoryStates = states.filter(
              (state) => state.category === value,
            );
            const isCreatingInCategory =
              selectedCategory === value && isCreateOpen;

            return (
              <Box className="px-6" key={value}>
                <Flex align="center" className="mb-2" justify="between">
                  <Text color="muted">{label}</Text>
                  <Tooltip title="Add Status">
                    <Button
                      color="tertiary"
                      leftIcon={<PlusIcon />}
                      onClick={() => {
                        setSelectedCategory(value);
                        setIsCreateOpen(true);
                      }}
                      size="sm"
                      variant="naked"
                    >
                      <span className="sr-only">Add Status</span>
                    </Button>
                  </Tooltip>
                </Flex>

                <Flex direction="column" gap={3}>
                  {categoryStates.map((state) => (
                    <StateRow
                      key={state.id}
                      onDelete={handleDeleteState}
                      state={state}
                    />
                  ))}
                  {isCreatingInCategory ? (
                    <StateRow
                      isNew
                      onCreate={handleCreateState}
                      onCreateCancel={() => {
                        setIsCreateOpen(false);
                        setSelectedCategory(null);
                      }}
                      onDelete={() => {}}
                      state={{
                        id: "new",
                        name: "",
                        color: "#6366F1",
                        category: value,
                        orderIndex: 50,
                        teamId: "workspace",
                        workspaceId: "workspace",
                        createdAt: new Date().toISOString(),
                        updatedAt: new Date().toISOString(),
                      }}
                    />
                  ) : null}
                </Flex>
              </Box>
            );
          })}
        </Flex>
        <ConfirmDialog
          confirmText="Delete status"
          description="Are you sure you want to delete this status? Any objectives in this status will need to be moved to another status."
          isOpen={Boolean(stateToDelete)}
          onClose={() => {
            setStateToDelete(null);
          }}
          onConfirm={confirmDelete}
          title="Delete status"
        />
      </Box>
    </Box>
  );
};
