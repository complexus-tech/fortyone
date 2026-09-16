import type { UIMessage } from "ai";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  ActivityIndicator,
  AppState,
  FlatList,
  Keyboard,
  Pressable,
  StyleSheet,
  View,
} from "react-native";
import {
  useFocusEffect,
  useLocalSearchParams,
  useNavigation,
  useRouter,
} from "expo-router";
import { Ionicons } from "@expo/vector-icons";
import { KeyboardAvoidingView } from "react-native-keyboard-controller";
import { SafeAreaView as NativeTabSafeAreaView } from "react-native-screens/experimental";
import { Back, SafeContainer, Text, HeaderActions } from "@/components/ui";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { useStatuses } from "@/modules/statuses/hooks/use-statuses";
import { useMembers } from "@/modules/members/hooks/use-members";
import { useAuthStore } from "@/store/auth";
import { useMayaChat } from "./hooks/use-maya-chat";
import { useMayaVoice } from "./hooks/use-maya-voice";
import { useMayaAttachments } from "./hooks/use-maya-attachments";
import { useMayaDictation } from "./hooks/use-maya-dictation";
import { sendMayaDraft } from "./lib/send-draft";
import {
  captureMayaResponse,
  selectMayaResponseRows,
  type MayaResponseRow,
} from "./lib/response-presentation";
import { MayaMessage } from "./components/message";
import { MayaThinking } from "./components/thinking";
import { MayaHistory } from "./components/history";
import { ApprovalCard } from "./components/approval-card";
import { VoiceControls } from "./components/voice-controls";
import { MayaWelcome } from "./components/welcome";
import { MayaComposer } from "./components/composer";
import { MayaBackground } from "./components/gradients";
import { resultStories, record } from "./components/message-model";

export function Maya() {
  const sessionIdentity = useAuthStore((state) =>
    state.isAuthenticated && !state.isLoading
      ? `${state.userId}:${state.workspace}:${state.sessionEpoch}`
      : null,
  );
  if (!sessionIdentity) return null;
  return <MayaContent key={sessionIdentity} />;
}

