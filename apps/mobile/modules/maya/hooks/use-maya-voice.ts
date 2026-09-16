import {
  useCallback,
  useContext,
  useEffect,
  useState,
  useSyncExternalStore,
} from "react";
import { AppState, Appearance } from "react-native";
import {
  useFocusEffect,
  useNavigation,
  useRouter,
  type Href,
} from "expo-router";
import { useQueryClient } from "@tanstack/react-query";
import { getStoredSession } from "@/lib/auth";
import { get, post } from "@/lib/http";
import { getApiURL } from "@/lib/http/config";
import { SessionMutationContext } from "@/lib/use-session-mutation";
import { scopedQueryKey } from "@/lib/query-scope";
import { useAuthStore } from "@/store/auth";
import { useTheme } from "@/hooks/theme";
import {
  MayaVoiceController,
  type VoiceContext,
  type VoiceIdentity,
  type VoiceLease,
} from "@/lib/voice-controller";
import {
  assertNativeVoiceAvailable,
  openNativeVoiceTransport,
} from "@/lib/voice-native-transport";
import {
  nativeVoiceRoute,
  voiceSessionRequest,
  type VoiceToolResult,
} from "@/lib/voice-protocol";
import type { ApiResponse } from "@/types";

export type {
  VoiceTranscriptMessage,
  VoicePendingAction,
} from "@/lib/voice-protocol";

const pathFor = (identity: VoiceIdentity, path: string) =>
  `workspaces/${encodeURIComponent(identity.workspace)}/${path}`;
const readData = <T>(response: ApiResponse<T>): T => {
  if (!response.data)
    throw new Error(
      response.error?.message ?? "Maya returned an unreadable response.",
    );
  return response.data;
};

export function useMayaVoice(options: VoiceContext = {}) {
  const queryClient = useQueryClient();
  const admission = useContext(SessionMutationContext);
  const router = useRouter();
  const navigation = useNavigation();
  const { setTheme } = useTheme();
  const [controller] = useState(() => {
    const isCurrent = (identity: VoiceIdentity) => {
      const state = useAuthStore.getState();
      return (
        admission?.isActive() === true &&
        state.isAuthenticated &&
        !state.isLoading &&
        state.userId === identity.userId &&
        state.workspace === identity.workspace &&
        state.sessionEpoch === identity.sessionEpoch
      );
    };
    const assertCurrent = (identity: VoiceIdentity) => {
      admission?.assertCanMutate();
      if (!isCurrent(identity))
        throw new Error("Your session changed. Start voice again.");
    };
    return new MayaVoiceController({
      prepareTransport: assertNativeVoiceAvailable,
      captureIdentity: async () => {
        admission?.assertCanMutate();
        const current = useAuthStore.getState();
        const session = await getStoredSession();
        if (
          !session ||
          !current.userId ||
          !current.workspace ||
          session.userId !== current.userId ||
          session.apiOrigin !== getApiURL().origin
        )
          throw new Error("Sign in again before starting voice.");
        const identity = {
          userId: current.userId,
          workspace: current.workspace,
          sessionEpoch: current.sessionEpoch,
          cookie: session.cookie,
        };
        assertCurrent(identity);
        return identity;
      },
      isCurrent,
      startSession: async (identity, context, signal) => {
        assertCurrent(identity);
        const input = voiceSessionRequest(context);
        return readData(
          await post<typeof input, ApiResponse<VoiceLease>>(
            pathFor(identity, "maya/realtime-session"),
            input,
            { useWorkspace: false, sessionCookie: identity.cookie, signal },
          ),
        );
      },
      endSession: async (identity, sessionId) => {
        await post(
          pathFor(identity, "maya/realtime-session/end"),
          { sessionId },
          {
            useWorkspace: false,
            sessionCookie: identity.cookie,
            handleUnauthorized: false,
          },
        );
      },
      tool: async (identity, input, signal) => {
        assertCurrent(identity);
        return readData(
          await post<typeof input, ApiResponse<VoiceToolResult>>(
            pathFor(identity, "maya/realtime-tool"),
            input,
            { useWorkspace: false, sessionCookie: identity.cookie, signal },
          ),
        );
      },
      openTransport: openNativeVoiceTransport,
      invalidate: (identity) => {
        if (!isCurrent(identity)) return;
        for (const resource of [
          "stories",
          "search",
          "home",
          "notifications",
          "sprints",
          "objectives",
        ]) {
          void queryClient.invalidateQueries({
            queryKey: scopedQueryKey(identity, resource),
          });
        }
      },
      clientAction: async (identity, action) => {
        if (!isCurrent(identity) || !action || typeof action !== "object")
          return false;
        const command = action as {
          type?: unknown;
          path?: unknown;
          theme?: unknown;
        };
        if (
          command.type === "theme" &&
          ["light", "dark", "system", "toggle"].includes(String(command.theme))
        ) {
          const theme =
            command.theme === "toggle"
              ? Appearance.getColorScheme() === "dark"
                ? "light"
                : "dark"
              : (command.theme as "light" | "dark" | "system");
          setTheme(theme);
          return true;
        }
        if (command.type !== "navigate" || typeof command.path !== "string")
          return false;
        let target = nativeVoiceRoute(command.path);
        const reference = command.path.match(
          /^\/work\/([A-Za-z0-9]+-\d+)$/,
        )?.[1];
        if (!target && reference) {
          const response = readData(
            await get<ApiResponse<{ id: string }>>(
              pathFor(
                identity,
                `story-by-ref/${encodeURIComponent(reference)}`,
              ),
              { useWorkspace: false, sessionCookie: identity.cookie },
            ),
          );
          target = nativeVoiceRoute(`/work/${response.id}`);
        }
        if (!target || !isCurrent(identity)) return false;
        router.push(target as Href);
        return true;
      },
    });
  });
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
    controller.getSnapshot,
  );
  useFocusEffect(
    useCallback(() => {
      // Native tabs keep their screens mounted. A blur must release the mic and
      // session immediately, while preserving completed text for history saving.
      return controller.stop;
    }, [controller]),
  );
  useEffect(() => {
    const unsubscribe = useAuthStore.subscribe(controller.checkIdentity);
    const subscription = AppState.addEventListener("change", (next) => {
      if (
        next === "background" ||
        (next === "inactive" && controller.getSnapshot().status === "connected")
      )
        controller.stop();
    });
    return () => {
      unsubscribe();
      subscription.remove();
      controller.clearTranscript();
    };
  }, [controller]);
  return {
    ...state,
    // A permission/confirmation flow may finish after the user switches tabs.
    start: () =>
      navigation.isFocused() ? controller.start(options) : Promise.resolve(),
    stop: controller.stop,
    toggleMute: controller.toggleMute,
    approveAction: controller.approveAction,
    cancelAction: controller.cancelAction,
    clearTranscript: controller.clearTranscript,
    clearError: controller.clearError,
  };
}
