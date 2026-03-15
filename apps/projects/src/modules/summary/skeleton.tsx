"use client";

import { Box, Container, Text } from "ui";
import { useSession } from "@/lib/auth/client";
import { useTerminology } from "@/hooks";
import { BodyContainer } from "@/components/shared/body";
import { Header } from "@/modules/summary/components/header";
import { OverviewSkeleton } from "./components/overview-skeleton";
import { MyStoriesSkeleton } from "./components/my-stories-skeleton";
import { ActivitiesSkeleton } from "./components/activities-skeleton";
import { PrioritySkeleton } from "./components/priority-skeleton";
import { StatusSkeleton } from "./components/status-skeleton";
import { ContributionsSkeleton } from "./components/contributions-skeleton";

export const SummarySkeleton = () => {
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
        <Container className="pt-3 pb-4 @container">
          <Text
            as="h2"
            className="mb-1 text-2xl @6xl:text-3xl"
            fontWeight="medium"
          >
            Good {timeOfDay()}, {session?.user?.name}.
          </Text>
          <Text color="muted" fontSize="lg">
            Here&rsquo;s what&rsquo;s happening with your{" "}
            {getTermDisplay("storyTerm", { variant: "plural" })}.
          </Text>
          <OverviewSkeleton />
          <Box className="my-4 grid grid-cols-1 gap-4 @3xl:grid-cols-2 @7xl:grid-cols-3">
            <PrioritySkeleton />
            <StatusSkeleton />
            <ContributionsSkeleton />
          </Box>
          <Box className="my-4 grid grid-cols-1 gap-4 @6xl:grid-cols-2">
            <MyStoriesSkeleton />
            <ActivitiesSkeleton />
          </Box>
        </Container>
      </BodyContainer>
    </>
  );
};
