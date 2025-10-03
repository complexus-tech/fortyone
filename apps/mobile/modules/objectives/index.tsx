import React from "react";
import { SafeContainer } from "@/components/ui";
import { Header } from "./components/header";
import { useGlobalSearchParams } from "expo-router";
import { useTeamObjectives } from "./hooks";
import { Card } from "./components/card";

export const Objectives = () => {
  const { teamId } = useGlobalSearchParams<{ teamId: string }>();
  const { data: objectives = [] } = useTeamObjectives(teamId);
  return (
    <SafeContainer isFull>
      <Header />
      {objectives.map((objective) => (
        <Card key={objective.id} objective={objective} />
      ))}
    </SafeContainer>
  );
};
