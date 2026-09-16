import { ScrollView, View, RefreshControl } from "react-native";
import { useLocalSearchParams, useRouter } from "expo-router";
import {
  Back,
  Button,
  SafeContainer,
  ScreenHeader,
  Text,
} from "@/components/ui";
import { QueryState } from "@/components/ui/query-state";
import { RichTextViewer } from "@/components/rich-text/viewer";
import { useTerminology } from "@/hooks/use-terminology";
import { useWorkspaceSettings } from "@/hooks/use-workspace-settings";
import { useMembers } from "@/modules/members";
import { teamStoriesHref } from "@/modules/teams/stories/team-story-navigation";
import { formatContextDates } from "@/modules/teams/stories/context-dates";
import { useObjective, useObjectiveStatuses } from "./hooks/use-objectives";

export function ObjectiveDetails() {
  const { teamId, objectiveId } = useLocalSearchParams<{
    teamId: string;
    objectiveId: string;
  }>();
  const settings = useWorkspaceSettings();
  const { getTermDisplay } = useTerminology();
  const term = getTermDisplay("objectiveTerm", { capitalize: true });
  return (
    <SafeContainer isFull>
      <ScreenHeader title={term} leading={<Back />} compact />
      {settings.isPending || settings.error ? (
        <QueryState
          loading={settings.isPending}
          title={
            settings.isPending ? "Loading settings" : "Could not load settings"
          }
          message={settings.error?.message}
          onRetry={
            settings.error
              ? () => {
                  void settings.refetch();
                }
              : undefined
          }
        />
      ) : !settings.data?.objectiveEnabled ? (
        <QueryState
          title={`${getTermDisplay("objectiveTerm", { variant: "plural", capitalize: true })} are not enabled`}
        />
      ) : (
        <ObjectiveContent teamId={teamId} objectiveId={objectiveId} />
      )}
    </SafeContainer>
  );
}

function ObjectiveContent({
  teamId,
  objectiveId,
}: {
  teamId: string;
  objectiveId: string;
}) {
  const query = useObjective(objectiveId);
  const { data: statuses = [] } = useObjectiveStatuses();
  const { data: members = [] } = useMembers();
  const { getTermDisplay } = useTerminology();
  const router = useRouter();
  const item = query.data;
  if (query.isPending || query.error || !item || item.teamId !== teamId) {
    return (
      <QueryState
        loading={query.isPending}
        title={query.isPending ? "Loading details" : "This item is unavailable"}
        message={query.error?.message}
        onRetry={
          query.error
            ? () => {
                void query.refetch();
              }
            : undefined
        }
      />
    );
  }
  const status = statuses.find((candidate) => candidate.id === item.statusId);
  const lead = members.find((member) => member.id === item.leadUser);
  return (
    <ScrollView
      contentContainerStyle={{
        paddingHorizontal: 20,
        paddingBottom: 32,
        gap: 16,
      }}
      refreshControl={
        <RefreshControl
          refreshing={query.isRefetching}
          onRefresh={() => {
            void query.refetch();
          }}
        />
      }
    >
      <Text fontSize="2xl" fontWeight="bold">
        {item.name}
      </Text>
      <View style={{ gap: 6 }}>
        {status ? <Text color="muted">Status: {status.name}</Text> : null}
        {item.health ? <Text color="muted">Health: {item.health}</Text> : null}
        {item.priority ? (
          <Text color="muted">Priority: {item.priority}</Text>
        ) : null}
        {lead ? <Text color="muted">Lead: {lead.fullName}</Text> : null}
        <Text color="muted">
          {formatContextDates(item.startDate, item.endDate)}
        </Text>
        {item.stats ? (
          <Text color="muted">
            {item.stats.completed} of {item.stats.total} completed
          </Text>
        ) : null}
      </View>
      {item.description ? (
        <RichTextViewer html={item.description} />
      ) : (
        <Text color="muted">No description yet.</Text>
      )}
      <Button
        color="tertiary"
        onPress={() => router.push(teamStoriesHref({ teamId, objectiveId }))}
      >
        View {getTermDisplay("storyTerm", { variant: "plural" })}
      </Button>
    </ScrollView>
  );
}
