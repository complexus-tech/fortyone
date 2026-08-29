"use client";

import { Avatar, Box, Button, Dialog, Flex, Select, Skeleton, Text } from "ui";
import type {
  FeedbackBoard,
  FeedbackReviewer,
  FeedbackReviewerEmailFrequency,
} from "./types";
import {
  useFeedbackBoardReviewers,
  useUpdateFeedbackBoardReviewerMutation,
} from "./hooks";

const frequencies: {
  label: string;
  value: FeedbackReviewerEmailFrequency;
}[] = [
  { label: "Off", value: "off" },
  { label: "Daily", value: "daily" },
  { label: "Weekly", value: "weekly" },
];

const formatRole = (role: FeedbackReviewer["role"]) =>
  role === "admin" ? "Admin" : "Member";

const ReviewerRow = ({
  boardId,
  reviewer,
}: {
  boardId: string;
  reviewer: FeedbackReviewer;
}) => {
  const updateReviewer = useUpdateFeedbackBoardReviewerMutation(boardId);

  const updateFrequency = (emailFrequency: FeedbackReviewerEmailFrequency) => {
    if (reviewer.emailFrequency === emailFrequency) return;
    updateReviewer.mutate({
      input: { emailFrequency },
      userId: reviewer.userId,
    });
  };

  return (
    <Flex
      align="center"
      className="border-border gap-4 border-b py-3.5 last:border-b-0"
      justify="between"
    >
      <Flex align="center" className="min-w-0" gap={3}>
        <Avatar
          className="shrink-0"
          name={reviewer.name}
          size="sm"
          src={reviewer.avatarUrl}
        />
        <Box className="min-w-0">
          <Flex align="center" className="min-w-0 gap-2">
            <Text className="truncate font-medium">{reviewer.name}</Text>
            <Text as="span" className="shrink-0 text-[0.9rem]" color="muted">
              {formatRole(reviewer.role)}
            </Text>
          </Flex>
          <Text className="truncate text-[0.95rem]" color="muted">
            {reviewer.email}
          </Text>
        </Box>
      </Flex>
      <Select
        disabled={updateReviewer.isPending}
        onValueChange={(value) => {
          updateFrequency(value as FeedbackReviewerEmailFrequency);
        }}
        value={reviewer.emailFrequency}
      >
        <Select.Trigger
          aria-label={`Email summary for ${reviewer.name}`}
          className="h-9 w-28 shrink-0 px-2"
        >
          <Select.Input />
        </Select.Trigger>
        <Select.Content align="end">
          {frequencies.map((frequency) => (
            <Select.Option key={frequency.value} value={frequency.value}>
              {frequency.label}
            </Select.Option>
          ))}
        </Select.Content>
      </Select>
    </Flex>
  );
};

const ReviewersSkeleton = () => (
  <Box aria-label="Loading reviewers" className="space-y-4 py-2">
    {Array.from({ length: 3 }, (_, index) => (
      <Flex align="center" gap={3} justify="between" key={index}>
        <Flex align="center" className="flex-1" gap={3}>
          <Skeleton className="size-8 shrink-0 rounded-full" />
          <Box className="flex-1 space-y-2">
            <Skeleton className="h-4 w-32" />
            <Skeleton className="h-3 w-48 max-w-full" />
          </Box>
        </Flex>
        <Skeleton className="h-9 w-28" />
      </Flex>
    ))}
  </Box>
);

export const FeedbackReviewersDialog = ({
  board,
  onOpenChange,
  open,
  teamName,
}: {
  board: FeedbackBoard;
  onOpenChange: (open: boolean) => void;
  open: boolean;
  teamName: string;
}) => {
  const reviewersQuery = useFeedbackBoardReviewers(board.id, open);
  const reviewers = reviewersQuery.data ?? [];
  const subscribedCount = reviewers.reduce(
    (count, reviewer) =>
      reviewer.emailFrequency === "off" ? count : count + 1,
    0,
  );

  return (
    <Dialog onOpenChange={onOpenChange} open={open}>
      <Dialog.Content className="max-w-xl">
        <Dialog.Header className="space-y-2 px-6 pt-5 pb-3">
          <Dialog.Title className="pr-10 text-lg">Reviewers</Dialog.Title>
          <Dialog.Description className="px-0">
            Choose who receives grouped email summaries for {board.name}.
          </Dialog.Description>
        </Dialog.Header>
        <Dialog.Body className="pt-0 pb-6">
          <Box className="bg-surface-muted/70 mb-5 rounded-xl px-4 py-3.5 dark:bg-white/[0.04]">
            <Text className="font-medium">Feedback stays immediate</Text>
            <Text className="mt-1 leading-relaxed" color="muted">
              New submissions appear in the {teamName} Feedback queue as they
              arrive. Reviewers can also receive one grouped summary on their
              chosen schedule.
            </Text>
          </Box>

          {reviewersQuery.isLoading ? <ReviewersSkeleton /> : null}
          {reviewersQuery.isError ? (
            <Flex
              align="center"
              className="border-border rounded-xl border px-4 py-6 text-center"
              direction="column"
              gap={3}
            >
              <Text className="font-medium">Reviewers could not be loaded</Text>
              <Text color="muted">Please try again.</Text>
              <Button
                color="tertiary"
                onClick={() => {
                  void reviewersQuery.refetch();
                }}
                size="sm"
              >
                Try again
              </Button>
            </Flex>
          ) : null}
          {!reviewersQuery.isLoading && !reviewersQuery.isError ? (
            <Box>
              <Flex align="center" className="mb-1" justify="between">
                <Text className="font-medium">Team members</Text>
                <Text className="text-[0.95rem]" color="muted">
                  {subscribedCount} subscribed
                </Text>
              </Flex>
              {reviewers.length === 0 ? (
                <Text className="py-7 text-center" color="muted">
                  No eligible team members found.
                </Text>
              ) : (
                reviewers.map((reviewer) => (
                  <ReviewerRow
                    boardId={board.id}
                    key={reviewer.userId}
                    reviewer={reviewer}
                  />
                ))
              )}
            </Box>
          ) : null}
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
};
