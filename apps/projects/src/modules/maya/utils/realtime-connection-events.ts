import { getMicrophoneErrorDescription } from "@/lib/microphone-error";

export const acquireRealtimeMicrophone = () =>
  navigator.mediaDevices
    .getUserMedia({
      audio: {
        autoGainControl: true,
        echoCancellation: true,
        noiseSuppression: true,
      },
    })
    .catch((microphoneError: unknown) => {
      throw new Error(getMicrophoneErrorDescription(microphoneError));
    });

type RealtimeConnectionEvents = {
  peerConnection: RTCPeerConnection;
  localStream: MediaStream;
  remoteAudio: HTMLAudioElement;
  isActive: () => boolean;
  onFailure: (message: string) => void;
  onMessage: (data: string) => void;
  onOpen: () => void;
};

export const configureRealtimeConnectionEvents = ({
  peerConnection,
  localStream,
  remoteAudio,
  isActive,
  onFailure,
  onMessage,
  onOpen,
}: RealtimeConnectionEvents) => {
  peerConnection.ontrack = (event) => {
    if (!isActive()) return;
    remoteAudio.srcObject = event.streams[0] ?? null;
    void remoteAudio.play().catch(() => {
      onFailure(
        "Maya's audio could not play. Check your audio output and start Live voice again.",
      );
    });
  };
  peerConnection.onconnectionstatechange = () => {
    if (
      peerConnection.connectionState === "failed" ||
      peerConnection.connectionState === "closed"
    ) {
      onFailure(
        "Live voice disconnected. Check your internet connection and try again.",
      );
    }
  };
  localStream.getAudioTracks().forEach((track) => {
    peerConnection.addTrack(track, localStream);
  });

  const dataChannel = peerConnection.createDataChannel("oai-events");
  dataChannel.onmessage = (event) => {
    if (isActive() && typeof event.data === "string") {
      onMessage(event.data);
    }
  };
  dataChannel.onclose = () => {
    onFailure(
      "Live voice disconnected. Check your internet connection and try again.",
    );
  };
  dataChannel.onerror = () => {
    onFailure(
      "Live voice could not connect. Check your internet connection and try again.",
    );
  };
  dataChannel.onopen = () => {
    if (isActive() && dataChannel.readyState === "open") onOpen();
  };
  return dataChannel;
};
