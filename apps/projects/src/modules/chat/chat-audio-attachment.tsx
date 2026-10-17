"use client";

import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { AudioAttachmentPlayer } from "./audio-attachment-player";
import { useAudioTranscription } from "./audio-transcription";
import { getChatAttachmentUrl, getChatAudioPreviewUrl } from "./attachment-url";
import type { ChatAttachment } from "./types";

export const ChatAudioAttachment = ({
  attachment,
  conversationId,
  messageId,
}: {
  attachment: ChatAttachment;
  conversationId: string;
  messageId: string;
}) => {
  const { workspaceSlug } = useWorkspacePath();
  const { transcript, request } = useAudioTranscription(
    conversationId,
    attachment.id,
  );
  return (
    <AudioAttachmentPlayer
      attachment={{
        id: attachment.id,
        fileName:
          attachment.kind === "voice-note" ? "Voice note" : attachment.fileName,
        messageId,
        url: getChatAttachmentUrl(workspaceSlug, conversationId, attachment.id),
        waveformUrl: getChatAudioPreviewUrl(
          workspaceSlug,
          conversationId,
          attachment.id,
        ),
        durationSeconds: attachment.durationMs
          ? attachment.durationMs / 1000
          : undefined,
      }}
      onRequestTranscript={request}
      transcript={transcript}
    />
  );
};
