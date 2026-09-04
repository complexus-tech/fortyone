"use client";

import { useRef } from "react";
import {
  ArrowRight2Icon,
  CheckIcon,
  CopyIcon,
  ImageIcon,
  RequestsIcon,
} from "icons";
import { Box, Button, Dialog, Flex, Input, Menu, Text, TextEditor } from "ui";
import { FeedbackAttachmentPreviews } from "@/components/ui/feedback-attachment-previews";
import { TeamColor } from "@/components/ui/team-color";
import { SimilarItemsPanel } from "@/components/ui/similar-items-panel";
import {
  FEEDBACK_ATTACHMENT_ACCEPT,
  MAX_FEEDBACK_ATTACHMENTS,
} from "@/shared/feedback-widget/attachments";
import { FeedbackGuestVerification } from "../guest-verification";
import type { NewFeedbackComposerState } from "../hooks/use-new-feedback-composer";
import { isContactableParticipant, isGuestParticipant } from "../participant";
import type { PublicPortal } from "../types";
import {
  copyAnonymousFeedbackTrackingLink,
  getFeedbackSubmitLabel,
  MAX_FEEDBACK_TITLE_LENGTH,
} from "../utils/feedback-controls";
import { SimilarFeedbackRow } from "./similar-feedback-row";

const AnonymousFeedbackSubmission = ({
  onDone,
  trackingUrl,
}: {
  onDone: () => void;
  trackingUrl: string;
}) => (
  <Box className="px-6 py-6">
    <Text as="h2" className="text-xl" fontWeight="semibold">
      Feedback submitted anonymously
    </Text>
    <Text className="mt-2 max-w-xl leading-6" color="muted">
      No name or email address was attached. Keep this tracking link to follow
      the public feedback as it is reviewed and worked on.
    </Text>
    <Box
      className="border-border bg-surface-muted/50 mt-5 rounded-xl border p-4"
      role="status"
    >
      <Text fontWeight="medium">Save your tracking link</Text>
      <Text className="mt-1 text-sm" color="muted">
        Anonymous feedback cannot receive personal notifications. Copy this
        public link before closing the window to check its status later.
      </Text>
    </Box>
    <Flex className="mt-6 flex-col-reverse gap-2 sm:flex-row" justify="end">
      <Button
        color="tertiary"
        leftIcon={<CopyIcon className="h-4" />}
        onClick={() => {
          void copyAnonymousFeedbackTrackingLink(trackingUrl);
        }}
        variant="outline"
      >
        Copy tracking link
      </Button>
      <Button color="invert" onClick={onDone}>
        Done
      </Button>
    </Flex>
  </Box>
);

const FeedbackParticipationChoice = ({
  composer,
}: {
  composer: NewFeedbackComposerState;
}) => (
  <Box className="px-6 py-6">
    <Text as="h2" className="text-xl" fontWeight="semibold">
      How would you like to submit?
    </Text>
    <Text className="mt-2 max-w-xl leading-6" color="muted">
      Your draft is saved. Choose whether you want a private email connection
      for replies and status updates.
    </Text>
    <Box className="mt-6 grid gap-3 sm:grid-cols-2">
      <button
        className="border-border hover:bg-state-hover focus-visible:ring-ring rounded-xl border p-4 text-left transition focus-visible:ring-2 focus-visible:outline-none"
        onClick={() => {
          void composer.continueWithEmail();
        }}
        type="button"
      >
        <Text fontWeight="semibold">
          {isContactableParticipant(composer.lockedParticipant)
            ? `Continue as ${composer.lockedParticipant.name}`
            : "Continue with email"}
        </Text>
        <Text className="mt-1 text-sm leading-5" color="muted">
          {isContactableParticipant(composer.lockedParticipant)
            ? "Attach this identity, follow the feedback, and receive meaningful updates."
            : "Verify privately, follow this feedback, and receive meaningful updates without creating an account."}
        </Text>
      </button>
      <button
        className="border-border hover:bg-state-hover focus-visible:ring-ring rounded-xl border p-4 text-left transition focus-visible:ring-2 focus-visible:outline-none"
        disabled={composer.isSubmitting}
        onClick={composer.submitAnonymously}
        type="button"
      >
        <Text fontWeight="semibold">Submit anonymously</Text>
        <Text className="mt-1 text-sm leading-5" color="muted">
          Attach no name or email. You will not receive personal notifications
          and must keep the public tracking link.
        </Text>
      </button>
    </Box>
    <Flex className="mt-6" justify="start">
      <Button
        color="tertiary"
        disabled={composer.isSubmitting}
        onClick={() => {
          composer.setComposerStep("draft");
        }}
        variant="naked"
      >
        Back to draft
      </Button>
    </Flex>
  </Box>
);

