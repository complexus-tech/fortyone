"use client";

import type { ReactNode } from "react";
import type { ChatAddToolApproveResponseFunction } from "ai";
import { Button, Flex } from "ui";
import { AnalyticsReport } from "./analytics-report";
import { EntityResults } from "./entity-results";
import { StoryResults } from "./story-results";
import { MayaWorkPlanResult } from "./maya-work-plan-result";
import type { ToolMessagePart } from "./tool-output-policy";
import {
  asToolOutputRecord,
  getMutationMessage,
  getStoryCreationApproval,
  getStoryCreationMessage,
  getToolSuggestions,
  isAnalyticsReportOutput,
  isEntityResultToolType,
  isMutationToolPart,
  isRenderableToolPart,
  isStoryCreationToolType,
  isStoryResultToolType,
} from "./tool-output-policy";

const GenerativeOutputFrame = ({ children }: { children: ReactNode }) => (
  <div className="mb-3 w-full max-w-full min-w-0">{children}</div>
);

const StoryCreationApproval = ({
  onToolApproval,
  part,
}: {
  onToolApproval: ChatAddToolApproveResponseFunction;
  part: ToolMessagePart;
}) => {
  const approval = getStoryCreationApproval(part);
  if (!approval) return null;

  if (part.state === "output-denied") {
    return <p className="text-text-muted text-base">Creation cancelled.</p>;
  }

  if (approval.approved !== undefined) {
    return (
      <p className="text-text-muted text-base">
        {approval.approved ? "Creating…" : "Cancelling…"}
      </p>
    );
  }

  const prompt = approval.title
    ? `Create “${approval.title}”?`
    : `Create ${approval.count} stories?`;

  return (
    <div className="border-border bg-surface-muted rounded-xl border p-4">
      <p className="text-base font-medium">{prompt}</p>
      <p className="text-text-muted mt-1 text-sm">
        Maya will create the prepared{" "}
        {approval.count === 1 ? "story" : "stories"}
        exactly as shown.
      </p>
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
}: {
  onToolApproval: ChatAddToolApproveResponseFunction;
  onPromptSelect: (prompt: string) => void;
  part: ToolMessagePart;
}) => {
  if (!isRenderableToolPart(part)) return null;

  if (getStoryCreationApproval(part)) {
    return (
      <GenerativeOutputFrame>
        <StoryCreationApproval onToolApproval={onToolApproval} part={part} />
      </GenerativeOutputFrame>
    );
  }

  if (isStoryResultToolType(part.type)) {
    return (
      <GenerativeOutputFrame>
        <StoryResults output={part.output} />
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

  if (part.type === "tool-mayaWorkPlanTool") {
    return (
      <GenerativeOutputFrame>
        <MayaWorkPlanResult output={part.output} />
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
