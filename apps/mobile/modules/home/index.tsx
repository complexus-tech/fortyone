import { useRef, useState } from "react";
import { RefreshControl, ScrollView } from "react-native";
import { useQueryClient } from "@tanstack/react-query";
import { SafeContainer } from "@/components/ui";
import { homeKeys, teamKeys } from "@/constants/keys";
import { Header } from "./components/header";
import { Overview } from "./components/overview";
import { Teams } from "./components/teams";

export const Home = () => {
  const client = useQueryClient();
  const [refreshing, setRefreshing] = useState(false);
  const refreshPending = useRef(false);

  const refresh = () => {
    if (refreshPending.current) return;
    refreshPending.current = true;
    setRefreshing(true);
    // Each query keeps its own error and retry UI; finish refreshing once both settle.
    void Promise.allSettled([
      client.invalidateQueries({ queryKey: homeKeys.overview() }),
      client.invalidateQueries({ queryKey: teamKeys.lists() }),
    ]).then(() => {
      refreshPending.current = false;
      setRefreshing(false);
    });
  };

  return (
    <SafeContainer isFull>
      <Header />
      <ScrollView
        className="flex-1"
        contentInsetAdjustmentBehavior="automatic"
        contentContainerStyle={{ paddingBottom: 32 }}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={refresh} />
        }
      >
        <Overview />
        <Teams />
      </ScrollView>
    </SafeContainer>
  );
};
