import { useLocalSearchParams } from "expo-router";
import { SafeContainer } from "@/components/ui";
import { QueryState } from "@/components/ui/query-state";
import { DetailHeader } from "@/modules/entity-details/header";
import { useFeedbackItem, useIntakeItem } from "./hooks";
import { FeedbackDetails } from "./feedback-details";
import { IntakeDetails } from "./intake-details";

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
  if (query.isPending || query.error || !query.data || itemTeamId !== teamId)
    return (
      <SafeContainer isFull>
        <DetailHeader reference={kind === "feedback" ? "Feedback" : "Intake"} />
        <QueryState
          loading={query.isPending}
          title={
            query.isPending ? "Loading details" : "This item is unavailable"
          }
          message={query.error?.message}
          onRetry={query.error ? () => void query.refetch() : undefined}
        />
      </SafeContainer>
    );
  if (kind === "feedback" && feedback.data)
    return (
      <FeedbackDetails
        key={feedbackId}
        item={feedback.data}
        teamId={teamId}
        refreshing={feedback.isRefetching}
        onRefresh={() => void feedback.refetch()}
      />
    );
  return intake.data ? (
    <IntakeDetails
      key={requestId}
      item={intake.data}
      refreshing={intake.isRefetching}
      onRefresh={() => void intake.refetch()}
    />
  ) : null;
}
