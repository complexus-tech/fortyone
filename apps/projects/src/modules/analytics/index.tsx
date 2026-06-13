"use client";

import { Box, Container, Flex, Tabs, Text } from "ui";
import { AnalyticsIcon, HealthIcon } from "icons";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import { useSession } from "@/lib/auth/client";
import { BodyContainer } from "@/components/shared/body";
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
import { PulseReportPanel } from "./components/pulse-report";

const tabs = ["overview", "pulse"] as const;

const getTimeOfDay = () => {
  const hour = new Date().getHours();
  if (hour < 12) return "morning";
  if (hour < 18) return "afternoon";
  return "evening";
};

export const AnalyticsPage = () => {
  const { data: session } = useSession();
  const [tab, setTab] = useQueryState(
    "tab",
    parseAsStringLiteral(tabs).withDefault("overview"),
  );

  return (
    <>
      <Header />
      <BodyContainer>
        <Container className="@container pt-3 pb-4">
          <Tabs
            onValueChange={(value) => setTab(value as (typeof tabs)[number])}
            value={tab}
          >
            <Tabs.List className="mx-0 mb-4 md:mx-0">
              <Tabs.Tab leftIcon={<AnalyticsIcon />} value="overview">
                Workspace analytics
              </Tabs.Tab>
              <Tabs.Tab leftIcon={<HealthIcon />} value="pulse">
                Workspace pulse
              </Tabs.Tab>
            </Tabs.List>
            <Tabs.Panel value="overview">
              <Flex className="flex flex-col items-start justify-between gap-3 @3xl:flex-row @3xl:items-end">
                <Box>
                  <Text
                    as="h2"
                    className="mb-1 text-2xl @3xl:text-3xl"
                    fontWeight="medium"
                  >
                    Good {getTimeOfDay()}, {session?.user.name}.
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
                <ErrorBoundary
                  fallback={<div>Error loading completion trend</div>}
                >
                  <CompletionTrend />
                </ErrorBoundary>
                <ErrorBoundary
                  fallback={<div>Error loading velocity trend</div>}
                >
                  <VelocityTrend />
                </ErrorBoundary>
                <ErrorBoundary
                  fallback={<div>Error loading team velocity</div>}
                >
                  <TeamVelocity />
                </ErrorBoundary>
              </Box>

              {/* Stories & Work Analysis */}
              <Box className="my-4 grid gap-4 md:grid-cols-3">
                <ErrorBoundary
                  fallback={<div>Error loading status breakdown</div>}
                >
                  <StatusBreakdown />
                </ErrorBoundary>
                <ErrorBoundary
                  fallback={<div>Error loading priority distribution</div>}
                >
                  <PriorityDistribution />
                </ErrorBoundary>
                <ErrorBoundary
                  fallback={<div>Error loading team allocation</div>}
                >
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
                <ErrorBoundary
                  fallback={<div>Error loading objective health</div>}
                >
                  <ObjectiveHealth />
                </ErrorBoundary>
                <ErrorBoundary
                  fallback={<div>Error loading sprint health</div>}
                >
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
            </Tabs.Panel>
            <Tabs.Panel value="pulse">
              {tab === "pulse" ? (
                <ErrorBoundary
                  fallback={<div>Error loading workspace pulse</div>}
                >
                  <PulseReportPanel />
                </ErrorBoundary>
              ) : null}
            </Tabs.Panel>
          </Tabs>
        </Container>
      </BodyContainer>
    </>
  );
};
