import type { ContextMenuAction } from "@/components/ui/context-menu.types";
import type { Objective } from "./types";
import type { ObjectivePatch } from "./detail-data";
import { useState } from "react";
import { View } from "react-native";
import { useLocalSearchParams, useRouter } from "expo-router";
import { Button, SafeContainer } from "@/components/ui";
import { QueryState } from "@/components/ui/query-state";
import { useTerminology } from "@/hooks/use-terminology";
import { useWorkspaceSettings } from "@/hooks/use-workspace-settings";
import { useMembers } from "@/modules/members";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { teamStoriesHref } from "@/modules/teams/stories/team-story-navigation";
import {
  DetailHeader,
  useDetailAccess,
  useLinkActions,
  confirmAction,
  destructiveColor,
} from "@/modules/entity-details/header";
import { DetailTitle } from "@/modules/entity-details/title";
import { DetailDescription } from "@/modules/entity-details/description";
import { DetailFeed } from "@/modules/entity-details/feed";
import { objectiveActivityRows } from "./activity-rows";
import { DetailCommentComposer } from "@/modules/entity-details/comment-composer";
import { useObjective, useObjectiveStatuses } from "./hooks/use-objectives";
import { ObjectiveProperties } from "./detail-properties";
import {
  readObjective,
  useObjectiveCommands,
  useObjectiveActivity,
} from "./detail-data";

export function ObjectiveDetails() {
  const { teamId, objectiveId } = useLocalSearchParams<{
    teamId: string;
    objectiveId: string;
  }>();
  const settings = useWorkspaceSettings();
  const query = useObjective(objectiveId);
  const { getTermDisplay } = useTerminology();
  const error = settings.error ?? query.error;
  const loading = settings.isPending || query.isPending;
  if (
    loading ||
    error ||
    !settings.data?.objectiveEnabled ||
    !query.data ||
    query.data.teamId !== teamId
  )
    return (
      <SafeContainer isFull>
        <DetailHeader
          reference={getTermDisplay("objectiveTerm", { capitalize: true })}
        />
        <QueryState
          loading={loading}
          title={
            loading
              ? "Loading details"
              : !settings.data?.objectiveEnabled && !error
                ? "This feature is not enabled"
                : "This item is unavailable"
          }
          message={error?.message}
          onRetry={
            error
              ? () => {
                  void settings.refetch();
                  void query.refetch();
                }
              : undefined
          }
        />
      </SafeContainer>
    );
  return (
    <ObjectiveContent
      key={objectiveId}
      item={query.data}
      refreshing={query.isRefetching}
      onRefresh={() => void query.refetch()}
    />
  );
}

function ObjectiveContent({
  item,
  refreshing,
  onRefresh,
}: {
  item: Objective;
  refreshing: boolean;
  onRefresh: () => void;
}) {
  const router = useRouter();
  const { canEdit } = useDetailAccess();
  const { getTermDisplay } = useTerminology();
  const { data: teams = [] } = useTeams();
  const { data: statuses = [] } = useObjectiveStatuses();
  const { data: members = [] } = useMembers();
  const [tab, setTab] = useState("updates");
  const activity = useObjectiveActivity(item.id, canEdit);
  const commands = useObjectiveCommands(item.id);
  const term = getTermDisplay("objectiveTerm");
  const reference = item.sequenceId
    ? `${teams.find((t) => t.id === item.teamId)?.code?.toUpperCase() ?? "OBJ"}-${item.sequenceId}`
    : getTermDisplay("objectiveTerm", { capitalize: true });
  const linkActions = useLinkActions(
    `/teams/${item.teamId}/objectives/${item.id}`,
    item.name,
  );
  const actions: ContextMenuAction[] = [...linkActions];
  if (canEdit && !commands.isPending)
    actions.push({
      label: `Delete ${term}`,
      systemImage: "trash",
      color: destructiveColor,
      onPress: () =>
        confirmAction(
          `Delete ${term}`,
          `This permanently deletes this ${term} and its key results.`,
          async () => {
            await commands.execute({ type: "delete" });
            router.back();
          },
          true,
        ),
    });
  const update = async (patch: ObjectivePatch) => {
    await commands.execute({ type: "update", patch });
  };
  const rows = objectiveActivityRows({
    activities: (activity.data?.pages ?? []).flatMap((page) => page.activities),
    tab,
    members,
    statuses,
    objectiveTerm: term,
    keyResultTerm: getTermDisplay("keyResultTerm"),
  });
  return (
    <SafeContainer isFull>
      <DetailHeader reference={reference} actions={actions} />
      <DetailFeed
        rows={rows}
        tab={tab}
        onTabChange={setTab}
        options={
          canEdit
            ? [
                { value: "updates", label: "Updates" },
                { value: "comments", label: "Comments" },
              ]
            : []
        }
        loading={canEdit && activity.isPending}
        error={activity.error}
        onRetry={() => {
          if (activity.isFetchNextPageError) void activity.fetchNextPage();
          else void activity.refetch();
        }}
        hasMore={activity.hasNextPage}
        loadingMore={activity.isFetchingNextPage}
        onLoadMore={() => {
          if (!activity.isFetching) void activity.fetchNextPage();
        }}
        refreshing={refreshing}
        onRefresh={() => {
          onRefresh();
          if (canEdit) void activity.refetch();
        }}
      >
        <DetailTitle
          title={item.name}
          draftKey={`objective:${item.id}:title`}
          onSave={
            canEdit
              ? async (name, baseline) => {
                  const latest = await readObjective(item.id);
                  if (latest.name !== baseline && latest.name !== name)
                    throw new Error(
                      "This title changed elsewhere. Your draft is preserved.",
                    );
                  if (latest.name !== name)
                    await update({ name, expectedUpdatedAt: latest.updatedAt });
                }
              : undefined
          }
        />
        <ObjectiveProperties
          item={item}
          disabled={!canEdit || commands.isPending}
          onUpdate={update}
        />
        <DetailDescription
          html={item.description ?? ""}
          draftKey={`objective:${item.id}:description`}
          onSave={
            canEdit
              ? async (description, baseline) => {
                  const latest = await readObjective(item.id);
                  if (
                    (latest.description ?? "") !== baseline &&
                    latest.description !== description
                  )
                    throw new Error(
                      "This description changed elsewhere. Your draft is preserved.",
                    );
                  if ((latest.description ?? "") !== description)
                    await update({
                      description,
                      expectedUpdatedAt: latest.updatedAt,
                    });
                }
              : undefined
          }
        />
        <View className="px-[20px] mb-[16px]">
          <Button
            color="tertiary"
            onPress={() =>
              router.push(
                teamStoriesHref({ teamId: item.teamId, objectiveId: item.id }),
              )
            }
          >
            View {getTermDisplay("storyTerm", { variant: "plural" })}
            {item.stats ? ` (${item.stats.total})` : ""}
          </Button>
        </View>
      </DetailFeed>
      {canEdit ? (
        <DetailCommentComposer
          draftKey={`objective:${item.id}:comment`}
          onSend={async (value) => {
            await update({ comment: value.markdown ?? value.text });
            setTab("comments");
          }}
        />
      ) : null}
    </SafeContainer>
  );
}
