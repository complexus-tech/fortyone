import { Box, Container } from "ui";
import { Suspense } from "react";
import { BodyContainer } from "@/components/shared/body";
import { Overview } from "./components/overview";
import { Activities } from "./components/activities";
import { MyStories } from "./components/my-stories";
import { Contributions } from "./components/contributions";
import { Header } from "./components/header";

export const SummaryPage = () => {
  return (
    <>
      <Header />
      <BodyContainer>
        <Container className="pb-4 pt-3">
          <Overview />
          <Contributions />
          <Box className="my-4 grid min-h-[30rem] grid-cols-2 gap-4">
            <Suspense fallback={<div>Loading...</div>}>
              <MyStories />
            </Suspense>
            <Activities />
          </Box>
        </Container>
      </BodyContainer>
    </>
  );
};
