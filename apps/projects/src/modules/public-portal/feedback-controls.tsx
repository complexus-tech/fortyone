"use client";

import { useRef, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import Underline from "@tiptap/extension-underline";
import { useEditor } from "@tiptap/react";
import {
  ArrowRight2Icon,
  CheckIcon,
  CopyIcon,
  PlusIcon,
  RequestsIcon,
  ThumbsDownIcon,
  ThumbsUpIcon,
} from "icons";
import {
  Avatar,
  Box,
  Button,
  Dialog,
  Flex,
  Input,
  Menu,
  Text,
  TextEditor,
} from "ui";
import { toast } from "sonner";
import { cn } from "lib";
import { TeamColor } from "@/components/ui/team-color";
import { SimilarItemsPanel } from "@/components/ui/similar-items-panel";
import { useDebouncedCallback } from "@/hooks/debounce";
import { createRichTextStarterKit } from "@/lib/tiptap/starter-kit";
import type {
  PublicPortal,
  PublicPortalViewer,
  PublicRequest,
  SimilarPublicFeedback,
} from "./types";
import { useSimilarPublicFeedback } from "./client-query";
import {
  useCreateAnonymousPublicFeedback,
  useCreatePublicFeedback,
  usePublicFeedbackVote,
} from "./feedback-mutations";
import { NEW_FEEDBACK_QUERY_PARAM } from "./query-params";
import { getRequestLoginUrl, getRequestPathBySlug } from "./utils";
import { requestStatusMeta } from "./status";
import { getPublicAvatarColor } from "./avatar-color";
import { getAnonymousFeedbackTrackingUrl } from "./anonymous-tracking";

const MAX_FEEDBACK_TITLE_LENGTH = 200;

const copyTrackingLink = async (trackingUrl: string) => {
  try {
    await navigator.clipboard.writeText(trackingUrl);
    toast.success("Tracking link copied");
  } catch {
    toast.error("Unable to copy tracking link");
  }
};

const getSubmitLabel = ({
  hasDuplicate,
  isSubmitting,
}: {
  hasDuplicate: boolean;
  isSubmitting: boolean;
}) => {
  if (hasDuplicate) return "View existing feedback";
  if (isSubmitting) return "Submitting...";
  return "Submit feedback";
};

const SimilarFeedbackRow = ({
  item,
  onOpen,
  portal,
  viewer,
}: {
  item: SimilarPublicFeedback;
  onOpen: () => void;
  portal: PublicPortal;
  viewer?: PublicPortalViewer | null;
}) => {
  const itemStatus = item.status ?? "pending";
  const authorName = item.authorName || "Unknown contributor";
  const status = requestStatusMeta[itemStatus];
  const request: PublicRequest = {
    authorAvatar: item.authorAvatar,
    authorId: item.authorId,
    authorName,
    boardId: "",
    commentCount: item.commentCount,
    comments: [],
    createdAtLabel: "",
    description: "",
    id: item.id,
    slug: item.slug,
    status: itemStatus,
    storyLinks: [],
    title: item.title,
    voteCount: item.voteCount,
  };

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
          {authorName}
        </Text>
      </button>
      <div className="flex shrink-0 items-center gap-3">
        <FeedbackVoteButton
          compact
          portal={portal}
          request={request}
          viewer={viewer}
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
        <Avatar
          name={authorName}
          size="xs"
          src={item.authorAvatar}
          style={{ backgroundColor: getPublicAvatarColor(authorName) }}
        />
      </div>
    </div>
  );
};

