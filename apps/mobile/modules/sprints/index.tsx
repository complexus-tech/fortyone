import React from "react";
import { SafeContainer } from "@/components/ui";
import { Header } from "./components/header";
import { Card } from "./components/card";
import { EmptyState } from "./components/empty-state";
import { useGlobalSearchParams } from "expo-router";
import { useTeamSprints } from "./hooks";
import { ScrollView } from "react-native";

export const Sprints = () => {
  const { teamId } = useGlobalSearchParams<{ teamId: string }>();
  const { data: sprints = [] } = useTeamSprints(teamId);

  if (sprints.length === 0) {
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
        {sprints.map((sprint) => (
          <Card key={sprint.id} sprint={sprint} />
        ))}
      </ScrollView>
    </SafeContainer>
  );
};
