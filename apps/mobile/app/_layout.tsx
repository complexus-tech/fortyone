import { Stack } from "expo-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import "../styles/global.css";
import "react-native-svg";
import { useAuthStore } from "@/store";
import { useEffect } from "react";

export default function RootLayout() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 1000 * 60 * 5, // 5 minutes
      },
    },
  });
  const loadAuthData = useAuthStore((state) => state.loadAuthData);

  useEffect(() => {
    loadAuthData();
  }, [loadAuthData]);

  return (
    <QueryClientProvider client={queryClient}>
      <Stack screenOptions={{ headerShown: false }}>
        <Stack.Screen name="(tabs)" />
        <Stack.Screen name="team/[teamId]" />
        <Stack.Screen name="story/[storyId]" />
      </Stack>
    </QueryClientProvider>
  );
}
