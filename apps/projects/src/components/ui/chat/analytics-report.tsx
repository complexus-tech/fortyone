"use client";

import { useTerminology } from "@/hooks/use-terminology-display";
import { WorkspaceCommandCenterReport } from "./analytics-report/command-center-report";
import {
  GitHubIntegrationReport,
  GitHubStoryReport,
  GitHubTeamAutomationReport,
} from "./analytics-report/github-reports";
import type { AnalyticsReportOutput } from "./analytics-report/model";
import {
  ObjectiveProgressReport,
  PulseReport,
  SprintPerformanceReport,
  StoryPerformanceReport,
  TeamPerformanceReport,
  TimelineTrendsReport,
  WorkloadAnalysisReport,
  WorkspacePerformanceReport,
} from "./analytics-report/performance-reports";
import { SingleSprintAnalyticsReport } from "./analytics-report/sprint-report";

export const AnalyticsReport = ({
  output,
}: {
  output: AnalyticsReportOutput;
}) => {
  const { getTermDisplay } = useTerminology();
  const storyTerm = getTermDisplay("storyTerm");
  const storyTermCapitalized = getTermDisplay("storyTerm", {
    capitalize: true,
  });
  const storyTermPlural = getTermDisplay("storyTerm", {
    capitalize: true,
    variant: "plural",
  });
  const kind = output.kind;
  const title = String(
    output.title ??
      (kind === "workload-analysis-report"
        ? "Workload analysis"
        : "Performance report"),
  );

  if (kind === "github-integration-report") {
    return <GitHubIntegrationReport output={output} title={title} />;
  }

  if (kind === "github-team-automation-report") {
    return <GitHubTeamAutomationReport output={output} title={title} />;
  }

  if (kind === "github-story-report") {
    return (
      <GitHubStoryReport output={output} storyTerm={storyTerm} title={title} />
    );
  }

  if (kind === "workspace-performance-report") {
    return (
      <WorkspacePerformanceReport
        output={output}
        storyTermPlural={storyTermPlural}
        title={title}
      />
    );
  }

  if (kind === "workspace-command-center-report") {
    return <WorkspaceCommandCenterReport output={output} title={title} />;
  }

  if (kind === "pulse-report") {
    return <PulseReport output={output} title={title} />;
  }

  if (kind === "workload-analysis-report") {
    return <WorkloadAnalysisReport output={output} title={title} />;
  }

  if (kind === "story-performance-report") {
    return <StoryPerformanceReport output={output} title={title} />;
  }

  if (kind === "objective-progress-report") {
    return <ObjectiveProgressReport output={output} title={title} />;
  }

  if (kind === "team-performance-report") {
    return <TeamPerformanceReport output={output} title={title} />;
  }

  if (kind === "sprint-performance-report") {
    return <SprintPerformanceReport output={output} title={title} />;
  }

  if (kind === "single-sprint-analytics-report") {
    return <SingleSprintAnalyticsReport output={output} title={title} />;
  }

  if (kind === "timeline-trends-report") {
    return (
      <TimelineTrendsReport
        output={output}
        storyTermCapitalized={storyTermCapitalized}
        title={title}
      />
    );
  }

  return null;
};
