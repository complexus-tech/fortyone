import type { ReactNode } from "react";
import { Avatar, Box, Text, Flex, Button, Tooltip } from "ui";
import { cn } from "lib";
import type { ChatStatus } from "ai";
import { useState } from "react";
import {
  CheckIcon,
  CopyIcon,
  PlusIcon,
  ReloadIcon,
  SprintsIcon,
  StoryIcon,
  TeamIcon,
  ObjectiveIcon,
  SearchIcon,
  NotificationsIcon,
  CommentIcon,
  TagsIcon,
  LinkIcon,
  AttachmentIcon,
  BrainIcon,
  SunIcon,
  WorkflowIcon,
  UserIcon,
  HistoryIcon,
} from "icons";
import { usePathname } from "next/navigation";
import { Streamdown } from "streamdown";
import type { User } from "@/types";
import { BurndownChart } from "@/modules/sprints/stories/burndown";
import { useCopyToClipboard, useTerminology } from "@/hooks";
import type { MayaUIMessage } from "@/lib/ai/tools/types";
import { NewStoryDialog } from "../new-story-dialog";
import { AiIcon } from "./ai";
import { Thinking } from "./thinking";
import { AttachmentsDisplay } from "./attachments-display";
import { Reasoning } from "./reasoning";
import { Sources } from "./sources";

type ChatMessageProps = {
  isLast: boolean;
  message: MayaUIMessage;
  profile: User | undefined;
  status: ChatStatus;
  regenerate: (messageId?: string) => void;
  onPromptSelect: (prompt: string) => void;
};

const ICON_CLASS = "size-3.5";

