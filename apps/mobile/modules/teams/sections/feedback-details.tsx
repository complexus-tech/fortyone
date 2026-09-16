import type { FeedbackItem } from "./types";
import { useState } from "react";
import { View } from "react-native";
import { toast } from "sonner-native";
import { useRouter } from "expo-router";
import { Avatar, Button, Col, SafeContainer, Text } from "@/components/ui";
import { feedbackKeys } from "@/constants/keys";
import { useTerminology } from "@/hooks/use-terminology";
import { DetailHeader } from "@/modules/entity-details/header";
import { DetailTitle } from "@/modules/entity-details/title";
import {
  DetailProperties,
  SelectProperty,
} from "@/modules/entity-details/properties";
import { PropertyChip } from "@/modules/story/components/properties/property-chip";
import {
  DetailFeed,
  MarkdownContent,
  type DetailFeedRow,
} from "@/modules/entity-details/feed";
import { DetailCommentComposer } from "@/modules/entity-details/comment-composer";
import {
  TextActionDialog,
  SearchActionDialog,
} from "@/modules/entity-details/dialogs";
import { readEntity } from "@/modules/entity-details/http";
import { searchQuery } from "@/modules/search/queries/search";
import {
  useFeedbackActions,
  type FeedbackDialog,
} from "./use-feedback-actions";
import { FeedbackStatusIcon } from "./feedback-status";
import { feedbackStatusLabels } from "./types";
import { threadedFeedbackComments } from "./comment-threads";

