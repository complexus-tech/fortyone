import {
  prepareVoiceTool,
  redactVoiceToolResult,
  voiceConversationContext,
  voicePendingAction,
  voiceTranscriptUpdate,
  type VoiceEvent,
  type VoicePendingAction,
  type VoiceToolCall,
  type VoiceToolResult,
  type VoiceTranscriptMessage,
} from "./voice-protocol";

export type VoiceIdentity = {
  userId: string;
  workspace: string;
  sessionEpoch: number;
  cookie: string;
};
export type VoiceLease = {
  sessionId: string;
  clientSecret: string;
  maxSessionSeconds: number;
};
export type VoiceTransport = {
  send: (event: unknown) => void;
  close: () => void;
  mute: (muted: boolean) => void;
};
export type VoiceState = {
  status: "idle" | "connecting" | "connected";
  isMuted: boolean;
  isSpeaking: boolean;
  error: string | null;
  remainingSeconds: number | null;
  transcript: VoiceTranscriptMessage[];
  pendingAction: VoicePendingAction | null;
  isApproving: boolean;
};
export type VoiceContext = {
  currentPath?: string;
  conversationMessages?: readonly (Pick<
    VoiceTranscriptMessage,
    "role" | "text"
  > & { id?: string })[];
  onTranscriptFinalized?: (
    message: VoiceTranscriptMessage,
  ) => void | Promise<void>;
};
export type VoiceDependencies = {
  prepareTransport?: () => void;
  captureIdentity: () => Promise<VoiceIdentity>;
  isCurrent: (identity: VoiceIdentity) => boolean;
  startSession: (
    identity: VoiceIdentity,
    context: {
      currentPath: string;
      messages: ReturnType<typeof voiceConversationContext>;
    },
    signal: AbortSignal,
  ) => Promise<VoiceLease>;
  endSession: (identity: VoiceIdentity, sessionId: string) => Promise<void>;
  tool: (
    identity: VoiceIdentity,
    input: {
      sessionId: string;
      callId: string;
      name: string;
      arguments: Record<string, unknown>;
    },
    signal: AbortSignal,
  ) => Promise<VoiceToolResult>;
  openTransport: (
    lease: VoiceLease,
    signal: AbortSignal,
    callbacks: {
      onEvent: (event: VoiceEvent) => void;
      onOpen: () => void;
      onFailure: (error: Error) => void;
    },
  ) => Promise<VoiceTransport>;
  invalidate: (identity: VoiceIdentity, tool: string) => void;
  clientAction: (identity: VoiceIdentity, action: unknown) => Promise<boolean>;
};
type ActiveVoice = {
  identity: VoiceIdentity | null;
  lease: VoiceLease | null;
  abort: AbortController;
  transport: VoiceTransport | null;
  calls: Set<string>;
  orders: Map<string, number>;
  finalized: Set<string>;
  timers: ReturnType<typeof setTimeout>[];
  idleTimer?: ReturnType<typeof setTimeout>;
  preparingCall?: string;
  pending?: {
    call: VoiceToolCall;
    args: Record<string, unknown>;
    token: string;
    approvalAttempted?: boolean;
  };
  context: VoiceContext;
};

const MAX_VOICE_SESSION_SECONDS = 5 * 60;

const initialState = (): VoiceState => ({
  status: "idle",
  isMuted: false,
  isSpeaking: false,
  error: null,
  remainingSeconds: null,
  transcript: [],
  pendingAction: null,
  isApproving: false,
});
const errorMessage = (error: unknown) =>
  error instanceof Error
    ? error.message
    : "Voice could not complete this request. Please try again.";

