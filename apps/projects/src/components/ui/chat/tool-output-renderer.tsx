"use client";

import type { ReactNode } from "react";
import type { ChatAddToolApproveResponseFunction } from "ai";
import dynamic from "next/dynamic";
import { Button, Flex, Tooltip } from "ui";
import { useStatuses } from "@/lib/hooks/statuses";
import { PriorityIcon } from "../priority-icon";
import { Dot } from "../dot";
import { EntityResults } from "./entity-results";
import { StoryResults } from "./story-results";
import type { StoryResultStatus } from "./story-results";
import { MayaWorkPlanResult } from "./maya-work-plan-result";
import type { ToolMessagePart } from "./tool-output-policy";
import {
  asToolOutputRecord,
  getGitHubInstallSessionUrl,
  getMutationApproval,
  getMutationMessage,
  getStoryCreationMessage,
  getToolSuggestions,
  isAnalyticsReportOutput,
  isEntityResultToolType,
  isMutationToolPart,
  isRenderableToolPart,
  isStoryCreationToolType,
  isStoryResultToolType,
} from "./tool-output-policy";

const AnalyticsReport = dynamic(
  () => import("./analytics-report").then((module) => module.AnalyticsReport),
  {
    loading: () => (
      <p className="text-text-muted text-sm">Loading workspace report…</p>
    ),
  },
);

const GenerativeOutputFrame = ({ children }: { children: ReactNode }) => (
  <div className="mb-3 w-full max-w-full min-w-0">{children}</div>
);

const MutationApproval = ({
  onToolApproval,
  part,
  statusOverrides,
}: {
  onToolApproval: ChatAddToolApproveResponseFunction;
  part: ToolMessagePart;
  statusOverrides?: StoryResultStatus[];
}) => {
  const approval = getMutationApproval(part);
  const { data: workspaceStatuses = [] } = useStatuses();
  if (!approval) return null;

  const statuses = statusOverrides ?? workspaceStatuses;
  const statusesById = new Map(statuses.map((status) => [status.id, status]));

  if (part.state === "output-denied") {
    return (
      <p className="text-text-muted text-base">{approval.cancelledMessage}</p>
    );
  }

  if (approval.approved !== undefined) {
    let progressMessage = "Cancelling…";
    if (approval.approved) {
      progressMessage = approval.isStoryCreation
        ? "Creating…"
        : "Applying change…";
    }

    return <p className="text-text-muted text-base">{progressMessage}</p>;
  }

  let approvalDetails: ReactNode = null;
  if (approval.storyPreviews.length > 0) {
    approvalDetails = (
      <div className="border-border/70 mt-3 max-h-64 overflow-y-auto border-t dark:border-white/[0.09]">
        {approval.storyPreviews.map((story) => {
          const status = story.statusId
            ? statusesById.get(story.statusId)
            : undefined;

          return (
            <div
              aria-label={story.summary}
              className="border-border/70 flex min-h-11 w-full min-w-0 items-center gap-3 border-b px-px py-2 text-base leading-6 dark:border-white/[0.09]"
              key={story.id}
              title={story.summary}
            >
              <Tooltip title={`Priority: ${story.priority}`}>
                <span className="flex size-5 shrink-0 items-center justify-center">
                  <PriorityIcon
                    className="max-h-4 max-w-4"
                    priority={story.priority}
                  />
                </span>
              </Tooltip>
              <span className="min-w-0 flex-1 truncate">{story.title}</span>
              {status ? (
                <Tooltip title={`Status: ${status.name}`}>
                  <span className="text-text-muted flex max-w-[16ch] min-w-0 shrink-0 items-center gap-1.5">
                    <Dot
                      className="size-2.5 rounded-[2px]"
                      color={status.color}
                    />
                    <span className="truncate">{status.name}</span>
                  </span>
                </Tooltip>
              ) : null}
            </div>
          );
        })}
      </div>
    );
  } else if (approval.details.length > 0) {
    approvalDetails = (
      <dl className="border-border/70 mt-3 max-h-64 overflow-y-auto border-t text-base dark:border-white/[0.09]">
        {approval.details.map((detail) => (
          <div
            className="border-border/70 grid min-h-11 grid-cols-[minmax(0,0.38fr)_minmax(0,0.62fr)] items-center gap-3 border-b px-px py-2 dark:border-white/[0.09]"
            key={`${detail.label}:${detail.value}`}
          >
            <dt className="text-text-muted">{detail.label}</dt>
            <dd className="min-w-0 truncate" title={detail.value}>
              {detail.value}
            </dd>
          </div>
        ))}
      </dl>
    );
  }

  return (
    <div className="w-full min-w-0">
      <p className="text-base font-medium">{approval.prompt}</p>
      <p className="text-text-muted mt-1 text-base">{approval.description}</p>
      {approvalDetails}
      <Flex className="mt-4" gap={2}>
        <Button
          color="tertiary"
          onClick={() => onToolApproval({ approved: false, id: approval.id })}
          size="sm"
        >
          Cancel
        </Button>
        <Button
          onClick={() => onToolApproval({ approved: true, id: approval.id })}
          size="sm"
        >
          Confirm
        </Button>
      </Flex>
    </div>
  );
};

