import { useId, useState } from "react";
import { Button, Flex, Skeleton } from "ui";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { useTerminology, useUserRole } from "@/hooks";
import type { DetailedStory } from "../types";
import { useChecklistMutation, useStoryChecklist } from "../hooks/checklist";
import { ChecklistAddItem } from "./checklist-add-item";
import { ChecklistHeader } from "./checklist-header";
import { ChecklistItems } from "./checklist-items";

export const StoryChecklist = ({ story }: { story: DetailedStory }) => {
  const headingId = useId();
  const { userRole } = useUserRole();
  const { getTermDisplay } = useTerminology(story.teamId);
  const {
    data: items = [],
    isLoading,
    isError,
    isFetching,
    refetch,
  } = useStoryChecklist(story.id);
  const mutation = useChecklistMutation(story.id);
  const [isAdding, setIsAdding] = useState(false);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [isChecklistOpen, setIsChecklistOpen] = useState(true);
  const canEdit = Boolean(
    userRole && userRole !== "guest" && !story.deletedAt && !story.archivedAt,
  );
  const canMutate = canEdit && !isError;
  const completedCount = items.filter(({ completed }) => completed).length;
  if (isLoading) {
    return (
      <div
        aria-label="Loading checklist"
        className="mt-6 space-y-3"
        role="status"
      >
        <Skeleton className="h-5 w-32" />
        <Skeleton className="h-8 w-full" />
      </div>
    );
  }

  if (!items.length && !isAdding && !isDeleteOpen && !isError) {
    return canEdit ? (
      <Button
        className="mt-6"
        color="tertiary"
        disabled={mutation.isPending}
        loading={mutation.isPending}
        loadingText="Saving checklist..."
        onClick={() => {
          setIsChecklistOpen(true);
          setIsAdding(true);
        }}
        size="sm"
        variant="naked"
      >
        Add checklist
      </Button>
    ) : null;
  }

  return (
    <section
      aria-busy={mutation.isPending}
      aria-labelledby={headingId}
      className="mt-6"
    >
      <ChecklistHeader
        canEdit={canMutate}
        completedCount={completedCount}
        id={headingId}
        isOpen={isChecklistOpen}
        isPending={mutation.isPending}
        itemCount={items.length}
        onDelete={() => {
          setIsDeleteOpen(true);
        }}
        onToggle={() => {
          setIsChecklistOpen(!isChecklistOpen);
        }}
      />
      {isError ? (
        <Flex align="center" className="py-3" gap={3}>
          <p className="text-text-muted" role="alert">
            Could not load this checklist.
          </p>
          <Button
            color="tertiary"
            disabled={isFetching}
            onClick={() => {
              void refetch();
            }}
            size="sm"
            variant="outline"
          >
            {isFetching ? "Retrying..." : "Retry"}
          </Button>
        </Flex>
      ) : null}
      <div hidden={!isChecklistOpen}>
        <ChecklistItems
          canEdit={canMutate}
          id={headingId}
          items={items}
          mutation={mutation}
          storyTerm={getTermDisplay("storyTerm")}
        />
        {canMutate ? (
          <ChecklistAddItem
            isAdding={isAdding}
            isPending={mutation.isPending}
            itemCount={items.length}
            onAdd={() => {
              setIsAdding(true);
            }}
            onCancel={() => {
              setIsAdding(false);
            }}
            onSave={(title) => {
              mutation.mutate(
                { type: "create", title },
                {
                  onSuccess: () => {
                    setIsAdding(false);
                  },
                },
              );
            }}
          />
        ) : null}
      </div>
      <ConfirmDialog
        confirmText="Delete checklist"
        description="All checklist items will be removed. Any stories created from these items will be kept."
        isLoading={mutation.isPending}
        isOpen={Boolean(isDeleteOpen && canMutate)}
        loadingText="Deleting..."
        onClose={() => {
          setIsDeleteOpen(false);
        }}
        onConfirm={() => {
          mutation.mutate(
            { type: "clear" },
            {
              onSuccess: () => {
                setIsDeleteOpen(false);
                setIsAdding(false);
              },
            },
          );
        }}
        title="Delete checklist?"
      />
    </section>
  );
};
