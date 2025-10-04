import React from "react";
import { SafeContainer } from "@/components/ui";
import { Header } from "./components/header";
import { useGlobalSearchParams } from "expo-router";
import { useTeamObjectives } from "./hooks";
import { List } from "./components/list";

export const Objectives = () => {
  const { teamId } = useGlobalSearchParams<{ teamId: string }>();
  const { data: objectives = [] } = useTeamObjectives(teamId);

  return (
    <SafeContainer isFull>
      <Header />
      <List objectives={objectives} />
    </SafeContainer>
  );
};
