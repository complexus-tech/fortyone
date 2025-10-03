import React from "react";
import { Sprint } from "../types";
import { Text } from "@/components/ui";

export const Card = ({ sprint }: { sprint: Sprint }) => {
  return <Text>{sprint.name}</Text>;
};
