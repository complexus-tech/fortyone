"use client";

import { useState } from "react";
import { Box, Container, Tabs } from "ui";
import { BodyContainer } from "@/components/shared/body";
import { ErrorBoundary } from "@/components/shared";
import { CommandCenterReport } from "./components/command-center-report";
import { Header } from "./components/header";

export const AnalyticsPage = () => {
  const [activeTab, setActiveTab] = useState("overview");
  return (
    <Tabs
      className="flex h-full min-h-0 flex-col"
      onValueChange={setActiveTab}
      value={activeTab}
    >
      <Header showFilters={activeTab !== "custom-fields"} />
      <BodyContainer>
        <Container className="@container min-w-0 pt-6 pb-4">
          <ErrorBoundary fallback={<Box>Error loading analytics</Box>}>
            <CommandCenterReport />
          </ErrorBoundary>
        </Container>
      </BodyContainer>
    </Tabs>
  );
};