/** Maps tool part types to label + icon */
const TOOL_THINKING_META: Record<string, { label: string; icon: ReactNode }> = {
  // Stories
  "tool-listTeamStories": { label: "Fetching stories", icon: <StoryIcon className={ICON_CLASS} /> },
  "tool-searchStories": { label: "Searching stories", icon: <StoryIcon className={ICON_CLASS} /> },
  "tool-getStoryDetails": { label: "Getting story details", icon: <StoryIcon className={ICON_CLASS} /> },
  "tool-createStory": { label: "Creating story", icon: <StoryIcon className={ICON_CLASS} /> },
  "tool-updateStory": { label: "Updating story", icon: <StoryIcon className={ICON_CLASS} /> },
  "tool-deleteStory": { label: "Deleting story", icon: <StoryIcon className={ICON_CLASS} /> },
  "tool-bulkCreateStories": { label: "Creating stories", icon: <StoryIcon className={ICON_CLASS} /> },
  "tool-bulkUpdateStories": { label: "Updating stories", icon: <StoryIcon className={ICON_CLASS} /> },
  "tool-bulkDeleteStories": { label: "Deleting stories", icon: <StoryIcon className={ICON_CLASS} /> },
  "tool-assignStoriesToUser": { label: "Assigning stories", icon: <StoryIcon className={ICON_CLASS} /> },
  "tool-duplicateStory": { label: "Duplicating story", icon: <StoryIcon className={ICON_CLASS} /> },
  "tool-restoreStory": { label: "Restoring story", icon: <StoryIcon className={ICON_CLASS} /> },
  "tool-addStoryAssociation": { label: "Linking stories", icon: <LinkIcon className={ICON_CLASS} /> },
  "tool-removeStoryAssociation": { label: "Unlinking stories", icon: <LinkIcon className={ICON_CLASS} /> },
  // Sprints
  "tool-listSprints": { label: "Loading sprints", icon: <SprintsIcon className={ICON_CLASS} /> },
  "tool-listRunningSprints": { label: "Getting active sprints", icon: <SprintsIcon className={ICON_CLASS} /> },
  "tool-getSprintDetailsTool": { label: "Getting sprint details", icon: <SprintsIcon className={ICON_CLASS} /> },
  "tool-getSprintAnalyticsTool": { label: "Analyzing sprint data", icon: <SprintsIcon className={ICON_CLASS} /> },
  "tool-updateSprintSettings": { label: "Updating sprint settings", icon: <SprintsIcon className={ICON_CLASS} /> },
  // Teams
  "tool-listTeams": { label: "Loading teams", icon: <TeamIcon className={ICON_CLASS} /> },
  "tool-listPublicTeams": { label: "Loading public teams", icon: <TeamIcon className={ICON_CLASS} /> },
  "tool-getTeamDetails": { label: "Getting team details", icon: <TeamIcon className={ICON_CLASS} /> },
  "tool-listTeamMembers": { label: "Loading team members", icon: <UserIcon className={ICON_CLASS} /> },
  "tool-createTeamTool": { label: "Creating team", icon: <TeamIcon className={ICON_CLASS} /> },
  "tool-updateTeam": { label: "Updating team", icon: <TeamIcon className={ICON_CLASS} /> },
  "tool-joinTeam": { label: "Joining team", icon: <TeamIcon className={ICON_CLASS} /> },
  "tool-leaveTeam": { label: "Leaving team", icon: <TeamIcon className={ICON_CLASS} /> },
  "tool-deleteTeam": { label: "Deleting team", icon: <TeamIcon className={ICON_CLASS} /> },
  "tool-getTeamSettingsTool": { label: "Loading team settings", icon: <TeamIcon className={ICON_CLASS} /> },
  // Objectives & Key Results
  "tool-listObjectivesTool": { label: "Loading objectives", icon: <ObjectiveIcon className={ICON_CLASS} /> },
  "tool-listTeamObjectivesTool": { label: "Loading team objectives", icon: <ObjectiveIcon className={ICON_CLASS} /> },
  "tool-createObjectiveTool": { label: "Creating objective", icon: <ObjectiveIcon className={ICON_CLASS} /> },
  "tool-updateObjectiveTool": { label: "Updating objective", icon: <ObjectiveIcon className={ICON_CLASS} /> },
  "tool-deleteObjectiveTool": { label: "Deleting objective", icon: <ObjectiveIcon className={ICON_CLASS} /> },
  "tool-getObjectiveDetailsTool": { label: "Getting objective details", icon: <ObjectiveIcon className={ICON_CLASS} /> },
  "tool-objectiveAnalyticsTool": { label: "Analyzing objective data", icon: <ObjectiveIcon className={ICON_CLASS} /> },
  "tool-getObjectiveActivitiesTool": { label: "Loading objective activity", icon: <HistoryIcon className={ICON_CLASS} /> },
  "tool-listKeyResultsTool": { label: "Loading key results", icon: <ObjectiveIcon className={ICON_CLASS} /> },
  "tool-createKeyResultTool": { label: "Creating key result", icon: <ObjectiveIcon className={ICON_CLASS} /> },
  "tool-updateKeyResultTool": { label: "Updating key result", icon: <ObjectiveIcon className={ICON_CLASS} /> },
  "tool-deleteKeyResultTool": { label: "Deleting key result", icon: <ObjectiveIcon className={ICON_CLASS} /> },
  "tool-getKeyResultActivitiesTool": { label: "Loading key result activity", icon: <HistoryIcon className={ICON_CLASS} /> },
  // Other
  "tool-navigation": { label: "Navigating", icon: <WorkflowIcon className={ICON_CLASS} /> },
  "tool-search": { label: "Searching", icon: <SearchIcon className={ICON_CLASS} /> },
  "tool-members": { label: "Loading members", icon: <UserIcon className={ICON_CLASS} /> },
  "tool-comments": { label: "Loading comments", icon: <CommentIcon className={ICON_CLASS} /> },
  "tool-notifications": { label: "Checking notifications", icon: <NotificationsIcon className={ICON_CLASS} /> },
  "tool-statuses": { label: "Loading statuses", icon: <WorkflowIcon className={ICON_CLASS} /> },
  "tool-objectiveStatuses": { label: "Loading objective statuses", icon: <WorkflowIcon className={ICON_CLASS} /> },
  "tool-links": { label: "Loading links", icon: <LinkIcon className={ICON_CLASS} /> },
  "tool-labels": { label: "Loading labels", icon: <TagsIcon className={ICON_CLASS} /> },
  "tool-storyLabels": { label: "Managing labels", icon: <TagsIcon className={ICON_CLASS} /> },
  "tool-storyActivities": { label: "Loading activity", icon: <HistoryIcon className={ICON_CLASS} /> },
  "tool-listAttachments": { label: "Loading attachments", icon: <AttachmentIcon className={ICON_CLASS} /> },
  "tool-deleteAttachment": { label: "Deleting attachment", icon: <AttachmentIcon className={ICON_CLASS} /> },
  "tool-listMemories": { label: "Checking memory", icon: <BrainIcon className={ICON_CLASS} /> },
  "tool-createMemory": { label: "Saving to memory", icon: <BrainIcon className={ICON_CLASS} /> },
  "tool-updateMemory": { label: "Updating memory", icon: <BrainIcon className={ICON_CLASS} /> },
  "tool-deleteMemory": { label: "Removing memory", icon: <BrainIcon className={ICON_CLASS} /> },
  "tool-theme": { label: "Changing theme", icon: <SunIcon className={ICON_CLASS} /> },
};

