import type { SearchResponse } from "../types";
import type { Story } from "@/modules/stories/types";
import type { Objective } from "@/modules/objectives/types";
import { DEFAULT_STORY_DISPLAY_COLUMNS } from "@/types/stories-view-options";
import { memo, useCallback, useMemo } from "react";
import { FlatList, Pressable, View } from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useRouter } from "expo-router";
import { Button, Col, Row, Text } from "@/components/ui";
import { QueryState } from "@/components/ui/query-state";
import { themeColors } from "@/constants/colors";
import { useTheme, useTerminology } from "@/hooks";
import { useStatuses } from "@/modules/statuses";
import { useMembers } from "@/modules/members";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { StoryRow } from "@/modules/stories/components/story-row";
import { getCompletionStatus } from "@/modules/stories/utils/quick-actions";
import { teamStoriesHref } from "@/modules/teams/stories/team-story-navigation";

type SearchResultsProps = {
  results: SearchResponse;
  type: "stories" | "objectives";
  query: string;
  hasMore: boolean;
  loadingMore: boolean;
  error: Error | null;
  loadMoreError: boolean;
  onLoadMore: () => void;
  onRetry: () => void;
};
type SearchResult =
  | { type: "story"; item: Story }
  | { type: "objective"; item: Objective };
const resultKey = (result: SearchResult) => `${result.type}:${result.item.id}`;

const ObjectiveResult = memo(function ObjectiveResult({
  objective,
  teamName,
}: {
  objective: Objective;
  teamName?: string;
}) {
  const router = useRouter();
  const { resolvedTheme } = useTheme();
  const { getTermDisplay } = useTerminology();
  const theme = themeColors[resolvedTheme];
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={objective.name}
      accessibilityHint={`View ${getTermDisplay("storyTerm", { variant: "plural" })} for this ${getTermDisplay("objectiveTerm")}`}
      className="min-h-[64px] px-[20px] py-3"
      style={({ pressed }) => ({
        backgroundColor: pressed ? theme.stateHover : "transparent",
      })}
      onPress={() =>
        router.push(
          teamStoriesHref({
            teamId: objective.teamId,
            objectiveId: objective.id,
          }),
        )
      }
    >
      <Row align="center" gap={3}>
        <Ionicons name="flag-outline" size={19} color={theme.textMuted} />
        <Col flex={1} gap={1}>
          <Text numberOfLines={1} ellipsizeMode="tail">
            {objective.name}
          </Text>
          {teamName || objective.health ? (
            <Text fontSize="xs" color="muted" numberOfLines={1}>
              {[teamName, objective.health].filter(Boolean).join(" · ")}
            </Text>
          ) : null}
        </Col>
        <Ionicons name="chevron-forward" size={14} color={theme.textMuted} />
      </Row>
    </Pressable>
  );
});

export function SearchResults({
  results,
  type,
  query,
  hasMore,
  loadingMore,
  error,
  loadMoreError,
  onLoadMore,
  onRetry,
}: SearchResultsProps) {
  const { getTermDisplay } = useTerminology();
  const { data: statuses = [] } = useStatuses();
  const { data: members = [] } = useMembers();
  const { data: teams = [] } = useTeams();
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
  const items = useMemo<SearchResult[]>(
    () =>
      type === "stories"
        ? results.stories.map((item) => ({ type: "story", item }))
        : results.objectives.map((item) => ({ type: "objective", item })),
    [type, results.stories, results.objectives],
  );
  const total =
    type === "stories" ? results.totalStories : results.totalObjectives;
  const renderResult = useCallback(
    ({ item: result }: { item: SearchResult }) =>
      result.type === "story" ? (
        <StoryRow
          story={result.item}
          visibleColumns={DEFAULT_STORY_DISPLAY_COLUMNS}
          showStatusLabel={false}
          status={statusById.get(result.item.statusId)}
          assignee={memberById.get(result.item.assigneeId ?? "")}
          team={teamById.get(result.item.teamId)}
          completionStatus={getCompletionStatus(statuses, result.item.teamId)}
        />
      ) : (
        <ObjectiveResult
          objective={result.item}
          teamName={teamById.get(result.item.teamId)?.name}
        />
      ),
    [statusById, memberById, teamById, statuses],
  );

  if (items.length === 0) {
    return (
      <QueryState
        title={
          error
            ? "Search could not be updated"
            : `No ${getTermDisplay(type === "stories" ? "storyTerm" : "objectiveTerm", { variant: "plural" })} found`
        }
        message={
          error
            ? error.message
            : `No matches for “${query}”. Try a different name or keyword.`
        }
        onRetry={error ? onRetry : undefined}
      />
    );
  }

  return (
    <FlatList
      data={items}
      renderItem={renderResult}
      keyExtractor={resultKey}
      keyboardShouldPersistTaps="handled"
      keyboardDismissMode="on-drag"
      initialNumToRender={12}
      maxToRenderPerBatch={10}
      ListHeaderComponent={
        <View className="px-[20px] pb-2 pt-1">
          <Text fontSize="xs" color="muted" accessibilityLiveRegion="polite">
            {total.toLocaleString()} {total === 1 ? "result" : "results"}
          </Text>
        </View>
      }
      ListFooterComponent={
        error ? (
          <QueryState
            title={
              loadMoreError
                ? "Could not load more results"
                : "Could not update results"
            }
            message={error.message}
            onRetry={loadMoreError ? onLoadMore : onRetry}
          />
        ) : hasMore ? (
          <View style={{ paddingHorizontal: 20, paddingTop: 16 }}>
            <Button
              color="tertiary"
              onPress={onLoadMore}
              disabled={loadingMore}
              loading={loadingMore}
            >
              {loadingMore ? "Loading results…" : "Load more results"}
            </Button>
          </View>
        ) : null
      }
      showsVerticalScrollIndicator={false}
      contentContainerStyle={{ paddingBottom: 16 }}
      style={{ flex: 1 }}
    />
  );
}
