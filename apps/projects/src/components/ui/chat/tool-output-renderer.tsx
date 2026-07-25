"use client";

import { Box, Button, Flex, Text } from "ui";
import { cn } from "lib";
import { usePathname } from "next/navigation";
import { BurndownChart } from "@/modules/sprints/stories/burndown";
import { AnalyticsReport } from "./analytics-report";
import { StoryResults } from "./story-results";
import type { ToolMessagePart } from "./tool-output-policy";
import {
  asToolOutputRecord,
  getToolSuggestions,
  isAnalyticsReportOutput,
  isRenderableToolPart,
  isStoryResultToolType,
} from "./tool-output-policy";

const getSprintBurndownData = (output: unknown) => {
  const outputRecord = asToolOutputRecord(output);
  const analyticsReport = asToolOutputRecord(outputRecord.analyticsReport);
  return Array.isArray(analyticsReport.burndown)
    ? analyticsReport.burndown
    : [];
};

export const ToolOutputRenderer = ({
  onPromptSelect,
  part,
}: {
  onPromptSelect: (prompt: string) => void;
  part: ToolMessagePart;
}) => {
  const pathname = usePathname();

  if (!isRenderableToolPart(part)) return null;

  if (isStoryResultToolType(part.type)) {
    return <StoryResults output={part.output} />;
  }

  const output = asToolOutputRecord(part.output);
  const report = isAnalyticsReportOutput(part.output) ? (
    <AnalyticsReport output={output} />
  ) : null;
  const sprintAnalytics =
    part.type === "tool-getSprintAnalyticsTool" ? (
      <Box className="mb-3">
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
    ) : null;

  if (part.type === "tool-suggestions") {
    return (
      <Flex className="mt-4" gap={2} wrap>
        {getToolSuggestions(part.output).map((suggestion) => (
          <Button
            className="truncate"
            color="tertiary"
            key={suggestion}
            onClick={() => {
              onPromptSelect(suggestion);
            }}
            size="sm"
          >
            {suggestion}
          </Button>
        ))}
      </Flex>
    );
  }

  return (
    <>
      {report}
      {sprintAnalytics}
    </>
  );
};