export const NewFeedbackButton = ({
  initialOpen = false,
  portal,
  viewer,
}: {
  initialOpen?: boolean;
  portal: PublicPortal;
  viewer?: PublicPortalViewer | null;
}) => {
  const router = useRouter();
  const [open, setOpen] = useState(initialOpen);
  const [title, setTitle] = useState("");
  const [anonymousSubmission, setAnonymousSubmission] = useState<{
    trackingUrl: string;
  } | null>(null);
  const titleRef = useRef("");
  const [similarityInput, setSimilarityInput] = useState({
    description: "",
    title: "",
  });
  const [boardId, setBoardId] = useState(
    portal.boards.length === 1 ? portal.boards[0]?.id ?? "" : "",
  );
  const createFeedback = useCreatePublicFeedback({ portal, viewer });
  const createAnonymousFeedback = useCreateAnonymousPublicFeedback({ portal });
  const isSubmitting =
    createFeedback.isPending || createAnonymousFeedback.isPending;
  const selectedBoard = portal.boards.find((board) => board.id === boardId);
  const { callback: checkForSimilarFeedback, cancel: cancelSimilarityCheck } =
    useDebouncedCallback(setSimilarityInput, 300);
  const descriptionEditor = useEditor({
    content: "",
    editable: true,
    editorProps: {
      attributes: {
        "aria-label": "Feedback description",
        class: "min-h-24 outline-none",
      },
    },
    extensions: [
      createRichTextStarterKit(),
      Underline,
      Link.configure({ autolink: true }),
      Placeholder.configure({
        placeholder: "Describe the feedback, context, or expected outcome...",
      }),
    ],
    immediatelyRender: false,
    onUpdate: ({ editor }) => {
      checkForSimilarFeedback({
        description: editor.getText(),
        title: titleRef.current,
      });
    },
  });
  const similarFeedback = useSimilarPublicFeedback({
    description: similarityInput.description,
    portalSlug: portal.slug,
    title: open ? similarityInput.title : "",
  });
  const similarFeedbackItems =
    title.trim() === similarityInput.title.trim()
      ? similarFeedback.data ?? []
      : [];
  const blockingMatch = similarFeedbackItems.find((item) => item.isDuplicate);

  const clearInitialOpenIntent = () => {
    if (!initialOpen) return;

    const url = new URL(window.location.href);
    url.searchParams.delete(NEW_FEEDBACK_QUERY_PARAM);
    window.history.replaceState(window.history.state, "", url);
  };

  const close = () => {
    clearInitialOpenIntent();
    cancelSimilarityCheck();
    setAnonymousSubmission(null);
    setOpen(false);
  };

  const resetDraft = () => {
    setTitle("");
    titleRef.current = "";
    setSimilarityInput({ description: "", title: "" });
    descriptionEditor?.commands.setContent("");
  };

  const openExistingFeedback = (slug: string, isDuplicate = false) => {
    close();
    router.push(getRequestPathBySlug(portal, slug));
    if (isDuplicate) {
      toast.info("This feedback was already reported", {
        description: "Add your context as a comment on the existing feedback.",
      });
    }
  };

  const submit = () => {
    if (blockingMatch) {
      openExistingFeedback(blockingMatch.slug, true);
      return;
    }
    const input = {
      boardId,
      description: descriptionEditor?.getText() ?? "",
      portalSlug: portal.slug,
      title,
    };

    if (!viewer) {
      createAnonymousFeedback.mutate(input, {
        onSuccess: (result) => {
          const trackingUrl = getAnonymousFeedbackTrackingUrl(
            portal,
            result.request,
          );
          resetDraft();
          setAnonymousSubmission({ trackingUrl });
          toast.success("Feedback submitted anonymously");
        },
      });
      return;
    }

    close();
    resetDraft();
    createFeedback.mutate(input, {
      onError: async () => {
        setTitle(input.title);
        titleRef.current = input.title;
        descriptionEditor?.commands.setContent(input.description);
        setOpen(true);
        const refreshed = await similarFeedback.refetch();
        const duplicate = refreshed.data?.find((item) => item.isDuplicate);
        if (duplicate) openExistingFeedback(duplicate.slug, true);
      },
      onSuccess: () => {
        toast.success("Feedback submitted");
      },
    });
  };

  return (
    <Dialog
      onOpenChange={(nextOpen) => {
        if (!nextOpen && isSubmitting) return;

        setOpen(nextOpen);
        if (!nextOpen) {
          clearInitialOpenIntent();
          cancelSimilarityCheck();
          setAnonymousSubmission(null);
        }
      }}
      open={open}
    >
      <Button
        className="h-12 w-full justify-center text-[1rem]"
        color="invert"
        leftIcon={<PlusIcon className="h-4 text-current" />}
        onClick={() => {
          setAnonymousSubmission(null);
          setOpen(true);
          checkForSimilarFeedback({
            description: descriptionEditor?.getText() ?? "",
            title: titleRef.current,
          });
        }}
        size="lg"
      >
        New Feedback
      </Button>
      <Dialog.Content className="max-w-4xl overflow-visible" hideClose>
        {anonymousSubmission ? (
          <Box className="px-6 py-6">
            <Text as="h2" className="text-xl" fontWeight="semibold">
              Feedback submitted anonymously
            </Text>
            <Text className="mt-2 max-w-xl leading-6" color="muted">
              No name or email address was attached. Keep this tracking link to
              follow the public feedback as it is reviewed and worked on.
            </Text>
            <Box
              className="border-border bg-surface-muted/50 mt-5 rounded-xl border p-4"
              role="status"
            >
              <Text fontWeight="medium">Save your tracking link</Text>
              <Text className="mt-1 text-sm" color="muted">
                Anonymous feedback cannot receive personal notifications. Copy
                this public link before closing the window to check its status
                later.
              </Text>
            </Box>
            <Flex
              className="mt-6 flex-col-reverse gap-2 sm:flex-row"
              justify="end"
            >
              <Button
                color="tertiary"
                leftIcon={<CopyIcon className="h-4" />}
                onClick={() => {
                  void copyTrackingLink(anonymousSubmission.trackingUrl);
                }}
                variant="outline"
              >
                Copy tracking link
              </Button>
              <Button color="invert" onClick={close}>
                Done
              </Button>
            </Flex>
          </Box>
        ) : (
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
                        selectedBoard ? (
                          <TeamColor color={selectedBoard.color} />
                        ) : (
                          <RequestsIcon className="h-4" />
                        )
                      }
                      size="sm"
                    >
                      {selectedBoard?.name ?? "Select board"}
                    </Button>
                  </Menu.Button>
                  <Menu.Items align="start" className="w-60">
                    <Menu.Group>
                      {portal.boards.map((board) => (
                        <Menu.Item
                          active={board.id === boardId}
                          className="justify-between gap-3"
                          key={board.id}
                          onSelect={() => {
                            setBoardId(board.id);
                          }}
                        >
                          <span className="flex min-w-0 items-center gap-1.5">
                            <TeamColor
                              className="shrink-0"
                              color={board.color}
                            />
                            <span className="truncate">{board.name}</span>
                          </span>
                          {board.id === boardId ? (
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
              {!isSubmitting ? <Dialog.Close /> : null}
            </Dialog.Header>
            <Dialog.Body className="pt-3 pb-3">
              <Box>
                <Input
                  aria-label="Feedback title"
                  autoFocus
                  className="h-auto border-0 bg-transparent px-0 pt-1 pb-1 text-2xl leading-tight font-medium focus-visible:ring-0 dark:bg-transparent"
                  maxLength={MAX_FEEDBACK_TITLE_LENGTH}
                  onChange={(event) => {
                    const nextTitle = event.target.value;
                    setTitle(nextTitle);
                    titleRef.current = nextTitle;
                    checkForSimilarFeedback({
                      description: descriptionEditor?.getText() ?? "",
                      title: nextTitle,
                    });
                  }}
                  placeholder="Feedback title"
                  value={title}
                />
              </Box>
              <TextEditor
                aria-label="Feedback description"
                className="min-h-24"
                editor={descriptionEditor}
              />
              {!viewer ? (
                <Box
                  className="border-border/70 bg-surface-muted/40 mt-4 rounded-xl border px-4 py-3"
                  role="note"
                >
                  <Text fontWeight="medium">Posting anonymously</Text>
                  <Text className="mt-1 text-sm leading-5" color="muted">
                    No name or email address will be attached. You won&apos;t
                    receive email updates, so keep the tracking link we provide
                    after submission.
                  </Text>
                </Box>
              ) : null}
            </Dialog.Body>
            <Dialog.Footer className="justify-end gap-2">
              <Button color="tertiary" disabled={isSubmitting} onClick={close}>
                Cancel
              </Button>
              <Button
                color="invert"
                disabled={!boardId || title.trim().length === 0 || isSubmitting}
                onClick={submit}
              >
                {getSubmitLabel({
                  hasDuplicate: Boolean(blockingMatch),
                  isSubmitting,
                })}
              </Button>
            </Dialog.Footer>
            <SimilarItemsPanel heading="Similar submissions">
              {similarFeedbackItems.map((item) => (
                <SimilarFeedbackRow
                  item={item}
                  key={item.id}
                  onOpen={() => {
                    openExistingFeedback(item.slug, item.isDuplicate);
                  }}
                  portal={portal}
                  viewer={viewer}
                />
              ))}
            </SimilarItemsPanel>
          </>
        )}
      </Dialog.Content>
    </Dialog>
  );
};

export const FeedbackVoteButton = ({
  compact = false,
  portal,
  request,
  showDownvote = false,
  viewer,
}: {
  compact?: boolean;
  portal: PublicPortal;
  request: PublicRequest;
  showDownvote?: boolean;
  viewer?: PublicPortalViewer | null;
}) => {
  const { mutation, vote, voteCount } = usePublicFeedbackVote({
    portalSlug: portal.slug,
    request,
  });

  return (
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
          href={viewer ? undefined : getRequestLoginUrl(portal, request)}
          onClick={
            viewer
              ? () => {
                  mutation.mutate(-1);
                }
              : undefined
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
        href={viewer ? undefined : getRequestLoginUrl(portal, request)}
        leftIcon={
          <ThumbsUpIcon className={compact ? "h-3.5" : "h-4"} strokeWidth={2} />
        }
        onClick={
          viewer
            ? () => {
                mutation.mutate(1);
              }
            : undefined
        }
        size="sm"
        title={vote === 1 ? "Remove upvote" : "Upvote"}
        variant="naked"
      >
        {voteCount}
      </Button>
    </Box>
  );
};
