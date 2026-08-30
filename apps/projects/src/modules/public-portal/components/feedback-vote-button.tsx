"use client";

import { useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { ThumbsDownIcon, ThumbsUpIcon } from "icons";
import { Box, Button } from "ui";
import { toast } from "sonner";
import { cn } from "lib";
import { updateFeedbackFollowAction } from "../actions";
import { usePublicFeedbackVote } from "../feedback-mutations";
import { FeedbackGuestVerificationDialog } from "../guest-verification";
import { canVerifyAsGuest, isContactableParticipant } from "../participant";
import type {
  PublicPortal,
  PublicPortalParticipant,
  PublicRequest,
} from "../types";
import { getRequestLoginUrl } from "../utils";

export const FeedbackVoteButton = ({
  compact = false,
  participant,
  portal,
  request,
  showDownvote = false,
}: {
  compact?: boolean;
  participant: PublicPortalParticipant;
  portal: PublicPortal;
  request: PublicRequest;
  showDownvote?: boolean;
}) => {
  const router = useRouter();
  const [participantOverride, setParticipantOverride] =
    useState<PublicPortalParticipant | null>(null);
  const activeParticipant = participantOverride ?? participant;
  const [verificationOpen, setVerificationOpen] = useState(false);
  const pendingDirectionRef = useRef<-1 | 1>(1);
  const { mutation, vote, voteCount } = usePublicFeedbackVote({
    participant: activeParticipant,
    portalSlug: portal.slug,
    request,
  });
  const requiresAccount = portal.participationMode === "account_required";

  const offerUpdateNotifications = (
    contactableParticipant: Exclude<
      PublicPortalParticipant,
      { kind: "anonymous" }
    >,
  ) => {
    if (request.following) return;

    toast.info("Want progress updates?", {
      description: "Following is optional and separate from your vote.",
      action: {
        label: "Notify me",
        onClick: () => {
          void updateFeedbackFollowAction({
            following: true,
            itemId: request.id,
            itemSlug: request.slug,
            participantKind: contactableParticipant.kind,
            portalSlug: portal.slug,
          }).then((response) => {
            if (response.error?.message) {
              toast.error("Unable to enable updates", {
                description: response.error.message,
              });
              return;
            }
            toast.success("Updates enabled");
          });
        },
      },
    });
  };

  const voteOrVerify = (direction: -1 | 1) => {
    if (isContactableParticipant(activeParticipant)) {
      mutation.mutate(
        { direction },
        {
          onSuccess: () => {
            offerUpdateNotifications(activeParticipant);
          },
        },
      );
      return;
    }
    pendingDirectionRef.current = direction;
    setVerificationOpen(true);
  };

  return (
    <>
      <Box className="flex shrink-0 items-center gap-0.5">
        {showDownvote ? (
          <Button
            aria-label={vote === -1 ? "Remove downvote" : "Downvote"}
            asIcon
            className={cn("text-text-muted hover:text-foreground h-9", {
              "text-foreground": vote === -1,
            })}
            color="tertiary"
            disabled={mutation.isPending}
            href={
              requiresAccount && !isContactableParticipant(activeParticipant)
                ? getRequestLoginUrl(portal, request)
                : undefined
            }
            onClick={
              requiresAccount && !isContactableParticipant(activeParticipant)
                ? undefined
                : () => {
                    voteOrVerify(-1);
                  }
            }
            size="sm"
            title={vote === -1 ? "Remove downvote" : "Downvote"}
            variant="naked"
          >
            <ThumbsDownIcon className="h-4" strokeWidth={2} />
          </Button>
        ) : null}
        <Button
          aria-label={vote === 1 ? "Remove upvote" : "Upvote"}
          className={cn(
            "text-text-muted hover:text-foreground",
            compact ? "h-7 gap-1 px-1.5" : "h-9 gap-1.5 px-2.5",
            { "text-foreground": vote === 1 },
          )}
          color="tertiary"
          disabled={mutation.isPending}
          href={
            requiresAccount && !isContactableParticipant(activeParticipant)
              ? getRequestLoginUrl(portal, request)
              : undefined
          }
          leftIcon={
            <ThumbsUpIcon
              className={compact ? "h-3.5" : "h-4"}
              strokeWidth={2}
            />
          }
          onClick={
            requiresAccount && !isContactableParticipant(activeParticipant)
              ? undefined
              : () => {
                  voteOrVerify(1);
                }
          }
          size="sm"
          title={vote === 1 ? "Remove upvote" : "Upvote"}
          variant="naked"
        >
          {voteCount}
        </Button>
      </Box>
      {canVerifyAsGuest(portal.participationMode) ? (
        <FeedbackGuestVerificationDialog
          onOpenChange={setVerificationOpen}
          onVerified={(verifiedParticipant) => {
            setParticipantOverride(verifiedParticipant);
            mutation.mutate(
              {
                direction: pendingDirectionRef.current,
                participant: verifiedParticipant,
              },
              {
                onSuccess: () => {
                  offerUpdateNotifications(verifiedParticipant);
                },
              },
            );
            router.refresh();
          }}
          open={verificationOpen}
          portal={portal}
          purpose="vote"
        />
      ) : null}
    </>
  );
};