export const ToolOutputRenderer = ({
  onToolApproval,
  onPromptSelect,
  part,
  statusOverrides,
}: {
  onToolApproval: ChatAddToolApproveResponseFunction;
  onPromptSelect: (prompt: string) => void;
  part: ToolMessagePart;
  statusOverrides?: StoryResultStatus[];
}) => {
  if (!isRenderableToolPart(part)) return null;

  if (getMutationApproval(part)) {
    return (
      <GenerativeOutputFrame>
        <MutationApproval
          onToolApproval={onToolApproval}
          part={part}
          statusOverrides={statusOverrides}
        />
      </GenerativeOutputFrame>
    );
  }

  if (isStoryResultToolType(part.type)) {
    return (
      <GenerativeOutputFrame>
        <StoryResults output={part.output} statusOverrides={statusOverrides} />
      </GenerativeOutputFrame>
    );
  }

  if (isStoryCreationToolType(part.type)) {
    return (
      <GenerativeOutputFrame>
        <p className="text-text-muted text-base">
          {getStoryCreationMessage(part.output)}
        </p>
      </GenerativeOutputFrame>
    );
  }

  if (
    part.type === "tool-mayaWorkPlanTool" ||
    part.type === "tool-applyMayaWorkPlanTool"
  ) {
    return (
      <GenerativeOutputFrame>
        <MayaWorkPlanResult output={part.output} />
      </GenerativeOutputFrame>
    );
  }

  if (part.type === "tool-createGitHubInstallSessionTool") {
    const installUrl = getGitHubInstallSessionUrl(part.output);
    return (
      <GenerativeOutputFrame>
        <div className="grid gap-3">
          <p className="text-text-muted text-base">
            {getMutationMessage(part.output)}
          </p>
          {installUrl ? (
            <Button
              onClick={() => {
                window.location.assign(installUrl);
              }}
              size="sm"
              type="button"
            >
              Continue to GitHub
            </Button>
          ) : null}
        </div>
      </GenerativeOutputFrame>
    );
  }

  if (isMutationToolPart(part)) {
    return (
      <GenerativeOutputFrame>
        <p className="text-text-muted text-base">
          {getMutationMessage(part.output)}
        </p>
      </GenerativeOutputFrame>
    );
  }

  if (isEntityResultToolType(part.type)) {
    return (
      <GenerativeOutputFrame>
        <EntityResults output={part.output} toolType={part.type} />
      </GenerativeOutputFrame>
    );
  }

  const output = asToolOutputRecord(part.output);
  const report = isAnalyticsReportOutput(part.output) ? (
    <AnalyticsReport output={output} />
  ) : null;

  if (part.type === "tool-suggestions") {
    return (
      <GenerativeOutputFrame>
        <Flex className="mt-4" gap={2} wrap>
          {getToolSuggestions(part.output).map((suggestion) => (
            <Button
              className="truncate"
              color="tertiary"
              key={suggestion}
              onClick={() => {
                onPromptSelect(suggestion);
              }}
              size="md"
            >
              {suggestion}
            </Button>
          ))}
        </Flex>
      </GenerativeOutputFrame>
    );
  }

  return <GenerativeOutputFrame>{report}</GenerativeOutputFrame>;
};
