import React from "react";
import { SafeContainer } from "@/components/ui";
import { Header } from "@/modules/sprints/components/header";
import { Card } from "@/modules/sprints/components/card";
import { ScrollView } from "react-native";
import { Sprint } from "@/modules/sprints/types";
import { EmptyState } from "./empty-state";

export const List = ({ sprints }: { sprints: Sprint[] }) => {
  return (
    <SafeContainer isFull>
      <Header />
      <ScrollView
        showsVerticalScrollIndicator={false}
        contentContainerStyle={{ paddingBottom: 100 }}
      >
        {sprints.length === 0 && <EmptyState />}
        {sprints.map((sprint) => (
          <Card key={sprint.id} sprint={sprint} />
        ))}
      </ScrollView>
    </SafeContainer>
  );
};
