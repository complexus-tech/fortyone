import React from "react";
import { Card } from "./card";
import { ScrollView } from "react-native";
import { Objective } from "../types";
import { EmptyState } from "./empty-state";

export const List = ({ objectives }: { objectives: Objective[] }) => {
  return (
    <ScrollView showsVerticalScrollIndicator={false}>
      {objectives.length === 0 && <EmptyState />}
      {objectives.map((objective) => (
        <Card key={objective.id} objective={objective} />
      ))}
    </ScrollView>
  );
};
