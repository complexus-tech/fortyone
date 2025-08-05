"use client";

import { Box, Container, Flex, Text } from "ui";
import { useSession } from "next-auth/react";
import { BodyContainer } from "@/components/shared/body";
// import { useTerminology } from "@/hooks";
import { ErrorBoundary } from "@/components/shared";
import { Header } from "./components/header";
import { Overview } from "./components/overview";
import { CompletionTrend } from "./components/completion-trend";
import { VelocityTrend } from "./components/velocity-trend";
import { StatusBreakdown } from "./components/status-breakdown";
import { PriorityDistribution } from "./components/priority-distribution";
// import { BurndownChart } from "./components/burndown-chart";
// import { TeamWorkload } from "./components/team-workload";
import { TeamVelocity } from "./components/team-velocity";
// import { MemberContributions } from "./components/member-contributions";
import { ObjectiveHealth } from "./components/objective-health";
// import { KeyResultsProgress } from "./components/key-results-progress";
import { SprintHealth } from "./components/sprint-health";
import { TeamAllocation } from "./components/team-allocation";
// import { TimelineTrends } from "./components/timeline-trends";
import { Filters } from "./components/filters";

export const AnalyticsPage = () => {
  // const { getTermDisplay } = useTerminology();
  const { data: session } = useSession();

  const timeOfDay = () => {
    const hour = new Date().getHours();
    if (hour < 12) return "morning";
    if (hour < 18) return "afternoon";
    return "evening";
  };

  return (
    <>
      <Header />
      <BodyContainer>
        <Container className="pb-4 pt-3">
          <Flex align="end" justify="between">
            <Box>
              <Text
                as="h2"
                className="mb-1 text-2xl md:text-3xl"
                fontWeight="medium"
              >
                Good {timeOfDay()}, {session?.user?.name}.
              </Text>
              <Text color="muted" fontSize="lg">
                Here&rsquo;s your workspace analytics and insights.
              </Text>
            </Box>
            <Filters />
          </Flex>
          <ErrorBoundary fallback={<div>Error loading overview</div>}>
            <Overview />
          </ErrorBoundary>
          <Box className="my-4 grid gap-4 md:grid-cols-3">
            <ErrorBoundary fallback={<div>Error loading completion trend</div>}>
              <CompletionTrend />
            </ErrorBoundary>
            <ErrorBoundary fallback={<div>Error loading velocity trend</div>}>
              <VelocityTrend />
            </ErrorBoundary>
            <ErrorBoundary fallback={<div>Error loading team velocity</div>}>
              <TeamVelocity />
            </ErrorBoundary>
          </Box>

          {/* Stories & Work Analysis */}
          <Box className="my-4 grid gap-4 md:grid-cols-3">
            <ErrorBoundary fallback={<div>Error loading status breakdown</div>}>
              <StatusBreakdown />
            </ErrorBoundary>
            <ErrorBoundary
              fallback={<div>Error loading priority distribution</div>}
            >
              <PriorityDistribution />
            </ErrorBoundary>
            <ErrorBoundary fallback={<div>Error loading team allocation</div>}>
              <TeamAllocation />
            </ErrorBoundary>
          </Box>

          {/* Key Results Progress */}
          {/* <ErrorBoundary
            fallback={<div>Error loading key results progress</div>}
          >
            <KeyResultsProgress startDate={dateRange.startDate} endDate={dateRange.endDate} />
          </ErrorBoundary> */}

          {/* Health & Allocation Analytics */}
          <Box className="my-4 grid gap-4 md:grid-cols-3">
            <ErrorBoundary fallback={<div>Error loading objective health</div>}>
              <ObjectiveHealth />
            </ErrorBoundary>
            <ErrorBoundary fallback={<div>Error loading sprint health</div>}>
              <SprintHealth />
            </ErrorBoundary>
          </Box>

          {/* Team Performance & Contributions */}
          {/* <Box className="my-4 grid gap-4 md:grid-cols-3">
            <ErrorBoundary
              fallback={<div>Error loading member contributions</div>}
            >
              <MemberContributions startDate={dateRange.startDate} endDate={dateRange.endDate} />
            </ErrorBoundary>
            <ErrorBoundary fallback={<div>Error loading burndown chart</div>}>
              <Box className="md:col-span-2">
                <BurndownChart startDate={dateRange.startDate} endDate={dateRange.endDate} />
              </Box>
            </ErrorBoundary>
          </Box> */}
          {/* 
          <Box className="my-4 grid gap-4 md:grid-cols-2">
            <ErrorBoundary fallback={<div>Error loading team workload</div>}>
              <TeamWorkload startDate={dateRange.startDate} endDate={dateRange.endDate} />
            </ErrorBoundary>
            <ErrorBoundary fallback={<div>Error loading timeline trends</div>}>
              <TimelineTrends startDate={dateRange.startDate} endDate={dateRange.endDate} />
            </ErrorBoundary>
          </Box> */}
        </Container>
      </BodyContainer>
    </>
  );
};
