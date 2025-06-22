"use client";

import { useState } from "react";
import { Box, Text, Button, Flex, Tooltip, Wrapper } from "ui";
import { PlusIcon, WarningIcon } from "icons";
import { toast } from "sonner";
import { DndContext, type DragEndEvent } from "@dnd-kit/core";
import {
  SortableContext,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";
import { SectionHeader } from "@/modules/settings/components/section-header";
import { ConfirmDialog, FeatureGuard } from "@/components/ui";
import type { StateCategory } from "@/types/states";
import {
  useCreateObjectiveStatusMutation,
  useDeleteObjectiveStatusMutation,
  useUpdateObjectiveStatusMutation,
} from "@/modules/objectives/hooks/statuses";
import type { ObjectiveStatus } from "@/modules/objectives/types";
import { useTerminology, useUserRole } from "@/hooks";
import { StateRow } from "./components/state-row";

const categories: {
  label: string;
  value: StateCategory;
  minOrderIndex: number;
  maxOrderIndex: number;
}[] = [
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
    label: "Paused",
    value: "paused",
    minOrderIndex: 5000,
    maxOrderIndex: 5999,
  },
  {
    label: "Completed",
    value: "completed",
    minOrderIndex: 4000,
    maxOrderIndex: 4999,
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
  const { data: statuses = [] } = useObjectiveStatuses();
  const deleteMutation = useDeleteObjectiveStatusMutation();
  const createMutation = useCreateObjectiveStatusMutation();
  const updateMutation = useUpdateObjectiveStatusMutation();
  const [selectedCategory, setSelectedCategory] =
    useState<StateCategory | null>(null);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [statusToDelete, setStatusToDelete] = useState<ObjectiveStatus | null>(
    null,
  );
  const { getTermDisplay } = useTerminology();
  const { userRole } = useUserRole();

  const handleDeleteState = (status: ObjectiveStatus) => {
    const categoryStatuses = statuses.filter(
      (s) => s.category === status.category,
    );
    if (categoryStatuses.length <= 1) {
      toast.warning("Status cannot be removed", {
        description:
          "Create another status in this category before deleting this one. You can also update the status instead.",
      });
      return;
    }
    setStatusToDelete(status);
  };

  const confirmDelete = () => {
    if (statusToDelete) {
      deleteMutation.mutate(statusToDelete.id);
      setStatusToDelete(null);
    }
  };

  const handleCreateState = (status: ObjectiveStatus) => {
    if (selectedCategory) {
      createMutation.mutate(
        {
          name: status.name,
          category: selectedCategory,
          color: status.color,
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

    const activeStatus = statuses.find((status) => status.id === active.id);
    if (!activeStatus) return;

    // Find the category configuration
    const categoryConfig = categories.find(
      (cat) => cat.value === activeStatus.category,
    );
    if (!categoryConfig) return;

    const categoryStatuses = statuses
      .filter((status) => status.category === activeStatus.category)
      .sort((a, b) => a.orderIndex - b.orderIndex);

    const oldIndex = categoryStatuses.findIndex(
      (status) => status.id === active.id,
    );
    const newIndex = categoryStatuses.findIndex(
      (status) => status.id === over.id,
    );

    if (oldIndex === -1 || newIndex === -1 || oldIndex === newIndex) return;

    const { minOrderIndex, maxOrderIndex } = categoryConfig;

    // Check if any existing statuses are outside the expected range (legacy statuses)
    const hasLegacyStatuses = categoryStatuses.some(
      (status) =>
        status.orderIndex < minOrderIndex || status.orderIndex > maxOrderIndex,
    );

    // If there are legacy statuses, always do full reorder to fix them
    if (hasLegacyStatuses) {
      const reorderedStatuses = arrayMove(categoryStatuses, oldIndex, newIndex);
      const range = maxOrderIndex - minOrderIndex;
      const idealStep = Math.floor(range / (reorderedStatuses.length + 1));
      const step = Math.max(50, idealStep);

      reorderedStatuses.forEach((status, index) => {
        const newOrderIndex = minOrderIndex + (index + 1) * step;
        const clampedOrderIndex = Math.min(
          newOrderIndex,
          maxOrderIndex - (reorderedStatuses.length - index - 1) * 50,
        );

        if (status.orderIndex !== clampedOrderIndex) {
          updateMutation.mutate({
            statusId: status.id,
            payload: { orderIndex: clampedOrderIndex },
          });
        }
      });
      return;
    }

    // Create a copy without the dragged item to check positions
    const otherStatuses = categoryStatuses.filter(
      (status) => status.id !== activeStatus.id,
    );

    // Calculate the desired orderIndex based on drop position
    let targetOrderIndex: number;

    if (oldIndex < newIndex) {
      // Moving down
      const targetItem = otherStatuses[newIndex - 1]; // Adjust index since we removed active item
      targetOrderIndex = targetItem.orderIndex + 10;
    } else {
      // Moving up
      const targetItem = otherStatuses[newIndex];
      targetOrderIndex = targetItem.orderIndex - 10;
    }

    // Ensure within bounds
    targetOrderIndex = Math.max(
      minOrderIndex,
      Math.min(maxOrderIndex, targetOrderIndex),
    );

    // Check for conflicts and boundary violations
    const hasExactConflict = categoryStatuses.some(
      (status) =>
        status.id !== activeStatus.id && status.orderIndex === targetOrderIndex,
    );

    const isOutOfBounds =
      targetOrderIndex < minOrderIndex || targetOrderIndex > maxOrderIndex;

    const needsFullReorder =
      targetOrderIndex === -1 || hasExactConflict || isOutOfBounds;

    if (needsFullReorder) {
      // Fall back to full reordering with guaranteed spacing
      const reorderedStatuses = arrayMove(categoryStatuses, oldIndex, newIndex);
      const range = maxOrderIndex - minOrderIndex;
      const idealStep = Math.floor(range / (reorderedStatuses.length + 1));
      const step = Math.max(50, idealStep); // Ensure minimum 50-point spacing

      reorderedStatuses.forEach((status, index) => {
        const newOrderIndex = minOrderIndex + (index + 1) * step;
        // Ensure we don't exceed max boundary
        const clampedOrderIndex = Math.min(
          newOrderIndex,
          maxOrderIndex - (reorderedStatuses.length - index - 1) * 50,
        );

        if (status.orderIndex !== clampedOrderIndex) {
          updateMutation.mutate({
            statusId: status.id,
            payload: { orderIndex: clampedOrderIndex },
          });
        }
      });
    } else {
      // Simple update - only change the dragged item
      if (activeStatus.orderIndex !== targetOrderIndex) {
        updateMutation.mutate({
          statusId: activeStatus.id,
          payload: { orderIndex: targetOrderIndex },
        });
      }
    }
  };

  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        {getTermDisplay("objectiveTerm", {
          variant: "plural",
          capitalize: true,
        })}{" "}
        workflow
      </Text>
      <FeatureGuard
        fallback={
          <Wrapper className="mb-6 flex items-center justify-between gap-2 rounded-[0.6rem] border border-warning bg-warning/10 p-4 dark:border-warning/20 dark:bg-warning/10">
            <Flex align="center" gap={2}>
              <WarningIcon className="text-warning dark:text-warning" />
              <Text>
                {userRole === "admin" ? "Upgrade" : "Ask your admin to upgrade"}{" "}
                to a business or enterprise plan to customize workflow states
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
        <Box className="mb-6 rounded-2xl border border-gray-100 bg-white pb-6 dark:border-dark-100 dark:bg-dark-100/40">
          <SectionHeader
            className="mb-4"
            description={`Configure custom workflow states to track the progress of ${getTermDisplay(
              "objectiveTerm",
              {
                variant: "plural",
              },
            )} across your workspace. Each category represents a different phase in your workflow process.`}
          />
          <DndContext onDragEnd={handleDragEnd}>
            <Flex direction="column" gap={4}>
              {categories.map(({ label, value }) => {
                const categoryStatuses = statuses
                  .filter((status) => status.category === value)
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
                          <span className="sr-only">Add Status</span>
                        </Button>
                      </Tooltip>
                    </Flex>

                    <SortableContext
                      items={categoryStatuses.map((status) => status.id)}
                      strategy={verticalListSortingStrategy}
                    >
                      <Flex direction="column" gap={3}>
                        {categoryStatuses.map((status) => (
                          <StateRow
                            key={status.id}
                            onDelete={handleDeleteState}
                            status={status}
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
                            status={{
                              id: "new",
                              name: "",
                              color: "#6366F1",
                              isDefault: false,
                              category: value,
                              orderIndex: 9999, // Temporary value, backend will set the actual orderIndex
                              workspaceId: "",
                              createdAt: new Date().toISOString(),
                              updatedAt: new Date().toISOString(),
                            }}
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
            confirmText="Delete status"
            description="Are you sure you want to delete this status? Any objectives in this status will need to be moved to another status."
            isOpen={Boolean(statusToDelete)}
            onClose={() => {
              setStatusToDelete(null);
            }}
            onConfirm={confirmDelete}
            title="Delete status"
          />
        </Box>
      </FeatureGuard>
    </Box>
  );
};
