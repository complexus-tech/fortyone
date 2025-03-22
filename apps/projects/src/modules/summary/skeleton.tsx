"use client";

import { Box, Container, Text } from "ui";
import { useSession } from "next-auth/react";
import { useTerminology } from "@/hooks";
import { BodyContainer } from "@/components/shared/body";
import { Header } from "@/modules/summary/components/header";
import { OverviewSkeleton } from "./components/overview-skeleton";
import { ContributionsSkeleton } from "./components/contributions-skeleton";
import { MyStoriesSkeleton } from "./components/my-stories-skeleton";
import { ActivitiesSkeleton } from "./components/activities-skeleton";

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
        <Container className="pb-4 pt-3">
          <Text as="h2" className="mb-2" fontSize="3xl" fontWeight="medium">
            Good {timeOfDay()}, {session?.user?.name}.
          </Text>
          <Text color="muted" fontSize="lg">
            Here&rsquo;s what&rsquo;s happening with your{" "}
            {getTermDisplay("storyTerm", { variant: "plural" })}.
          </Text>
          <OverviewSkeleton />
          <ContributionsSkeleton />
          <Box className="my-4 grid min-h-[30rem] grid-cols-2 gap-4">
            <MyStoriesSkeleton />
            <ActivitiesSkeleton />
          </Box>
        </Container>
      </BodyContainer>
    </>
  );
};
