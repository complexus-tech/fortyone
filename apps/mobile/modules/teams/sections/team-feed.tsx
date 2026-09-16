import type { StatusCategory } from "@/types/statuses";
import type { FeedbackItem } from "./types";
import { FlatList, View } from "react-native";
import { useRouter } from "expo-router";
import { Button, Text, Avatar } from "@/components/ui";
import { QueryState } from "@/components/ui/query-state";
import { WorkItemRow } from "../stories/components/work-item-row";
import { StatusIcon } from "@/components/icons";
import { IntegrationProviderIcon } from "@/components/icons/integration-provider";
import { colors } from "@/constants/colors";
import { useTeamFeedback, useTeamIntake } from "./hooks";
import { feedbackStatusLabels } from "./types";

const FEEDBACK_STATUS_ICONS: Record<
  FeedbackItem["status"],
  { category: StatusCategory; color?: string }
> = {
  pending: { category: "backlog" },
  reviewing: { category: "started", color: colors.info },
  planned: { category: "unstarted", color: colors.primary },
  in_progress: { category: "started", color: colors.warning },
  completed: { category: "completed", color: colors.success },
  closed: { category: "cancelled", color: colors.danger },
};
const PROVIDER_LABELS = {
  github: "GitHub",
  slack: "Slack",
  intercom: "Intercom",
};

export function TeamFeed({
  teamId,
  kind,
}: {
  teamId: string;
  kind: "feedback" | "intake";
}) {
  const router = useRouter();
  const feedback = useTeamFeedback(teamId, kind === "feedback");
  const intake = useTeamIntake(teamId, kind === "intake");
  const feed = kind === "feedback" ? feedback : intake;
  const rows =
    kind === "feedback"
      ? (feedback.data?.pages ?? []).flatMap((page) =>
          page.feedback.map((item) => ({
            id: item.id,
            title: item.title,
            subtitle: `${feedbackStatusLabels[item.status]} · ${item.voteCount} votes · ${item.commentCount} comments`,
            accessibilityLabel: `${item.title}, ${feedbackStatusLabels[item.status]}, ${item.voteCount} votes, ${item.commentCount} comments, submitted by ${item.authorName}`,
            leading: (
              <StatusIcon {...FEEDBACK_STATUS_ICONS[item.status]} size={20} />
            ),
            trailing: (
              <Avatar
                name={item.authorName}
                src={item.authorAvatar}
                style={{ width: 24, height: 24 }}
              />
            ),
          })),
        )
      : (intake.data?.pages ?? []).flatMap((page) =>
          page.requests.map((item) => ({
            id: item.id,
            title: item.title,
            subtitle: `${PROVIDER_LABELS[item.provider]} · ${item.priority} · Pending`,
            accessibilityLabel: `${item.title}, ${PROVIDER_LABELS[item.provider]}, ${item.priority}, Pending`,
            leading: <IntegrationProviderIcon provider={item.provider} />,
            trailing: null,
          })),
        );
  const items = [...new Map(rows.map((item) => [item.id, item])).values()];
  if (feed.isPending || (feed.error && !items.length)) {
    return (
      <QueryState
        loading={feed.isPending}
        title={feed.isPending ? `Loading ${kind}` : `Could not load ${kind}`}
        message={feed.error?.message}
        onRetry={
          feed.isPending
            ? undefined
            : () => {
                void feed.refetch();
              }
        }
      />
    );
  }
  return (
    <FlatList
      data={items}
      keyExtractor={(item) => item.id}
      renderItem={({ item }) => (
        <WorkItemRow
          title={item.title}
          subtitle={item.subtitle}
          leading={item.leading}
          trailing={item.trailing}
          accessibilityLabel={item.accessibilityLabel}
          accessibilityHint={`View ${kind} details`}
          onPress={() => {
            if (kind === "feedback")
              router.push({
                pathname: "/teams/[teamId]/feedback/[feedbackId]",
                params: { teamId, feedbackId: item.id },
              });
            else
              router.push({
                pathname: "/teams/[teamId]/intake/[requestId]",
                params: { teamId, requestId: item.id },
              });
          }}
        />
      )}
      contentContainerStyle={{ flexGrow: 1, paddingBottom: 32 }}
      initialNumToRender={12}
      refreshing={feed.isRefetching && !feed.isFetchingNextPage}
      onRefresh={() => {
        void feed.refetch();
      }}
      ListEmptyComponent={
        <QueryState
          title={kind === "feedback" ? "No feedback yet" : "No pending intake"}
          message="New items for this team will appear here."
        />
      }
      ListFooterComponent={
        feed.hasNextPage || feed.error ? (
          <View style={{ padding: 20, gap: 8 }}>
            {feed.error ? (
              <Text color="muted" fontSize="sm">
                {feed.error.message}
              </Text>
            ) : null}
            <Button
              color="tertiary"
              disabled={feed.isFetching}
              onPress={() => {
                if (feed.error && !feed.isFetchNextPageError)
                  void feed.refetch();
                else void feed.fetchNextPage();
              }}
            >
              {feed.isFetching
                ? "Loading…"
                : feed.error
                  ? "Try again"
                  : "Load more"}
            </Button>
          </View>
        ) : null
      }
    />
  );
}
