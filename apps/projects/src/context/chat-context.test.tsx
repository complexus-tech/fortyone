/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { generateId } from "ai";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { useMayaChat } from "@/modules/maya";
import type { MayaChatConfig } from "@/modules/maya";
import { ChatProvider, useChatContext } from "./chat-context";

jest.mock("ai", () => ({ generateId: jest.fn() }));

jest.mock("@/components/walkthrough/walkthrough-provider", () => ({
  useWalkthrough: () => ({ completeWalkthroughAction: jest.fn() }),
}));

jest.mock("@/modules/maya", () => ({ useMayaChat: jest.fn() }));

const generateIdMock = jest.mocked(generateId);
const useMayaChatMock = jest.mocked(useMayaChat);
const addGoogleDriveFile = jest.fn();
const setInput = jest.fn();

const GoogleDriveContextProbe = ({
  name = "Launch plan",
}: {
  name?: string;
}) => {
  const { openChatWithGoogleDriveFile } = useChatContext();
  return (
    <button
      onClick={() => {
        openChatWithGoogleDriveFile(
          {
            id: "00000000-0000-4000-8000-000000000001",
            mimeType: "application/vnd.google-apps.document",
            name,
          },
          "Review this plan",
        );
      }}
      type="button"
    >
      Ask Maya
    </button>
  );
};

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
      addGoogleDriveFile,
      handleSuggestedPrompt: jest.fn(),
      setInput,
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

  it("opens Maya with a validated linked-file reference and optional draft", () => {
    render(
      <ChatProvider>
        <GoogleDriveContextProbe />
      </ChatProvider>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Ask Maya" }));

    expect(addGoogleDriveFile).toHaveBeenCalledWith({
      referenceId: "00000000-0000-4000-8000-000000000001",
      mimeType: "application/vnd.google-apps.document",
      name: "Launch plan",
    });
    expect(setInput).toHaveBeenCalledWith("Review this plan");
  });

  it("bounds a linked-file display name before adding it to Maya", () => {
    const longName = "x".repeat(700);
    render(
      <ChatProvider>
        <GoogleDriveContextProbe name={longName} />
      </ChatProvider>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Ask Maya" }));

    expect(addGoogleDriveFile).toHaveBeenCalledWith(
      expect.objectContaining({ name: longName.slice(0, 500) }),
    );
  });
});
