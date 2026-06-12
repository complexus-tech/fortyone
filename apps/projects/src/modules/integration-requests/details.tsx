"use client";

import { useEffect, useState } from "react";
import { useEditor } from "@tiptap/react";
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
import { format, formatISO } from "date-fns";
import { cn } from "lib";
import {
  CalendarIcon,
  ChatIcon,
  CheckIcon,
  CloseIcon,
  ClockIcon,
  CopyIcon,
  GitHubIcon,
  LinkIcon,
  MoreHorizontalIcon,
  ObjectiveIcon,
  SprintsIcon,
} from "icons";
import {
  Avatar,
  Box,
  Button,
  Container,
  DatePicker,
  Divider,
  Flex,
  Menu,
  Skeleton,
  Tabs,
  Text,
  TextEditor,
  TimeAgo,
} from "ui";
import {
  AssigneesMenu,
  ConfirmDialog,
  PrioritiesMenu,
  PriorityIcon,
  StatusesMenu,
  StoryStatusIcon,
} from "@/components/ui";
import { ObjectivesMenu } from "@/components/ui/story/objectives-menu";
import { SprintsMenu } from "@/components/ui/story/sprints-menu";
import { useDebounce, useTerminology, useWorkspacePath } from "@/hooks";
import {
  formatEstimate,
  getEstimateOptions,
  type EstimateScheme,
} from "@/lib/estimate";
import { createRichTextStarterKit } from "@/lib/tiptap/starter-kit";
import { useSession } from "@/lib/auth/client";
import { useMembers } from "@/lib/hooks/members";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { BodyContainer } from "@/components/shared";
import type { Member } from "@/types";
import type { State } from "@/types/states";
import type { GitHubComment } from "@/modules/settings/workspace/integrations/github/types";
import type { StoryPriority } from "@/modules/stories/types";
import { Option } from "@/modules/story/components/options";
import { useObjective } from "@/modules/objectives/hooks/use-objective";
import { useSprint } from "@/modules/sprints/hooks/sprint-details";
import { useTeamSettings } from "@/modules/teams/hooks/use-team-settings";
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
import type {
  IntegrationRequest,
  UpdateIntegrationRequestInput,
} from "./types";

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
      placeholder: "Leave a comment...",
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

const RequestGitHubBanner = ({
  canEditRequest,
  issueNumber,
  onAccept,
  onDecline,
  repositoryName,
  sourceUrl,
}: {
  canEditRequest: boolean;
  issueNumber: string;
  onAccept: () => void;
  onDecline: () => void;
  repositoryName: string | null;
  sourceUrl?: string;
}) => (
  <Box className="mb-3 space-y-2">
    <Flex
      align="center"
      className="border-primary/20 bg-primary/5 rounded-xl border px-4 py-3"
      justify="between"
    >
      <Flex align="center" className="min-w-0" gap={2}>
        <GitHubIcon className="text-primary h-5 shrink-0" />
        <Text className="line-clamp-1" color="primary" fontWeight="medium">
          Issue synced with GitHub {issueNumber}
        </Text>
        {repositoryName ? (
          <Text className="line-clamp-1" color="muted">
            {repositoryName}
          </Text>
        ) : null}
      </Flex>
      {sourceUrl ? (
        <Flex align="center" gap={1}>
          <a
            className="text-primary hover:text-primary/80 rounded-md p-1 transition"
            href={sourceUrl}
            rel="noopener noreferrer"
            target="_blank"
            title="Open on GitHub"
          >
            <LinkIcon className="text-current" />
          </a>
          <Menu>
            <Menu.Button>
              <button
                className="text-primary hover:text-primary/80 rounded-md p-1 transition"
                type="button"
              >
                <MoreHorizontalIcon className="h-5 text-current" />
              </button>
            </Menu.Button>
            <Menu.Items align="end">
              <Menu.Group>
                <Menu.Item
                  onSelect={() => {
                    window.open(sourceUrl, "_blank", "noopener,noreferrer");
                  }}
                >
                  <GitHubIcon className="text-icon h-5 w-auto" />
                  Open on GitHub
                </Menu.Item>
                <Menu.Item
                  onSelect={() => {
                    navigator.clipboard.writeText(sourceUrl);
                  }}
                >
                  <CopyIcon className="text-icon h-5 w-auto" />
                  Copy link
                </Menu.Item>
                <Menu.Item disabled={!canEditRequest} onSelect={onAccept}>
                  <CheckIcon className="text-icon h-5 w-auto" />
                  Accept request
                </Menu.Item>
                <Menu.Item
                  className="text-danger"
                  disabled={!canEditRequest}
                  onSelect={onDecline}
                >
                  <CloseIcon className="text-danger" />
                  Decline request...
                </Menu.Item>
              </Menu.Group>
            </Menu.Items>
          </Menu>
        </Flex>
      ) : null}
    </Flex>
  </Box>
);