const DEFAULT_TOOL_META = { label: "Working on it", icon: <WorkflowIcon className={ICON_CLASS} /> };

const getToolThinkingMeta = (toolType: string) => {
  return TOOL_THINKING_META[toolType] ?? DEFAULT_TOOL_META;
};

const isToolPart = (type: string): boolean => type.startsWith("tool-");

const RenderMessage = ({
  message,
  onPromptSelect,
  status,
  isLast,
}: {
  isLast: boolean;
  message: MayaUIMessage;
  status: ChatStatus;
  onPromptSelect: (prompt: string) => void;
}) => {
  const pathname = usePathname();
  const isStreaming = status === "streaming";
  const isAssistant = message.role === "assistant";
  const hasText = message.parts.some((p) => p.type === "text");

  const totalSources = message.parts.filter(
    (part) => part.type === "source-url",
  ).length;

  // Track whether a tool-specific thinking label is being rendered
  // so we can avoid showing the generic "Maya is thinking" at the same time
  let hasActiveToolThinking = false;
  if (isLast && isAssistant && status !== "ready") {
    hasActiveToolThinking = message.parts.some(
      (p) =>
        isToolPart(p.type) &&
        "state" in p &&
        (p.state === "input-available" || p.state === "input-streaming"),
    );
  }

  return (
    <>
      {/* Show generic thinking when streaming/submitted, no text yet,
          and no tool-specific thinking label is active */}
      {status !== "ready" &&
        isLast &&
        isAssistant &&
        !hasText &&
        !hasActiveToolThinking && <Thinking />}

      {message.parts.map((part, index) => {
        // Text content
        if (part.type === "text") {
          return (
            <Streamdown
              className={cn("chat-tables", {
                "text-foreground-inverse": message.role === "user",
              })}
              controls={{
                table: true,
                code: true,
                mermaid: {
                  download: true,
                  copy: true,
                  fullscreen: true,
                  panZoom: true,
                },
              }}
              isAnimating={isStreaming && isAssistant}
              key={index}
            >
              {part.text}
            </Streamdown>
          );
        }

        // Reasoning (visible while streaming and after completion)
        if (part.type === "reasoning") {
          return (
            <Reasoning
              className="mb-2"
              content={part.text}
              isStreaming={isStreaming && isLast}
              key={index}
            />
          );
        }

        // Generic tool thinking — show for any tool in progress
        if (isToolPart(part.type) && "state" in part) {
          // Show thinking indicator while tool is executing
          if (
            isLast &&
            (part.state === "input-available" ||
              part.state === "input-streaming")
          ) {
            const { label, icon } = getToolThinkingMeta(part.type);
            return (
              <Thinking key={index} message={label} icon={icon} />
            );
          }

          // Burndown chart — custom output for sprint analytics
          if (
            part.type === "tool-getSprintAnalyticsTool" &&
            part.state === "output-available"
          ) {
            return (
              <Box className="mb-3" key={index}>
                <Text
                  as="h3"
                  className="mt-4 mb-1 text-xl font-semibold antialiased"
                >
                  Burndown graph
                </Text>
                <BurndownChart
                  burndownData={part.output.analyticsReport?.burndown ?? []}
                  className={cn("h-72", {
                    "h-80": pathname.includes("/maya"),
                  })}
                />
              </Box>
            );
          }

          // Suggestions — shown when ready
          if (
            part.type === "tool-suggestions" &&
            part.state === "output-available"
          ) {
            return (
              <Flex className="mt-2" gap={2} key={index} wrap>
                {part.output.suggestions.map(
                  (suggestion: string, i: number) => (
                    <Button
                      color="tertiary"
                      className="truncate"
                      key={i}
                      onClick={() => onPromptSelect(suggestion)}
                      size="sm"
                    >
                      {suggestion}
                    </Button>
                  ),
                )}
              </Flex>
            );
          }
        }

        return null;
      })}

      {totalSources > 0 && (
        <Sources>
          <Sources.Trigger count={totalSources} />
          <Sources.Content>
            {message.parts.map((part, index) => {
              if (part.type === "source-url") {
                return (
                  <Sources.Source
                    href={part.url}
                    key={`${message.id}-${index}`}
                    title={part.title}
                  />
                );
              }
              return null;
            })}
          </Sources.Content>
        </Sources>
      )}
    </>
  );
};

