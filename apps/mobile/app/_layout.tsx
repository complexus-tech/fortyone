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
import { GestureHandlerRootView } from "react-native-gesture-handler";
import { Toaster } from "sonner-native";
import { fetchGlobalQueries } from "@/lib/utils";
import { useTheme } from "@/hooks";
import { SymbolView } from "expo-symbols";
import { Ionicons } from "@expo/vector-icons";
import { colors } from "@/constants";
import { BottomSheetModalProvider } from "@gorhom/bottom-sheet";

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
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  fetchGlobalQueries(queryClient);

  return (
    <Stack screenOptions={{ headerShown: false }}>
      <Stack.Protected guard={isAuthenticated}>
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
            sheetAllowedDetents: [0.95],
          }}
        />
      </Stack.Protected>

      <Stack.Protected guard={!isAuthenticated}>
        <Stack.Screen name="login" />
      </Stack.Protected>
    </Stack>
  );
};

export default function RootLayout() {
  const { resolvedTheme } = useTheme();
  const iconColor =
    resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[300];
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
        <GestureHandlerRootView>
          <BottomSheetModalProvider>
            <RenderApp />
          </BottomSheetModalProvider>
          <Toaster
            theme={resolvedTheme}
            closeButton
            toastOptions={{
              style: {
                backgroundColor:
                  resolvedTheme === "dark" ? colors.dark[200] : colors.white,
              },
            }}
            icons={{
              success: (
                <SymbolView
                  name="checkmark.circle.fill"
                  size={20}
                  tintColor={iconColor}
                  fallback={
                    <Ionicons
                      name="checkmark-circle"
                      size={20}
                      color={iconColor}
                    />
                  }
                />
              ),
              error: (
                <SymbolView
                  name="xmark.circle.fill"
                  size={20}
                  tintColor={iconColor}
                  fallback={
                    <Ionicons name="close-circle" size={20} color={iconColor} />
                  }
                />
              ),
              warning: (
                <SymbolView
                  name="exclamationmark.triangle.fill"
                  size={20}
                  tintColor={iconColor}
                  fallback={
                    <Ionicons name="warning" size={20} color={iconColor} />
                  }
                />
              ),
              info: (
                <SymbolView
                  name="info.circle.fill"
                  size={20}
                  tintColor={iconColor}
                  fallback={
                    <Ionicons
                      name="information-circle"
                      size={20}
                      color={iconColor}
                    />
                  }
                />
              ),
            }}
          />
        </GestureHandlerRootView>
      </KeyboardProvider>
    </QueryClientProvider>
  );
}