/** The controller owns the session, timers and secrets; React only subscribes to safe UI state. */
export class MayaVoiceController {
  private state = initialState();
  private listeners = new Set<() => void>();
  private active: ActiveVoice | null = null;
  private owner: VoiceIdentity | null = null;
  constructor(private readonly deps: VoiceDependencies) {}
  getSnapshot = () => this.state;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private update(change: Partial<VoiceState>) {
    this.state = { ...this.state, ...change };
    this.listeners.forEach((listener) => listener());
  }
  private current(active: ActiveVoice) {
    return (
      this.active === active &&
      !active.abort.signal.aborted &&
      (!active.identity || this.deps.isCurrent(active.identity))
    );
  }
  private send(active: ActiveVoice, event: unknown) {
    if (this.current(active)) active.transport?.send(event);
  }
  private output(active: ActiveVoice, callId: string, output: unknown) {
    this.send(active, {
      type: "conversation.item.create",
      item: {
        type: "function_call_output",
        call_id: callId,
        output: JSON.stringify(redactVoiceToolResult(output)),
      },
    });
    this.send(active, { type: "response.create" });
  }
  private touch(active: ActiveVoice) {
    clearTimeout(active.idleTimer);
    active.idleTimer = setTimeout(() => {
      if (this.current(active)) this.stop();
    }, 60_000);
  }
  start = async (context: VoiceContext = {}) => {
    if (this.active) return;
    const active: ActiveVoice = {
      identity: null,
      lease: null,
      abort: new AbortController(),
      transport: null,
      calls: new Set(),
      orders: new Map(),
      finalized: new Set(),
      timers: [],
      context,
    };
    this.active = active;
    this.update({
      status: "connecting",
      error: null,
      pendingAction: null,
      isApproving: false,
    });
    active.timers.push(
      setTimeout(() => {
        if (this.current(active))
          this.fail(
            active,
            new Error(
              "Voice connection timed out. Check your connection and try again.",
            ),
          );
      }, 30_000),
    );
    try {
      this.deps.prepareTransport?.();
      active.identity = await this.deps.captureIdentity();
      if (!this.current(active)) {
        this.stopIfOwned(active);
        return;
      }
      this.owner = active.identity;
      const startedAt = Date.now();
      const lease = await this.deps.startSession(
        active.identity,
        {
          currentPath: (context.currentPath ?? "/maya").slice(0, 512),
          messages: voiceConversationContext([
            ...(context.conversationMessages ?? []),
            ...this.state.transcript,
          ]),
        },
        active.abort.signal,
      );
      active.lease = lease;
      if (!this.current(active)) {
        void this.deps
          .endSession(active.identity, lease.sessionId)
          .catch(() => undefined);
        this.stopIfOwned(active);
        return;
      }
      if (
        !lease.sessionId ||
        !lease.clientSecret ||
        !Number.isFinite(lease.maxSessionSeconds) ||
        lease.maxSessionSeconds <= 0
      ) {
        throw new Error("The server returned an invalid voice session.");
      }
      const expiresAt =
        startedAt +
        Math.min(lease.maxSessionSeconds, MAX_VOICE_SESSION_SECONDS) * 1000;
      // Keep the hard cutoff separate from countdown refreshes, which may drift
      // when the JS thread is busy. Activity never extends this deadline.
      active.timers.push(
        setTimeout(
          () => {
            if (this.current(active)) this.stop();
          },
          Math.max(0, expiresAt - Date.now()),
        ),
      );
      const tick = () => {
        if (!this.current(active)) return;
        const seconds = Math.max(0, Math.ceil((expiresAt - Date.now()) / 1000));
        this.update({ remainingSeconds: seconds });
        if (!seconds) this.stop();
        else active.timers.push(setTimeout(tick, 1000));
      };
      tick();
      if (!this.current(active)) return;
      const transport = await this.deps.openTransport(
        lease,
        active.abort.signal,
        {
          onEvent: (event) => {
            if (this.current(active)) this.event(active, event);
          },
          onOpen: () => {
            if (!this.current(active)) return;
            clearTimeout(active.timers[0]);
            this.update({ status: "connected" });
            this.touch(active);
          },
          onFailure: (error) => this.fail(active, error),
        },
      );
      if (!this.current(active)) {
        transport.close();
        this.stopIfOwned(active);
        return;
      }
      active.transport = transport;
      transport.mute(this.state.isMuted);
      // Response instructions replace the session prompt, including the
      // authenticated user's identity. Keep the server's greeting policy intact.
      this.send(active, { type: "response.create" });
    } catch (error) {
      this.fail(active, error);
    }
  };
  private fail(active: ActiveVoice, error: unknown) {
    if (!this.current(active)) {
      this.stopIfOwned(active);
      return;
    }
    this.stop();
    this.update({ error: errorMessage(error) });
  }
  private stopIfOwned(active: ActiveVoice) {
    if (this.active === active) this.stop();
  }
  stop = () => {
    const active = this.active;
    this.active = null;
    if (active) {
      active.abort.abort();
      active.timers.forEach(clearTimeout);
      clearTimeout(active.idleTimer);
      active.transport?.close();
      active.pending = undefined;
      if (active.identity && active.lease) {
        // Best effort accounting; an unanswered lease remains conservatively reserved.
        void this.deps
          .endSession(active.identity, active.lease.sessionId)
          .catch(() => undefined);
      }
    }
    this.update({
      status: "idle",
      isMuted: false,
      isSpeaking: false,
      remainingSeconds: null,
      pendingAction: null,
      isApproving: false,
    });
  };
  clearTranscript = () => {
    this.stop();
    this.owner = null;
    this.update({ transcript: [], error: null });
  };
  clearError = () => this.update({ error: null });
  toggleMute = () => {
    if (!this.active || this.state.status !== "connected") return;
    const isMuted = !this.state.isMuted;
    this.active.transport?.mute(isMuted);
    this.update({ isMuted });
  };
  checkIdentity = () => {
    if (this.owner && !this.deps.isCurrent(this.owner)) this.clearTranscript();
  };
  private event(active: ActiveVoice, event: VoiceEvent) {
    this.touch(active);
    const itemID = event.item?.id ?? event.item_id;
    if (itemID && !active.orders.has(itemID))
      active.orders.set(itemID, active.orders.size);
    const update = voiceTranscriptUpdate(event);
    if (update && active.lease) {
      const id = `voice-${active.lease.sessionId}-${update.id}`;
      const existing = this.state.transcript.find(
        (message) => message.id === id,
      );
      const message: VoiceTranscriptMessage = {
        id,
        role: update.role,
        text: update.final ? update.text : (existing?.text ?? "") + update.text,
        createdAt: existing?.createdAt ?? Date.now(),
        order: active.orders.get(update.id),
      };
      const transcript = existing
        ? this.state.transcript.map((item) => (item.id === id ? message : item))
        : [...this.state.transcript, message];
      const prefix = `voice-${active.lease.sessionId}-`;
      transcript.sort((a, b) => {
        if (!a.id.startsWith(prefix) || !b.id.startsWith(prefix))
          return a.createdAt - b.createdAt;
        return (
          (active.orders.get(a.id.slice(prefix.length)) ?? 0) -
          (active.orders.get(b.id.slice(prefix.length)) ?? 0)
        );
      });
      this.update({ transcript });
      if (update.final && !active.finalized.has(id)) {
        active.finalized.add(id);
        // Persistence is provided by the shared chat controller. A failed save
        // must remain visible without aborting the microphone event handler.
        const reportSaveError = () => {
          if (this.current(active))
            this.update({
              error:
                "This voice message could not be saved. It remains visible until you leave this conversation.",
            });
        };
        try {
          void Promise.resolve(
            active.context.onTranscriptFinalized?.(message),
          ).catch(reportSaveError);
        } catch {
          reportSaveError();
        }
      }
    }
    if (
      [
        "output_audio_buffer.started",
        "response.output_audio.delta",
        "response.audio.delta",
        "response.output_audio_transcript.delta",
        "response.audio_transcript.delta",
      ].includes(event.type ?? "")
    )
      this.update({ isSpeaking: true });
    if (
      [
        "output_audio_buffer.stopped",
        "output_audio_buffer.cleared",
        "input_audio_buffer.speech_started",
      ].includes(event.type ?? "")
    )
      this.update({ isSpeaking: false });
    if (event.type === "error")
      this.update({
        error:
          event.error?.message ?? "The voice connection encountered an error.",
      });
    if (event.type === "response.done") {
      for (const call of event.response?.output ?? []) {
        if (call.type === "function_call" && call.call_id && call.name) {
          void this.runTool(active, {
            call_id: call.call_id,
            name: call.name,
            arguments: call.arguments ?? "{}",
          });
        }
      }
    }
  }
  private async runTool(active: ActiveVoice, call: VoiceToolCall) {
    if (
      !this.current(active) ||
      !active.identity ||
      !active.lease ||
      active.calls.has(call.call_id)
    )
      return;
    active.calls.add(call.call_id);
    try {
      const { args, mutation } = prepareVoiceTool(call.name, call.arguments);
      if (
        mutation &&
        (active.pending || active.preparingCall || this.state.isApproving)
      ) {
        this.output(active, call.call_id, {
          success: false,
          message: "Wait for the user to review the pending change in the app.",
        });
        return;
      }
      if (mutation) active.preparingCall = call.call_id;
      if (call.name === "end_conversation") {
        this.stop();
        return;
      }
      const output = await this.deps.tool(
        active.identity,
        {
          sessionId: active.lease.sessionId,
          callId: call.call_id,
          name: call.name,
          arguments: args,
        },
        active.abort.signal,
      );
      if (!this.current(active)) return;
      if (output.requiresConfirmation) {
        if (
          !mutation ||
          typeof output.confirmationToken !== "string" ||
          !output.confirmationToken
        )
          throw new Error(
            "Maya could not prepare a safe confirmation. Please try again.",
          );
        active.pending = { call, args, token: output.confirmationToken };
        this.update({
          pendingAction: voicePendingAction(
            call.call_id,
            call.name,
            args,
            output,
          ),
          error: null,
        });
        // Keep the provider call open until a real UI decision. It never sees
        // the token, cannot self-confirm, and cannot race a second proposal.
        return;
      }
      if (output.success && output.clientAction) {
        const applied = await this.deps.clientAction(
          active.identity,
          output.clientAction,
        );
        if (!this.current(active)) return;
        if (!applied) {
          this.output(active, call.call_id, {
            success: false,
            error: "That destination is not available in the mobile app.",
          });
          return;
        }
      }
      this.output(active, call.call_id, output);
    } catch (error) {
      if (!this.current(active)) return;
      this.update({ error: errorMessage(error) });
      this.output(active, call.call_id, {
        success: false,
        error: "The request could not be completed. Ask the user to try again.",
      });
    } finally {
      if (active.preparingCall === call.call_id)
        active.preparingCall = undefined;
    }
  }
  approveAction = async () => {
    const active = this.active;
    const pending = active?.pending;
    if (
      !active ||
      !pending ||
      !this.current(active) ||
      !active.identity ||
      !active.lease ||
      this.state.isApproving
    )
      return;
    this.update({ isApproving: true, error: null });
    pending.approvalAttempted = true;
    this.touch(active);
    try {
      const output = await this.deps.tool(
        active.identity,
        {
          sessionId: active.lease.sessionId,
          callId: `${pending.call.call_id.slice(0, 100)}:approved`,
          name: pending.call.name,
          arguments: {
            ...pending.args,
            confirmed: true,
            confirmationToken: pending.token,
          },
        },
        active.abort.signal,
      );
      if (!this.current(active)) return;
      if (output.requiresConfirmation) {
        // A changed server proposal is never auto-approved under an old gesture.
        throw new Error(
          "The task changed. Cancel this proposal and ask Maya to prepare it again.",
        );
      }
      active.pending = undefined;
      this.update({ pendingAction: null, isApproving: false });
      if (output.success)
        this.deps.invalidate(active.identity, pending.call.name);
      else
        this.update({
          error:
            output.error ?? output.message ?? "The change was not applied.",
        });
      this.output(active, pending.call.call_id, output);
    } catch (error) {
      if (!this.current(active)) return;
      this.deps.invalidate(active.identity, pending.call.name);
      // Retain the exact approved call ID for safe receipt-based retry; never
      // claim success or create a second mutation after a lost response.
      this.update({
        error: `${errorMessage(error)} Check the task before trying again.`,
        isApproving: false,
      });
    }
  };
  cancelAction = () => {
    const active = this.active;
    if (!active?.pending || !this.current(active) || this.state.isApproving)
      return;
    const { call, approvalAttempted } = active.pending;
    active.pending = undefined;
    this.update({ pendingAction: null, error: null });
    this.output(active, call.call_id, {
      success: false,
      cancelled: !approvalAttempted,
      message: approvalAttempted
        ? "The app could not verify whether the approved action completed. The user dismissed it. Do not retry; ask them to check the task."
        : "The user cancelled this action. Do not retry it without a new request.",
    });
  };
}