const FeedbackDraft = ({
  composer,
  portal,
}: {
  composer: NewFeedbackComposerState;
  portal: PublicPortal;
}) => {
  const attachmentInputRef = useRef<HTMLInputElement>(null);

  return (
    <>
      <Dialog.Header className="flex items-center justify-between px-6 pt-5 pb-1">
        <Dialog.Title className="flex items-center gap-1 text-lg">
          <Menu>
            <Menu.Button>
              <Button
                className="dark:bg-surface-elevated/90 gap-1.5 text-[0.95rem] font-semibold"
                color="tertiary"
                disabled={portal.boards.length === 0}
                leftIcon={
                  composer.selectedBoard ? (
                    <TeamColor color={composer.selectedBoard.color} />
                  ) : (
                    <RequestsIcon className="h-4" />
                  )
                }
                size="sm"
              >
                {composer.selectedBoard?.name ?? "Select board"}
              </Button>
            </Menu.Button>
            <Menu.Items align="start" className="w-60">
              <Menu.Group>
                {portal.boards.map((board) => (
                  <Menu.Item
                    active={board.id === composer.boardId}
                    className="justify-between gap-3"
                    key={board.id}
                    onSelect={() => {
                      composer.setBoardId(board.id);
                    }}
                  >
                    <span className="flex min-w-0 items-center gap-1.5">
                      <TeamColor className="shrink-0" color={board.color} />
                      <span className="truncate">{board.name}</span>
                    </span>
                    {board.id === composer.boardId ? (
                      <CheckIcon className="h-[1.1rem] w-auto" />
                    ) : null}
                  </Menu.Item>
                ))}
              </Menu.Group>
            </Menu.Items>
          </Menu>
          <ArrowRight2Icon
            className="h-4.5 w-auto opacity-30"
            strokeWidth={3}
          />
          <Text color="muted">New feedback</Text>
        </Dialog.Title>
        {!composer.isSubmitting ? <Dialog.Close /> : null}
      </Dialog.Header>
      <Dialog.Body className="pt-3 pb-3">
        <Box>
          <Input
            aria-label="Feedback title"
            autoFocus
            className="h-auto border-0 bg-transparent px-0 pt-1 pb-1 text-2xl leading-tight font-medium focus-visible:ring-0 dark:bg-transparent"
            maxLength={MAX_FEEDBACK_TITLE_LENGTH}
            onChange={(event) => {
              composer.updateTitle(event.target.value);
            }}
            placeholder="Feedback title"
            value={composer.title}
          />
        </Box>
        <TextEditor
          aria-label="Feedback description"
          className="min-h-24"
          editor={composer.descriptionEditor}
        />
        <FeedbackAttachmentPreviews
          files={composer.attachments}
          layout="page"
          onRemove={composer.removeAttachment}
        />
        {!isContactableParticipant(composer.lockedParticipant) ? (
          <Box
            className="border-border/70 bg-surface-muted/40 mt-4 rounded-xl border px-4 py-3"
            role="note"
          >
            <Text fontWeight="medium">Choose how you participate</Text>
            <Text className="mt-1 text-sm leading-5" color="muted">
              {portal.participationMode === "anonymous_allowed"
                ? "Continue with a private verified email to receive updates, or submit with no identity and no personal notifications."
                : "A private email verification is required. It lets you receive replies and updates without creating an account."}
            </Text>
          </Box>
        ) : null}
        {isGuestParticipant(composer.lockedParticipant) ? (
          <Box
            className="border-border/70 bg-surface-muted/40 mt-4 rounded-xl border px-4 py-3"
            role="note"
          >
            <Text fontWeight="medium">
              Continuing as {composer.lockedParticipant.displayName}
            </Text>
            <Text className="mt-1 text-sm leading-5" color="muted">
              {composer.lockedParticipant.masked
                ? "Your verified identity stays private and your public name is masked."
                : "Your email stays private. You will follow this feedback and can receive meaningful updates."}
            </Text>
          </Box>
        ) : null}
      </Dialog.Body>
      <Dialog.Footer className="justify-between gap-3">
        <div>
          <input
            accept={FEEDBACK_ATTACHMENT_ACCEPT}
            aria-label="Select feedback attachments"
            className="sr-only"
            multiple
            onChange={(event) => {
              composer.addAttachments(Array.from(event.target.files ?? []));
              event.target.value = "";
            }}
            ref={attachmentInputRef}
            type="file"
          />
          <Button
            color="tertiary"
            disabled={
              composer.isSubmitting ||
              composer.attachments.length >= MAX_FEEDBACK_ATTACHMENTS
            }
            leftIcon={<ImageIcon className="h-4" />}
            onClick={() => attachmentInputRef.current?.click()}
            variant="outline"
          >
            Attach files
            {composer.attachments.length > 0
              ? ` (${composer.attachments.length}/${MAX_FEEDBACK_ATTACHMENTS})`
              : ""}
          </Button>
        </div>
        <Flex gap={2}>
          <Button
            color="tertiary"
            disabled={composer.isSubmitting}
            onClick={composer.close}
          >
            Cancel
          </Button>
          <Button
            color="invert"
            disabled={
              !composer.boardId ||
              composer.title.trim().length === 0 ||
              composer.isSubmitting
            }
            onClick={composer.submit}
          >
            {getFeedbackSubmitLabel({
              hasDuplicate: Boolean(composer.blockingMatch),
              isSubmitting: composer.isSubmitting,
              requiresIdentity:
                portal.participationMode === "anonymous_allowed" ||
                !isContactableParticipant(composer.lockedParticipant),
            })}
          </Button>
        </Flex>
      </Dialog.Footer>
      <SimilarItemsPanel heading="Similar submissions">
        {composer.similarFeedbackItems.map((item) => (
          <SimilarFeedbackRow
            item={item}
            key={item.id}
            onOpen={() => {
              composer.openExistingFeedback(item.slug, item.isDuplicate);
            }}
            participant={composer.lockedParticipant}
            portal={portal}
          />
        ))}
      </SimilarItemsPanel>
    </>
  );
};

export const NewFeedbackComposerContent = ({
  composer,
  portal,
}: {
  composer: NewFeedbackComposerState;
  portal: PublicPortal;
}) => {
  if (composer.anonymousSubmission) {
    return (
      <AnonymousFeedbackSubmission
        onDone={composer.close}
        trackingUrl={composer.anonymousSubmission.trackingUrl}
      />
    );
  }

  if (composer.composerStep === "participation") {
    return <FeedbackParticipationChoice composer={composer} />;
  }

  if (composer.composerStep === "verification") {
    return (
      <FeedbackGuestVerification
        onBack={() => {
          composer.setComposerStep(
            portal.participationMode === "anonymous_allowed"
              ? "participation"
              : "draft",
          );
        }}
        onVerified={(verifiedParticipant) => {
          composer.setLockedParticipant(verifiedParticipant);
          composer.submitAsContactableParticipant(verifiedParticipant);
        }}
        portal={portal}
        purpose="submit this feedback"
      />
    );
  }

  return <FeedbackDraft composer={composer} portal={portal} />;
};
