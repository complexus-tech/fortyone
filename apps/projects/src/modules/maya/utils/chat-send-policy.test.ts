/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";
import type { ChatStatus } from "ai";
import {
  isChatResponseInProgress,
  runWithAtomicSendGuard,
  runWithChatSendGuard,
} from "./chat-send-policy";

const readSource = (path: string) =>
  readFileSync(join(process.cwd(), path), "utf8");

describe("Maya chat send policy", () => {
  it.each<ChatStatus>(["submitted", "streaming"])(
    "blocks a second send while the chat is %s",
    (status) => {
      expect(isChatResponseInProgress(status)).toBe(true);
    },
  );

  it.each<ChatStatus>(["ready", "error"])(
    "allows a send while the chat is %s",
    (status) => {
      expect(isChatResponseInProgress(status)).toBe(false);
    },
  );

  it("admits only one in-flight send and releases the guard afterward", async () => {
    const guard = { current: false };
    let finishFirstSend = () => {};
    const firstSendTask = new Promise<void>((resolve) => {
      finishFirstSend = resolve;
    });
    const firstSend = runWithAtomicSendGuard(guard, () => firstSendTask);
    const secondTask = jest.fn(async () => {});

    await expect(runWithAtomicSendGuard(guard, secondTask)).resolves.toBe(
      false,
    );
    expect(secondTask).not.toHaveBeenCalled();

    finishFirstSend();
    await expect(firstSend).resolves.toBe(true);
    expect(guard.current).toBe(false);
  });

  it("releases the atomic guard when send preparation fails", async () => {
    const guard = { current: false };

    await expect(
      runWithAtomicSendGuard(guard, async () => {
        throw new Error("Attachment conversion failed");
      }),
    ).rejects.toThrow("Attachment conversion failed");

    expect(guard.current).toBe(false);
  });

  it.each<ChatStatus>(["submitted", "streaming"])(
    "does not start a guarded retry while the chat is %s",
    async (status) => {
      const guard = { current: false };
      const task = jest.fn(async () => {});

      await expect(
        runWithChatSendGuard({ sendGuard: guard, status, task }),
      ).resolves.toBe(false);
      expect(task).not.toHaveBeenCalled();
      expect(guard.current).toBe(false);
    },
  );

  it.each<ChatStatus>(["ready", "error"])(
    "allows one guarded send while the chat is %s",
    async (status) => {
      const guard = { current: false };
      const task = jest.fn(async () => {});

      await expect(
        runWithChatSendGuard({ sendGuard: guard, status, task }),
      ).resolves.toBe(true);
      expect(task).toHaveBeenCalledTimes(1);
      expect(guard.current).toBe(false);
    },
  );

  it("awaits the guarded SDK send in the Maya hook", () => {
    const source = readSource("src/modules/maya/hooks/use-maya-chat.ts");

    expect(source).toContain("const isSendingRef = useRef(false)");
    expect(source).toContain("if (isSendingRef.current)");
    expect(source).toContain("await runWithChatSendGuard({");
    expect(source).toContain("await sendMessage({");
    expect(source).toContain("await regenerate({ messageId });");
    expect(source).toContain(
      "super({ prepareSendMessagesRequest: prepareMayaChatSendRequest })",
    );
  });

  it("blocks busy Enter submits while preserving drafting and Stop", () => {
    const source = readSource("src/components/ui/chat/chat-input.tsx");

    expect(source).toContain(
      "if (isLiveVoiceActive || isChatResponseInProgress(status))",
    );
    expect(source).toContain("disabled={isLiveVoiceActive}");
    expect(source).toContain("onStop();");
  });
});
