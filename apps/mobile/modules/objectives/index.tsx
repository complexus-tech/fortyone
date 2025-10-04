import React from "react";
import { SafeContainer } from "@/components/ui";
import { Header } from "./components/header";
import { Card } from "./components/card";
import { EmptyState } from "./components/empty-state";
import { useGlobalSearchParams } from "expo-router";
import { useTeamObjectives } from "./hooks";
import { ScrollView } from "react-native";

export const Objectives = () => {
  const { teamId } = useGlobalSearchParams<{ teamId: string }>();
  const { data: objectives = [] } = useTeamObjectives(teamId);

  if (objectives.length === 0) {
    return (
      <SafeContainer isFull>
        <Header />
        <EmptyState />
      </SafeContainer>
    );
  }

  return (
    <SafeContainer isFull>
      <Header />
      <ScrollView showsVerticalScrollIndicator={false}>
        {objectives.map((objective) => (
          <Card key={objective.id} objective={objective} />
        ))}
      </ScrollView>
    </SafeContainer>
  );
};
