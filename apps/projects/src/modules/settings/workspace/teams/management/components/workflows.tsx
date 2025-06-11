"use client";

import { useState } from "react";
import { Box, Text, Button, Flex, Tooltip, Wrapper } from "ui";
import { PlusIcon, WarningIcon } from "icons";
import { useParams } from "next/navigation";
import { toast } from "sonner";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { SectionHeader } from "@/modules/settings/components/section-header";
import { useDeleteStateMutation } from "@/lib/hooks/states/delete-mutation";
import { useCreateStateMutation } from "@/lib/hooks/states/create-mutation";
import { ConfirmDialog, FeatureGuard } from "@/components/ui";
import type { State, StateCategory } from "@/types/states";
import { useTeamStories } from "@/modules/stories/hooks/team-stories";
import { useUserRole } from "@/hooks";
import { StateRow } from "./state-row";

const categories: { label: string; value: StateCategory }[] = [
  { label: "Backlog", value: "backlog" },
  { label: "Unstarted", value: "unstarted" },
  { label: "Started", value: "started" },
  { label: "Completed", value: "completed" },
  { label: "Paused", value: "paused" },
  { label: "Cancelled", value: "cancelled" },
];

export const WorkflowSettings = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const { userRole } = useUserRole();
  const { data: states = [] } = useTeamStatuses(teamId);
  const { data: stories = [] } = useTeamStories(teamId);
  const deleteMutation = useDeleteStateMutation();
  const createMutation = useCreateStateMutation();
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

  return (
    <FeatureGuard
      fallback={
        <Wrapper className="mb-6 flex items-center justify-between gap-2 rounded-lg border border-warning bg-warning/10 p-4 dark:border-warning/20 dark:bg-warning/10">
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
      <Box className="mb-6 rounded-lg border border-gray-100 bg-white pb-6 dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          className="mb-4"
          description="Configure custom workflow states to track the progress of your team's work. Each category represents a different phase in your workflow process."
          title="Team Workflow"
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
                      <span className="sr-only">Add State</span>
                    </Button>
                  </Tooltip>
                </Flex>

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
                        color: "#6366F1",
                        isDefault: false,
                        category: value,
                        orderIndex: 50,
                        teamId,
                        workspaceId: "",
                        createdAt: new Date().toISOString(),
                        updatedAt: new Date().toISOString(),
                      }}
                      storyCount={0}
                    />
                  ) : null}
                </Flex>
              </Box>
            );
          })}
        </Flex>
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