const RequestSourceBanner = ({
  canEditRequest,
  channel,
  onAccept,
  onDecline,
  provider,
  sourceUrl,
}: {
  canEditRequest: boolean;
  sourceUrl?: string;
  channel: string | null;
  onAccept: () => void;
  onDecline: () => void;
  provider: "slack" | "intercom";
}) => (
  <Box className="mb-3 space-y-2">
    <Flex
      align="center"
      className="border-border bg-surface-muted/40 rounded-xl border px-4 py-3"
      justify="between"
    >
      <Flex align="center" className="min-w-0" gap={2}>
        <ChatIcon className="h-5 shrink-0" />
        <Text className="line-clamp-1 font-medium">
          Story from {provider === "slack" ? "Slack" : "Intercom"}
        </Text>
        {channel ? (
          <Text className="line-clamp-1" color="muted">
            #{channel}
          </Text>
        ) : null}
      </Flex>
      {sourceUrl ? (
        <Flex align="center" gap={1}>
          <a
            className="text-muted hover:text-foreground rounded-md p-1 transition"
            href={sourceUrl}
            rel="noopener noreferrer"
            target="_blank"
            title="Open source"
          >
            <LinkIcon className="h-5 text-current" />
          </a>
          <Menu>
            <Menu.Button>
              <button
                className="text-muted hover:text-foreground rounded-md p-1 transition"
                type="button"
              >
                <MoreHorizontalIcon className="h-5 text-current" />
              </button>
            </Menu.Button>
            <Menu.Items align="end">
              <Menu.Group>
                <Menu.Item
                  onSelect={() => {
                    window.open(sourceUrl, "_blank", "noopener,noreferrer");
                  }}
                >
                  <LinkIcon className="text-icon h-5 w-auto" />
                  Open source
                </Menu.Item>
                <Menu.Item
                  onSelect={() => {
                    navigator.clipboard.writeText(sourceUrl);
                  }}
                >
                  <CopyIcon className="text-icon h-5 w-auto" />
                  Copy link
                </Menu.Item>
                <Menu.Item disabled={!canEditRequest} onSelect={onAccept}>
                  <CheckIcon className="text-icon h-5 w-auto" />
                  Accept request
                </Menu.Item>
                <Menu.Item
                  className="text-danger"
                  disabled={!canEditRequest}
                  onSelect={onDecline}
                >
                  <CloseIcon className="text-danger" />
                  Decline request...
                </Menu.Item>
              </Menu.Group>
            </Menu.Items>
          </Menu>
        </Flex>
      ) : null}
    </Flex>
  </Box>
);

