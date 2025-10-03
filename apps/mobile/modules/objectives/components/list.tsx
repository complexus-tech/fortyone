import React from "react";
import { SafeContainer } from "@/components/ui";
import { Header } from "./header";
import { useTeamObjectives } from "../hooks/use-objectives";
import { useGlobalSearchParams } from "expo-router";
import { Card } from "./card";

export const List = () => {
  const { teamId } = useGlobalSearchParams<{ teamId: string }>();
  const { data: objectives = [], isPending } = useTeamObjectives(teamId);

  if (isPending) {
    return (
      <SafeContainer isFull>
        <Header />
      </SafeContainer>
    );
  }

  return (
    <SafeContainer isFull>
      <Header />
      {objectives.map((objective) => (
        <Card key={objective.id} objective={objective} />
      ))}
    </SafeContainer>
  );
};
