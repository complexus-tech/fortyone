import {
  getRealtimeErrorMessage,
  parseRealtimeCallAnswer,
  parseRealtimeSessionResponse,
  parseRealtimeToolOutput,
} from "./maya-realtime-voice-contract";

describe("Maya realtime voice contract", () => {
  it("preserves a useful runtime error and falls back for opaque failures", () => {
    expect(
      getRealtimeErrorMessage(new Error("Microphone unavailable"), "Failed"),
    ).toBe("Microphone unavailable");
    expect(getRealtimeErrorMessage({ reason: "unknown" }, "Failed")).toBe(
      "Failed",
    );
  });

  it("normalizes tool response failures before they reach the data channel", async () => {
    const createResponse = (ok: boolean, payload: unknown) =>
      ({
        json: async () => payload,
        ok,
      }) as Response;

    await expect(
      parseRealtimeToolOutput(
        createResponse(false, { error: { message: "Not allowed" } }),
      ),
    ).resolves.toEqual({ error: "Not allowed", success: false });

    await expect(
      parseRealtimeToolOutput(createResponse(true, { data: null })),
    ).resolves.toEqual({
      error: "Tool returned an unreadable response.",
      success: false,
    });
  });

  it("checks the session response status before accepting its payload", async () => {
    const createResponse = (ok: boolean, payload: unknown) =>
      ({
        json: async () => payload,
        ok,
      }) as Response;

    await expect(
      parseRealtimeSessionResponse(
        createResponse(false, { error: { message: "Monthly limit reached" } }),
      ),
    ).rejects.toThrow("Monthly limit reached");

    await expect(
      parseRealtimeSessionResponse(
        createResponse(true, {
          data: {
            clientSecret: "secret",
            sessionId: "session-1",
          },
        }),
      ),
    ).resolves.toMatchObject({
      clientSecret: "secret",
      sessionId: "session-1",
    });
  });

  it("does not consume an unsuccessful Realtime answer", async () => {
    let wasBodyRead = false;
    const response = {
      ok: false,
      text: async () => {
        wasBodyRead = true;
        return "error";
      },
    } as Response;

    await expect(parseRealtimeCallAnswer(response)).rejects.toThrow(
      "Failed to connect voice session.",
    );
    expect(wasBodyRead).toBe(false);
  });
});
