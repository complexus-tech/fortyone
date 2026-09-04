"use client";
import type { Dispatch, SetStateAction } from "react";
import { cn } from "lib";
import { Box, Button, Checkbox, Flex, Input, Text } from "ui";
import type { ImportDraft } from "../schema";
import { REVIEW_PAGE_SIZE } from "./use-import-selection";
import { useImportTerms } from "./use-import-terms";

export type ImportTaskReviewProps = {
  draft: ImportDraft;
  archivedTrelloTaskIndexes: Set<number>;
  includeArchivedTrelloCards: boolean;
  toggleArchivedTrelloCards: (included: boolean) => void;
  selectedTasks: ImportDraft["tasks"];
  visibleReviewTasks: {
    task: ImportDraft["tasks"][number];
    taskIndex: number;
  }[];
  excludedRows: Set<number>;
  toggleTask: (index: number, checked: boolean) => void;
  updateTaskTitle: (index: number, title: string) => void;
  reviewTasks: { task: ImportDraft["tasks"][number]; taskIndex: number }[];
  reviewPageStart: number;
  reviewPage: number;
  reviewPageCount: number;
  setReviewPage: Dispatch<SetStateAction<number>>;
};
export const ImportTaskReview = ({
  draft,
  archivedTrelloTaskIndexes,
  includeArchivedTrelloCards,
  toggleArchivedTrelloCards,
  selectedTasks,
  visibleReviewTasks,
  excludedRows,
  toggleTask,
  updateTaskTitle,
  reviewTasks,
  reviewPageStart,
  reviewPage,
  reviewPageCount,
  setReviewPage,
}: ImportTaskReviewProps) => {
  const { storyTermCapitalized } = useImportTerms();
  return (
    <>
      {draft.tasks.length ? (
        <Box className="border-border bg-surface mt-5 overflow-hidden rounded-xl border-[0.5px]">
          <Flex
            align="center"
            className="border-border border-b-[0.5px] px-4 py-3"
            justify="between"
          >
            <Text className="font-medium">{storyTermCapitalized} review</Text>
            <Flex align="center" className="flex-wrap justify-end" gap={4}>
              {archivedTrelloTaskIndexes.size > 0 ? (
                <Flex align="center" gap={2}>
                  <Checkbox
                    aria-label={`Include ${archivedTrelloTaskIndexes.size} archived cards`}
                    checked={includeArchivedTrelloCards}
                    id="include-archived-trello-cards"
                    onCheckedChange={(checked) => {
                      toggleArchivedTrelloCards(checked === true);
                    }}
                  />
                  <label
                    className="cursor-pointer whitespace-nowrap"
                    htmlFor="include-archived-trello-cards"
                  >
                    Include archived ({archivedTrelloTaskIndexes.size})
                  </label>
                </Flex>
              ) : null}
              <Text color="muted">{selectedTasks.length} selected</Text>
            </Flex>
          </Flex>
          <Box className="divide-border max-h-96 divide-y overflow-y-auto">
            {visibleReviewTasks.map(({ task, taskIndex }) => {
              const isExcluded = excludedRows.has(taskIndex);
              const titleMissing =
                !isExcluded && task.title.trim().length === 0;
              return (
                <Flex
                  align="center"
                  className={cn(
                    "gap-2.5 px-4 py-2",
                    isExcluded && "opacity-55",
                  )}
                  key={`${task.sourceId}-${taskIndex}`}
                >
                  <Checkbox
                    aria-label={`Import ${task.title}`}
                    checked={!isExcluded}
                    onCheckedChange={(checked) => {
                      toggleTask(taskIndex, checked === true);
                    }}
                  />
                  <Box className="min-w-0 flex-1">
                    <Input
                      aria-invalid={titleMissing}
                      aria-label={`Title for ${task.sourceId}`}
                      className={cn(
                        "h-9 bg-transparent px-2 text-base font-medium dark:bg-transparent",
                        titleMissing
                          ? "border-danger"
                          : "hover:border-input border-transparent",
                      )}
                      disabled={isExcluded}
                      hasError={titleMissing}
                      maxLength={255}
                      onChange={(event) => {
                        updateTaskTitle(taskIndex, event.target.value);
                      }}
                      placeholder="Add a title"
                      value={task.title}
                    />
                  </Box>
                </Flex>
              );
            })}
          </Box>
          <Flex
            className="border-border flex-col items-stretch border-t-[0.5px] px-4 py-3 sm:flex-row sm:items-center"
            gap={3}
            justify="between"
          >
            <Text color="muted">
              {reviewTasks.length > 0
                ? `Showing ${reviewPageStart + 1}–${Math.min(
                    reviewPageStart + REVIEW_PAGE_SIZE,
                    reviewTasks.length,
                  )} of ${reviewTasks.length}`
                : "No active cards to review"}
            </Text>
            {reviewPageCount > 1 ? (
              <Flex className="justify-end" gap={2}>
                <Button
                  color="tertiary"
                  disabled={reviewPage === 0}
                  onClick={() => {
                    setReviewPage((current) => Math.max(0, current - 1));
                  }}
                  variant="outline"
                >
                  Previous
                </Button>
                <Button
                  color="tertiary"
                  disabled={reviewPage >= reviewPageCount - 1}
                  onClick={() => {
                    setReviewPage((current) =>
                      Math.min(reviewPageCount - 1, current + 1),
                    );
                  }}
                  variant="outline"
                >
                  Next
                </Button>
              </Flex>
            ) : null}
          </Flex>
        </Box>
      ) : null}
    </>
  );
};