export function FeedbackDetails({
  item,
  teamId,
  refreshing,
  onRefresh,
}: {
  item: FeedbackItem;
  teamId: string;
  refreshing: boolean;
  onRefresh: () => void;
}) {
  const router = useRouter();
  const { getTermDisplay } = useTerminology();
  const [dialog, setDialog] = useState<FeedbackDialog>(null);
  const [replyTo, setReplyTo] = useState<{ id: string; name: string } | null>(
    null,
  );
  const {
    commands,
    actions,
    canPlan,
    canTriage,
    canComment,
    primary,
    openStory,
    prioritize,
  } = useFeedbackActions(item, teamId, setDialog);
  const mergedIntoItemId = item.mergedIntoItemId;
  const storyTerm = getTermDisplay("storyTerm");
  const rows: DetailFeedRow[] = threadedFeedbackComments(
    item.comments ?? [],
  ).map((comment) => ({
    id: comment.id,
    kind: "comment",
    author: comment.authorName,
    avatar: comment.authorAvatar,
    body: comment.body,
    format: "markdown",
    createdAt: comment.createdAt,
    parentId: comment.parentId,
    onReply: canComment
      ? () =>
          setReplyTo({
            id: comment.parentId ?? comment.id,
            name: comment.authorName,
          })
      : undefined,
  }));
  const statusOptions = (
    Object.keys(feedbackStatusLabels) as FeedbackItem["status"][]
  ).filter((status) => status !== "closed");
  return (
    <SafeContainer isFull>
      <DetailHeader reference="Feedback" actions={actions} />
      <DetailFeed
        rows={rows}
        tab="comments"
        options={[{ value: "comments", label: "Comments" }]}
        refreshing={refreshing}
        onRefresh={onRefresh}
      >
        {item.deletedAt ? (
          <Col asContainer>
            <Text color="muted">In trash</Text>
          </Col>
        ) : null}
        {item.mergedIntoItemId ? (
          <View className="px-[20px] mb-3">
            <Button
              color="tertiary"
              onPress={() =>
                router.replace({
                  pathname: "/teams/[teamId]/feedback/[feedbackId]",
                  params: { teamId, feedbackId: mergedIntoItemId ?? item.id },
                })
              }
            >
              Open merged feedback
            </Button>
          </View>
        ) : null}
        <DetailTitle
          title={item.title}
          draftKey={`feedback:${item.id}:title`}
        />
        <DetailProperties>
          <SelectProperty
            title="Status"
            label={feedbackStatusLabels[item.status]}
            icon={<FeedbackStatusIcon status={item.status} size={18} />}
            disabled={!canTriage || item.status === "closed"}
            searchable={false}
            options={statusOptions.map((status) => ({
              id: status,
              label: feedbackStatusLabels[status],
              icon: <FeedbackStatusIcon status={status} />,
            }))}
            selectedIds={[item.status]}
            onSelect={async (value) => {
              const status = statusOptions.find((status) => status === value);
              if (status) await commands.execute({ type: "status", status });
            }}
          />
          <PropertyChip>
            <Avatar size="xs" name={item.authorName} src={item.authorAvatar} />
            <Text fontSize="sm" numberOfLines={1}>
              {item.authorName}
            </Text>
          </PropertyChip>
          <PropertyChip>
            <Text fontSize="sm" color="muted">
              {item.voteCount} votes
            </Text>
          </PropertyChip>
        </DetailProperties>
        {primary ? (
          <Col asContainer>
            <Text color="muted" fontSize="sm">
              Status follows the linked {storyTerm}.
            </Text>
          </Col>
        ) : null}
        <View className="px-[20px] mb-[12px]">
          {item.description ? (
            <MarkdownContent>{item.description}</MarkdownContent>
          ) : (
            <Text color="muted">No description yet.</Text>
          )}
        </View>
        <View className="px-[20px] mb-[16px] gap-2">
          {canPlan ? (
            <Button
              color="tertiary"
              loading={commands.isPending}
              onPress={() =>
                void prioritize().catch((error: Error) =>
                  toast.error(error.message),
                )
              }
            >
              Prioritize
            </Button>
          ) : null}
          {(item.storyLinks ?? []).map((link) => (
            <Button
              key={link.id}
              color="tertiary"
              onPress={() => openStory(link.storyId)}
            >
              {link.storyTitle || `View ${storyTerm}`}
            </Button>
          ))}
        </View>
      </DetailFeed>
      {canComment ? (
        <>
          {replyTo ? (
            <View className="px-[20px]">
              <Button
                fullWidth={false}
                color="tertiary"
                onPress={() => setReplyTo(null)}
              >
                Cancel reply to {replyTo.name}
              </Button>
            </View>
          ) : null}
          <DetailCommentComposer
            key={replyTo?.id ?? "new"}
            draftKey={`feedback:${item.id}:comment:${replyTo?.id ?? "new"}`}
            notice="Comments are visible in the feedback portal."
            label={replyTo ? `Reply to ${replyTo.name}…` : undefined}
            onSent={() => setReplyTo(null)}
            onSend={async (value) => {
              await commands.execute({
                type: "comment",
                body: value.markdown ?? value.text,
                parentId: replyTo?.id,
              });
            }}
          />
        </>
      ) : null}
      {dialog === "close" ? (
        <TextActionDialog
          title="Close feedback"
          placeholder="Public explanation (optional)"
          description="Closing removes this item from active feedback. Followers are notified only when you include an explanation."
          onClose={() => setDialog(null)}
          onSubmit={async (explanation) => {
            await commands.execute({
              type: "status",
              status: "closed",
              explanation: explanation || null,
            });
          }}
        />
      ) : null}
      {dialog === "link" ? (
        <SearchActionDialog
          title={`Link ${storyTerm}`}
          description={`Search this team’s ${getTermDisplay("storyTerm", { variant: "plural" })} to connect this feedback.`}
          queryKey={[...feedbackKeys.detail(item.id), "link-search", teamId]}
          search={async (query, signal) => {
            if (!query) return [];
            const result = await searchQuery(
              { query, teamId, type: "stories", pageSize: 20 },
              signal,
            );
            return (result.stories ?? []).map((story) => ({
              id: story.id,
              title: story.title,
            }));
          }}
          onClose={() => setDialog(null)}
          onSelect={async (storyId) => {
            await commands.execute({ type: "plan", teamId, storyId });
          }}
        />
      ) : null}
      {dialog === "merge" ? (
        <SearchActionDialog
          title="Merge feedback"
          description="Select the feedback to keep. This request’s followers and linked updates move to it. This cannot be undone."
          queryKey={[...feedbackKeys.detail(item.id), "merge-search"]}
          search={async (query, signal) => {
            const result = await readEntity<{
              candidates: { id: string; title: string }[];
            }>(
              `feedback/items/${encodeURIComponent(item.id)}/merge-candidates?limit=30&search=${encodeURIComponent(query)}`,
              signal,
            );
            return result.candidates;
          }}
          onClose={() => setDialog(null)}
          onSelect={async (targetItemId) => {
            const result = await commands.execute({
              type: "merge",
              targetItemId,
            });
            if (result?.target?.board?.teamId)
              router.replace({
                pathname: "/teams/[teamId]/feedback/[feedbackId]",
                params: {
                  teamId: result.target.board.teamId,
                  feedbackId: result.target.id,
                },
              });
            else router.back();
          }}
        />
      ) : null}
    </SafeContainer>
  );
}
