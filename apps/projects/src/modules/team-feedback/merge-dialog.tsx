"use client";

import { useState } from "react";
import { SearchIcon } from "icons";
import { Box, Button, Dialog, Flex, Input, Text } from "ui";
import type { TeamFeedbackItem, TeamFeedbackMergeCandidate } from "./types";
import { useMergeTeamFeedback } from "./hooks/use-merge";
import { useTeamFeedbackMergeCandidates } from "./hooks/use-team-feedback";
import { FeedbackStatus } from "./status";

export const MergeTeamFeedbackDialog = ({
  onClose,
  onMerged,
  open,
  source,
}: {
  onClose: () => void;
  onMerged: (target: TeamFeedbackItem) => void;
  open: boolean;
  source: TeamFeedbackItem;
}) => {
  const [search, setSearch] = useState("");
  const [target, setTarget] = useState<TeamFeedbackMergeCandidate | null>(null);
  const candidates = useTeamFeedbackMergeCandidates(source.id, search, open);
  const merge = useMergeTeamFeedback();

  const close = () => {
    if (merge.isPending) return;
    setSearch("");
    setTarget(null);
    onClose();
  };

  return (
    <Dialog
      onOpenChange={(nextOpen) => {
        if (!nextOpen) close();
      }}
      open={open}
    >
      <Dialog.Content className="max-w-xl">
        <Dialog.Header>
          <Dialog.Title className="px-6 pt-1 text-lg">
            Merge feedback
          </Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="space-y-4">
          <Text className="leading-6" color="muted">
            Choose the canonical request. Followers and linked Updates move to
            it, and people opening this request are sent to the target.
          </Text>
          <Input
            leftIcon={<SearchIcon className="h-4" />}
            onChange={(event) => {
              setSearch(event.target.value);
              setTarget(null);
            }}
            placeholder="Search feedback to merge into"
            value={search}
          />
          <Box className="border-border max-h-72 overflow-y-auto rounded-xl border">
            {candidates.isLoading ? (
              <Text className="px-4 py-8 text-center" color="muted">
                Loading feedback…
              </Text>
            ) : null}
            {!candidates.isLoading &&
            (candidates.data?.candidates.length ?? 0) === 0 ? (
              <Text className="px-4 py-8 text-center" color="muted">
                No matching feedback
              </Text>
            ) : null}
            {candidates.data?.candidates.map((candidate) => (
              <button
                className={`border-border/60 hover:bg-state-hover flex w-full items-start gap-3 border-b px-4 py-3 text-left last:border-b-0 ${
                  target?.id === candidate.id ? "bg-state-hover" : ""
                }`}
                key={candidate.id}
                onClick={() => {
                  setTarget(candidate);
                }}
                type="button"
              >
                <Box className="min-w-0 flex-1">
                  <Text className="truncate" fontWeight="medium">
                    {candidate.title}
                  </Text>
                  <Flex align="center" className="mt-1 text-sm" gap={3}>
                    <FeedbackStatus status={candidate.status} />
                    <Text color="muted">
                      {candidate.voteCount} votes · {candidate.commentCount}{" "}
                      comments
                    </Text>
                  </Flex>
                </Box>
                {target?.id === candidate.id ? (
                  <Text className="text-sm" color="primary">
                    Selected
                  </Text>
                ) : null}
              </button>
            ))}
          </Box>
          {target ? (
            <Box className="border-warning/30 bg-warning/5 rounded-xl border p-3">
              <Text className="text-sm" fontWeight="medium">
                {source.title} will merge into {target.title}.
              </Text>
              <Text className="mt-1 text-sm" color="muted">
                This cannot be changed to a different target later.
              </Text>
            </Box>
          ) : null}
        </Dialog.Body>
        <Dialog.Footer>
          <Flex className="w-full" gap={3} justify="end">
            <Button color="tertiary" disabled={merge.isPending} onClick={close}>
              Cancel
            </Button>
            <Button
              color="primary"
              disabled={!target || merge.isPending}
              loading={merge.isPending}
              onClick={() => {
                if (!target) return;
                merge.mutate(
                  { sourceItemId: source.id, targetItemId: target.id },
                  {
                    onSuccess: (result) => {
                      onMerged(result.target);
                    },
                  },
                );
              }}
            >
              Merge feedback
            </Button>
          </Flex>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
