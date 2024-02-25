"use client";
import { Box, Container } from "ui";
import { BodyContainer } from "@/components/layout";
import {
  Activities,
  Contributions,
  Header,
  MyIssues,
  Overview,
} from "@/components/(dashboard)";

export default function Page(): JSX.Element {
  return (
    <>
      <Header />
      <BodyContainer>
        <Container className="py-4">
          <Overview />
          <Contributions />
          <Box className="my-4 grid min-h-[30rem] grid-cols-2 gap-4">
            <MyIssues />
            <Activities />
          </Box>
        </Container>
      </BodyContainer>
    </>
  );
}
