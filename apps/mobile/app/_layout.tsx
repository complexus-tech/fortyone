import { Stack, type ErrorBoundaryProps } from "expo-router";
import {
  useQueryClient,
  focusManager,
  onlineManager,
} from "@tanstack/react-query";
import "../styles/global.css";
import "react-native-svg";
import { useAuthStore } from "@/store";
import { useEffect } from "react";
import { SessionQueryProvider } from "@/lib/query-provider";
import { QueryState } from "@/components/ui/query-state";
import { KeyboardProvider } from "react-native-keyboard-controller";
import {
  AppState,
  Platform,
  Pressable,
  Text as NativeText,
  View,
} from "react-native";
import { Text, Button } from "@/components/ui";
import NetInfo from "@react-native-community/netinfo";
import { GestureHandlerRootView } from "react-native-gesture-handler";
import { Toaster } from "sonner-native";
import { fetchGlobalQueries } from "@/lib/utils";
import { useTheme } from "@/hooks";
import { SymbolView } from "expo-symbols";
import { Ionicons } from "@expo/vector-icons";
import { themeColors } from "@/constants/colors";
import { BottomSheetModalProvider } from "@gorhom/bottom-sheet";
import { captureAppError, initializeObservability } from "@/lib/observability";

initializeObservability();

// Keep the recovery screen independent of the theme, auth and query providers:
// a failure in one of those providers must not break the fallback as well.
export function ErrorBoundary({ error, retry }: ErrorBoundaryProps) {
  useEffect(() => {
    captureAppError(error);
  }, [error]);

  return (
    <View
      style={{
        flex: 1,
        justifyContent: "center",
        padding: 28,
        gap: 16,
        backgroundColor: themeColors.dark.background,
      }}
    >
      <NativeText
        accessibilityRole="header"
        style={{
          color: themeColors.dark.foreground,
          fontSize: 24,
          fontWeight: "700",
        }}
      >
        Something went wrong
      </NativeText>
      <NativeText
        style={{ color: themeColors.dark.textSecondary, fontSize: 16 }}
      >
        Please try reopening this screen.
      </NativeText>
      <Pressable
        accessibilityRole="button"
        accessibilityLabel="Try reopening this screen"
        onPress={retry}
        style={{
          minHeight: 48,
          alignItems: "center",
          justifyContent: "center",
          borderRadius: 12,
          backgroundColor: themeColors.dark.backgroundInverse,
        }}
      >
        <NativeText
          style={{
            color: themeColors.dark.foregroundInverse,
            fontSize: 16,
            fontWeight: "700",
          }}
        >
          Try again
        </NativeText>
      </Pressable>
    </View>
  );
}

function useReactQueryAppLifecycle() {
  useEffect(() => {
    focusManager.setFocused(AppState.currentState === "active");
    const unsubscribeNetInfo = NetInfo.addEventListener((state) => {
      onlineManager.setOnline(
        state.isConnected === true && state.isInternetReachable !== false,
      );
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
  const isLoading = useAuthStore((state) => state.isLoading);
  const workspace = useAuthStore((state) => state.workspace);
  const sessionError = useAuthStore((state) => state.sessionError);

  useEffect(() => {
    if (isAuthenticated && workspace) {
      fetchGlobalQueries(queryClient);
    }
  }, [isAuthenticated, workspace, queryClient]);

  if (isLoading) {
    return <QueryState loading title="Restoring your session" />;
  }

  if (isAuthenticated && !workspace) {
    return (
      <QueryState
        title="No workspace available"
        message="Create or join a workspace on the FortyOne website, then sign in again."
        onRetry={() => {
          void useAuthStore.getState().clearAuth();
        }}
        retryLabel="Sign out"
      />
    );
  }

  return (
    <View className="flex-1">
      {isAuthenticated && sessionError ? (
        <View className="gap-2 px-4 py-3" accessibilityLiveRegion="polite">
          <Text color="muted">{sessionError}</Text>
          <Button
            onPress={() => {
              void useAuthStore.getState().loadAuthData();
            }}
          >
            Retry connection
          </Button>
        </View>
      ) : null}
      <Stack screenOptions={{ headerShown: false }}>
        <Stack.Protected guard={isAuthenticated}>
          <Stack.Screen name="(tabs)" />
          <Stack.Screen name="teams/[teamId]" />
          <Stack.Screen name="team/[teamId]/objectives/[objectiveId]" />
          <Stack.Screen name="team/[teamId]/sprints/[sprintId]" />
          <Stack.Screen name="story/[storyId]" />
          <Stack.Screen
            name="settings"
            options={
              Platform.OS === "ios"
                ? {
                    presentation: "transparentModal",
                    animation: "none",
                    gestureEnabled: false,
                    contentStyle: { backgroundColor: "transparent" },
                  }
                : {
                    presentation: "formSheet",
                    gestureDirection: "vertical",
                    animation: "slide_from_bottom",
                    sheetGrabberVisible: true,
                    sheetExpandsWhenScrolledToEdge: true,
                    sheetElevation: 24,
                    sheetInitialDetentIndex: 0,
                    sheetAllowedDetents: [0.8, 1],
                  }
            }
          />
          <Stack.Screen
            name="new"
            options={{
              presentation: "formSheet",
              gestureDirection: "vertical",
              animation: "slide_from_bottom",
              sheetGrabberVisible: false,
              sheetCornerRadius: 28,
              sheetExpandsWhenScrolledToEdge: true,
              sheetElevation: 24,
              sheetInitialDetentIndex: 0,
              sheetAllowedDetents: [1],
            }}
          />
        </Stack.Protected>

        <Stack.Protected guard={!isAuthenticated}>
          <Stack.Screen name="login" />
        </Stack.Protected>
      </Stack>
    </View>
  );
};

export default function RootLayout() {
  const { resolvedTheme } = useTheme();
  const iconColor = themeColors[resolvedTheme].textMuted;

  const userId = useAuthStore((state) => state.userId);
  const workspace = useAuthStore((state) => state.workspace);
  const sessionEpoch = useAuthStore((state) => state.sessionEpoch);
  const loadAuthData = useAuthStore((state) => state.loadAuthData);
  useReactQueryAppLifecycle();

  useEffect(() => {
    loadAuthData();
  }, [loadAuthData]);

  return (
    <SessionQueryProvider
      key={JSON.stringify([userId, workspace, sessionEpoch])}
      scope={{ userId, workspace }}
    >
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
                backgroundColor: themeColors[resolvedTheme].surface,
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
    </SessionQueryProvider>
  );
}