const GitHubComments = ({ requestId }: { requestId: string }) => {
  const { data: comments = [], isLoading } =
    useRequestGitHubComments(requestId);

  return (
    <Box>
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

type RequestPropertiesProps = {
  assignee?: Member;
  canEditRequest: boolean;
  onUpdate: (payload: UpdateIntegrationRequestInput) => void;
  priority: StoryPriority;
  request: IntegrationRequest;
  selectedStatus?: State;
  statusId?: string;
  teamId: string;
  variant?: "sidebar" | "inline";
};

const RequestProperties = ({
  assignee,
  canEditRequest,
  onUpdate,
  priority,
  request,
  selectedStatus,
  statusId,
  teamId,
  variant = "sidebar",
}: RequestPropertiesProps) => {
  const isInline = variant === "inline";
  const { getTermDisplay } = useTerminology();
  const { data: teamSettings } = useTeamSettings(teamId);
  const { data: selectedObjective } = useObjective(
    request.objectiveId ?? null,
    teamId,
  );
  const { data: selectedSprint } = useSprint(request.sprintId ?? null, teamId);
  const estimateScheme = (teamSettings?.estimationSettings.scheme ??
    "points") as EstimateScheme;
  const requestEstimateLabel = formatEstimate(
    estimateScheme,
    request.estimateValue,
    "compact",
  );

  return (
    <Container
      className={cn("text-text-muted px-0.5 pt-4 md:px-6", {
        "px-0 pt-0 md:px-0": isInline,
      })}
    >
      {!isInline ? (
        <Box className="mb-0 grid grid-cols-[9rem_auto] items-center gap-3 md:mb-6">
          <Text className="hidden md:block" fontWeight="semibold">
            Properties
          </Text>
        </Box>
      ) : null}

      <Box className={cn("flex flex-wrap gap-2", { "md:block": !isInline })}>
        <Option
          isCompact={isInline}
          isNotifications={isInline}
          label="Status"
          value={
            <StatusesMenu>
              <StatusesMenu.Trigger>
                <Button
                  color="tertiary"
                  disabled={!canEditRequest}
                  leftIcon={<StoryStatusIcon statusId={statusId} />}
                  size="sm"
                  variant={isInline ? "solid" : "naked"}
                >
                  {selectedStatus?.name ?? "Todo"}
                </Button>
              </StatusesMenu.Trigger>
              <StatusesMenu.Items
                setStatusId={(nextStatusId) => {
                  onUpdate({ statusId: nextStatusId });
                }}
                statusId={statusId}
                teamId={teamId}
              />
            </StatusesMenu>
          }
        />
        <Option
          isCompact={isInline}
          isNotifications={isInline}
          label="Priority"
          value={
            <PrioritiesMenu>
              <PrioritiesMenu.Trigger>
                <Button
                  color="tertiary"
                  disabled={!canEditRequest}
                  leftIcon={<PriorityIcon priority={priority} />}
                  size="sm"
                  variant={isInline ? "solid" : "naked"}
                >
                  {priority}
                </Button>
              </PrioritiesMenu.Trigger>
              <PrioritiesMenu.Items
                priority={priority}
                setPriority={(nextPriority: StoryPriority) => {
                  onUpdate({ priority: nextPriority });
                }}
              />
            </PrioritiesMenu>
          }
        />
        <Option
          isCompact={isInline}
          isNotifications={isInline}
          label="Assignee"
          value={
            <AssigneesMenu>
              <AssigneesMenu.Trigger>
                <Button
                  className="font-medium"
                  color="tertiary"
                  disabled={!canEditRequest}
                  leftIcon={
                    <Avatar
                      className="text-foreground/80"
                      name={assignee?.fullName}
                      size="xs"
                      src={assignee?.avatarUrl}
                    />
                  }
                  size="sm"
                  variant={isInline ? "solid" : "naked"}
                >
                  {assignee?.username ?? (
                    <Text as="span" color="muted">
                      Assign
                    </Text>
                  )}
                </Button>
              </AssigneesMenu.Trigger>
              <AssigneesMenu.Items
                assigneeId={request.assigneeId}
                onAssigneeSelected={(assigneeId) => {
                  onUpdate({ assigneeId: assigneeId ?? undefined });
                }}
                teamId={teamId}
              />
            </AssigneesMenu>
          }
        />
        <Option
          isCompact={isInline}
          isNotifications={isInline}
          label="Estimate"
          value={
            <Menu>
              <Menu.Button>
                <Button
                  color="tertiary"
                  disabled={!canEditRequest}
                  size="sm"
                  variant={isInline ? "solid" : "naked"}
                >
                  {requestEstimateLabel}
                </Button>
              </Menu.Button>
              <Menu.Items align="start">
                <Menu.Group>
                  {getEstimateOptions(estimateScheme).map(
                    ({ label, value }) => (
                      <Menu.Item
                        key={value}
                        onSelect={() => {
                          onUpdate({ estimateValue: value });
                        }}
                      >
                        {label}
                      </Menu.Item>
                    ),
                  )}
                </Menu.Group>
              </Menu.Items>
            </Menu>
          }
        />
        <Option
          isCompact={isInline}
          isNotifications={isInline}
          label={getTermDisplay("objectiveTerm", { capitalize: true })}
          value={
            <ObjectivesMenu>
              <ObjectivesMenu.Trigger>
                <Button
                  color="tertiary"
                  disabled={!canEditRequest}
                  leftIcon={<ObjectiveIcon className="h-4" />}
                  size="sm"
                  variant={isInline ? "solid" : "naked"}
                >
                  <span className="max-w-40 truncate">
                    {selectedObjective?.name ??
                      getTermDisplay("objectiveTerm", { capitalize: true })}
                  </span>
                </Button>
              </ObjectivesMenu.Trigger>
              <ObjectivesMenu.Items
                objectiveId={request.objectiveId}
                setObjectiveId={(objectiveId) => {
                  if (objectiveId) {
                    onUpdate({ objectiveId });
                  }
                }}
                teamId={teamId}
              />
            </ObjectivesMenu>
          }
        />
        <Option
          isCompact={isInline}
          isNotifications={isInline}
          label={getTermDisplay("sprintTerm", { capitalize: true })}
          value={
            <SprintsMenu>
              <SprintsMenu.Trigger>
                <Button
                  color="tertiary"
                  disabled={!canEditRequest}
                  leftIcon={<SprintsIcon className="h-[1.05rem]" />}
                  size="sm"
                  variant={isInline ? "solid" : "naked"}
                >
                  <span className="max-w-40 truncate">
                    {selectedSprint?.name ??
                      getTermDisplay("sprintTerm", { capitalize: true })}
                  </span>
                </Button>
              </SprintsMenu.Trigger>
              <SprintsMenu.Items
                setSprintId={(sprintId) => {
                  if (sprintId) {
                    onUpdate({ sprintId });
                  }
                }}
                sprintId={request.sprintId}
                teamId={teamId}
              />
            </SprintsMenu>
          }
        />
        <Option
          isCompact={isInline}
          isNotifications={isInline}
          label="Start"
          value={
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  color="tertiary"
                  disabled={!canEditRequest}
                  leftIcon={<CalendarIcon className="h-4" />}
                  size="sm"
                  variant={isInline ? "solid" : "naked"}
                >
                  {request.startDate
                    ? format(new Date(request.startDate), "MMM d")
                    : "Start"}
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar
                onDayClick={(day) => {
                  onUpdate({
                    startDate: formatISO(day, { representation: "date" }),
                  });
                }}
                selected={
                  request.startDate ? new Date(request.startDate) : undefined
                }
              />
            </DatePicker>
          }
        />
        <Option
          isCompact={isInline}
          isNotifications={isInline}
          label="Deadline"
          value={
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  color="tertiary"
                  disabled={!canEditRequest}
                  leftIcon={<CalendarIcon className="h-4" />}
                  size="sm"
                  variant={isInline ? "solid" : "naked"}
                >
                  {request.endDate
                    ? format(new Date(request.endDate), "MMM d")
                    : "Deadline"}
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar
                onDayClick={(day) => {
                  onUpdate({
                    endDate: formatISO(day, { representation: "date" }),
                  });
                }}
                selected={
                  request.endDate ? new Date(request.endDate) : undefined
                }
              />
            </DatePicker>
          }
        />
      </Box>
    </Container>
  );
};

type RequestAttachment = {
  name: string;
  url?: string;
};

type RequestExternalLink = {
  title: string;
  url: string;
};

const metadataLinks = (
  metadata: Record<string, unknown>,
): RequestExternalLink[] => {
  const raw = metadata.links ?? metadata.urls ?? metadata.external_links;
  if (!Array.isArray(raw)) return [];

  return raw.flatMap((item): RequestExternalLink[] => {
    if (typeof item === "string" && item.trim()) {
      return [{ title: item, url: item }];
    }
    if (!item || typeof item !== "object") {
      return [];
    }
    const record = item as Record<string, unknown>;
    const url = record.url ?? record.href;
    if (typeof url !== "string" || !url.trim()) {
      return [];
    }
    const title = record.title ?? record.name ?? record.label ?? url;
    return [
      {
        title: typeof title === "string" && title.trim() ? title : url,
        url,
      },
    ];
  });
};

const RequestExternalLinks = ({
  metadata,
}: {
  metadata: Record<string, unknown>;
}) => {
  const links = metadataLinks(metadata);
  if (links.length === 0) return null;

  return (
    <Box className="border-border mt-5 border-t-[0.5px] pt-4">
      <Text as="h4" className="mb-3" fontWeight="medium">
        External links
      </Text>
      <Box className="space-y-2">
        {links.map((link) => (
          <a
            className="border-border hover:bg-surface-muted flex items-center gap-2 rounded-lg border px-3 py-2 transition"
            href={link.url}
            key={`${link.title}-${link.url}`}
            rel="noopener noreferrer"
            target="_blank"
          >
            <LinkIcon className="h-4 shrink-0" />
            <Text className="line-clamp-1">{link.title}</Text>
          </a>
        ))}
      </Box>
    </Box>
  );
};

const metadataAttachments = (
  metadata: Record<string, unknown>,
): RequestAttachment[] => {
  const raw = metadata.attachments ?? metadata.files;
  if (!Array.isArray(raw)) return [];

  return raw.flatMap((item): RequestAttachment[] => {
    if (typeof item === "string") {
      return [{ name: item, url: item }];
    }
    if (!item || typeof item !== "object") {
      return [];
    }
    const record = item as Record<string, unknown>;
    const name = record.name ?? record.filename ?? record.title ?? record.url;
    if (typeof name !== "string" || !name.trim()) {
      return [];
    }
    return [
      {
        name,
        url: typeof record.url === "string" ? record.url : undefined,
      },
    ];
  });
};

const RequestAttachments = ({
  metadata,
}: {
  metadata: Record<string, unknown>;
}) => {
  const attachments = metadataAttachments(metadata);
  if (attachments.length === 0) return null;

  return (
    <Box className="border-border mt-5 border-t-[0.5px] pt-4">
      <Text as="h4" className="mb-3" fontWeight="medium">
        Attachments
      </Text>
      <Box className="space-y-2">
        {attachments.map((attachment) =>
          attachment.url ? (
            <a
              className="border-border hover:bg-surface-muted flex items-center gap-2 rounded-lg border px-3 py-2 transition"
              href={attachment.url}
              key={`${attachment.name}-${attachment.url}`}
              rel="noopener noreferrer"
              target="_blank"
            >
              <LinkIcon className="h-4 shrink-0" />
              <Text className="line-clamp-1">{attachment.name}</Text>
            </a>
          ) : (
            <Flex
              align="center"
              className="border-border rounded-lg border px-3 py-2"
              gap={2}
              key={attachment.name}
            >
              <LinkIcon className="h-4 shrink-0" />
              <Text className="line-clamp-1">{attachment.name}</Text>
            </Flex>
          ),
        )}
      </Box>
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
  const { data: members = [] } = useMembers();
  const updateRequest = useUpdateIntegrationRequest();
  const acceptRequest = useAcceptIntegrationRequest();
  const declineRequest = useDeclineIntegrationRequest();

  const defaultStatus =
    statuses.find((status) => status.category === "unstarted") ||
    statuses.at(0);
  const statusId = request?.statusId ?? defaultStatus?.id;
  const priority: StoryPriority = request?.priority ?? "No Priority";

  const handleUpdate = (payload: UpdateIntegrationRequestInput) => {
    if (!request) return;
    updateRequest.mutate({ requestId: request.id, payload });
  };
  const debouncedHandleUpdate = useDebounce(handleUpdate, DEBOUNCE_DELAY);

  const descriptionEditor = useEditor({
    extensions: [
      createRichTextStarterKit(),
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
  const slackChannel = metadataText(request.metadata.slack_channel);
  const issueNumber = request.sourceNumber ? `#${request.sourceNumber}` : "";
  const selectedStatus = statuses.find((status) => status.id === statusId);
  const assignee = members.find((member) => member.id === request.assigneeId);
  const canEditRequest = request.status === "pending";

  const handleAccept = () => {
    acceptRequest.mutate(request.id, {
      onSuccess: (res) => {
        if (res.data?.acceptedStoryId) {
          router.push(withWorkspace(`/story/${res.data.acceptedStoryId}`));
        }
      },
    });
  };

  const handleDecline = () => {
    setIsDeclining(true);
  };

  return (
    <Box className="h-dvh">
      <Box className="notification-story-container hidden h-full md:flex">
        <Box className="min-w-0 flex-1">
          <BodyContainer className="h-dvh overflow-y-auto pb-8">
            <Container className="max-w-7xl pt-7">
              {request.provider === "github" ? (
                <RequestGitHubBanner
                  canEditRequest={canEditRequest}
                  issueNumber={issueNumber}
                  onAccept={handleAccept}
                  onDecline={handleDecline}
                  repositoryName={repositoryName}
                  sourceUrl={request.sourceUrl}
                />
              ) : null}
              {request.provider === "slack" ||
              request.provider === "intercom" ? (
                <RequestSourceBanner
                  canEditRequest={canEditRequest}
                  channel={slackChannel}
                  onAccept={handleAccept}
                  onDecline={handleDecline}
                  provider={request.provider}
                  sourceUrl={request.sourceUrl}
                />
              ) : null}
              <TextEditor
                asTitle
                className="text-foreground mb-8 text-3xl md:text-4xl"
                editor={titleEditor}
              />
              <TextEditor className="text-lg" editor={descriptionEditor} />
              <Box className="notification-story-inline-options mt-6 hidden">
                <RequestProperties
                  assignee={assignee}
                  canEditRequest={canEditRequest}
                  onUpdate={handleUpdate}
                  priority={priority}
                  request={request}
                  selectedStatus={selectedStatus}
                  statusId={statusId}
                  teamId={teamId}
                  variant="inline"
                />
              </Box>
              <RequestExternalLinks metadata={request.metadata} />
              <RequestAttachments metadata={request.metadata} />
              <Divider className="my-6" />
              {request.provider === "github" ? (
                <Box>
                  <Text
                    as="h4"
                    className="mb-4 flex items-center gap-1"
                    fontWeight="medium"
                  >
                    <ClockIcon className="relative -top-px" />
                    Activity feed
                  </Text>
                  <Tabs defaultValue="github">
                    <Tabs.List className="mx-0 mb-5 md:mx-0">
                      <Tabs.Tab
                        className="gap-1 px-2"
                        leftIcon={<GitHubIcon className="h-[1.05rem]" />}
                        value="github"
                      >
                        GitHub
                      </Tabs.Tab>
                    </Tabs.List>
                    <Tabs.Panel value="github">
                      <GitHubComments requestId={request.id} />
                    </Tabs.Panel>
                  </Tabs>
                </Box>
              ) : (
                <Box>
                  <Text
                    as="h4"
                    className="mb-2 flex items-center gap-1"
                    fontWeight="medium"
                  >
                    <ClockIcon className="relative -top-px" />
                    Activity feed
                  </Text>
                  <Text color="muted">
                    Slack story details are captured in the source link and
                    metadata.
                  </Text>
                </Box>
              )}
            </Container>
          </BodyContainer>
        </Box>

        <Box className="notification-story-sidebar from-sidebar/70 to-sidebar/40 border-border w-(--story-sidebar-width) shrink-0 border-l-[0.5px] bg-linear-to-br md:h-dvh md:overflow-y-auto md:pb-6">
          <RequestProperties
            assignee={assignee}
            canEditRequest={canEditRequest}
            onUpdate={handleUpdate}
            priority={priority}
            request={request}
            selectedStatus={selectedStatus}
            statusId={statusId}
            teamId={teamId}
          />
        </Box>
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
