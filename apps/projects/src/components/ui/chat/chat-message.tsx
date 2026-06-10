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
import { AttachmentsDisplay } from "./attachments-display";
import { AnalyticsReport } from "./analytics-report";

type ChatMessageProps = {
  isLast: boolean;
  message: MayaUIMessage;
  profile: User | undefined;
  status: ChatStatus;
  deferToolOutputs?: boolean;
  regenerate: (messageId?: string) => void;
  onPromptSelect: (prompt: string) => void;
};

/** Maps tool part types to the single progress label shown below the chat. */
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
  "tool-workspacePerformanceReportTool": "Building workspace report",
  "tool-storyPerformanceReportTool": "Building story report",
  "tool-objectiveProgressReportTool": "Building objective report",
  "tool-teamPerformanceReportTool": "Building team report",
  "tool-sprintPerformanceReportTool": "Building sprint report",
  "tool-timelineTrendsReportTool": "Building trends report",
  "tool-getGitHubIntegrationTool": "Checking GitHub integration",
  "tool-createGitHubInstallSessionTool": "Creating GitHub install link",
  "tool-resyncGitHubRepositoriesTool": "Resyncing GitHub repositories",
  "tool-createGitHubIssueSyncLinkTool": "Linking GitHub repository",
  "tool-deleteGitHubIssueSyncLinkTool": "Removing GitHub sync link",
  "tool-updateGitHubWorkspaceSettingsTool": "Updating GitHub settings",
  "tool-getGitHubTeamSettingsTool": "Checking GitHub automation",
  "tool-updateGitHubTeamSettingsTool": "Updating GitHub automation",
  "tool-getStoryGitHubLinksTool": "Checking story GitHub links",
  "tool-getStoryGitHubCommentsTool": "Reading GitHub comments",
  "tool-postStoryGitHubCommentTool": "Posting GitHub comment",
  "tool-deleteStoryGitHubLinkTool": "Removing story GitHub link",
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

const DEFAULT_PROGRESS_LABEL = "Working on it";

const isToolPart = (type: string): boolean => type.startsWith("tool-");

type ToolMessagePart = MayaUIMessage["parts"][number] & {
  state: string;
  output?: unknown;
};

const isToolMessagePart = (
  part: MayaUIMessage["parts"][number],
): part is ToolMessagePart => isToolPart(part.type) && "state" in part;

const isAnalyticsReportOutput = (
  output: unknown,
): output is Record<string, unknown> => {
  if (!output || typeof output !== "object" || !("kind" in output)) {
    return false;
  }

  const kind = (output as { kind?: unknown }).kind;
  return typeof kind === "string" && kind.endsWith("-report");
};

const asRecord = (value: unknown): Record<string, unknown> =>
  value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : {};

const getSprintBurndownData = (output: unknown) => {
  const outputRecord = asRecord(output);
  const analyticsReport = asRecord(outputRecord.analyticsReport);
  return Array.isArray(analyticsReport.burndown)
    ? analyticsReport.burndown
    : [];
};

const getSuggestions = (output: unknown) => {
  const outputRecord = asRecord(output);
  return Array.isArray(outputRecord.suggestions)
    ? outputRecord.suggestions.filter(
        (suggestion): suggestion is string => typeof suggestion === "string",
      )
    : [];
};

export const getMessageProgressLabel = (message: MayaUIMessage) => {
  const lastPart = message.parts.at(-1);

  if (lastPart?.type === "text" && lastPart.text.trim()) {
    return undefined;
  }

  const latestToolPart = message.parts.filter(isToolMessagePart).at(-1);

  if (!latestToolPart) {
    return "Thinking";
  }

  return TOOL_THINKING_LABELS[latestToolPart.type] ?? DEFAULT_PROGRESS_LABEL;
};

export const getMessageText = (message: MayaUIMessage) =>
  message.parts
    .filter((part) => part.type === "text")
    .map((part) => part.text)
    .join("");

