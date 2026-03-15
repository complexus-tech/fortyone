"use client";
import { Box, Container, Flex, Text } from "ui";
import { useSession } from "@/lib/auth/client";
import { BodyContainer } from "@/components/shared/body";
import { useTerminology } from "@/hooks";
import { ErrorBoundary } from "@/components/shared";
import { DateRangeFilter } from "@/modules/analytics/components/filters/date-range-filter";
import { Overview } from "./components/overview";
import { Activities } from "./components/activities";
import { MyStories } from "./components/my-stories";
import { Header } from "./components/header";
import { Priority } from "./components/priority";
import { Status } from "./components/status";
import { Contributions } from "./components/contributions";

export const SummaryPage = () => {
  const { getTermDisplay } = useTerminology();
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
        <Container className="@container pt-3 pb-4">
          <Flex className="flex flex-col items-start justify-between gap-3 @3xl:flex-row @3xl:items-center">
            <Box>
              <Text
                as="h2"
                className="mb-1 text-2xl @3xl:text-3xl"
                fontWeight="medium"
              >
                Good {timeOfDay()}, {session?.user?.name}.
              </Text>
              <Text color="muted" fontSize="lg">
                Here&rsquo;s what&rsquo;s happening with your{" "}
                {getTermDisplay("storyTerm", { variant: "plural" })}.
              </Text>
            </Box>
            <DateRangeFilter />
          </Flex>
          <Overview />
          <Box className="my-4 grid grid-cols-1 gap-4 @3xl:grid-cols-2 @7xl:grid-cols-3">
            <ErrorBoundary fallback={<div>Error loading priority</div>}>
              <Priority />
            </ErrorBoundary>
            <ErrorBoundary fallback={<div>Error loading status</div>}>
              <Status />
            </ErrorBoundary>
            <ErrorBoundary fallback={<div>Error loading contributions</div>}>
              <Contributions />
            </ErrorBoundary>
          </Box>
          <Box className="my-4 grid grid-cols-1 gap-4 @5xl:grid-cols-2">
            <ErrorBoundary fallback={<div>Error loading stories</div>}>
              <MyStories />
            </ErrorBoundary>
            <ErrorBoundary fallback={<div>Error loading activities</div>}>
              <Activities />
            </ErrorBoundary>
          </Box>
        </Container>
      </BodyContainer>
    </>
  );
};
