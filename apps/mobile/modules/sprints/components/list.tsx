import type { Sprint } from "../types";
import { FlatList } from "react-native";
import { Card } from "./card";
import { EmptyState } from "./empty-state";

export const List = ({
  sprints,
  refreshing,
  onRefresh,
}: {
  sprints: Sprint[];
  refreshing: boolean;
  onRefresh: () => void;
}) => (
  <FlatList
    data={sprints}
    keyExtractor={(item) => item.id}
    renderItem={({ item }) => <Card sprint={item} />}
    ListEmptyComponent={<EmptyState />}
    contentContainerStyle={{ flexGrow: 1, paddingBottom: 32 }}
    showsVerticalScrollIndicator={false}
    initialNumToRender={12}
    refreshing={refreshing}
    onRefresh={onRefresh}
  />
);
