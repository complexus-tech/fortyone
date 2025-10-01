import { Stack } from "expo-router";
import { SafeAreaProvider } from "react-native-safe-area-context";
import "../styles/global.css";
import "react-native-svg";

export default function RootLayout() {
  return (
    <SafeAreaProvider>
      <Stack screenOptions={{ headerShown: false }}>
        <Stack.Screen name="(tabs)" />
        <Stack.Screen name="team/[teamId]" />
        <Stack.Screen name="story/[storyId]" />
      </Stack>
    </SafeAreaProvider>
  );
}
