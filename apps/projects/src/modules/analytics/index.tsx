"use client";

import { Box, Container } from "ui";
import { BodyContainer } from "@/components/shared/body";
import { ErrorBoundary } from "@/components/shared";
import { CommandCenterReport } from "./components/command-center-report";
import { Header } from "./components/header";

export const AnalyticsPage = () => {
  return (
    <>
      <Header />
      <BodyContainer>
        <Container className="@container pt-3 pb-4">
          <ErrorBoundary fallback={<Box>Error loading analytics</Box>}>
            <CommandCenterReport />
          </ErrorBoundary>
        </Container>
      </BodyContainer>
    </>
  );
};
