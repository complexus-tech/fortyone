import { useState } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import type { useMayaRealtimeVoice } from "@/modules/maya/hooks/use-maya-realtime-voice";
import { insertMayaSkillInstructions } from "@/modules/maya-skills/starters";
import { MayaSkillSlotProvider } from "@/shared/maya/skill-slot";
import type { MayaSkillSlotProps } from "@/shared/maya/skill-slot";
import { ChatInput } from "./chat-input";

jest.mock("@/hooks", () => ({
  useTerminology: () => ({ getTermDisplay: () => "tasks" }),
}));
jest.mock("@/hooks/use-voice-recording", () => ({
  useVoiceRecording: () => ({
    isRecording: false,
    recordingState: "idle",
    recordingDuration: 0,
    formatDuration: jest.fn(),
    MAX_RECORDING_TIME: 60,
    startRecording: jest.fn(),
    stopRecording: jest.fn(),
    getAudioBlob: jest.fn(),
    resetRecording: jest.fn(),
  }),
}));
jest.mock("react-dropzone", () => ({
  useDropzone: () => ({ getInputProps: () => ({}), open: jest.fn() }),
}));
jest.mock("@/modules/maya/components/realtime-voice-control", () => ({
  RealtimeVoiceControl: () => null,
}));
jest.mock("@/modules/story/components/story-attachment-preview", () => ({
  StoryAttachmentPreview: () => null,
}));

const TestSkillPicker = ({
  open,
  onOpenChange,
  onValueChange,
  value,
  disabled,
}: MayaSkillSlotProps) => (
  <>
    <button
      disabled={disabled}
      onClick={() => {
        onOpenChange(true);
      }}
      type="button"
    >
      Skills
    </button>
    {open ? (
      <button
        onClick={() => {
          onValueChange(
            insertMayaSkillInstructions(
              value,
              "Draft a weekly update from current work.",
            ),
          );
          onOpenChange(false);
        }}
        type="button"
      >
        Apply weekly update
      </button>
    ) : null}
  </>
);
const idleVoice = { status: "idle", error: null } as ReturnType<
  typeof useMayaRealtimeVoice
>;
const Composer = ({
  initial = "",
  enabled = true,
  onSend,
}: {
  initial?: string;
  enabled?: boolean;
  onSend: () => void;
}) => {
  const [value, setValue] = useState(initial);
  return (
    <MayaSkillSlotProvider Picker={TestSkillPicker}>
      <ChatInput
        attachments={[]}
        googleDriveFiles={[]}
        messagesCount={0}
        onAttachmentsChange={jest.fn()}
        onChange={(event) => {
          setValue(event.target.value);
        }}
        onGoogleDriveFileRemove={jest.fn()}
        onSend={onSend}
        onStop={jest.fn()}
        onValueChange={enabled ? setValue : undefined}
        realtimeVoice={idleVoice}
        status="ready"
        value={value}
      />
    </MayaSkillSlotProvider>
  );
};

it("opens skills with a slash and inserts instructions without sending", () => {
  const send = jest.fn();
  render(<Composer onSend={send} />);
  const input = screen.getByRole("textbox", { name: "Chat message" });
  fireEvent.keyDown(input, { key: "/" });
  fireEvent.click(screen.getByRole("button", { name: "Apply weekly update" }));
  expect(input).toHaveValue("Draft a weekly update from current work.");
  expect(send).not.toHaveBeenCalled();
  fireEvent.keyDown(input, { key: "Enter" });
  expect(send).toHaveBeenCalledTimes(1);
});

it("preserves draft content when choosing through the toolbar", () => {
  const send = jest.fn();
  render(<Composer initial="For the support team" onSend={send} />);
  fireEvent.click(screen.getByRole("button", { name: "Skills" }));
  fireEvent.click(screen.getByRole("button", { name: "Apply weekly update" }));
  expect(screen.getByRole("textbox", { name: "Chat message" })).toHaveValue(
    "For the support team\n\nDraft a weekly update from current work.",
  );
  expect(send).not.toHaveBeenCalled();
});

it("leaves other composer callers without a skills control or slash interception", () => {
  render(<Composer enabled={false} onSend={jest.fn()} />);
  expect(
    screen.queryByRole("button", { name: "Skills" }),
  ).not.toBeInTheDocument();
  const cancelled = fireEvent.keyDown(
    screen.getByRole("textbox", { name: "Chat message" }),
    { key: "/" },
  );
  expect(cancelled).toBe(true);
});
