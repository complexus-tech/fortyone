"use client";
import { Box, Container, Text } from "ui";
import { Suspense } from "react";
import { useSession } from "next-auth/react";
import { BodyContainer } from "@/components/shared/body";
import { useTerminology } from "@/hooks";
import { Overview } from "./components/overview";
import { Activities } from "./components/activities";
import { MyStories } from "./components/my-stories";
import { Contributions } from "./components/contributions";
import { Header } from "./components/header";
import { MyStoriesSkeleton } from "./components/my-stories-skeleton";
import { OverviewSkeleton } from "./components/overview-skeleton";
import { ContributionsSkeleton } from "./components/contributions-skeleton";
import { ActivitiesSkeleton } from "./components/activities-skeleton";

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
        <Container className="pb-4 pt-3">
          <Text as="h2" className="mb-2" fontSize="3xl" fontWeight="medium">
            Good {timeOfDay()}, {session?.user?.name}.
          </Text>
          <Text color="muted" fontSize="lg">
            Here&rsquo;s what&rsquo;s happening with your{" "}
            {getTermDisplay("storyTerm", { variant: "plural" })}.
          </Text>
          <Suspense fallback={<OverviewSkeleton />}>
            <Overview />
          </Suspense>
          <Suspense fallback={<ContributionsSkeleton />}>
            <Contributions />
          </Suspense>
          <Box className="my-4 grid min-h-[30rem] grid-cols-2 gap-4">
            <Suspense fallback={<MyStoriesSkeleton />}>
              <MyStories />
            </Suspense>
            <Suspense fallback={<ActivitiesSkeleton />}>
              <Activities />
            </Suspense>
          </Box>
        </Container>
      </BodyContainer>
    </>
  );
};
