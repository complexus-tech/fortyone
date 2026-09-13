import type { GroupedStoriesResponse, GroupStoryParams, Story } from "../types";
import type { DisplayColumn } from "@/types/stories-view-options";
import { useCallback, useMemo, useState } from "react";
import { SectionList, View } from "react-native";
import { useQueries } from "@tanstack/react-query";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { Button, StoriesListSkeleton, Text } from "@/components/ui";
import { useStatuses } from "@/modules/statuses";
import { useMembers } from "@/modules/members";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { useTerminology } from "@/hooks/use-terminology";
import { storyKeys } from "@/constants/keys";
import { getGroupStories } from "../queries/get-group-stories";
import { mergeStoryPages } from "../utils/pages";
import { StoryRow } from "./story-row";
import { EmptyState } from "./empty-state";
import { SectionFooter } from "./section-footer";

type StoriesBoardProps = {
  groupedStories?: GroupedStoriesResponse | null;
  groupFilters: Omit<GroupStoryParams, "groupKey">;
  isLoading?: boolean;
  error?: Error | null;
  onRetry?: () => void;
  emptyTitle?: string;
  emptyMessage?: string;
  visibleColumns: DisplayColumn[];
  onRefresh?: () => void;
  isRefreshing?: boolean;
};
type PageRequest = { groupKey: string; page: number };

export const StoriesBoard = (props: StoriesBoardProps) => (
  <GroupedStoriesList key={JSON.stringify(props.groupFilters)} {...props} />
);

function GroupedStoriesList({
  groupedStories,
  groupFilters,
  isLoading = false,
  error,
  onRetry,
  emptyTitle,
  emptyMessage,
  visibleColumns,
  onRefresh,
  isRefreshing = false,
}: StoriesBoardProps) {
  const { data: statuses = [] } = useStatuses();
  const { data: members = [] } = useMembers();
  const { data: teams = [] } = useTeams();
  const { getTermDisplay } = useTerminology();
  const { bottom } = useSafeAreaInsets();
  const [requestedPages, setRequestedPages] = useState<PageRequest[]>([]);
  const groups = groupedStories?.groups ?? [];
  const requests = requestedPages.filter((request) =>
    groups.some((group) => group.key === request.groupKey),
  );
  const pageQueries = useQueries({
    queries: requests.map(({ groupKey, page }) => ({
      queryKey: storyKeys.group(groupKey, { ...groupFilters, page }),
      queryFn: ({ signal }: { signal: AbortSignal }) =>
        getGroupStories({ ...groupFilters, groupKey, page }, signal),
    })),
  });
  const statusById = useMemo(
    () => new Map(statuses.map((status) => [status.id, status])),
    [statuses],
  );
  const memberById = useMemo(
    () => new Map(members.map((member) => [member.id, member])),
    [members],
  );
  const teamById = useMemo(
    () => new Map(teams.map((team) => [team.id, team])),
    [teams],
  );

  const sections = groups.map((group) => {
    const queries = requests.flatMap((request, index) =>
      request.groupKey === group.key ? [pageQueries[index]] : [],
    );
    const pages = queries.flatMap((query) => (query.data ? [query.data] : []));
    const lastPage = pages.at(-1);
    const failed = queries.find((query) => query.isError);
    const isFetchingPage = queries.some(
      (query) => query.isPending || query.isFetching,
    );
    const nextPage = lastPage?.pagination.nextPage ?? group.nextPage;
    const hasMore = lastPage?.pagination.hasMore ?? group.hasMore;
    const data = mergeStoryPages([
      group.stories,
      ...pages.map((page) => page.stories),
    ]);
    const title =
      groupFilters.groupBy === "status"
        ? statusById.get(group.key)?.name ?? "No status"
        : groupFilters.groupBy === "assignee"
          ? memberById.get(group.key)?.fullName ?? "Unassigned"
          : groupFilters.groupBy === "priority"
            ? group.key
            : getTermDisplay("storyTerm", { variant: "plural" });
    const loadMore = () => {
      if (failed) {
        void failed.refetch();
        return;
      }
      if (isFetchingPage || !hasMore || !nextPage) return;
      setRequestedPages((previous) =>
        previous.some(
          (page) => page.groupKey === group.key && page.page === nextPage,
        )
          ? previous
          : [...previous, { groupKey: group.key, page: nextPage }],
      );
    };
    return {
      ...group,
      title,
      data,
      hasMore,
      isFetchingPage,
      failed: Boolean(failed),
      loadMore,
    };
  });

  const renderItem = useCallback(
    ({ item }: { item: Story }) => (
      <StoryRow
        story={item}
        visibleColumns={visibleColumns}
        status={statusById.get(item.statusId)}
        assignee={memberById.get(item.assigneeId ?? "")}
        team={teamById.get(item.teamId)}
      />
    ),
    [visibleColumns, statusById, memberById, teamById],
  );

  if (isLoading && !groupedStories) return <StoriesListSkeleton />;
  if (error && !groupedStories) {
    return (
      <View className="flex-1 items-center justify-center p-6 gap-4">
        <Text fontWeight="semibold">Couldn’t load your work</Text>
        <Text color="muted" align="center">
          Check your connection and try again.
        </Text>
        <Button onPress={onRetry ?? onRefresh} color="tertiary">
          Try again
        </Button>
      </View>
    );
  }
  return (
    <SectionList
      sections={sections}
      renderItem={renderItem}
      keyExtractor={(story) => story.id}
      stickySectionHeadersEnabled={false}
      initialNumToRender={15}
      maxToRenderPerBatch={12}
      windowSize={7}
      onRefresh={onRefresh}
      refreshing={isRefreshing}
      keyboardShouldPersistTaps="handled"
      contentContainerStyle={{ paddingBottom: bottom + 100, flexGrow: 1 }}
      ListHeaderComponent={
        error ? (
          <View className="p-4">
            <Text color="muted">
              Showing saved work. Refresh failed; pull down to retry.
            </Text>
          </View>
        ) : null
      }
      ListEmptyComponent={
        <EmptyState
          title={
            emptyTitle ||
            `No ${getTermDisplay("storyTerm", { variant: "plural" })} found`
          }
          message={emptyMessage || "Your work will appear here."}
        />
      }
      renderSectionHeader={({ section }) => (
        <Text
          accessibilityRole="header"
          className="px-4.5 pt-3 pb-1"
          fontSize="sm"
          fontWeight="semibold"
          color="muted"
        >
          {section.title}
        </Text>
      )}
      renderSectionFooter={({ section }) =>
        section.failed ? (
          <View className="px-4 pb-3">
            <Button color="tertiary" onPress={section.loadMore}>
              Retry loading more
            </Button>
          </View>
        ) : (
          <SectionFooter
            hasMore={section.hasMore}
            loadedCount={section.data.length}
            totalCount={section.totalCount}
            isLoading={section.isFetchingPage}
            onLoadMore={section.loadMore}
          />
        )
      }
      progressViewOffset={8}
      style={{ flex: 1 }}
    />
  );
}
