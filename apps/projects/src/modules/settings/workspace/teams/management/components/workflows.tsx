"use client";

import { useState } from "react";
import { Box, Text, Button, Flex, Tooltip, Wrapper } from "ui";
import { PlusIcon, WarningIcon } from "icons";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import { DndContext, type DragEndEvent } from "@dnd-kit/core";
import {
  SortableContext,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { SectionHeader } from "@/modules/settings/components/section-header";
import { useDeleteStateMutation } from "@/lib/hooks/states/delete-mutation";
import { useCreateStateMutation } from "@/lib/hooks/states/create-mutation";
import { useUpdateStateMutation } from "@/lib/hooks/states/update-mutation";
import { ConfirmDialog, FeatureGuard } from "@/components/ui";
import type { State, StateCategory } from "@/types/states";
import { useTeamStories } from "@/modules/stories/hooks/team-stories";
import { useUserRole } from "@/hooks";
import { StateRow } from "./state-row";

const categories: {
  label: string;
  value: StateCategory;
  minOrderIndex: number;
  maxOrderIndex: number;
}[] = [
  { label: "Backlog", value: "backlog", minOrderIndex: 1, maxOrderIndex: 1999 },
  {
    label: "Unstarted",
    value: "unstarted",
    minOrderIndex: 2000,
    maxOrderIndex: 2999,
  },
  {
    label: "Started",
    value: "started",
    minOrderIndex: 3000,
    maxOrderIndex: 3999,
  },
  {
    label: "Completed",
    value: "completed",
    minOrderIndex: 4000,
    maxOrderIndex: 4999,
  },
  {
    label: "Paused",
    value: "paused",
    minOrderIndex: 5000,
    maxOrderIndex: 5999,
  },
  {
    label: "Cancelled",
    value: "cancelled",
    minOrderIndex: 6000,
    maxOrderIndex: 6999,
  },
];

// Helper function to reorder array items
const arrayMove = <T,>(array: T[], from: number, to: number): T[] => {
  const newArray = [...array];
  const [removed] = newArray.splice(from, 1);
  newArray.splice(to, 0, removed);
  return newArray;
};

export const WorkflowSettings = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const { userRole } = useUserRole();
  const { data: states = [] } = useTeamStatuses(teamId);
  const { data: stories = [] } = useTeamStories(teamId);
  const deleteMutation = useDeleteStateMutation();
  const createMutation = useCreateStateMutation();
  const updateMutation = useUpdateStateMutation();
  const [selectedCategory, setSelectedCategory] =
    useState<StateCategory | null>(null);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [stateToDelete, setStateToDelete] = useState<State | null>(null);

  const handleDeleteState = (state: State) => {
    const categoryStates = states.filter((s) => s.category === state.category);
    if (
      categoryStates.length <= 1 &&
      ["unstarted", "started"].includes(state.category)
    ) {
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
          teamId,
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

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;

    if (!over || active.id === over.id) {
      return;
    }

    const activeState = states.find((state) => state.id === active.id);
    if (!activeState) return;

    // Find the category configuration
    const categoryConfig = categories.find(
      (cat) => cat.value === activeState.category,
    );
    if (!categoryConfig) return;

    const categoryStates = states
      .filter((state) => state.category === activeState.category)
      .sort((a, b) => a.orderIndex - b.orderIndex);

    const oldIndex = categoryStates.findIndex(
      (state) => state.id === active.id,
    );
    const newIndex = categoryStates.findIndex((state) => state.id === over.id);

    if (oldIndex === -1 || newIndex === -1 || oldIndex === newIndex) return;

    const { minOrderIndex, maxOrderIndex } = categoryConfig;

    // Check if any existing statuses are outside the expected range (legacy statuses)
    const hasLegacyStatuses = categoryStates.some(
      (state) =>
        state.orderIndex < minOrderIndex || state.orderIndex > maxOrderIndex,
    );

    // If there are legacy statuses, always do full reorder to fix them
    if (hasLegacyStatuses) {
      const reorderedStates = arrayMove(categoryStates, oldIndex, newIndex);
      const range = maxOrderIndex - minOrderIndex;
      const idealStep = Math.floor(range / (reorderedStates.length + 1));
      const step = Math.max(50, idealStep);

      reorderedStates.forEach((state, index) => {
        const newOrderIndex = minOrderIndex + (index + 1) * step;
        const clampedOrderIndex = Math.min(
          newOrderIndex,
          maxOrderIndex - (reorderedStates.length - index - 1) * 50,
        );

        if (state.orderIndex !== clampedOrderIndex) {
          updateMutation.mutate({
            stateId: state.id,
            payload: { orderIndex: clampedOrderIndex },
          });
        }
      });
      return;
    }

    // Create a copy without the dragged item to check positions
    const otherStates = categoryStates.filter(
      (state) => state.id !== activeState.id,
    );

    // Calculate the desired orderIndex based on drop position
    let targetOrderIndex: number;

    if (oldIndex < newIndex) {
      // Moving down
      const targetItem = otherStates[newIndex - 1]; // Adjust index since we removed active item
      targetOrderIndex = targetItem.orderIndex + 10;
    } else {
      // Moving up
      const targetItem = otherStates[newIndex];
      targetOrderIndex = targetItem.orderIndex - 10;
    }

    // Ensure within bounds
    targetOrderIndex = Math.max(
      minOrderIndex,
      Math.min(maxOrderIndex, targetOrderIndex),
    );

    // Check for conflicts and boundary violations
    const hasExactConflict = categoryStates.some(
      (state) =>
        state.id !== activeState.id && state.orderIndex === targetOrderIndex,
    );

    const isOutOfBounds =
      targetOrderIndex < minOrderIndex || targetOrderIndex > maxOrderIndex;

    const needsFullReorder =
      targetOrderIndex === -1 || hasExactConflict || isOutOfBounds;

    if (needsFullReorder) {
      // Fall back to full reordering with guaranteed spacing
      const reorderedStates = arrayMove(categoryStates, oldIndex, newIndex);
      const range = maxOrderIndex - minOrderIndex;
      const idealStep = Math.floor(range / (reorderedStates.length + 1));
      const step = Math.max(50, idealStep); // Ensure minimum 50-point spacing

      reorderedStates.forEach((state, index) => {
        const newOrderIndex = minOrderIndex + (index + 1) * step;
        // Ensure we don't exceed max boundary
        const clampedOrderIndex = Math.min(
          newOrderIndex,
          maxOrderIndex - (reorderedStates.length - index - 1) * 50,
        );

        if (state.orderIndex !== clampedOrderIndex) {
          updateMutation.mutate({
            stateId: state.id,
            payload: { orderIndex: clampedOrderIndex },
          });
        }
      });
    } else {
      // Simple update - only change the dragged item
      if (activeState.orderIndex !== targetOrderIndex) {
        updateMutation.mutate({
          stateId: activeState.id,
          payload: { orderIndex: targetOrderIndex },
        });
      }
    }
  };

  return (
    <FeatureGuard
      fallback={
        <Wrapper className="border-warning bg-warning/10 dark:border-warning/20 dark:bg-warning/10 mb-6 flex items-center justify-between gap-2 border p-4">
          <Flex align="center" gap={2}>
            <WarningIcon className="text-warning dark:text-warning" />
            <Text>
              {userRole === "admin" ? "Upgrade" : "Ask your admin to upgrade"}{" "}
              to create custom team workflows
            </Text>
          </Flex>
          {userRole === "admin" && (
            <Button color="warning" href="/settings/workspace/billing">
              Upgrade now
            </Button>
          )}
        </Wrapper>
      }
      feature="customWorkflows"
    >
      <Box className="mb-6 rounded-2xl border border-border bg-surface-elevated pb-6">
        <SectionHeader
          className="mb-4"
          description="Configure custom workflow states to track the progress of your team's work. Each category represents a different phase in your workflow process."
          title="Team Workflow"
        />
        <DndContext onDragEnd={handleDragEnd}>
          <Flex direction="column" gap={4}>
            {categories.map(({ label, value }) => {
              const categoryStates = states
                .filter((state) => state.category === value)
                .sort((a, b) => a.orderIndex - b.orderIndex);
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
                        <span className="sr-only">Add State</span>
                      </Button>
                    </Tooltip>
                  </Flex>

                  <SortableContext
                    items={categoryStates.map((state) => state.id)}
                    strategy={verticalListSortingStrategy}
                  >
                    <Flex direction="column" gap={3}>
                      {categoryStates.map((state) => {
                        const storyCount = stories.filter(
                          (story) => story.statusId === state.id,
                        ).length;

                        return (
                          <StateRow
                            key={state.id}
                            onDelete={handleDeleteState}
                            state={state}
                            storyCount={storyCount}
                          />
                        );
                      })}
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
                            color: categoryStates[0]?.color || "#6366F1",
                            isDefault: false,
                            category: value,
                            orderIndex: 9999, // Temporary value, backend will set the actual orderIndex
                            teamId,
                            workspaceId: "",
                            createdAt: new Date().toISOString(),
                            updatedAt: new Date().toISOString(),
                          }}
                          storyCount={0}
                        />
                      ) : null}
                    </Flex>
                  </SortableContext>
                </Box>
              );
            })}
          </Flex>
        </DndContext>
        <ConfirmDialog
          confirmText="Delete state"
          description="Are you sure you want to delete this state? Any stories in this state will need to be moved to another state."
          isOpen={Boolean(stateToDelete)}
          onClose={() => {
            setStateToDelete(null);
          }}
          onConfirm={confirmDelete}
          title="Delete state"
        />
      </Box>
    </FeatureGuard>
  );
};
