import React from "react";
import { SafeContainer } from "@/components/ui";
import { Header } from "./components/header";
import { useGlobalSearchParams } from "expo-router";
import { useTeamSprints } from "./hooks";
import { List } from "./components/list";

export const Sprints = () => {
  const { teamId } = useGlobalSearchParams<{ teamId: string }>();
  const { data: sprints = [] } = useTeamSprints(teamId);

  return (
    <SafeContainer isFull>
      <Header />
      <List sprints={sprints} />
    </SafeContainer>
  );
};