export const ChatMessage = ({
  isLast,
  message,
  profile,
  status,
  regenerate,
  onPromptSelect,
}: ChatMessageProps) => {
  const [_, copy] = useCopyToClipboard();
  const [hasCopied, setHasCopied] = useState(false);
  const [isOpen, setIsOpen] = useState(false);
  const { getTermDisplay } = useTerminology();
  const content = message.parts.find((p) => p.type === "text")?.text ?? "";
  return (
    <>
      <Flex
        className={cn({
          "flex-row-reverse": message.role === "user",
        })}
        gap={3}
      >
        {message.role === "assistant" ? (
          <AiIcon />
        ) : (
          <Avatar
            color="tertiary"
            name={profile?.fullName || profile?.username}
            src={profile?.avatarUrl}
          />
        )}
        <Flex
          className={cn("max-w-[80%] flex-1", {
            "items-end": message.role === "user",
            "max-w-[85%]": message.role === "assistant",
          })}
          direction="column"
        >
          <Box
            className={cn("mb-2 rounded-2xl px-4 py-3", {
              "bg-background-inverse rounded-tr-md": message.role === "user",
              "bg-transparent p-0": message.role === "assistant",
            })}
          >
            <RenderMessage
              isLast={isLast}
              message={message}
              onPromptSelect={onPromptSelect}
              status={status}
            />
          </Box>
          <AttachmentsDisplay message={message} />
          <Flex className="mt-2 px-0.5" justify="between">
            {message.role === "assistant" && status !== "streaming" && (
              <Flex gap={2} justify="end">
                <Tooltip title={`Create ${getTermDisplay("storyTerm")}`}>
                  <Button
                    asIcon
                    color="tertiary"
                    onClick={() => {
                      setIsOpen(true);
                    }}
                    size="sm"
                    variant="naked"
                  >
                    <PlusIcon />
                  </Button>
                </Tooltip>
                <Tooltip title="Copy">
                  <Button
                    asIcon
                    color="tertiary"
                    onClick={() => {
                      copy(content).then(() => {
                        setHasCopied(true);
                        setTimeout(() => {
                          setHasCopied(false);
                        }, 1500);
                      });
                    }}
                    size="sm"
                    variant="naked"
                  >
                    {hasCopied ? <CheckIcon /> : <CopyIcon />}
                  </Button>
                </Tooltip>
                {isLast ? (
                  <Tooltip title="Retry">
                    <Button
                      asIcon
                      color="tertiary"
                      onClick={() => {
                        regenerate();
                      }}
                      size="sm"
                      variant="naked"
                    >
                      <ReloadIcon strokeWidth={2.8} />
                    </Button>
                  </Tooltip>
                ) : null}
              </Flex>
            )}
          </Flex>
        </Flex>
      </Flex>
      <NewStoryDialog
        description={content}
        isOpen={isOpen}
        setIsOpen={setIsOpen}
      />
    </>
  );
};
