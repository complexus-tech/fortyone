import type { Objective } from "../types";
import { useMemo } from "react";
import { SectionList } from "react-native";
import { QueryState } from "@/components/ui/query-state";
import { GroupHeader } from "@/modules/stories/components/group-header";
import { useObjectiveStatuses } from "../hooks/use-objectives";
import { groupObjectivesByStatus } from "../group-by-status";
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
}) => {
  const statuses = useObjectiveStatuses();
  const sections = useMemo(
    () => groupObjectivesByStatus(objectives, statuses.data ?? []),
    [objectives, statuses.data],
  );
  if (statuses.isPending || statuses.error) {
    return (
      <QueryState
        loading={statuses.isPending}
        title={
          statuses.isPending ? "Loading statuses" : "Could not load statuses"
        }
        message={statuses.error?.message}
        onRetry={
          statuses.error
            ? () => {
                void statuses.refetch();
              }
            : undefined
        }
      />
    );
  }
  return (
    <SectionList
      sections={sections}
      keyExtractor={(item) => item.id}
      renderItem={({ item, section }) => (
        <Card objective={item} status={section.status} />
      )}
      renderSectionHeader={({ section }) => (
        <GroupHeader title={section.title} count={section.data.length} />
      )}
      stickySectionHeadersEnabled={false}
      ListEmptyComponent={<EmptyState />}
      contentContainerStyle={{ flexGrow: 1, paddingBottom: 32 }}
      showsVerticalScrollIndicator={false}
      initialNumToRender={12}
      refreshing={refreshing || statuses.isRefetching}
      onRefresh={() => {
        onRefresh();
        void statuses.refetch();
      }}
    />
  );
};
