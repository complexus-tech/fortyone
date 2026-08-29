"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { BellIcon, CheckIcon } from "icons";
import { Box, Button, Text } from "ui";
import { toast } from "sonner";
import {
  getFeedbackFollowStateAction,
  updateFeedbackFollowAction,
} from "./actions";
import { FeedbackGuestVerificationDialog } from "./guest-verification";
import { canVerifyAsGuest, isContactableParticipant } from "./participant";
import { publicPortalKeys } from "./query-keys";
import type {
  PublicPortal,
  PublicPortalParticipant,
  PublicRequest,
} from "./types";
import { getRequestLoginUrl } from "./utils";

const getActionData = <T,>(response: {
  data?: T | null;
  error?: { message: string };
}) => {
  if (response.error?.message) throw new Error(response.error.message);
  if (!response.data) throw new Error("Unable to update this subscription");
  return response.data;
};

export const FeedbackFollowControl = ({
  participant,
  portal,
  request,
}: {
  participant: PublicPortalParticipant;
  portal: PublicPortal;
  request: PublicRequest;
}) => {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [verifiedParticipant, setVerifiedParticipant] = useState<
    PublicPortalParticipant | undefined
  >();
  const activeParticipant = verifiedParticipant ?? participant;
  const [verificationOpen, setVerificationOpen] = useState(false);
  const followQueryKey = publicPortalKeys.feedbackFollow(
    portal.slug,
    request.id,
    activeParticipant.kind,
  );
  const followState = useQuery({
    queryKey: followQueryKey,
    queryFn: async () =>
      getActionData(
        await getFeedbackFollowStateAction({
          itemId: request.id,
          participantKind: activeParticipant.kind,
          portalSlug: portal.slug,
        }),
      ),
    enabled: isContactableParticipant(activeParticipant),
    initialData:
      request.following === undefined
        ? undefined
        : { following: request.following },
    staleTime: 60_000,
  });
  const following = followState.data?.following ?? false;

  const mutation = useMutation({
    mutationFn: async ({
      following: nextFollowing,
      participant: participantOverride,
    }: {
      following: boolean;
      participant: Exclude<PublicPortalParticipant, { kind: "anonymous" }>;
    }) =>
      getActionData(
        await updateFeedbackFollowAction({
          following: nextFollowing,
          itemId: request.id,
          itemSlug: request.slug,
          participantKind: participantOverride.kind,
          portalSlug: portal.slug,
        }),
      ),
    onError: (error) => {
      toast.error("Updates", { description: error.message });
    },
    onSuccess: (result) => {
      queryClient.setQueryData(followQueryKey, result);
      let description =
        activeParticipant.kind === "account"
          ? "You will no longer receive updates for this feedback."
          : "You will no longer receive email updates for this feedback.";
      if (result.following) {
        description =
          activeParticipant.kind === "account"
            ? "We will notify you about meaningful progress and replies."
            : "We will email you about meaningful progress and replies.";
      }
      toast.success(result.following ? "Updates enabled" : "Updates disabled", {
        description,
      });
    },
  });

  const followAs = (
    contactableParticipant: Exclude<
      PublicPortalParticipant,
      { kind: "anonymous" }
    >,
  ) => {
    mutation.mutate({
      following: true,
      participant: contactableParticipant,
    });
  };

  if (!isContactableParticipant(activeParticipant)) {
    const guestEnabled = canVerifyAsGuest(portal.participationMode);
    return (
      <Box className="border-border bg-surface rounded-xl border-[0.5px] p-5">
        <Text fontWeight="semibold">Get progress updates</Text>
        <Text className="mt-1 text-sm leading-5" color="muted">
          {guestEnabled
            ? "Verify your email to follow this feedback. This does not create an account."
            : "Log in to follow this feedback and receive progress updates."}
        </Text>
        <Button
          className="mt-4 w-full justify-center"
          color="tertiary"
          href={guestEnabled ? undefined : getRequestLoginUrl(portal, request)}
          leftIcon={<BellIcon className="h-4" />}
          onClick={
            guestEnabled
              ? () => {
                  setVerificationOpen(true);
                }
              : undefined
          }
          variant="outline"
        >
          {guestEnabled ? "Continue with email" : "Login to follow"}
        </Button>
        {guestEnabled ? (
          <FeedbackGuestVerificationDialog
            onOpenChange={setVerificationOpen}
            onVerified={(nextParticipant) => {
              setVerifiedParticipant(nextParticipant);
              followAs(nextParticipant);
              router.refresh();
            }}
            open={verificationOpen}
            portal={portal}
            purpose="follow this feedback"
          />
        ) : null}
      </Box>
    );
  }

  return (
    <Box className="border-border bg-surface rounded-xl border-[0.5px] p-5">
      <Text fontWeight="semibold">Progress updates</Text>
      <Text className="mt-1 text-sm leading-5" color="muted">
        {following
          ? "You will receive meaningful status changes, replies, and published updates."
          : "Follow this feedback to receive meaningful progress and reply emails."}
      </Text>
      <Button
        className="mt-4 w-full justify-center"
        color="tertiary"
        disabled={followState.isPending || mutation.isPending}
        leftIcon={
          following ? (
            <CheckIcon className="h-4" />
          ) : (
            <BellIcon className="h-4" />
          )
        }
        onClick={() => {
          mutation.mutate({
            following: !following,
            participant: activeParticipant,
          });
        }}
        variant="outline"
      >
        {following ? "Following" : "Follow feedback"}
      </Button>
    </Box>
  );
};
