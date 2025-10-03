import React from "react";
import { SafeContainer } from "@/components/ui";
import { Header } from "./components/header";
import { useGlobalSearchParams } from "expo-router";
import { useTeamSprints } from "./hooks";
import { Card } from "./components/card";

export const Sprints = () => {
  const { teamId } = useGlobalSearchParams<{ teamId: string }>();
  const { data: sprints = [] } = useTeamSprints(teamId);
  return (
    <SafeContainer isFull>
      <Header />
      {sprints.map((sprint) => (
        <Card key={sprint.id} sprint={sprint} />
      ))}
    </SafeContainer>
  );
};
