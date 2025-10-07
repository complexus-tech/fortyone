import { Stack } from "expo-router";
import { QueryClient } from "@tanstack/react-query";
import { PersistQueryClientProvider } from "@tanstack/react-query-persist-client";
import { createAsyncStoragePersister } from "@tanstack/query-async-storage-persister";
import AsyncStorage from "@react-native-async-storage/async-storage";
import "../styles/global.css";
import "react-native-svg";
import { useAuthStore } from "@/store";
import { useEffect } from "react";

const persister = createAsyncStoragePersister({
  storage: AsyncStorage,
});

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
      </Stack>
    </PersistQueryClientProvider>
  );
}