export const hasVisibleMessageContent = (
  message: MayaUIMessage,
  deferToolOutputs = false,
) => {
  if (message.role === "user") {
    return true;
  }

  if (getMessageText(message).trim()) {
    return true;
  }

  if (message.parts.some((part) => part.type === "file")) {
    return true;
  }

  if (deferToolOutputs) {
    return false;
  }

  return message.parts.some(
    (part) => isToolMessagePart(part) && part.state === "output-available",
  );
};

const RenderMessage = ({
  message,
  onPromptSelect,
  status,
  deferToolOutputs = false,
}: {
  message: MayaUIMessage;
  status: ChatStatus;
  deferToolOutputs?: boolean;
  onPromptSelect: (prompt: string) => void;
}) => {
  const pathname = usePathname();
  const isStreaming = status === "streaming";
  const isAssistant = message.role === "assistant";

  const textParts = message.parts.filter((part) => part.type === "text");
  const toolParts = message.parts.filter(isToolMessagePart);
  const reportParts = deferToolOutputs
    ? []
    : toolParts.filter(
        (part) =>
          part.state === "output-available" &&
          isAnalyticsReportOutput(part.output),
      );
  const sprintAnalyticsParts = deferToolOutputs
    ? []
    : toolParts.filter(
        (part) =>
          part.type === "tool-getSprintAnalyticsTool" &&
          part.state === "output-available",
      );
  const suggestionParts = deferToolOutputs
    ? []
    : toolParts.filter(
        (part) =>
          part.type === "tool-suggestions" && part.state === "output-available",
      );

  return (
    <>
      {textParts.map((part, index) => (
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
          key={`${message.id}-text-${index}`}
        >
          {part.text}
        </Streamdown>
      ))}

      {reportParts.map((part, index) => (
        <AnalyticsReport
          key={`${message.id}-report-${index}`}
          output={asRecord(part.output)}
        />
      ))}

      {sprintAnalyticsParts.map((part, index) => (
        <Box className="mb-3" key={`${message.id}-sprint-${index}`}>
          <Text as="h3" className="mt-4 mb-1 text-xl font-semibold antialiased">
            Burndown graph
          </Text>
          <BurndownChart
            burndownData={getSprintBurndownData(part.output)}
            className={cn("h-72", {
              "h-80": pathname.includes("/maya"),
            })}
          />
        </Box>
      ))}

      {suggestionParts.map((part, index) => (
        <Flex
          className="mt-4"
          gap={2}
          key={`${message.id}-suggestions-${index}`}
          wrap
        >
          {getSuggestions(part.output).map((suggestion, i) => (
            <Button
              color="tertiary"
              className="truncate"
              key={i}
              onClick={() => onPromptSelect(suggestion)}
              size="sm"
            >
              {suggestion}
            </Button>
          ))}
        </Flex>
      ))}
    </>
  );
};

export const ChatMessage = ({
  isLast,
  message,
  profile,
  status,
  deferToolOutputs = false,
  regenerate,
  onPromptSelect,
}: ChatMessageProps) => {
  const [_, copy] = useCopyToClipboard();
  const [hasCopied, setHasCopied] = useState(false);
  const [isOpen, setIsOpen] = useState(false);
  const { getTermDisplay } = useTerminology();
  const content = getMessageText(message);
  return (
    <>
      <Flex
        className={cn({
          "flex-row-reverse": message.role === "user",
        })}
        gap={message.role === "user" ? 3 : 0}
      >
        {message.role === "user" ? (
          <Avatar
            color="tertiary"
            name={profile?.fullName || profile?.username}
            src={profile?.avatarUrl}
          />
        ) : null}
        <Flex
          className={cn("flex-1", {
            "items-end": message.role === "user",
            "max-w-[80%]": message.role === "user",
            "max-w-full": message.role === "assistant",
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
              deferToolOutputs={deferToolOutputs}
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
