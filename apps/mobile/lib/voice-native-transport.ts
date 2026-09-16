import { NativeModules, Platform } from "react-native";
import { fetch as expoFetch } from "expo/fetch";
import type { MediaStream, RTCPeerConnection } from "react-native-webrtc";
import type { VoiceDependencies, VoiceTransport } from "./voice-controller";
import type { VoiceEvent } from "./voice-protocol";

const REBUILD_MESSAGE =
  "Live voice needs the latest FortyOne native build. Rebuild or update the app, then try again. Expo Go does not include voice support.";

const assertNotAborted = (signal: AbortSignal) => {
  if (signal.aborted) throw new Error("Voice connection ended.");
};

export const assertNativeVoiceAvailable = () => {
  if (Platform.OS === "web" || !NativeModules.WebRTCModule)
    throw new Error(REBUILD_MESSAGE);
};

export const openNativeVoiceTransport: VoiceDependencies["openTransport"] =
  async (lease, signal, callbacks) => {
    assertNativeVoiceAvailable();
    // Loading this module eagerly crashes older installed native binaries before
    // the user can update. Only initialize WebRTC after checking the native bridge.
    let rtc: typeof import("react-native-webrtc");
    try {
      rtc = await import("react-native-webrtc");
    } catch {
      throw new Error(REBUILD_MESSAGE);
    }
    assertNotAborted(signal);
    let stream: MediaStream | null = null;
    let peer: RTCPeerConnection | null = null;
    let channel: ReturnType<RTCPeerConnection["createDataChannel"]> | null =
      null;
    let closed = false;
    let rejectOpen: ((error: Error) => void) | undefined;
    const close = () => {
      if (closed) return;
      closed = true;
      signal.removeEventListener("abort", abort);
      rejectOpen?.(new Error("Voice connection ended."));
      channel?.close();
      peer?.close();
      stream?.getTracks().forEach((track) => track.stop());
      stream?.release();
    };
    const abort = () => close();
    signal.addEventListener("abort", abort, { once: true });
    try {
      stream = await rtc.mediaDevices.getUserMedia({
        audio: true,
        video: false,
      });
      if (closed || signal.aborted) {
        stream.getTracks().forEach((track) => track.stop());
        stream.release();
        throw new Error("Voice connection ended.");
      }
      peer = new rtc.RTCPeerConnection();
      const connection = peer;
      stream.getAudioTracks().forEach((track) => {
        track.onended = () => {
          if (!closed)
            callbacks.onFailure(
              new Error(
                "Microphone access ended. Start voice again when you are ready.",
              ),
            );
        };
        connection.addTrack(track, stream!);
      });
      connection.onconnectionstatechange = () => {
        if (
          !closed &&
          ["failed", "closed", "disconnected"].includes(
            connection.connectionState,
          )
        ) {
          callbacks.onFailure(
            new Error(
              "Voice disconnected. Check your connection and start again.",
            ),
          );
        }
      };
      channel = connection.createDataChannel("oai-events");
      const data = channel;
      const opened = new Promise<void>((resolve, reject) => {
        rejectOpen = reject;
        data.onopen = () => {
          if (closed) return;
          callbacks.onOpen();
          resolve();
        };
        data.onclose = () => {
          if (!closed)
            callbacks.onFailure(
              new Error("Voice disconnected. Please start again."),
            );
        };
        data.onerror = () => {
          if (!closed)
            callbacks.onFailure(
              new Error("The voice connection failed. Please try again."),
            );
        };
      });
      // Attach a rejection handler immediately while SDP negotiation is pending.
      void opened.catch(() => undefined);
      data.onmessage = (event: unknown) => {
        if (
          closed ||
          !event ||
          typeof event !== "object" ||
          !("data" in event) ||
          typeof event.data !== "string"
        )
          return;
        try {
          callbacks.onEvent(JSON.parse(event.data) as VoiceEvent);
        } catch {
          /* Ignore malformed provider events, without logging sensitive content. */
        }
      };
      const offer = await connection.createOffer({
        offerToReceiveAudio: true,
        offerToReceiveVideo: false,
      });
      assertNotAborted(signal);
      await connection.setLocalDescription(offer);
      assertNotAborted(signal);
      // Never use the first-party API wrapper here: only the ephemeral OpenAI
      // credential goes to this fixed provider origin, with native cookies omitted.
      const response = await expoFetch(
        "https://api.openai.com/v1/realtime/calls",
        {
          method: "POST",
          body: offer.sdp,
          headers: {
            Authorization: `Bearer ${lease.clientSecret}`,
            "Content-Type": "application/sdp",
          },
          credentials: "omit",
          redirect: "error",
          signal,
        },
      );
      if (!response.ok)
        throw new Error("Maya could not connect to voice. Please try again.");
      const sdp = await response.text();
      assertNotAborted(signal);
      await connection.setRemoteDescription(
        new rtc.RTCSessionDescription({ type: "answer", sdp }),
      );
      await opened;
      assertNotAborted(signal);
      const transport: VoiceTransport = {
        send: (event) => {
          if (!closed && data.readyState === "open")
            data.send(JSON.stringify(event));
        },
        mute: (muted) =>
          stream?.getAudioTracks().forEach((track) => {
            track.enabled = !muted;
          }),
        close,
      };
      // Native WebRTC plays received audio directly; no DOM audio view is needed.
      return transport;
    } catch (error) {
      close();
      if (
        error instanceof Error &&
        /permission|denied|notallowed/i.test(`${error.name} ${error.message}`)
      ) {
        throw new Error(
          "Allow microphone access in Settings to speak with Maya.",
        );
      }
      throw error;
    }
  };
