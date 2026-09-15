import { useLocalSearchParams } from "expo-router";
import { SafeContainer } from "@/components/ui";
import { QueryState } from "@/components/ui/query-state";
import { useTerminology } from "@/hooks/use-terminology";
import { useTeamSprints } from "./hooks";
import { Header } from "./components/header";
import { List } from "./components/list";

export const Sprints = () => {
  const { teamId } = useLocalSearchParams<{ teamId: string }>();
  const { getTermDisplay } = useTerminology();
  const {
    data: items = [],
    isPending,
    isRefetching,
    error,
    refetch,
  } = useTeamSprints(teamId);
  const term = getTermDisplay("sprintTerm", { variant: "plural" });
  return (
    <SafeContainer isFull>
      <Header />
      {isPending ? (
        <QueryState loading title={`Loading ${term}`} />
      ) : error ? (
        <QueryState
          title={`Could not load ${term}`}
          message={error.message}
          onRetry={() => {
            void refetch();
          }}
        />
      ) : (
        <List
          sprints={items}
          refreshing={isRefetching}
          onRefresh={() => {
            void refetch();
          }}
        />
      )}
    </SafeContainer>
  );
};
