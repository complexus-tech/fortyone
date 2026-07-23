"use client";

import { useState } from "react";
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
import { Box, Button, Dialog, Input, Menu, Text, TextEditor } from "ui";
import { toast } from "sonner";
import { cn } from "lib";
import { TeamColor } from "@/components/ui/team-color";
import { createRichTextStarterKit } from "@/lib/tiptap/starter-kit";
import type { PublicPortal, PublicPortalViewer, PublicRequest } from "./types";
import {
  useCreatePublicFeedback,
  usePublicFeedbackVote,
} from "./feedback-mutations";

const MAX_FEEDBACK_TITLE_LENGTH = 200;

export const NewFeedbackButton = ({
  portal,
  viewer,
}: {
  portal: PublicPortal;
  viewer: PublicPortalViewer;
}) => {
  const [open, setOpen] = useState(false);
  const [title, setTitle] = useState("");
  const [boardId, setBoardId] = useState(
    portal.boards.length === 1 ? portal.boards[0]?.id ?? "" : "",
  );
  const createFeedback = useCreatePublicFeedback({ portal, viewer });
  const selectedBoard = portal.boards.find((board) => board.id === boardId);
  const descriptionEditor = useEditor({
    content: "",
    editable: true,
    editorProps: {
      attributes: {
        "aria-label": "Feedback description",
        class: "min-h-40 outline-none",
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
  });

  const submit = () => {
    const input = {
      boardId,
      description: descriptionEditor?.getText() ?? "",
      portalSlug: portal.slug,
      title,
    };

    setOpen(false);
    setTitle("");
    descriptionEditor?.commands.setContent("");
    createFeedback.mutate(input, {
      onError: () => {
        setTitle(input.title);
        descriptionEditor?.commands.setContent(input.description);
        setOpen(true);
      },
      onSuccess: () => {
        toast.success("Feedback submitted");
      },
    });
  };

  return (
    <Dialog onOpenChange={setOpen} open={open}>
      <Button
        className="h-12 w-full justify-center text-[1rem]"
        color="invert"
        leftIcon={<PlusIcon className="h-4 text-current" />}
        onClick={() => {
          setOpen(true);
        }}
        size="lg"
      >
        New Feedback
      </Button>
      <Dialog.Content className="max-w-4xl" hideClose>
        <Dialog.Header className="flex items-center justify-between px-6 pt-6">
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
        <Dialog.Body className="pt-2">
          <Box>
            <Input
              aria-label="Feedback title"
              autoFocus
              className="h-auto border-0 bg-transparent px-0 py-3 text-2xl leading-tight font-medium focus-visible:ring-0 dark:bg-transparent"
              maxLength={MAX_FEEDBACK_TITLE_LENGTH}
              onChange={(event) => {
                setTitle(event.target.value);
              }}
              placeholder="Feedback title"
              value={title}
            />
          </Box>
          <TextEditor
            aria-label="Feedback description"
            className="min-h-40"
            editor={descriptionEditor}
          />
        </Dialog.Body>
        <Dialog.Footer className="justify-end gap-2">
          <Button
            color="tertiary"
            onClick={() => {
              setOpen(false);
            }}
          >
            Cancel
          </Button>
          <Button
            color="invert"
            disabled={
              !boardId || title.trim().length === 0 || createFeedback.isPending
            }
            onClick={submit}
          >
            Submit feedback
          </Button>
        </Dialog.Footer>
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
