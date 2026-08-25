"use client";

import type { ReactNode } from "react";
import { Button, Flex } from "ui";
import { AnalyticsReport } from "./analytics-report";
import { EntityResults } from "./entity-results";
import { StoryResults } from "./story-results";
import { MayaWorkPlanResult } from "./maya-work-plan-result";
import type { ToolMessagePart } from "./tool-output-policy";
import {
  asToolOutputRecord,
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

const GenerativeOutputFrame = ({ children }: { children: ReactNode }) => (
  <div className="mb-3 w-full max-w-full min-w-0">{children}</div>
);

export const ToolOutputRenderer = ({
  onPromptSelect,
  part,
}: {
  onPromptSelect: (prompt: string) => void;
  part: ToolMessagePart;
}) => {
  if (!isRenderableToolPart(part)) return null;

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
