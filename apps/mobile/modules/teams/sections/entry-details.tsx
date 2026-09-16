import type { FeedbackItem, IntakeItem } from "./types";
import {
  FlatList,
  ScrollView,
  View,
  RefreshControl,
  Linking,
} from "react-native";
import { useLocalSearchParams, useRouter } from "expo-router";
import Markdown from "react-native-markdown-display";
import { toast } from "sonner-native";
import {
  Back,
  Button,
  SafeContainer,
  ScreenHeader,
  Text,
} from "@/components/ui";
import { QueryState } from "@/components/ui/query-state";
import { RichTextViewer } from "@/components/rich-text/viewer";
import { isSafeLink } from "@/components/rich-text/content";
import { useTheme } from "@/hooks/theme";
import { themeColors } from "@/constants/colors";
import { useTerminology } from "@/hooks/use-terminology";
import { useFeedbackItem, useIntakeItem } from "./hooks";
import { feedbackStatusLabels } from "./types";

function FeedbackMarkdown({ children }: { children: string }) {
  const { resolvedTheme } = useTheme();
  const theme = themeColors[resolvedTheme];
  return (
    <Markdown
      style={{
        body: {
          color: theme.foreground,
          fontSize: 17,
          lineHeight: 25,
          fontWeight: "500",
        },
        link: { color: theme.foreground, textDecorationLine: "underline" },
      }}
      onLinkPress={(url) => {
        if (isSafeLink(url))
          void Linking.openURL(url).catch(() =>
            toast.error("Could not open this link."),
          );
        return false;
      }}
    >
      {children}
    </Markdown>
  );
}

export function TeamEntryDetails({ kind }: { kind: "feedback" | "intake" }) {
  const {
    teamId,
    feedbackId = "",
    requestId = "",
  } = useLocalSearchParams<{
    teamId: string;
    feedbackId?: string;
    requestId?: string;
  }>();
  const feedback = useFeedbackItem(feedbackId, kind === "feedback");
  const intake = useIntakeItem(requestId, kind === "intake");
  const query = kind === "feedback" ? feedback : intake;
  const itemTeamId =
    kind === "feedback" ? feedback.data?.board?.teamId : intake.data?.teamId;
  const ready = Boolean(query.data) && itemTeamId === teamId && !query.error;
  return (
    <SafeContainer isFull>
      <ScreenHeader
        title={kind === "feedback" ? "Feedback" : "Intake"}
        leading={<Back />}
        compact
      />
      {!ready ? (
        <QueryState
          loading={query.isPending}
          title={
            query.isPending ? "Loading details" : "This item is unavailable"
          }
          message={query.error?.message}
          onRetry={
            query.error
              ? () => {
                  void query.refetch();
                }
              : undefined
          }
        />
      ) : kind === "feedback" && feedback.data ? (
        <FeedbackContent
          item={feedback.data}
          refreshing={feedback.isRefetching}
          onRefresh={() => {
            void feedback.refetch();
          }}
        />
      ) : intake.data ? (
        <IntakeContent
          item={intake.data}
          refreshing={intake.isRefetching}
          onRefresh={() => {
            void intake.refetch();
          }}
        />
      ) : null}
    </SafeContainer>
  );
}

type RefreshProps = { refreshing: boolean; onRefresh: () => void };

function FeedbackContent({
  item,
  refreshing,
  onRefresh,
}: RefreshProps & { item: FeedbackItem }) {
  const router = useRouter();
  const { getTermDisplay } = useTerminology();
  const comments = item.comments ?? [];
  return (
    <FlatList
      data={comments}
      keyExtractor={(comment) => comment.id}
      contentContainerStyle={{
        paddingHorizontal: 20,
        paddingBottom: 32,
        gap: 16,
      }}
      refreshing={refreshing}
      onRefresh={onRefresh}
      ListHeaderComponent={
        <View style={{ gap: 16 }}>
          <Text fontSize="2xl" fontWeight="bold">
            {item.title}
          </Text>
          <Text color="muted">
            {feedbackStatusLabels[item.status]} · {item.authorName}
          </Text>
          <Text color="muted">
            {item.voteCount} votes · {item.commentCount} comments
          </Text>
          {item.description ? (
            <FeedbackMarkdown>{item.description}</FeedbackMarkdown>
          ) : (
            <Text color="muted">No description yet.</Text>
          )}
          {(item.storyLinks ?? []).map((link) => (
            <Button
              key={link.id}
              color="tertiary"
              onPress={() =>
                router.push({
                  pathname: "/story/[storyId]",
                  params: { storyId: link.storyId },
                })
              }
            >
              {link.storyTitle || `View ${getTermDisplay("storyTerm")}`}
            </Button>
          ))}
          {comments.length ? <Text fontWeight="bold">Comments</Text> : null}
        </View>
      }
      renderItem={({ item: comment }) => (
        <View style={{ gap: 4 }}>
          <Text fontWeight="semibold">{comment.authorName}</Text>
          <FeedbackMarkdown>{comment.body}</FeedbackMarkdown>
        </View>
      )}
    />
  );
}

function IntakeContent({
  item,
  refreshing,
  onRefresh,
}: RefreshProps & { item: IntakeItem }) {
  const router = useRouter();
  const { getTermDisplay } = useTerminology();
  const acceptedStoryId = item.acceptedStoryId;
  return (
    <ScrollView
      contentContainerStyle={{
        paddingHorizontal: 20,
        paddingBottom: 32,
        gap: 16,
      }}
      refreshControl={
        <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
      }
    >
      <Text fontSize="2xl" fontWeight="bold">
        {item.title}
      </Text>
      <Text color="muted">
        {item.provider} · {item.status} · {item.priority}
      </Text>
      {item.description ? (
        <RichTextViewer html={item.description} />
      ) : (
        <Text color="muted">No description yet.</Text>
      )}
      {acceptedStoryId ? (
        <Button
          color="tertiary"
          onPress={() =>
            router.push({
              pathname: "/story/[storyId]",
              params: { storyId: acceptedStoryId },
            })
          }
        >
          View {getTermDisplay("storyTerm")}
        </Button>
      ) : null}
    </ScrollView>
  );
}
