/* global describe, expect, it, jest -- Jest globals. */
import type { UIMessage } from "ai";
import { appendVoiceHistory, type VoiceHistoryDependencies } from "./append";

const user = {
  id: "voice-user-1",
  role: "user" as const,
  text: "Create a launch task",
};
const assistant = {
  id: "voice-assistant-1",
  role: "assistant" as const,
  text: "Review the proposed task first.",
};
const stored = (message: typeof user | typeof assistant): UIMessage => ({
  id: message.id,
  role: message.role,
  parts: [{ type: "text", text: message.text }],
});
const dependencies = (current: UIMessage[] = []) =>
  ({
    read: jest.fn(async () => current),
    begin: jest.fn(
      async (): Promise<
        Awaited<ReturnType<VoiceHistoryDependencies["begin"]>>
      > => ({ generation: 1, token: "opaque" }),
    ),
    finalize: jest.fn(async () => ({ applied: true })),
  }) satisfies VoiceHistoryDependencies;

describe("voice transcript append", () => {
  it("reserves the user turn then finalizes its assistant reply without overwriting history", async () => {
    const previous: UIMessage = {
      id: "previous",
      role: "assistant",
      parts: [{ type: "text", text: "Earlier work" }],
    };
    const io = dependencies([previous]);
    expect(await appendVoiceHistory([user, assistant], io)).toEqual([
      previous,
      stored(user),
      stored(assistant),
    ]);
    expect(io.begin).toHaveBeenCalledWith([previous, stored(user)]);
    expect(io.finalize).toHaveBeenCalledWith(
      [previous, stored(user), stored(assistant)],
      { generation: 1, token: "opaque" },
    );
  });
  it("deduplicates exact retries before opening a generation", async () => {
    const io = dependencies([stored(user), stored(assistant)]);
    await appendVoiceHistory([user, assistant], io);
    expect(io.begin).not.toHaveBeenCalled();
    expect(io.finalize).not.toHaveBeenCalled();
  });
  it("completes a user transcript saved before a delayed final reply", async () => {
    const io = dependencies([stored(user)]);
    expect(await appendVoiceHistory([user, assistant], io)).toEqual([
      stored(user),
      stored(assistant),
    ]);
  });
  it("rejects rewriting an existing ID", async () => {
    const io = dependencies([stored(user)]);
    await expect(
      appendVoiceHistory([{ ...user, text: "Forged replacement" }], io),
    ).rejects.toMatchObject({ status: 409 });
    expect(io.begin).not.toHaveBeenCalled();
  });
  it("does not attach a late assistant-only reply to a completed unrelated turn", async () => {
    const io = dependencies([stored(user), stored(assistant)]);
    await expect(
      appendVoiceHistory([{ ...assistant, id: "late" }], io),
    ).rejects.toMatchObject({ status: 409 });
    expect(io.begin).not.toHaveBeenCalled();
  });
  it("never attaches a delayed reply to a newer user turn", async () => {
    const nextUser = {
      ...user,
      id: "next-user",
      text: "An unrelated text request",
    };
    const io = dependencies([stored(user), stored(nextUser)]);
    await expect(
      appendVoiceHistory([user, assistant], io),
    ).rejects.toMatchObject({ status: 409 });
    expect(io.begin).not.toHaveBeenCalled();
  });
  it("requires the originating user ID for an assistant tail", async () => {
    const io = dependencies([stored(user)]);
    await expect(appendVoiceHistory([assistant], io)).rejects.toMatchObject({
      status: 409,
    });
    expect(io.begin).not.toHaveBeenCalled();
  });
  it("surfaces a superseded generation rather than acknowledging lost history", async () => {
    const io = dependencies();
    io.finalize.mockResolvedValue({ applied: false });
    await expect(
      appendVoiceHistory([user, assistant], io),
    ).rejects.toMatchObject({ status: 409 });
  });
  it("does not finalize when the server rejects an open mutation approval", async () => {
    const io = dependencies();
    io.begin.mockRejectedValue(new Error("approval open"));
    await expect(appendVoiceHistory([user, assistant], io)).rejects.toThrow(
      "approval open",
    );
    expect(io.finalize).not.toHaveBeenCalled();
  });
  it("keeps server-repaired approval receipts in its canonical prefix", async () => {
    const io = dependencies();
    const repaired: UIMessage = {
      id: "old",
      role: "assistant",
      parts: [{ type: "text", text: "Confirmed server receipt" }],
    };
    io.begin.mockResolvedValue({
      generation: 1,
      token: "opaque",
      messages: [repaired, stored(user)],
    } as Awaited<ReturnType<VoiceHistoryDependencies["begin"]>>);
    expect(await appendVoiceHistory([user, assistant], io)).toEqual([
      repaired,
      stored(user),
      stored(assistant),
    ]);
  });
});
