import { Stack } from "expo-router";
import "../styles/global.css";
import "react-native-svg";
import { useAuthStore } from "@/store";
import { useEffect } from "react";

export default function RootLayout() {
  const loadAuthData = useAuthStore((state) => state.loadAuthData);

  useEffect(() => {
    loadAuthData();
  }, [loadAuthData]);
  return (
    <Stack screenOptions={{ headerShown: false }}>
      <Stack.Screen name="(tabs)" />
      <Stack.Screen name="team/[teamId]" />
      <Stack.Screen name="story/[storyId]" />
    </Stack>
  );
}
