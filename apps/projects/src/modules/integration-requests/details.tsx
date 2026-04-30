"use client";

import { useEffect, useState } from "react";
import { useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Underline from "@tiptap/extension-underline";
import { TaskItem, TaskList } from "@tiptap/extension-list";
import Link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import TextExtension from "@tiptap/extension-text";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { useParams, useRouter } from "next/navigation";
import { CheckIcon, CloseIcon, GitHubIcon, LinkIcon, NewTabIcon } from "icons";
import {
  Avatar,
  Box,
  Button,
  Container,
  Divider,
  Flex,
  Skeleton,
  Text,
  TextEditor,
  TimeAgo,
} from "ui";
import {
  ConfirmDialog,
  PrioritiesMenu,
  PriorityIcon,
  StatusesMenu,
  StoryStatusIcon,
} from "@/components/ui";
import { useDebounce, useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import type { GitHubComment } from "@/modules/settings/workspace/integrations/github/types";
import type { StoryPriority } from "@/modules/stories/types";
import {
  getStoryCommentEditorExtensions,
  serializeStoryCommentToGitHubMarkdown,
} from "@/modules/story/components/story-comment-editor";
import { useAcceptIntegrationRequest } from "./hooks/use-accept-request";
import { useDeclineIntegrationRequest } from "./hooks/use-decline-request";
import { useIntegrationRequest } from "./hooks/use-request";
import { usePostRequestGitHubComment } from "./hooks/use-post-request-github-comment";
import { useRequestGitHubComments } from "./hooks/use-request-github-comments";
import { useUpdateIntegrationRequest } from "./hooks/use-update-request";
import type { UpdateIntegrationRequestInput } from "./types";

const DEBOUNCE_DELAY = 1000;

const metadataText = (value: unknown) =>
  typeof value === "string" && value.trim() ? value : null;

const CommentRow = ({ comment }: { comment: GitHubComment }) => (
  <Box className="relative pb-4">
    <Flex align="center" gap={1}>
      <Box className="bg-surface relative top-px flex aspect-square items-center rounded-full p-[0.3rem]">
        <Avatar
          className="relative top-0.5"
          name={comment.userLogin}
          size="xs"
          src={comment.userAvatar || undefined}
        />
      </Box>
      <Text className="ml-1 text-black dark:text-white">
        {comment.userLogin}
      </Text>
      <Text className="mx-0.5 text-[0.95rem]" color="muted">
        ·
      </Text>
      <Text className="text-[0.95rem]" color="muted">
        <TimeAgo timestamp={comment.createdAt} />
      </Text>
      <a
        className="text-muted hover:text-foreground rounded-md p-0.5 transition"
        href={comment.htmlUrl}
        rel="noopener noreferrer"
        target="_blank"
        title="View on GitHub"
      >
        <LinkIcon className="h-4" />
      </a>
    </Flex>
    <Box className="prose prose-stone dark:prose-invert prose-headings:font-semibold prose-a:text-primary prose-pre:bg-surface-muted prose-pre:text-foreground mt-0.5 ml-9 max-w-full leading-6">
      <Markdown remarkPlugins={[remarkGfm]}>{comment.body}</Markdown>
    </Box>
  </Box>
);

const GitHubCommentInput = ({ requestId }: { requestId: string }) => {
  const { data: session } = useSession();
  const { mutate: postComment, isPending } = usePostRequestGitHubComment();
  const editor = useEditor({
    content: "",
    editable: !isPending,
    extensions: getStoryCommentEditorExtensions({
      placeholder: "Leave a GitHub comment...",
    }),
    immediatelyRender: false,
  });

  const handleSubmit = () => {
    if (!editor || editor.isEmpty) return;
    const body = serializeStoryCommentToGitHubMarkdown(editor.getJSON());
    if (!body) return;

    postComment(
      { requestId, body },
      {
        onSuccess: () => {
          editor.commands.clearContent();
        },
      },
    );
  };

  return (
    <Flex align="start" className="mb-3">
      <Box className="bg-surface z-1 flex aspect-square items-center rounded-full p-[0.3rem]">
        <Avatar
          name={session?.user?.name ?? undefined}
          size="xs"
          src={session?.user?.image ?? undefined}
        />
      </Box>
      <Flex
        className="border-border/40 bg-surface-muted/40 ml-1 min-h-24 w-full rounded-2xl border px-4 pb-4"
        direction="column"
        gap={2}
        justify="between"
      >
        <TextEditor
          className="prose-base prose-a:text-foreground leading-6 antialiased"
          editor={editor}
        />
        <Flex gap={2} justify="end">
          <Button
            className="px-3"
            color="tertiary"
            disabled={isPending}
            onClick={handleSubmit}
            size="sm"
            variant="outline"
          >
            Comment
          </Button>
        </Flex>
      </Flex>
    </Flex>
  );
};

const GitHubComments = ({ requestId }: { requestId: string }) => {
  const { data: comments = [], isLoading } =
    useRequestGitHubComments(requestId);

  return (
    <Box>
      <Flex align="center" className="mb-4" gap={2}>
        <GitHubIcon className="text-primary h-4" />
        <Text className="font-medium">GitHub comments</Text>
      </Flex>
      <GitHubCommentInput requestId={requestId} />
      {isLoading ? (
        <Box className="space-y-4">
          {Array.from({ length: 2 }).map((_, index) => (
            <Box className="flex gap-3" key={index}>
              <Skeleton className="h-8 w-8 rounded-full" />
              <Box className="flex-1 space-y-2">
                <Skeleton className="h-4 w-32" />
                <Skeleton className="h-12 w-full" />
              </Box>
            </Box>
          ))}
        </Box>
      ) : comments.length > 0 ? (
        <Box>
          {comments.map((comment) => (
            <CommentRow comment={comment} key={comment.id} />
          ))}
        </Box>
      ) : null}
    </Box>
  );
};

export const IntegrationRequestDetails = ({
  requestId,
}: {
  requestId: string;
}) => {
  const { teamId } = useParams<{ teamId: string }>();
  const router = useRouter();
  const { withWorkspace } = useWorkspacePath();
  const [isDeclining, setIsDeclining] = useState(false);
  const { data: request, isPending } = useIntegrationRequest(requestId);
  const { data: statuses = [] } = useTeamStatuses(teamId);
  const updateRequest = useUpdateIntegrationRequest();
  const acceptRequest = useAcceptIntegrationRequest();
  const declineRequest = useDeclineIntegrationRequest();

  const defaultStatus =
    statuses.find((status) => status.category === "unstarted") ||
    statuses.at(0);
  const statusId = request?.statusId ?? defaultStatus?.id;
  const priority = request?.priority ?? "No Priority";

  const handleUpdate = (payload: UpdateIntegrationRequestInput) => {
    if (!request) return;
    updateRequest.mutate({ requestId: request.id, payload });
  };
  const debouncedHandleUpdate = useDebounce(handleUpdate, DEBOUNCE_DELAY);

  const descriptionEditor = useEditor({
    extensions: [
      StarterKit,
      Underline,
      TaskList,
      TaskItem.configure({ nested: true }),
      Link.configure({ autolink: true }),
      Placeholder.configure({ placeholder: "Add description..." }),
    ],
    content: request?.description ?? "",
    editable: request?.status === "pending",
    onUpdate: ({ editor }) => {
      debouncedHandleUpdate({
        description: editor.getHTML(),
      });
    },
    immediatelyRender: false,
  });

  const titleEditor = useEditor({
    extensions: [
      Document,
      Paragraph,
      TextExtension,
      Placeholder.configure({ placeholder: "Add title..." }),
    ],
    content: request?.title ?? "",
    editable: request?.status === "pending",
    onUpdate: ({ editor }) => {
      debouncedHandleUpdate({
        title: editor.getText(),
      });
    },
    immediatelyRender: false,
  });

  useEffect(() => {
    if (
      titleEditor &&
      request?.title &&
      titleEditor.getText() !== request.title
    ) {
      titleEditor.commands.setContent(request.title);
    }
    if (
      descriptionEditor &&
      request?.description &&
      descriptionEditor.getHTML() !== request.description
    ) {
      descriptionEditor.commands.setContent(request.description);
    }
  }, [descriptionEditor, request?.description, request?.title, titleEditor]);

  if (isPending) {
    return (
      <Box className="h-dvh px-8 py-7">
        <Box className="bg-surface-muted mb-8 h-8 w-2/5 rounded" />
        <Box className="bg-surface-muted mb-4 h-4 w-4/5 rounded" />
        <Box className="bg-surface-muted h-4 w-3/5 rounded" />
      </Box>
    );
  }

  if (!request) {
    return (
      <Box className="flex h-dvh items-center justify-center px-6">
        <Box>
          <Text align="center" className="mb-3" fontSize="xl">
            Request not found
          </Text>
          <Text align="center" color="muted">
            This request may have already been handled.
          </Text>
        </Box>
      </Box>
    );
  }

  const repositoryName = metadataText(request.metadata.repository_full_name);
  const issueNumber = request.sourceNumber ? `#${request.sourceNumber}` : "";
  const selectedStatus = statuses.find((status) => status.id === statusId);

  return (
    <Box className="grid h-dvh md:grid-cols-[minmax(0,1fr)_280px]">
      <Box className="h-dvh overflow-y-auto pb-8">
        <Container className="max-w-5xl pt-6">
          <Flex
            align="center"
            className="border-border bg-surface-muted/30 mb-8 rounded-xl border px-4 py-3"
            gap={2}
            justify="between"
          >
            <Flex align="center" className="min-w-0" gap={2}>
              <GitHubIcon className="text-primary h-5 shrink-0" />
              <Text className="line-clamp-1 font-medium">
                Request from GitHub {issueNumber}
              </Text>
              <Text className="line-clamp-1" color="muted">
                {repositoryName}
              </Text>
            </Flex>
            {request.sourceUrl ? (
              <Button
                asIcon
                color="tertiary"
                href={request.sourceUrl}
                rounded="full"
                size="sm"
                target="_blank"
                title="Open on GitHub"
              >
                <NewTabIcon className="h-4" />
              </Button>
            ) : null}
          </Flex>

          <TextEditor
            asTitle
            className="text-foreground mb-8 text-3xl md:text-4xl"
            editor={titleEditor}
          />
          <TextEditor className="text-lg" editor={descriptionEditor} />

          <Divider className="my-8" />
          <GitHubComments requestId={request.id} />
        </Container>
      </Box>

      <Box className="border-border hidden h-dvh overflow-y-auto border-l-[0.5px] p-5 md:block">
        <Text className="mb-5 font-semibold">Properties</Text>
        <Box className="space-y-4">
          <Box>
            <Text className="mb-2" color="muted">
              Request state
            </Text>
            <Flex align="center" gap={2}>
              <GitHubIcon className="text-primary h-4" />
              <Text>GitHub request</Text>
            </Flex>
          </Box>

          <Box>
            <Text className="mb-2" color="muted">
              Status
            </Text>
            <StatusesMenu>
              <StatusesMenu.Trigger>
                <Button color="tertiary" size="sm">
                  <StoryStatusIcon statusId={statusId} />
                  {selectedStatus?.name ?? "Todo"}
                </Button>
              </StatusesMenu.Trigger>
              <StatusesMenu.Items
                setStatusId={(nextStatusId) => {
                  handleUpdate({ statusId: nextStatusId });
                }}
                statusId={statusId}
                teamId={teamId}
              />
            </StatusesMenu>
          </Box>

          <Box>
            <Text className="mb-2" color="muted">
              Priority
            </Text>
            <PrioritiesMenu>
              <PrioritiesMenu.Trigger>
                <Button color="tertiary" size="sm">
                  <PriorityIcon priority={priority} />
                  {priority}
                </Button>
              </PrioritiesMenu.Trigger>
              <PrioritiesMenu.Items
                priority={priority}
                setPriority={(nextPriority: StoryPriority) => {
                  handleUpdate({ priority: nextPriority });
                }}
              />
            </PrioritiesMenu>
          </Box>
        </Box>

        <Divider className="my-6" />

        <Flex direction="column" gap={2}>
          <Button
            disabled={request.status !== "pending"}
            fullWidth
            leftIcon={<CheckIcon className="h-4" />}
            loading={acceptRequest.isPending}
            loadingText="Accepting..."
            onClick={() => {
              acceptRequest.mutate(request.id, {
                onSuccess: (res) => {
                  if (res.data?.acceptedStoryId) {
                    router.push(
                      withWorkspace(`/story/${res.data.acceptedStoryId}`),
                    );
                  }
                },
              });
            }}
          >
            Accept
          </Button>
          <Button
            color="tertiary"
            disabled={request.status !== "pending"}
            fullWidth
            leftIcon={<CloseIcon className="h-4" />}
            onClick={() => {
              setIsDeclining(true);
            }}
          >
            Decline
          </Button>
        </Flex>
      </Box>

      <ConfirmDialog
        confirmText="Decline request"
        description="Declining removes this item from the team request queue. You can still find the original item in the source integration."
        isLoading={declineRequest.isPending}
        isOpen={isDeclining}
        loadingText="Declining..."
        onCancel={() => {
          setIsDeclining(false);
        }}
        onClose={() => {
          setIsDeclining(false);
        }}
        onConfirm={() => {
          declineRequest.mutate(request.id, {
            onSuccess: (res) => {
              if (!res.error?.message) {
                setIsDeclining(false);
                router.push(withWorkspace(`/teams/${teamId}/requests`));
              }
            },
          });
        }}
        title="Decline this request?"
      />
    </Box>
  );
};
