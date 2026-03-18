import { Avatar, Box, Text, Flex, Button, Tooltip } from "ui";
import { cn } from "lib";
import type { ChatStatus } from "ai";
import { useState } from "react";
import { CheckIcon, CopyIcon, PlusIcon, ReloadIcon } from "icons";
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

/** Maps tool part types to user-friendly thinking labels */
const TOOL_THINKING_LABELS: Record<string, string> = {
  // Stories
  "tool-listTeamStories": "Fetching stories",
  "tool-searchStories": "Searching stories",
  "tool-getStoryDetails": "Getting story details",
  "tool-createStory": "Creating story",
  "tool-updateStory": "Updating story",
  "tool-deleteStory": "Deleting story",
  "tool-bulkCreateStories": "Creating stories",
  "tool-bulkUpdateStories": "Updating stories",
  "tool-bulkDeleteStories": "Deleting stories",
  "tool-assignStoriesToUser": "Assigning stories",
  "tool-duplicateStory": "Duplicating story",
  "tool-restoreStory": "Restoring story",
  "tool-addStoryAssociation": "Linking stories",
  "tool-removeStoryAssociation": "Unlinking stories",
  // Sprints
  "tool-listSprints": "Loading sprints",
  "tool-listRunningSprints": "Getting active sprints",
  "tool-getSprintDetailsTool": "Getting sprint details",
  "tool-getSprintAnalyticsTool": "Analyzing sprint data",
  "tool-updateSprintSettings": "Updating sprint settings",
  // Teams
  "tool-listTeams": "Loading teams",
  "tool-listPublicTeams": "Loading public teams",
  "tool-getTeamDetails": "Getting team details",
  "tool-listTeamMembers": "Loading team members",
  "tool-createTeamTool": "Creating team",
  "tool-updateTeam": "Updating team",
  "tool-joinTeam": "Joining team",
  "tool-leaveTeam": "Leaving team",
  "tool-deleteTeam": "Deleting team",
  "tool-getTeamSettingsTool": "Loading team settings",
  // Objectives & Key Results
  "tool-listObjectivesTool": "Loading objectives",
  "tool-listTeamObjectivesTool": "Loading team objectives",
  "tool-createObjectiveTool": "Creating objective",
  "tool-updateObjectiveTool": "Updating objective",
  "tool-deleteObjectiveTool": "Deleting objective",
  "tool-getObjectiveDetailsTool": "Getting objective details",
  "tool-objectiveAnalyticsTool": "Analyzing objective data",
  "tool-getObjectiveActivitiesTool": "Loading objective activity",
  "tool-listKeyResultsTool": "Loading key results",
  "tool-createKeyResultTool": "Creating key result",
  "tool-updateKeyResultTool": "Updating key result",
  "tool-deleteKeyResultTool": "Deleting key result",
  "tool-getKeyResultActivitiesTool": "Loading key result activity",
  // Other
  "tool-navigation": "Navigating",
  "tool-search": "Searching",
  "tool-members": "Loading members",
  "tool-comments": "Loading comments",
  "tool-notifications": "Checking notifications",
  "tool-statuses": "Loading statuses",
  "tool-objectiveStatuses": "Loading objective statuses",
  "tool-links": "Loading links",
  "tool-labels": "Loading labels",
  "tool-storyLabels": "Managing labels",
  "tool-storyActivities": "Loading activity",
  "tool-listAttachments": "Loading attachments",
  "tool-deleteAttachment": "Deleting attachment",
  "tool-listMemories": "Checking memory",
  "tool-createMemory": "Saving to memory",
  "tool-updateMemory": "Updating memory",
  "tool-deleteMemory": "Removing memory",
  "tool-theme": "Changing theme",
};

const getToolThinkingLabel = (toolType: string): string => {
  return TOOL_THINKING_LABELS[toolType] ?? "Working on it";
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
            return (
              <Thinking
                key={index}
                message={getToolThinkingLabel(part.type)}
              />
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
