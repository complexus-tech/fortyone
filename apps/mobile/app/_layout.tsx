import { Stack } from "expo-router";
import { QueryClient, useQueryClient } from "@tanstack/react-query";
import { PersistQueryClientProvider } from "@tanstack/react-query-persist-client";
import { createAsyncStoragePersister } from "@tanstack/query-async-storage-persister";
import AsyncStorage from "@react-native-async-storage/async-storage";
import "../styles/global.css";
import "react-native-svg";
import { useAuthStore } from "@/store";
import { useEffect } from "react";
import { getProfile } from "@/modules/users/queries/get-profile";
import {
  memberKeys,
  objectiveKeys,
  statusKeys,
  subscriptionKeys,
  teamKeys,
  userKeys,
  workspaceKeys,
} from "@/constants/keys";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getTeams } from "@/modules/teams/queries/get-teams";
import { getSubscription } from "@/lib/queries/get-subscription";
import { getObjectiveStatuses } from "@/modules/objectives/queries/get-objectives";
import { getStatuses } from "@/modules/statuses/queries/get-statuses";
import { getMembers } from "@/modules/members/queries/get-members";

const persister = createAsyncStoragePersister({
  storage: AsyncStorage,
});

const RenderApp = () => {
  const queryClient = useQueryClient();
  queryClient.prefetchQuery({
    queryKey: userKeys.profile(),
    queryFn: getProfile,
  });
  queryClient.prefetchQuery({
    queryKey: workspaceKeys.lists(),
    queryFn: getWorkspaces,
  });
  queryClient.prefetchQuery({
    queryKey: teamKeys.lists(),
    queryFn: getTeams,
  });
  queryClient.prefetchQuery({
    queryKey: subscriptionKeys.details,
    queryFn: getSubscription,
  });
  queryClient.prefetchQuery({
    queryKey: memberKeys.lists(),
    queryFn: getMembers,
  });
  queryClient.prefetchQuery({
    queryKey: objectiveKeys.statuses(),
    queryFn: getObjectiveStatuses,
  });
  queryClient.prefetchQuery({
    queryKey: statusKeys.lists(),
    queryFn: getStatuses,
  });

  return (
    <Stack screenOptions={{ headerShown: false }}>
      <Stack.Screen name="(tabs)" />
      <Stack.Screen name="teams/[teamId]" />
      <Stack.Screen name="story/[storyId]" />
      <Stack.Screen
        name="settings"
        options={{
          presentation: "formSheet",
          gestureDirection: "vertical",
          animation: "slide_from_bottom",
          sheetGrabberVisible: true,
          sheetCornerRadius: 36,
          sheetExpandsWhenScrolledToEdge: true,
          sheetElevation: 24,
          sheetInitialDetentIndex: 0,
          sheetAllowedDetents: [0.95],
        }}
      />
      <Stack.Screen
        name="account"
        options={{
          presentation: "formSheet",
          gestureDirection: "vertical",
          animation: "slide_from_bottom",
          sheetGrabberVisible: true,
          sheetCornerRadius: 36,
          sheetExpandsWhenScrolledToEdge: true,
          sheetElevation: 24,
          sheetInitialDetentIndex: 0,
          sheetAllowedDetents: [0.91],
        }}
      />
    </Stack>
  );
};

export default function RootLayout() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        gcTime: 1000 * 60 * 60 * 24, // keep unused query data for 24 hours
        retry: 1,
        refetchOnReconnect: true,
        refetchOnWindowFocus: true,
      },
    },
  });
  const loadAuthData = useAuthStore((state) => state.loadAuthData);

  useEffect(() => {
    loadAuthData();
  }, [loadAuthData]);

  return (
    <PersistQueryClientProvider
      client={queryClient}
      persistOptions={{ persister }}
    >
      <RenderApp />
    </PersistQueryClientProvider>
  );
}
