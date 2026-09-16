import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useState,
  useSyncExternalStore,
} from "react";
import { AppState } from "react-native";
import { useFocusEffect } from "expo-router";
import { fetch as expoFetch } from "expo/fetch";
import { File } from "expo-file-system";
import {
  RecordingPresets,
  requestRecordingPermissionsAsync,
  setAudioModeAsync,
  useAudioRecorder,
  type AudioRecorder,
} from "expo-audio";
import { getStoredSession } from "@/lib/auth";
import { getApiURL, getApplicationURL } from "@/lib/http/config";
import { useAuthStore } from "@/store/auth";
import { createMayaCloudClient } from "../lib/cloud-client";
import {
  createMayaDictation,
  isCurrentDictationEvent,
  MAYA_DICTATION_SECONDS,
} from "../lib/dictation";
import { matchesMayaSession } from "../lib/session-scope";

const createNativeRecorder = (recorder: AudioRecorder) => {
  let ownsAudioMode = false;
  let uri: string | null = null;
  const ownedUris = new Set<string>();
  return {
    getUri: () => uri,
    prepare: async () => {
      uri = null;
      await setAudioModeAsync({
        allowsRecording: true,
        playsInSilentMode: true,
        shouldPlayInBackground: false,
        allowsBackgroundRecording: false,
      });
      ownsAudioMode = true;
      try {
        // Passing the options creates a fresh native recording file per run.
        await recorder.prepareToRecordAsync(RecordingPresets.HIGH_QUALITY);
      } finally {
        uri = recorder.uri;
        if (uri) ownedUris.add(uri);
      }
    },
    record: () => recorder.record({ forDuration: MAYA_DICTATION_SECONDS }),
    stop: async () => {
      if (ownsAudioMode) {
        try {
          await recorder.stop();
        } finally {
          await setAudioModeAsync({ allowsRecording: false });
          ownsAudioMode = false;
        }
      }
      return uri;
    },
    release: (recordingUri: string) => {
      if (!ownedUris.has(recordingUri)) return;
      const file = new File(recordingUri);
      if (file.exists) file.delete();
      ownedUris.delete(recordingUri);
    },
  };
};

const createDictationEnvironment = (
  initialScope: string,
  initialOnText: (text: string) => void,
) => {
  let scopeKey = initialScope;
  let onText = initialOnText;
  let focused = true;
  return {
    getScopeKey: () => scopeKey,
    isFocused: () => focused,
    onText: (text: string) => onText(text),
    focus: (value: boolean) => {
      focused = value;
    },
    update: (nextScope: string, nextOnText: (text: string) => void) => {
      const changed = nextScope !== scopeKey;
      scopeKey = nextScope;
      onText = nextOnText;
      return changed;
    },
  };
};

export const useMayaDictation = ({
  scopeKey,
  onText,
}: {
  scopeKey: string;
  onText: (text: string) => void;
}) => {
  const recorder = useAudioRecorder(RecordingPresets.HIGH_QUALITY);
  const [environment] = useState(() =>
    createDictationEnvironment(scopeKey, onText),
  );
  const [scope] = useState(() => {
    const state = useAuthStore.getState();
    return {
      userId: state.userId ?? "",
      workspace: state.workspace ?? "",
      sessionEpoch: state.sessionEpoch,
    };
  });
  const native = useMemo(() => createNativeRecorder(recorder), [recorder]);
  const cloud = useMemo(() => {
    const isCurrent = () =>
      environment.isFocused() &&
      matchesMayaSession(scope, useAuthStore.getState());
    return createMayaCloudClient({
      applicationURL: getApplicationURL(),
      apiOrigin: getApiURL().origin,
      scope,
      isCurrent,
      assertCurrent: () => {
        if (!isCurrent()) throw new Error("Dictation session changed.");
      },
      getSession: getStoredSession,
      expireSession: (cookie) => useAuthStore.getState().expireSession(cookie),
      fetch: expoFetch as typeof globalThis.fetch,
    });
  }, [scope, environment]);
  const controller = useMemo(
    () =>
      createMayaDictation({
        getScopeKey: environment.getScopeKey,
        isCurrent: () =>
          environment.isFocused() &&
          matchesMayaSession(scope, useAuthStore.getState()),
        permission: async () =>
          (await requestRecordingPermissionsAsync()).granted,
        ...native,
        transcribe: async (uri, signal) => {
          const file = new File(uri);
          if (!file.exists || file.size <= 0)
            throw new Error("No recording was captured. Try dictation again.");
          if (file.size > 5 * 1024 * 1024)
            throw new Error(
              "The recording is too large. Dictate a shorter message.",
            );
          // Expo's multipart encoder consumes File.bytes(), unlike RN's URI-only
          // objects. Supply the known recorder format even if native MIME is empty.
          const audio = new Proxy(file, {
            get(target, property) {
              if (property === "type") return "audio/mp4";
              const value = Reflect.get(target, property, target);
              return typeof value === "function" ? value.bind(target) : value;
            },
          });
          const body = new FormData();
          body.append("audio", audio);
          const response = await cloud.fetch(cloud.url("/api/transcribe"), {
            method: "POST",
            body,
            signal,
          });
          const result: unknown = await response.json();
          if (
            !result ||
            typeof result !== "object" ||
            !("text" in result) ||
            typeof result.text !== "string"
          )
            throw new Error(
              "Maya returned an invalid transcript. Please try again.",
            );
          return result.text;
        },
        onText: environment.onText,
      }),
    [cloud, native, scope, environment],
  );
  const snapshot = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
    controller.getSnapshot,
  );
  useLayoutEffect(() => {
    if (environment.update(scopeKey, onText)) {
      void controller.cancel().catch(() => undefined);
    }
  }, [scopeKey, onText, controller, environment]);
  useFocusEffect(
    useCallback(() => {
      environment.focus(true);
      return () => {
        environment.focus(false);
        void controller.cancel().catch(() => undefined);
      };
    }, [controller, environment]),
  );
  useEffect(() => {
    const appState = AppState.addEventListener("change", (state) => {
      // A permission dialog may make iOS inactive; leaving the app still cancels.
      if (
        state === "background" ||
        (state === "inactive" &&
          controller.getSnapshot().status !== "preparing")
      )
        void controller.cancel().catch(() => undefined);
    });
    const unsubscribe = useAuthStore.subscribe((state) => {
      if (!matchesMayaSession(scope, state))
        void controller.cancel().catch(() => undefined);
    });
    const recording = recorder.addListener("recordingStatusUpdate", (event) => {
      if (controller.getSnapshot().status !== "recording") return;
      // Android may enqueue an old completion event after another run starts.
      if (
        !isCurrentDictationEvent(event, native.getUri(), recorder.isRecording)
      )
        return;
      const operation =
        event.hasError || event.mediaServicesDidReset
          ? controller.interrupt()
          : event.isFinished
            ? controller.finish()
            : undefined;
      void operation?.catch(() => undefined);
    });
    return () => {
      appState.remove();
      recording.remove();
      unsubscribe();
      cloud.abort();
      void controller.cancel().catch(() => undefined);
    };
  }, [cloud, controller, native, recorder, scope]);
  useEffect(() => {
    if (snapshot.status !== "recording") return;
    const started = Date.now();
    const timer = setInterval(() => {
      void controller
        .progress((Date.now() - started) / 1000)
        .catch(() => undefined);
    }, 250);
    return () => clearInterval(timer);
  }, [controller, snapshot.status]);
  return {
    ...snapshot,
    start: controller.start,
    finish: controller.finish,
    cancel: controller.cancel,
    clearError: controller.clearError,
  };
};
