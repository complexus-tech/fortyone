import type { Objective } from "../types";
import { FlatList } from "react-native";
import { Card } from "./card";
import { EmptyState } from "./empty-state";

export const List = ({
  objectives,
  refreshing,
  onRefresh,
}: {
  objectives: Objective[];
  refreshing: boolean;
  onRefresh: () => void;
}) => (
  <FlatList
    data={objectives}
    keyExtractor={(item) => item.id}
    renderItem={({ item }) => <Card objective={item} />}
    ListEmptyComponent={<EmptyState />}
    contentContainerStyle={{ flexGrow: 1, paddingBottom: 32 }}
    showsVerticalScrollIndicator={false}
    initialNumToRender={12}
    refreshing={refreshing}
    onRefresh={onRefresh}
  />
);
