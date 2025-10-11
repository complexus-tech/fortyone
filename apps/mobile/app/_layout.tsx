import { Stack } from "expo-router";
import {
  QueryClient,
  QueryClientProvider,
  useQueryClient,
  focusManager,
  onlineManager,
} from "@tanstack/react-query";
// import { PersistQueryClientProvider } from "@tanstack/react-query-persist-client";
// import { createAsyncStoragePersister } from "@tanstack/query-async-storage-persister";
// import AsyncStorage from "@react-native-async-storage/async-storage";
import "../styles/global.css";
import "react-native-svg";
import { useAuthStore } from "@/store";
import { useEffect } from "react";
import { KeyboardProvider } from "react-native-keyboard-controller";
import { AppState } from "react-native";
import NetInfo from "@react-native-community/netinfo";
import { fetchGlobalQueries } from "@/lib/utils";

// const persister = createAsyncStoragePersister({
//   storage: AsyncStorage,
// });

function useReactQueryAppLifecycle() {
  useEffect(() => {
    const unsubscribeNetInfo = NetInfo.addEventListener((state) => {
      onlineManager.setOnline(!!state.isConnected);
    });

    const subscription = AppState.addEventListener("change", (status) => {
      focusManager.setFocused(status === "active");
    });

    return () => {
      unsubscribeNetInfo();
      subscription.remove();
    };
  }, []);
}
const RenderApp = () => {
  const queryClient = useQueryClient();
  fetchGlobalQueries(queryClient);

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
          sheetCornerRadius: 28,
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
          sheetCornerRadius: 28,
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
        retry: 1,
        refetchOnMount: true,
        refetchOnReconnect: true,
        refetchOnWindowFocus: true,
      },
    },
  });
  const loadAuthData = useAuthStore((state) => state.loadAuthData);
  useReactQueryAppLifecycle();

  useEffect(() => {
    loadAuthData();
  }, [loadAuthData]);

  return (
    <QueryClientProvider client={queryClient}>
      <KeyboardProvider>
        <RenderApp />
      </KeyboardProvider>
    </QueryClientProvider>
  );
}
