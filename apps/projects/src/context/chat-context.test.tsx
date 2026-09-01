/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { generateId } from "ai";
import { act, render } from "@testing-library/react";
import { useMayaChat } from "@/modules/maya";
import type { MayaChatConfig } from "@/modules/maya";
import { ChatProvider } from "./chat-context";

jest.mock("ai", () => ({ generateId: jest.fn() }));

jest.mock("@/components/walkthrough/walkthrough-provider", () => ({
  useWalkthrough: () => ({ completeWalkthroughAction: jest.fn() }),
}));

jest.mock("@/modules/maya", () => ({ useMayaChat: jest.fn() }));

const generateIdMock = jest.mocked(generateId);
const useMayaChatMock = jest.mocked(useMayaChat);

const getLatestConfig = (): MayaChatConfig => {
  const latestCall = useMayaChatMock.mock.calls.at(-1);
  if (!latestCall) {
    throw new Error("Expected ChatProvider to configure Maya chat.");
  }
  return latestCall[0];
};

describe("ChatProvider session lifecycle", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    generateIdMock.mockReturnValue("draft-chat-id-01");
    useMayaChatMock.mockReturnValue({
      handleSuggestedPrompt: jest.fn(),
      setInput: jest.fn(),
    } as unknown as ReturnType<typeof useMayaChat>);
  });

  it("distinguishes local drafts from selected persisted sessions", () => {
    render(
      <ChatProvider>
        <div>Workspace content</div>
      </ChatProvider>,
    );

    expect(getLatestConfig()).toEqual(
      expect.objectContaining({
        currentChatId: "draft-chat-id-01",
        hasSelectedChat: false,
      }),
    );

    act(() => {
      getLatestConfig().updateChatRef("saved-chat-id-01");
    });
    expect(getLatestConfig()).toEqual(
      expect.objectContaining({
        currentChatId: "saved-chat-id-01",
        hasSelectedChat: true,
      }),
    );

    act(() => {
      getLatestConfig().clearChatRef("draft-chat-id-02");
    });
    expect(getLatestConfig()).toEqual(
      expect.objectContaining({
        currentChatId: "draft-chat-id-02",
        hasSelectedChat: false,
      }),
    );
  });
});
