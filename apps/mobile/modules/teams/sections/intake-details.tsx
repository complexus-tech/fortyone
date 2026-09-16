import type { ContextMenuAction } from "@/components/ui/context-menu.types";
import type { IntakeItem } from "./types";
import type { IntakePatch } from "./detail-data";
import { Linking, View } from "react-native";
import { useRouter } from "expo-router";
import { toast } from "sonner-native";
import { Button, Row, SafeContainer, Text } from "@/components/ui";
import { IntegrationProviderIcon } from "@/components/icons/integration-provider";
import { isSafeLink } from "@/components/rich-text/content";
import { useTerminology } from "@/hooks/use-terminology";
import {
  DetailHeader,
  useDetailAccess,
  useLinkActions,
  confirmAction,
  destructiveColor,
} from "@/modules/entity-details/header";
import { DetailTitle } from "@/modules/entity-details/title";
import { DetailDescription } from "@/modules/entity-details/description";
import { DetailFeed, type DetailFeedRow } from "@/modules/entity-details/feed";
import { DetailCommentComposer } from "@/modules/entity-details/comment-composer";
import { useIntakeComments, useIntakeCommands } from "./detail-data";
import { IntakeProperties } from "./intake-properties";
import { getIntakeItem } from "./queries";

const PROVIDERS = { github: "GitHub", slack: "Slack", intercom: "Intercom" };
export function IntakeDetails({
  item,
  refreshing,
  onRefresh,
}: {
  item: IntakeItem;
  refreshing: boolean;
  onRefresh: () => void;
}) {
  const router = useRouter();
  const { canEdit } = useDetailAccess();
  const { getTermDisplay } = useTerminology();
  const commands = useIntakeCommands(item.id);
  const comments = useIntakeComments(item);
  const canUpdate = canEdit && item.status === "pending";
  const hasComments = item.provider === "github" || item.provider === "slack";
  const acceptedStoryId = item.acceptedStoryId;
  const provider = PROVIDERS[item.provider];
  const sourceUrl = item.sourceUrl;
  const update = async (patch: IntakePatch) => {
    await commands.execute({ type: "update", patch });
  };
  const accept = async () => {
    const result = await commands.execute({ type: "accept" });
    if (result?.acceptedStoryId)
      router.push({
        pathname: "/story/[storyId]",
        params: { storyId: result.acceptedStoryId },
      });
  };
  const actions: ContextMenuAction[] = [
    ...useLinkActions(`/teams/${item.teamId}/requests/${item.id}`, item.title),
  ];
  if (sourceUrl && isSafeLink(sourceUrl))
    actions.push({
      label: `Open in ${provider}`,
      systemImage: "arrow.up.right.square",
      onPress: () => {
        void Linking.openURL(sourceUrl).catch(() =>
          toast.error("Could not open the source."),
        );
      },
    });
  if (canUpdate && !commands.isPending)
    actions.push(
      {
        label: "Accept intake",
        systemImage: "checkmark.circle",
        onPress: () => {
          void accept().catch((error) => toast.error(error.message));
        },
      },
      {
        label: "Decline intake",
        systemImage: "xmark.circle",
        color: destructiveColor,
        onPress: () =>
          confirmAction(
            "Decline intake",
            "This removes the item from your team’s intake queue. The original stays in the source integration.",
            async () => {
              await commands.execute({ type: "decline" });
              router.back();
            },
            true,
          ),
      },
    );
  const rows: DetailFeedRow[] = (comments.data ?? []).map((comment) => ({
    id: comment.id,
    kind: "comment",
    author: comment.authorName,
    avatar: comment.authorAvatar,
    body: comment.body,
    format: item.provider === "github" ? "markdown" : "plain",
    createdAt: comment.createdAt,
    deliveryStatus: comment.deliveryStatus,
  }));
  return (
    <SafeContainer isFull>
      <DetailHeader
        reference={
          item.sourceNumber ? `${provider} #${item.sourceNumber}` : "Intake"
        }
        actions={actions}
      />
      <DetailFeed
        rows={rows}
        tab="comments"
        options={hasComments ? [{ value: "comments", label: provider }] : []}
        loading={hasComments && comments.isPending}
        error={comments.error}
        onRetry={() => void comments.refetch()}
        refreshing={refreshing}
        onRefresh={() => {
          onRefresh();
          if (hasComments) void comments.refetch();
        }}
      >
        <DetailTitle
          title={item.title}
          draftKey={`intake:${item.id}:title`}
          onSave={
            canUpdate
              ? async (title, baseline) => {
                  const latest = await getIntakeItem(item.id);
                  if (latest.status !== "pending")
                    throw new Error("This item has already been handled.");
                  if (latest.title !== baseline && latest.title !== title)
                    throw new Error(
                      "The title changed elsewhere. Your draft is preserved.",
                    );
                  if (latest.title !== title) await update({ title });
                }
              : undefined
          }
        />
        <IntakeProperties
          item={item}
          disabled={!canUpdate || commands.isPending}
          onUpdate={update}
        />
        <Row asContainer align="center" className="mb-3 gap-2">
          <IntegrationProviderIcon provider={item.provider} size={16} />
          <Text color="muted" fontSize="sm">
            {provider} ·{" "}
            {item.status === "pending"
              ? "Pending review"
              : item.status === "accepted"
                ? "Accepted"
                : "Declined"}
          </Text>
        </Row>
        <DetailDescription
          html={item.description ?? ""}
          draftKey={`intake:${item.id}:description`}
          onSave={
            canUpdate
              ? async (description, baseline) => {
                  const latest = await getIntakeItem(item.id);
                  if (latest.status !== "pending")
                    throw new Error("This item has already been handled.");
                  if (
                    (latest.description ?? "") !== baseline &&
                    latest.description !== description
                  )
                    throw new Error(
                      "The description changed elsewhere. Your draft is preserved.",
                    );
                  if ((latest.description ?? "") !== description)
                    await update({ description });
                }
              : undefined
          }
        />
        <View className="px-[20px] mb-[16px] gap-2">
          {canUpdate ? (
            <Button
              color="tertiary"
              loading={commands.isPending}
              onPress={() =>
                void accept().catch((error) => toast.error(error.message))
              }
            >
              Accept intake
            </Button>
          ) : null}
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
          {!hasComments ? (
            <Text color="muted" fontSize="sm">
              Conversation is available in {provider}.
            </Text>
          ) : null}
        </View>
      </DetailFeed>
      {canEdit && hasComments ? (
        <DetailCommentComposer
          draftKey={`intake:${item.id}:comment`}
          notice={`Replies are sent to ${provider}.`}
          onSend={async (value, idempotencyKey) => {
            await commands.execute({
              type: "comment",
              provider: item.provider,
              body:
                item.provider === "github"
                  ? value.markdown ?? value.text
                  : value.text,
              idempotencyKey,
            });
          }}
        />
      ) : null}
    </SafeContainer>
  );
}
