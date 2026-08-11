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
  PlusIcon,
  RequestsIcon,
  ThumbsDownIcon,
  ThumbsUpIcon,
} from "icons";
import { Avatar, Box, Button, Dialog, Input, Menu, Text, TextEditor } from "ui";
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
  useCreatePublicFeedback,
  usePublicFeedbackVote,
} from "./feedback-mutations";
import { NEW_FEEDBACK_QUERY_PARAM } from "./query-params";
import { getRequestPathBySlug } from "./utils";
import { requestStatusMeta } from "./status";
import { getPublicAvatarColor } from "./avatar-color";

const MAX_FEEDBACK_TITLE_LENGTH = 200;

const SimilarFeedbackRow = ({
  item,
  onOpen,
  portal,
}: {
  item: SimilarPublicFeedback;
  onOpen: () => void;
  portal: PublicPortal;
}) => {
  const itemStatus = item.status ?? "pending";
  const authorName = item.authorName || "Unknown contributor";
  const status = requestStatusMeta[itemStatus];
  const request: PublicRequest = {
    authorAvatar: item.authorAvatar,
    authorId: item.authorId ?? "",
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
      </button>
      <div className="flex shrink-0 items-center gap-3">
        <FeedbackVoteButton compact portal={portal} request={request} />
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
  viewer: PublicPortalViewer;
}) => {
  const router = useRouter();
  const [open, setOpen] = useState(initialOpen);
  const [title, setTitle] = useState("");
  const titleRef = useRef("");
  const [similarityInput, setSimilarityInput] = useState({
    description: "",
    title: "",
  });
  const [boardId, setBoardId] = useState(
    portal.boards.length === 1 ? portal.boards[0]?.id ?? "" : "",
  );
  const createFeedback = useCreatePublicFeedback({ portal, viewer });
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
    setOpen(false);
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

    close();
    setTitle("");
    titleRef.current = "";
    setSimilarityInput({ description: "", title: "" });
    descriptionEditor?.commands.setContent("");
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
        setOpen(nextOpen);
        if (!nextOpen) {
          clearInitialOpenIntent();
          cancelSimilarityCheck();
        }
      }}
      open={open}
    >
      <Button
        className="h-12 w-full justify-center text-[1rem]"
        color="invert"
        leftIcon={<PlusIcon className="h-4 text-current" />}
        onClick={() => {
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
                        <TeamColor className="shrink-0" color={board.color} />
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
          <Dialog.Close />
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
        </Dialog.Body>
        <Dialog.Footer className="justify-end gap-2">
          <Button color="tertiary" onClick={close}>
            Cancel
          </Button>
          <Button
            color="invert"
            disabled={
              !boardId || title.trim().length === 0 || createFeedback.isPending
            }
            onClick={submit}
          >
            {blockingMatch ? "View existing feedback" : "Submit feedback"}
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
            />
          ))}
        </SimilarItemsPanel>
      </Dialog.Content>
    </Dialog>
  );
};

export const FeedbackVoteButton = ({
  compact = false,
  portal,
  request,
  showDownvote = false,
}: {
  compact?: boolean;
  portal: PublicPortal;
  request: PublicRequest;
  showDownvote?: boolean;
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
          onClick={() => {
            mutation.mutate(-1);
          }}
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
        leftIcon={
          <ThumbsUpIcon className={compact ? "h-3.5" : "h-4"} strokeWidth={2} />
        }
        onClick={() => {
          mutation.mutate(1);
        }}
        size="sm"
        title={vote === 1 ? "Remove upvote" : "Upvote"}
        variant="naked"
      >
        {voteCount}
      </Button>
    </Box>
  );
};
