import { disposeRealtimePeerConnection } from "./realtime-peer-connection";

describe("Realtime peer connection lifecycle", () => {
  it("detaches event handlers before closing a connection", () => {
    let wasClosed = false;
    const peerConnection = {
      close: () => {
        wasClosed = true;
      },
      onconnectionstatechange: () => undefined,
      ontrack: () => undefined,
    } as unknown as RTCPeerConnection;

    disposeRealtimePeerConnection(peerConnection);

    expect(peerConnection.onconnectionstatechange).toBeNull();
    expect(peerConnection.ontrack).toBeNull();
    expect(wasClosed).toBe(true);
  });
});
