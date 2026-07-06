"use client";

import { useState, useTransition } from "react";
import { ArrowUpIcon, PlusIcon } from "icons";
import { Box, Button, Dialog, Flex, Input, Text, TextArea } from "ui";
import { toast } from "sonner";
import type { Team } from "@/modules/teams/types";
import type { PublicPortal, PublicRequest } from "./types";
import {
  createStoryFromFeedbackAction,
  createFeedbackAction,
  createFeedbackCommentAction,
  toggleFeedbackVoteAction,
} from "./actions";

export const NewFeedbackButton = ({ portal }: { portal: PublicPortal }) => {
  const [open, setOpen] = useState(false);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [boardId, setBoardId] = useState(portal.boards[0]?.id ?? "");
  const [isPending, startTransition] = useTransition();

  const submit = () => {
    startTransition(async () => {
      const response = await createFeedbackAction({
        boardId,
        description,
        portalId: portal.id,
        portalSlug: portal.slug,
        title,
        workspaceSlug: portal.workspace.slug,
      });
      if (response.error?.message) {
        toast.error("Feedback", { description: response.error.message });
        return;
      }
      setOpen(false);
      setTitle("");
      setDescription("");
      toast.success("Feedback submitted");
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
        rounded="full"
        size="lg"
      >
        New Feedback
      </Button>
      <Dialog.Content className="max-w-xl">
        <Dialog.Header>
          <Dialog.Title className="px-6 pt-1 text-lg">
            New Feedback
          </Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="space-y-4">
          <Box>
            <Text className="mb-2 text-sm" fontWeight="medium">
              Board
            </Text>
            <select
              className="border-border bg-surface h-11 w-full rounded-2xl border px-3"
              onChange={(event) => {
                setBoardId(event.target.value);
              }}
              value={boardId}
            >
              {portal.boards.map((board) => (
                <option key={board.id} value={board.id}>
                  {board.name}
                </option>
              ))}
            </select>
          </Box>
          <Input
            onChange={(event) => {
              setTitle(event.target.value);
            }}
            placeholder="Summarize the feedback"
            value={title}
          />
          <TextArea
            className="min-h-32"
            onChange={(event) => {
              setDescription(event.target.value);
            }}
            placeholder="Add details, context, or examples"
            value={description}
          />
        </Dialog.Body>
        <Dialog.Footer className="justify-end gap-3">
          <Button
            color="tertiary"
            onClick={() => {
              setOpen(false);
            }}
          >
            Cancel
          </Button>
          <Button
            color="primary"
            disabled={!boardId || title.trim().length === 0 || isPending}
            onClick={submit}
          >
            Submit
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};

export const FeedbackVoteButton = ({
  portal,
  request,
}: {
  portal: PublicPortal;
  request: PublicRequest;
}) => {
  const [isPending, startTransition] = useTransition();

  return (
    <Button
      className="h-8 rounded-xl px-3"
      color="tertiary"
      disabled={isPending}
      leftIcon={<ArrowUpIcon className="h-3.5" />}
      onClick={() => {
        startTransition(async () => {
          const response = await toggleFeedbackVoteAction({
            itemId: request.id,
            itemSlug: request.slug,
            portalSlug: portal.slug,
            workspaceSlug: portal.workspace.slug,
          });
          if (response.error?.message) {
            toast.error("Vote", { description: response.error.message });
          }
        });
      }}
      size="sm"
      variant="naked"
    >
      {request.voteCount}
    </Button>
  );
};

export const FeedbackCommentComposer = ({
  portal,
  request,
}: {
  portal: PublicPortal;
  request: PublicRequest;
}) => {
  const [body, setBody] = useState("");
  const [isPending, startTransition] = useTransition();

  return (
    <Box className="border-border bg-surface rounded-3xl border-[0.5px] p-4">
      <TextArea
        className="min-h-28 border-none p-0 shadow-none"
        onChange={(event) => {
          setBody(event.target.value);
        }}
        placeholder="Add a comment..."
        value={body}
      />
      <Flex align="end" className="mt-4" justify="end">
        <Button
          color="tertiary"
          disabled={body.trim().length === 0 || isPending}
          onClick={() => {
            startTransition(async () => {
              const response = await createFeedbackCommentAction({
                body,
                itemId: request.id,
                itemSlug: request.slug,
                portalSlug: portal.slug,
                workspaceSlug: portal.workspace.slug,
              });
              if (response.error?.message) {
                toast.error("Comment", { description: response.error.message });
                return;
              }
              setBody("");
            });
          }}
          size="sm"
        >
          Comment
        </Button>
      </Flex>
    </Box>
  );
};

export const CreateStoryFromFeedbackButton = ({
  portal,
  request,
  teams,
}: {
  portal: PublicPortal;
  request: PublicRequest;
  teams: Team[];
}) => {
  const [open, setOpen] = useState(false);
  const [teamId, setTeamId] = useState(teams[0]?.id ?? "");
  const [isPending, startTransition] = useTransition();

  if (teams.length === 0) return null;

  return (
    <Dialog onOpenChange={setOpen} open={open}>
      <Button
        className="w-full justify-center"
        color="tertiary"
        onClick={() => {
          setOpen(true);
        }}
        rounded="full"
      >
        Create internal story
      </Button>
      <Dialog.Content className="max-w-md">
        <Dialog.Header>
          <Dialog.Title className="px-6 pt-1 text-lg">
            Create Story
          </Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="space-y-4">
          <Text color="muted">
            This creates a FortyOne story and links it to this feedback item.
            Public feedback stays visible in the portal.
          </Text>
          <select
            className="border-border bg-surface h-11 w-full rounded-2xl border px-3"
            onChange={(event) => {
              setTeamId(event.target.value);
            }}
            value={teamId}
          >
            {teams.map((team) => (
              <option key={team.id} value={team.id}>
                {team.name}
              </option>
            ))}
          </select>
        </Dialog.Body>
        <Dialog.Footer className="justify-end gap-3">
          <Button
            color="tertiary"
            onClick={() => {
              setOpen(false);
            }}
          >
            Cancel
          </Button>
          <Button
            color="primary"
            disabled={!teamId || isPending}
            onClick={() => {
              startTransition(async () => {
                const response = await createStoryFromFeedbackAction({
                  itemId: request.id,
                  itemSlug: request.slug,
                  portalSlug: portal.slug,
                  teamId,
                  workspaceSlug: portal.workspace.slug,
                });
                if (response.error?.message) {
                  toast.error("Story", { description: response.error.message });
                  return;
                }
                setOpen(false);
                toast.success("Story created");
              });
            }}
          >
            Create
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};
