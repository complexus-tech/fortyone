import { Stack } from "expo-router";

export const unstable_settings = { initialRouteName: "index" };

export default function StoryLayout() {
  return (
    <Stack screenOptions={{ headerShown: false }}>
      <Stack.Screen name="index" />
      <Stack.Screen name="sub-stories" />
      <Stack.Screen name="links" />
    </Stack>
  );
}
