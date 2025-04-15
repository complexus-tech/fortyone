"use client";
import { Box, Container, Text } from "ui";
import { useSession } from "next-auth/react";
import { BodyContainer } from "@/components/shared/body";
import { useTerminology } from "@/hooks";
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
        <Container className="pb-4 pt-3">
          <Text
            as="h2"
            className="mb-1 text-2xl md:mb-2 md:text-3xl"
            fontWeight="medium"
          >
            Good {timeOfDay()}, {session?.user?.name}.
          </Text>
          <Text color="muted" fontSize="lg">
            Here&rsquo;s what&rsquo;s happening with your{" "}
            {getTermDisplay("storyTerm", { variant: "plural" })}.
          </Text>
          <Overview />
          <Box className="my-4 grid gap-4 md:grid-cols-3">
            <Priority />
            <Status />
            <Contributions />
          </Box>
          <Box className="my-4 grid gap-4 md:grid-cols-2">
            <MyStories />
            <Activities />
          </Box>
        </Container>
      </BodyContainer>
    </>
  );
};
