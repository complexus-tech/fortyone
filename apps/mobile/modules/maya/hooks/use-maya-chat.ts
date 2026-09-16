import {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { AppState } from "react-native";
import { useFocusEffect, useRouter } from "expo-router";
import { fetch as expoFetch } from "expo/fetch";
import { useChat } from "@ai-sdk/react";
import type { FileUIPart } from "ai";
import { generateId } from "ai";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { getStoredSession } from "@/lib/auth";
import { getApiURL, getApplicationURL } from "@/lib/http/config";
import { SessionMutationContext } from "@/lib/use-session-mutation";
import { useAuthStore } from "@/store/auth";
import { useTheme } from "@/hooks/theme";
import type { MayaMessage, MayaScreenContext } from "../types";
import { createMayaCloudClient } from "../lib/cloud-client";
import { createMayaChatRuntime } from "../lib/chat-runtime";
import { loadMayaRequestContext } from "../lib/load-request-context";
import {
  createMayaSessionGuard,
  type MayaSessionScope,
} from "../lib/session-scope";
import {
  deleteMayaSession,
  getMayaMessages,
  getMayaSessions,
} from "../lib/conversations";
import {
  getPendingMayaApprovals,
  getMayaScreenPath,
  isEntityId,
} from "../lib/chat-protocol";
import {
  getMayaClientActions,
  hasMayaMutationReceipt,
} from "../lib/client-actions";
import {
  createVoiceHistoryQueue,
  type VoiceHistoryMessage,
} from "../lib/voice-history-queue";

const asError = (error: unknown) =>
  error instanceof Error
    ? error
    : new Error("Maya could not complete this action.");

const runMayaOperation = async (
  operation: () => Promise<void>,
  options: {
    assertCurrent: () => void;
    isCurrent: () => boolean;
    isSelecting: boolean;
    onError: (error: Error | null) => void;
  },
) => {
  try {
    options.assertCurrent();
    if (options.isSelecting)
      throw new Error("Wait for the conversation to finish loading.");
    options.onError(null);
    await operation();
  } catch (error) {
    if (options.isCurrent()) options.onError(asError(error));
    throw error;
  }
};

export const useMayaChat = (context: MayaScreenContext = {}) => {
  const provider = useContext(SessionMutationContext);
  if (!provider) throw new Error("Maya requires an active session provider.");
  const [scope] = useState<MayaSessionScope>(() => {
    const state = useAuthStore.getState();
    if (
      !state.userId ||
      !state.workspace ||
      !state.isAuthenticated ||
      state.isLoading
    ) {
      throw new Error("Sign in to a workspace before opening Maya.");
    }
    return {
      userId: state.userId,
      workspace: state.workspace,
      sessionEpoch: state.sessionEpoch,
    };
  });
  const [selection, setSelection] = useState<{
    id: string;
    messages: MayaMessage[];
    persisted: boolean;
  }>(() => ({ id: generateId(), messages: [], persisted: false }));
  const [isHistoryLoading, setHistoryLoading] = useState(false);
  const [operationError, setOperationError] = useState<Error | null>(null);
  const queryClient = useQueryClient();
  const router = useRouter();
  const { theme, resolvedTheme } = useTheme();
  const selectionRequest = useRef<AbortController | null>(null);
  const deletionRequest = useRef<AbortController | null>(null);
  const navigationVersion = useRef(0);
  const lifetime = useMemo(
    () =>
      createMayaSessionGuard(scope, useAuthStore.getState, provider.isActive),
    [scope, provider],
  );

  // This namespace deliberately stays outside the persisted "session" cache.
  // Private conversation titles and transcript bodies remain in memory only.
  const sessionListKey = useMemo(
    () => ["maya-sessions", scope.userId, scope.workspace, scope.sessionEpoch],
    [scope],
  );
  const {
    data: sessions = [],
    isFetching: isSessionsLoading,
    error: sessionsError,
    refetch,
  } = useQuery({
    queryKey: sessionListKey,
    queryFn: async ({ signal }) => {
      lifetime.assertCurrent();
      const result = await getMayaSessions(scope.userId, signal);
      lifetime.assertCurrent();
      return result;
    },
    retry: false,
  });
  const refreshSessions = useCallback(async () => {
    lifetime.assertCurrent();
    const result = await refetch();
    if (result.error) throw result.error;
  }, [lifetime, refetch]);

  const runtime = useMemo(() => {
    const guard = createMayaSessionGuard(
      scope,
      useAuthStore.getState,
      provider.isActive,
    );
    const cloud = createMayaCloudClient({
      applicationURL: getApplicationURL(),
      apiOrigin: getApiURL().origin,
      scope,
      assertCurrent: guard.assertCurrent,
      isCurrent: guard.isCurrent,
      getSession: getStoredSession,
      expireSession: (cookie) => useAuthStore.getState().expireSession(cookie),
      fetch: expoFetch as typeof globalThis.fetch,
    });
    return createMayaChatRuntime({
      id: selection.id,
      messages: selection.messages,
      persisted: selection.persisted,
      workspace: scope.workspace,
      cloud,
      guard,
      getHistory: (signal) => getMayaMessages(selection.id, signal),
      onFinish: (message, interrupted) => {
        void queryClient.invalidateQueries({ queryKey: sessionListKey });
        if (hasMayaMutationReceipt(message)) {
          void queryClient.invalidateQueries({
            queryKey: ["session", scope.userId, scope.workspace],
          });
        }
        if (interrupted) return;
        for (const action of getMayaClientActions(message)) {
          if (action.type === "navigate") router.push(action.href);
        }
      },
    });
  }, [selection, scope, provider, sessionListKey, queryClient, router]);
  const {
    messages,
    status,
    error: chatError,
    clearError: clearChatError,
  } = useChat({ chat: runtime.chat, experimental_throttle: 50 });
  const voiceHistory = useMemo(
    () =>
      createVoiceHistoryQueue({
        assertCurrent: runtime.guard.assertCurrent,
        persist: async (batch) => {
          provider.assertCanMutate();
          await runtime.persistVoiceMessages(batch);
          void queryClient.invalidateQueries({ queryKey: sessionListKey });
        },
      }),
    [runtime, provider, queryClient, sessionListKey],
  );
  useFocusEffect(
    useCallback(
      () => () => {
        void runtime
          .stop()
          .then(() => {
            if (runtime.guard.isCurrent()) return voiceHistory.flush(true);
          })
          .catch((error: unknown) => {
            if (runtime.guard.isCurrent()) setOperationError(asError(error));
          });
      },
      [runtime, voiceHistory],
    ),
  );

  useEffect(() => {
    const release = runtime.retain();
    const unsubscribe = useAuthStore.subscribe(() => {
      if (lifetime.isCurrent()) return;
      lifetime.dispose();
      runtime.dispose();
      void queryClient.cancelQueries({ queryKey: sessionListKey });
      selectionRequest.current?.abort();
      deletionRequest.current?.abort();
    });
    const appState = AppState.addEventListener("change", (state) => {
      if (state !== "active") void runtime.stop();
    });
    return () => {
      unsubscribe();
      appState.remove();
      release();
    };
  }, [runtime, lifetime, queryClient, sessionListKey]);
  useEffect(() => {
    const release = lifetime.retain();
    return () => {
      release();
      void queryClient.cancelQueries({ queryKey: sessionListKey });
      selectionRequest.current?.abort();
      deletionRequest.current?.abort();
    };
  }, [lifetime, queryClient, sessionListKey]);

  const requestBody = async (signal: AbortSignal) => ({
    ...(await loadMayaRequestContext({
      scope,
      queryClient,
      signal,
      assertCurrent: runtime.guard.assertCurrent,
    })),
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    currentPath: getMayaScreenPath(context.storyReference),
    currentTheme: theme,
    resolvedTheme,
    ...(isEntityId(context.storyId)
      ? { screenContext: { storyId: context.storyId } }
      : {}),
  });
  const perform = (operation: () => Promise<void>) =>
    runMayaOperation(operation, {
      assertCurrent: lifetime.assertCurrent,
      isCurrent: lifetime.isCurrent,
      isSelecting: selectionRequest.current !== null,
      onError: setOperationError,
    });
  const newChat = async () => {
    lifetime.assertCurrent();
    const version = ++navigationVersion.current;
    selectionRequest.current?.abort();
    selectionRequest.current = null;
    await runtime.stop();
    if (!lifetime.isCurrent() || version !== navigationVersion.current) return;
    runtime.dispose();
    setOperationError(null);
    setHistoryLoading(false);
    setSelection({ id: generateId(), messages: [], persisted: false });
  };
  const selectChat = async (id: string) => {
    lifetime.assertCurrent();
    const version = ++navigationVersion.current;
    selectionRequest.current?.abort();
    const request = new AbortController();
    selectionRequest.current = request;
    setHistoryLoading(true);
    setOperationError(null);
    await runtime
      .stop()
      .then(async () => {
        lifetime.assertCurrent();
        const history = await getMayaMessages(id, request.signal);
        if (
          !lifetime.isCurrent() ||
          request.signal.aborted ||
          version !== navigationVersion.current
        )
          return;
        runtime.dispose();
        setSelection({ id, messages: history, persisted: true });
      })
      .catch((error: unknown) => {
        if (!request.signal.aborted && lifetime.isCurrent()) {
          setOperationError(asError(error));
          throw error;
        }
      })
      .finally(() => {
        if (selectionRequest.current === request && lifetime.isCurrent()) {
          selectionRequest.current = null;
          setHistoryLoading(false);
        }
      });
  };
  const deleteChat = (id: string) =>
    perform(async () => {
      provider.assertCanMutate();
      if (deletionRequest.current)
        throw new Error("Wait for the conversation deletion to finish.");
      const request = new AbortController();
      deletionRequest.current = request;
      const version = navigationVersion.current;
      if (id === selection.id) await runtime.stop();
      lifetime.assertCurrent();
      const unsubscribe = useAuthStore.subscribe(() => {
        if (!lifetime.isCurrent()) request.abort();
      });
      await deleteMayaSession(id, request.signal).finally(() => {
        unsubscribe();
        if (deletionRequest.current === request) deletionRequest.current = null;
      });
      lifetime.assertCurrent();
      if (id === selection.id && version === navigationVersion.current)
        await newChat();
      await refreshSessions();
    });

  return {
    messages,
    sessions,
    currentChatId: selection.id,
    status,
    error: operationError ?? chatError ?? sessionsError ?? null,
    isSessionsLoading,
    isHistoryLoading,
    pendingApprovals: getPendingMayaApprovals(messages),
    send: (text: string, files?: FileUIPart[]) =>
      perform(() => {
        provider.assertCanMutate();
        return runtime.send(text, requestBody, files);
      }),
    approve: (id: string, approved: boolean) =>
      perform(() => {
        provider.assertCanMutate();
        return runtime.approve(id, approved, requestBody);
      }),
    recordVoiceMessage: (message: VoiceHistoryMessage) => {
      voiceHistory.add(message);
    },
    flushVoiceHistory: () => perform(() => voiceHistory.flush(true)),
    stop: runtime.stop,
    newChat,
    selectChat,
    deleteChat,
    reloadHistory: () => selectChat(selection.id),
    refreshSessions,
    clearError: () => {
      setOperationError(null);
      clearChatError();
    },
  };
};
