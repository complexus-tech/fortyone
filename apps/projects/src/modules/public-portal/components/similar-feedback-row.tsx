"use client";

import { Avatar, Text } from "ui";
import { cn } from "lib";
import { requestStatusMeta } from "../status";
import type {
  PublicPortal,
  PublicPortalParticipant,
  SimilarPublicFeedback,
} from "../types";
import { toSimilarFeedbackRequest } from "../utils/feedback-controls";
import { FeedbackVoteButton } from "./feedback-vote-button";

export const SimilarFeedbackRow = ({
  item,
  onOpen,
  participant,
  portal,
}: {
  item: SimilarPublicFeedback;
  onOpen: () => void;
  participant: PublicPortalParticipant;
  portal: PublicPortal;
}) => {
  const request = toSimilarFeedbackRequest(item);
  const status = requestStatusMeta[request.status];

  return (
    <div className="hover:bg-state-hover/40 focus-within:bg-state-hover/40 group flex min-h-14 items-center gap-3 px-6 py-2.5 transition-colors">
      <button
        className="min-w-0 flex-1 text-left outline-none"
        onClick={onOpen}
        type="button"
      >
        <Text className="truncate" fontWeight="medium">
          {item.title}
        </Text>
        <Text className="truncate text-xs" color="muted">
          {request.authorName}
        </Text>
      </button>
      <div className="flex shrink-0 items-center gap-3">
        <FeedbackVoteButton
          compact
          participant={participant}
          portal={portal}
          request={request}
        />
        <span
          className={cn(
            "inline-flex h-7 items-center justify-center gap-2 rounded-lg border px-2 text-sm font-medium sm:min-w-24 sm:px-2.5",
            status.badgeClassName,
          )}
        >
          <span className={cn("size-2 rounded-sm", status.dotClassName)} />
          <span className="hidden sm:inline">{status.label}</span>
        </span>
        <Avatar name={request.authorName} size="xs" src={item.authorAvatar} />
      </div>
    </div>
  );
};
