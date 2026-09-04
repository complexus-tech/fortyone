"use client";

import type { Editor } from "@tiptap/core";
import { useEditor } from "@tiptap/react";
import { useCallback, useRef, useState } from "react";
import { ArrowLeft2Icon, CheckIcon, ImageIcon } from "icons";
import { Box, Button, Flex, Menu, Text, TextEditor } from "ui";
import { FeedbackAttachmentPreviews } from "@/components/ui/feedback-attachment-previews";
import { TeamColor } from "@/components/ui/team-color";
import { createRichTextExtensions } from "@/lib/tiptap/rich-text-extensions";
import { getPersistableRichTextContent } from "@/lib/tiptap/rich-text-media";
import { RichTextTableMenu } from "@/lib/tiptap/rich-text-table-menu";
import {
  addUniqueFeedbackAttachments,
  FEEDBACK_ATTACHMENT_ACCEPT,
  MAX_FEEDBACK_ATTACHMENTS,
} from "@/shared/feedback-widget/attachments";
import type {
  PublicPortal,
  PublicRequest,
} from "@/shared/feedback-widget/types";
import {
  createWidgetFeedbackAction,
  type CreateWidgetFeedbackInput,
  type CreateWidgetFeedbackResult,
} from "../actions";
import type { WidgetSubmissionIdentity } from "./types";
import { WidgetBackButton, WidgetIconButton } from "./widget-ui";

const boardPreferenceKey = (portalSlug: string) =>
  `fortyone-feedback-widget:${portalSlug}:board`;

const getInitialBoardId = (portal: PublicPortal) => {
  const fallbackBoardId = portal.boards[0]?.id ?? "";
  if (portal.boards.length < 2 || typeof window === "undefined") {
    return fallbackBoardId;
  }
  try {
    const savedBoardId = window.localStorage.getItem(
      boardPreferenceKey(portal.slug),
    );
    return portal.boards.some((board) => board.id === savedBoardId)
      ? savedBoardId ?? fallbackBoardId
      : fallbackBoardId;
  } catch {
    return fallbackBoardId;
  }
};

