export const createRealtimePeerConnection = () => new RTCPeerConnection();

export const disposeRealtimePeerConnection = (
  peerConnection: RTCPeerConnection | null,
) => {
  if (!peerConnection) {
    return;
  }

  peerConnection.onconnectionstatechange = null;
  peerConnection.ontrack = null;
  peerConnection.close();
};