function MayaContent() {
  const router = useRouter();
  const navigation = useNavigation();
  const { resolvedTheme } = useTheme();
  const theme = themeColors[resolvedTheme];
  const params = useLocalSearchParams<{
    storyId?: string;
    storyReference?: string;
    storyTitle?: string;
  }>();
  const chat = useMayaChat({
    storyId: params.storyId,
    storyReference: params.storyReference,
  });
  const attachments = useMayaAttachments(chat.currentChatId);
  const conversationMessages = useMemo(
    () =>
      chat.messages.flatMap((message) => {
        if (message.role !== "assistant" && message.role !== "user") return [];
        const text = message.parts
          .filter((part) => part.type === "text")
          .map((part) => part.text)
          .join("\n");
        return text ? [{ id: message.id, role: message.role, text }] : [];
      }),
    [chat.messages],
  );
  const voice = useMayaVoice({
    currentPath: params.storyReference
      ? `/work/${params.storyReference}`
      : "/my-work",
    conversationMessages,
    onTranscriptFinalized: chat.recordVoiceMessage,
  });
  const [draft, setDraft] = useState("");
  const appendDictation = useCallback((text: string) => {
    setDraft((current) =>
      current.trim() ? `${current.trimEnd()} ${text}` : text,
    );
  }, []);
  const dictation = useMayaDictation({
    scopeKey: chat.currentChatId,
    onText: appendDictation,
  });
  const [historyOpen, setHistoryOpen] = useState(false);
  const [uiError, setUiError] = useState<string | null>(null);
  const [approving, setApproving] = useState(false);
  const [isSending, setSending] = useState(false);
  const [response, setResponse] = useState<ReturnType<
    typeof captureMayaResponse
  > | null>(null);
  const [isChangingConversation, setChangingConversation] = useState(false);
  const listRef = useRef<FlatList<MayaResponseRow>>(null);
  const following = useRef(true);
  const submitPending = useRef(false);
  const sendGeneration = useRef(0);
  const transitionPending = useRef(false);
  const busy =
    isSending || chat.status === "submitted" || chat.status === "streaming";
  const voiceActive = voice.status !== "idle";
  const dictationActive = dictation.status !== "idle";
  const wasVoiceActive = useRef(false);
  const { data: teams = [] } = useTeams();
  const { data: statuses = [] } = useStatuses();
  const { data: members = [] } = useMembers();
  const names = useMemo(() => {
    const values = new Map<string, string>();
    for (const team of teams) values.set(team.id, team.name);
    for (const status of statuses) values.set(status.id, status.name);
    for (const member of members)
      values.set(member.id, member.fullName || member.username);
    for (const message of chat.messages)
      for (const part of message.parts) {
        for (const story of resultStories(record(part)?.output))
          values.set(
            story.id,
            `${story.reference ? `${story.reference} · ` : ""}${story.title}`,
          );
      }
    if (params.storyId && params.storyTitle)
      values.set(params.storyId, params.storyTitle);
    return values;
  }, [
    teams,
    statuses,
    members,
    chat.messages,
    params.storyId,
    params.storyTitle,
  ]);
  const messages = useMemo(() => {
    const seen = new Set(chat.messages.map((message) => message.id));
    const spoken: UIMessage[] = voice.transcript
      .filter((message) => !seen.has(message.id))
      .map((message) => ({
        id: message.id,
        role: message.role,
        parts: [{ type: "text", text: message.text }],
      }));
    return [...chat.messages, ...spoken];
  }, [chat.messages, voice.transcript]);
  const responseRows = useMemo(
    () =>
      selectMayaResponseRows({
        chatId: chat.currentChatId,
        messages,
        response,
        active: chat.status === "submitted" || chat.status === "streaming",
      }),
    [chat.currentChatId, chat.status, messages, response],
  );
  const error =
    uiError ?? chat.error?.message ?? dictation.error ?? voice.error;
  const report = useCallback(
    (cause: unknown) =>
      setUiError(cause instanceof Error ? cause.message : "Please try again."),
    [],
  );
  useFocusEffect(
    useCallback(
      () => () => {
        sendGeneration.current++;
      },
      [],
    ),
  );
  useEffect(() => {
    const subscription = AppState.addEventListener("change", (state) => {
      if (state !== "active") sendGeneration.current++;
    });
    return () => subscription.remove();
  }, []);
  useEffect(() => {
    const ended = wasVoiceActive.current && !voiceActive;
    wasVoiceActive.current = voiceActive;
    if (ended && !transitionPending.current)
      void chat.flushVoiceHistory().catch(report);
  }, [voiceActive, chat, report]);
  const send = useCallback(
    async (text: string) => {
      if (
        (!text.trim() && attachments.attachments.length === 0) ||
        attachments.picking ||
        submitPending.current ||
        transitionPending.current ||
        busy ||
        voiceActive ||
        dictationActive ||
        isChangingConversation
      )
        return;
      submitPending.current = true;
      const generation = ++sendGeneration.current;
      const isCurrent = () => generation === sendGeneration.current;
      setResponse(
        captureMayaResponse({
          chatId: chat.currentChatId,
          requestId: String(generation),
          mode: "send",
          messages: chat.messages,
        }),
      );
      Keyboard.dismiss();
      setSending(true);
      setUiError(null);
      following.current = true;
      await sendMayaDraft({
        text: text.trim(),
        prepare: attachments.prepare,
        flushVoiceHistory: chat.flushVoiceHistory,
        send: chat.send,
        isCurrent,
        onSent: () => {
          attachments.clear();
          setDraft("");
        },
      })
        .catch((cause: unknown) => {
          if (isCurrent()) report(cause);
        })
        .finally(() => {
          submitPending.current = false;
          setSending(false);
        });
    },
    [
      busy,
      voiceActive,
      dictationActive,
      isChangingConversation,
      chat,
      attachments,
      report,
    ],
  );
  const approve = useCallback(
    (id: string, approved: boolean) => {
      if (
        busy ||
        approving ||
        submitPending.current ||
        transitionPending.current
      )
        return;
      if (voiceActive || dictationActive || isChangingConversation) {
        setUiError(
          "Finish dictation or end your voice conversation before confirming a change.",
        );
        return;
      }
      setResponse(
        captureMayaResponse({
          chatId: chat.currentChatId,
          requestId: String(++sendGeneration.current),
          mode: "approval",
          messages: chat.messages,
        }),
      );
      setApproving(true);
      void chat
        .approve(id, approved)
        .catch(report)
        .finally(() => setApproving(false));
    },
    [
      chat,
      busy,
      approving,
      voiceActive,
      dictationActive,
      isChangingConversation,
      report,
    ],
  );
  const canSwitch = () => {
    if (isChangingConversation || attachments.picking) return false;
    if (voiceActive) {
      setUiError("End your voice conversation before switching chats.");
      return false;
    }
    return true;
  };
  const runConversationTransition = (operation: () => Promise<void> | void) => {
    if (transitionPending.current) return Promise.resolve(false);
    sendGeneration.current++;
    transitionPending.current = true;
    setChangingConversation(true);
    setUiError(null);
    voice.stop();
    return dictation
      .cancel()
      .then(() => chat.stop())
      .then(() => chat.flushVoiceHistory())
      .then(operation)
      .then(() => true)
      .catch((cause: unknown) => {
        report(cause);
        return false;
      })
      .finally(() => {
        transitionPending.current = false;
        setChangingConversation(false);
      });
  };
  const leaveMaya = () => {
    Keyboard.dismiss();
    void runConversationTransition(() => {
      if (router.canGoBack()) {
        router.back();
        return;
      }
      const state = navigation.getState();
      const homeIndex =
        state?.routes.findIndex((route) => route.name === "index") ?? -1;
      if (state?.type === "tab" && homeIndex >= 0) {
        // Direct links have no previous tab. Reset just the tab history so
        // Android Back cannot reopen Maya after returning home.
        navigation.dispatch({
          type: "RESET",
          target: state.key,
          payload: {
            ...state,
            index: homeIndex,
            history: [{ type: "route", key: state.routes[homeIndex].key }],
          },
        });
        return;
      }
      router.replace("/");
    });
  };
  const beginDictation = () => {
    if (dictation.status === "recording") {
      void dictation.finish().catch(report);
      return;
    }
    if (
      dictationActive ||
      busy ||
      voiceActive ||
      attachments.picking ||
      submitPending.current ||
      transitionPending.current
    )
      return;
    Keyboard.dismiss();
    setUiError(null);
    dictation.clearError();
    void dictation.start().catch(report);
  };
  const beginVoice = () => {
    if (
      voiceActive ||
      attachments.picking ||
      submitPending.current ||
      transitionPending.current
    )
      return;
    if (chat.pendingApprovals.length) {
      setUiError(
        "Confirm or cancel the proposed text change before starting voice.",
      );
      return;
    }
    Keyboard.dismiss();
    setUiError(null);
    void runConversationTransition(() => voice.start());
  };
  const renderMessage = useCallback(
    ({ item }: { item: MayaResponseRow }) =>
      item.kind === "thinking" ? (
        <View style={styles.response}>
          <MayaThinking />
        </View>
      ) : (
        <MayaMessage
          message={item.message}
          names={names}
          statuses={statuses}
          allowApproval={item.message.id === chat.messages.at(-1)?.id}
          busy={
            approving ||
            busy ||
            voiceActive ||
            dictationActive ||
            isChangingConversation
          }
          onApprove={approve}
        />
      ),
    [
      names,
      statuses,
      chat.messages,
      approving,
      busy,
      voiceActive,
      dictationActive,
      isChangingConversation,
      approve,
    ],
  );
  return (
    <SafeContainer isFull edges={["top"]}>
      <MayaBackground />
      <NativeTabSafeAreaView
        style={{ flex: 1 }}
        edges={{ top: false, bottom: true, left: false, right: false }}
      >
        <View style={styles.header}>
          <Back onPress={leaveMaya} />
          <HeaderActions
            createLabel="New conversation"
            onCreate={() => {
              if (!canSwitch()) return;
              runConversationTransition(async () => {
                await chat.newChat();
                voice.clearTranscript();
                setDraft("");
              });
            }}
            optionsIcon="time-outline"
            optionsSystemImage="clock.arrow.circlepath"
            onOptions={() => {
              if (canSwitch())
                void dictation
                  .cancel()
                  .then(() => setHistoryOpen(true))
                  .catch(report);
            }}
            optionsLabel="Conversation history"
          />
        </View>
        {params.storyReference ? (
          <View style={[styles.context, { borderColor: theme.border }]}>
            <Ionicons
              name="document-text-outline"
              size={14}
              color={theme.textMuted}
            />
            <Text
              fontSize="sm"
              color="muted"
              numberOfLines={1}
              style={{ flex: 1 }}
            >
              {params.storyReference}
              {params.storyTitle ? ` · ${params.storyTitle}` : ""}
            </Text>
          </View>
        ) : null}
        <KeyboardAvoidingView
          behavior="padding"
          automaticOffset
          style={{ flex: 1 }}
        >
          {chat.isHistoryLoading ? (
            <View style={styles.center}>
              <ActivityIndicator color={theme.textMuted} />
              <Text color="muted">Loading conversation…</Text>
            </View>
          ) : !messages.length && !voiceActive && !busy ? (
            <MayaWelcome
              disabled={
                isChangingConversation || attachments.picking || dictationActive
              }
              onPrompt={(prompt) => void send(prompt)}
            />
          ) : (
            <FlatList
              ref={listRef}
              data={responseRows}
              keyExtractor={(item) => item.id}
              renderItem={renderMessage}
              keyboardDismissMode="interactive"
              keyboardShouldPersistTaps="handled"
              contentContainerStyle={{
                paddingTop: 8,
                paddingBottom: 20,
                flexGrow: 1,
              }}
              onScroll={({ nativeEvent }) => {
                following.current =
                  nativeEvent.contentSize.height -
                    nativeEvent.layoutMeasurement.height -
                    nativeEvent.contentOffset.y <
                  100;
              }}
              scrollEventThrottle={64}
              onContentSizeChange={() => {
                if (following.current)
                  listRef.current?.scrollToEnd({ animated: false });
              }}
            />
          )}
          {error ? (
            <View
              style={[styles.error, { backgroundColor: theme.surfaceMuted }]}
            >
              <Text color="danger" fontSize="sm" style={{ flex: 1 }}>
                {typeof error === "string" ? error : "Voice could not connect."}
              </Text>
              <Pressable
                accessibilityLabel="Dismiss error"
                accessibilityRole="button"
                onPress={() => {
                  setUiError(null);
                  chat.clearError();
                  voice.clearError();
                  dictation.clearError();
                }}
                style={styles.smallButton}
              >
                <Ionicons name="close" size={18} color={theme.textMuted} />
              </Pressable>
            </View>
          ) : null}
          {voice.pendingAction ? (
            <View style={{ paddingHorizontal: 20, paddingVertical: 8 }}>
              <ApprovalCard
                title={voice.pendingAction.title}
                description={voice.pendingAction.description}
                details={voice.pendingAction.details}
                destructive={voice.pendingAction.isDestructive}
                busy={voice.isApproving}
                onConfirm={() => void voice.approveAction().catch(report)}
                onCancel={voice.cancelAction}
              />
            </View>
          ) : null}
          <View
            style={{
              paddingBottom: 12,
              paddingHorizontal: 16,
              paddingTop: 8,
            }}
          >
            {voiceActive ? (
              <VoiceControls
                status={voice.status}
                muted={voice.isMuted}
                speaking={voice.isSpeaking}
                remainingSeconds={voice.remainingSeconds}
                onMute={voice.toggleMute}
                onStop={() => runConversationTransition(() => undefined)}
              />
            ) : (
              <MayaComposer
                value={draft}
                attachments={attachments.attachments}
                picking={attachments.picking || attachments.preparing}
                dictationStatus={dictation.status}
                dictationSeconds={dictation.seconds}
                onDictation={beginDictation}
                onCancelDictation={() => void dictation.cancel().catch(report)}
                onAttach={() => void attachments.pick().catch(report)}
                onRemoveAttachment={attachments.remove}
                onChange={setDraft}
                busy={busy}
                disabled={chat.isHistoryLoading || isChangingConversation}
                onVoice={beginVoice}
                onStop={() => {
                  sendGeneration.current++;
                  void chat.stop().catch(report);
                }}
                onSend={() => void send(draft)}
              />
            )}
          </View>
        </KeyboardAvoidingView>
      </NativeTabSafeAreaView>
      <MayaHistory
        visible={historyOpen}
        onClose={() => setHistoryOpen(false)}
        sessions={chat.sessions}
        loading={chat.isSessionsLoading}
        currentId={chat.currentChatId}
        onSelect={(id) => {
          if (id === chat.currentChatId) {
            setHistoryOpen(false);
            return;
          }
          runConversationTransition(async () => {
            await chat.selectChat(id);
            voice.clearTranscript();
            setDraft("");
            setHistoryOpen(false);
          });
        }}
        onDelete={(id) => {
          const deletingCurrent = id === chat.currentChatId;
          runConversationTransition(async () => {
            await chat.deleteChat(id);
            if (deletingCurrent) {
              voice.clearTranscript();
              setDraft("");
            }
          });
        }}
      />
    </SafeContainer>
  );
}

const styles = StyleSheet.create({
  header: {
    minHeight: 56,
    paddingHorizontal: 20,
    marginBottom: 12,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
  },
  context: {
    marginHorizontal: 20,
    paddingVertical: 10,
    borderBottomWidth: StyleSheet.hairlineWidth,
    flexDirection: "row",
    gap: 7,
    alignItems: "center",
  },
  center: { flex: 1, alignItems: "center", justifyContent: "center", gap: 12 },
  response: {
    paddingHorizontal: 20,
    paddingVertical: 12,
    // Reserve the same trailing space as the first Markdown paragraph.
    paddingBottom: 22,
  },
  error: {
    flexDirection: "row",
    alignItems: "center",
    marginHorizontal: 16,
    borderRadius: 16,
    paddingLeft: 14,
    paddingVertical: 6,
  },
  smallButton: {
    width: 44,
    height: 44,
    justifyContent: "center",
    alignItems: "center",
  },
});