export const FeedbackComposer = ({
  canUseIdentity,
  identity,
  isWriteLocked,
  onBack,
  onCreated,
  onRequireIdentity,
  portal,
}: {
  canUseIdentity: (identity: WidgetSubmissionIdentity | null) => boolean;
  identity: WidgetSubmissionIdentity | null;
  isWriteLocked: boolean;
  onBack: () => void;
  onCreated: (result: CreateWidgetFeedbackResult) => void;
  onRequireIdentity: () => void;
  portal: PublicPortal;
}) => {
  const [boardId, setBoardId] = useState(() => getInitialBoardId(portal));
  const [title, setTitle] = useState("");
  const [attachments, setAttachments] = useState<File[]>([]);
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showIdentityChoice, setShowIdentityChoice] = useState(false);
  const mediaInputRef = useRef<HTMLInputElement>(null);
  const selectedBoard = portal.boards.find((board) => board.id === boardId);

  const addAttachments = useCallback((_: Editor, files: File[]) => {
    setAttachments((current) => addUniqueFeedbackAttachments(current, files));
  }, []);
  const openMediaPicker = useCallback(() => {
    mediaInputRef.current?.click();
  }, []);
  const descriptionEditor = useEditor({
    content: "",
    editable: true,
    editorProps: {
      attributes: {
        "aria-label": "Feedback details",
        class: "min-h-48 outline-none",
      },
    },
    extensions: createRichTextExtensions({
      onMediaFiles: addAttachments,
      onMediaRequest: openMediaPicker,
      placeholder:
        "Tell us a little more about the problem, idea, or improvement. Type / for commands.",
    }),
    immediatelyRender: false,
  });

  const selectBoard = (nextBoardId: string) => {
    setBoardId(nextBoardId);
    try {
      window.localStorage.setItem(boardPreferenceKey(portal.slug), nextBoardId);
    } catch {
      // The in-memory selection still works when storage is unavailable.
    }
  };

  const submit = async (activeIdentity: WidgetSubmissionIdentity | null) => {
    if (!boardId || !title.trim() || isSubmitting) return;
    if (!canUseIdentity(activeIdentity)) {
      setError(
        "Your feedback draft is safe. Wait for the new identity to finish before submitting.",
      );
      return;
    }
    setIsSubmitting(true);
    setError("");
    const participationIntent = activeIdentity?.kind ?? "anonymous";
    const description = descriptionEditor
      ? getPersistableRichTextContent(descriptionEditor)
      : { contentHtml: "", contentText: "" };
    const attachmentData = new FormData();
    attachments.forEach((file) => {
      attachmentData.append("files", file);
    });
    const feedbackInput: CreateWidgetFeedbackInput = {
      boardId,
      description: description.contentText.trim(),
      descriptionHTML: description.contentHtml,
      participationIntent,
      portalSlug: portal.slug,
      sessionToken:
        activeIdentity && activeIdentity.kind !== "account"
          ? activeIdentity.sessionToken
          : undefined,
      title: title.trim(),
    };
    const submission =
      attachments.length > 0
        ? createWidgetFeedbackAction(feedbackInput, attachmentData)
        : createWidgetFeedbackAction(feedbackInput);
    const response = await submission
      .catch(() => null)
      .finally(() => {
        setIsSubmitting(false);
      });
    if (!response) {
      setError("Unable to submit feedback");
      return;
    }
    if (!canUseIdentity(activeIdentity)) return;
    if (response.error?.message || !response.data) {
      setError(response.error?.message ?? "Unable to submit feedback");
      return;
    }
    if (response.data.participantKind !== participationIntent) {
      setError("The submission privacy setting could not be confirmed.");
      return;
    }
    onCreated(response.data);
  };

  if (showIdentityChoice && !identity) {
    return (
      <Box className="bg-background absolute inset-0 z-30 flex min-h-0 flex-col">
        <Flex align="center" className="h-16 shrink-0 px-4" gap={2}>
          <WidgetBackButton
            aria-label="Back to feedback draft"
            onClick={() => {
              setShowIdentityChoice(false);
            }}
          >
            <ArrowLeft2Icon className="h-5" />
          </WidgetBackButton>
          <Text className="text-[16px]" fontWeight="semibold">
            Choose how to submit
          </Text>
        </Flex>
        <Flex
          className="min-h-0 flex-1 px-6 py-8"
          direction="column"
          justify="between"
        >
          <Box>
            <Text className="text-[20px] leading-7" fontWeight="semibold">
              Get progress updates or stay anonymous
            </Text>
            <Text className="mt-2 text-[13px] leading-6" color="muted">
              Verify your email without creating an account to receive replies
              and meaningful status updates. Your feedback draft will stay here
              while you verify.
            </Text>
            <Button
              className="mt-6 w-full justify-center rounded-lg"
              color="invert"
              onClick={onRequireIdentity}
            >
              Continue with email
            </Button>
            {portal.participationMode === "anonymous_allowed" ? (
              <Button
                className="mt-3 w-full justify-center rounded-lg"
                color="tertiary"
                disabled={isSubmitting || isWriteLocked}
                onClick={() => {
                  void submit(null);
                }}
              >
                {isSubmitting ? "Posting…" : "Submit anonymously"}
              </Button>
            ) : null}
          </Box>
          <Text className="text-[11px] leading-5" color="muted">
            Anonymous submissions cannot receive personal replies or status
            emails. Administrators will not receive your name or email.
          </Text>
        </Flex>
      </Box>
    );
  }

  return (
    <Box className="bg-background absolute inset-0 z-30 flex min-h-0 flex-col">
      <Flex align="center" className="h-16 shrink-0 px-4" gap={2}>
        <WidgetBackButton aria-label="Back to feedback" onClick={onBack}>
          <ArrowLeft2Icon className="h-5" />
        </WidgetBackButton>
        <Text className="text-[16px]" fontWeight="semibold">
          Share feedback
        </Text>
        {portal.boards.length > 1 ? (
          <div className="ml-auto">
            <Menu>
              <Menu.Button>
                <Button
                  className="max-w-40 gap-1.5 rounded-lg px-2.5 text-[12px]"
                  color="tertiary"
                  leftIcon={
                    selectedBoard ? (
                      <TeamColor color={selectedBoard.color} />
                    ) : undefined
                  }
                  size="xs"
                  variant="outline"
                >
                  <span className="truncate">
                    {selectedBoard?.name ?? "Select board"}
                  </span>
                </Button>
              </Menu.Button>
              <Menu.Items align="end" className="w-56">
                <Menu.Group>
                  {portal.boards.map((board) => (
                    <Menu.Item
                      active={board.id === boardId}
                      className="justify-between gap-3"
                      key={board.id}
                      onSelect={() => {
                        selectBoard(board.id);
                      }}
                    >
                      <span className="flex min-w-0 items-center gap-1.5">
                        <TeamColor className="shrink-0" color={board.color} />
                        <span className="truncate">{board.name}</span>
                      </span>
                      {board.id === boardId ? (
                        <CheckIcon className="h-4" />
                      ) : null}
                    </Menu.Item>
                  ))}
                </Menu.Group>
              </Menu.Items>
            </Menu>
          </div>
        ) : null}
      </Flex>
      <Box className="min-h-0 flex-1 overflow-y-auto px-5 py-3">
        <input
          aria-label="Feedback title"
          className="text-foreground placeholder:text-text-muted/55 w-full border-0 bg-transparent p-0 text-[20px] leading-7 font-semibold outline-none"
          maxLength={200}
          onChange={(event) => {
            setTitle(event.target.value);
          }}
          placeholder="What could be better?"
          value={title}
        />
        <TextEditor
          className="rich-document-editor text-foreground mt-4 min-h-48 text-[14px] leading-6"
          editor={descriptionEditor}
        />
        <RichTextTableMenu editor={descriptionEditor} scrollTarget={null} />
        <FeedbackAttachmentPreviews
          files={attachments}
          layout="widget"
          onRemove={(file) => {
            setAttachments((current) =>
              current.filter((candidate) => candidate !== file),
            );
          }}
        />
        {error ? (
          <Text className="mt-4 text-[12px] text-red-600 dark:text-red-400">
            {error}
          </Text>
        ) : null}
      </Box>
      <Flex align="center" className="shrink-0 px-5 py-4" justify="between">
        <input
          accept={FEEDBACK_ATTACHMENT_ACCEPT}
          aria-label="Attach files"
          className="sr-only"
          multiple
          onChange={(event) => {
            const files = Array.from(event.target.files ?? []);
            event.target.value = "";
            if (descriptionEditor && files.length > 0) {
              addAttachments(descriptionEditor, files);
            }
          }}
          ref={mediaInputRef}
          type="file"
        />
        <WidgetIconButton
          aria-label="Attach files"
          disabled={attachments.length >= MAX_FEEDBACK_ATTACHMENTS}
          onClick={openMediaPicker}
        >
          <ImageIcon className="h-5" />
        </WidgetIconButton>
        <button
          className="bg-foreground text-background focus-visible:ring-ring inline-flex h-10 items-center justify-center rounded-lg px-5 text-[13px] font-semibold transition-opacity focus-visible:ring-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-40"
          disabled={!boardId || !title.trim() || isSubmitting || isWriteLocked}
          onClick={() => {
            if (identity) {
              void submit(identity);
              return;
            }
            setShowIdentityChoice(true);
          }}
          type="button"
        >
          {isSubmitting ? "Submitting…" : "Submit feedback"}
        </button>
      </Flex>
    </Box>
  );
};

export const SubmissionSuccess = ({
  onView,
  request,
}: {
  onView: () => void;
  request: PublicRequest;
}) => (
  <Box className="bg-background absolute inset-0 z-40 flex flex-col">
    <Flex align="center" className="h-16 shrink-0 px-5">
      <Text className="text-[16px]" fontWeight="semibold">
        Feedback received
      </Text>
    </Flex>
    <Flex
      align="center"
      className="min-h-0 flex-1 px-10 text-center"
      direction="column"
      justify="center"
    >
      <Flex
        align="center"
        className="mb-5 size-12 rounded-full bg-emerald-500/12 text-emerald-600 dark:text-emerald-400"
        justify="center"
      >
        <CheckIcon className="h-5" />
      </Flex>
      <Text className="text-[19px] leading-7" fontWeight="semibold">
        Thanks for helping shape what comes next
      </Text>
      <Text className="mt-2 max-w-72 text-[12px] leading-5" color="muted">
        “{request.title}” is now in the feedback list.
      </Text>
      <Button
        className="mt-7 rounded-lg"
        color="invert"
        onClick={onView}
        size="sm"
      >
        View feedback
      </Button>
    </Flex>
  </Box>
);
