/* eslint-disable no-nested-ternary -- ok for now */
import { Button, Box, Flex, Text, Tooltip } from "ui";
import type { ChangeEvent } from "react";
import { useRef, useEffect, useState } from "react";
import { cn } from "lib";
import {
  PlusIcon,
  MicrophoneIcon,
  CloseIcon,
  LoadingIcon,
  CheckIcon,
  StopIcon,
} from "icons";
import type { FileRejection } from "react-dropzone";
import { useDropzone } from "react-dropzone";
import { toast } from "sonner";
import ky from "ky";
import type { ChatStatus } from "ai";
import { StoryAttachmentPreview } from "@/modules/story/components/story-attachment-preview";
import { useVoiceRecording } from "@/hooks/use-voice-recording";

type ChatInputProps = {
  value: string;
  onChange: (e: ChangeEvent<HTMLTextAreaElement>) => void;
  onSend: () => void;
  onStop: () => void;
  status: ChatStatus;
  attachments: File[];
  onAttachmentsChange: (files: File[]) => void;
  isOnPage?: boolean;
  messagesCount: number;
};

const SendIcon = () => {
  return (
    <svg
      className="h-6 w-auto text-current"
      fill="none"
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M18 9.47326L16.5858 10.8813L13.0006 7.31184L13.0006 20.5H11.0006L11.0006 7.3114L7.41422 10.8814L6 9.47338L12.0003 3.5L18 9.47326Z"
        fill="currentColor"
      />
    </svg>
  );
};

export const ChatInput = ({
  value,
  onChange,
  onSend,
  status,
  onStop,
  attachments,
  onAttachmentsChange,
  isOnPage,
  messagesCount,
}: ChatInputProps) => {
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const [currentPlaceholderIndex, setCurrentPlaceholderIndex] = useState(0);
  const [isAnimating, setIsAnimating] = useState(false);
  const [isTranscribing, setIsTranscribing] = useState(false);

  const {
    isRecording,
    recordingState,
    recordingDuration,
    formatDuration,
    MAX_RECORDING_TIME,
    startRecording,
    stopRecording,
    getAudioBlob,
    resetRecording,
  } = useVoiceRecording(() => {
    processRecording();
  });

  const onDropRejected = (fileRejections: FileRejection[]) => {
    const errors: string[] = [];
    fileRejections.forEach((file) => {
      if (file.errors[0]?.code === "file-too-large") {
        errors.push(`File ${file.file.name} size exceeds 10MB limit`);
      } else if (file.errors[0]?.code === "file-invalid-type") {
        errors.push("Invalid file type");
      } else if (file.errors[0]?.code === "too-many-files") {
        errors.push("Too many files");
      } else if (file.errors[0]?.code === "file-too-small") {
        errors.push("File size is too small");
      }
    });
    if (errors.length > 0) {
      toast.error("Some files were not uploaded", {
        description: errors.join("\n"),
        duration: 2500,
        closeButton: false,
      });
    }
  };
  const { open } = useDropzone({
    noClick: true,
    noKeyboard: true,
    accept: {
      "image/jpeg": [".jpg", ".jpeg"],
      "image/png": [".png"],
      "image/webp": [".webp"],
      "image/gif": [".gif"],
      "application/pdf": [".pdf"],
    },
    maxFiles: 5,
    maxSize: 5 * 1024 * 1024, // 5MB
    minSize: 100, // 100 bytes
    onDrop: async (acceptedFiles) => {
      onAttachmentsChange([...attachments, ...acceptedFiles]);
    },
    onDropRejected,
  });

  const handleRemoveAttachment = (index: number) => {
    const updatedAttachments = attachments.filter((_, i) => i !== index);
    onAttachmentsChange(updatedAttachments);
  };

  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
      textareaRef.current.style.height = `${textareaRef.current.scrollHeight}px`;
    }
  }, [value]);

  useEffect(() => {
    const showDuration = 4000;
    const animationDuration = 300;

    const cycle = () => {
      // Show text for 4 seconds
      setTimeout(() => {
        setIsAnimating(true);

        // Change text at the midpoint of animation (when it's most faded)
        setTimeout(() => {
          setCurrentPlaceholderIndex(
            (prev) => (prev + 1) % placeholderTexts.length,
          );
        }, animationDuration / 2);

        // Stop animation after full duration
        setTimeout(() => {
          setIsAnimating(false);
        }, animationDuration);
      }, showDuration);
    };

    // Start the first cycle
    const timeout = setTimeout(cycle, showDuration);

    // Set up repeating interval
    const interval = setInterval(() => {
      cycle();
    }, showDuration + animationDuration);

    return () => {
      clearTimeout(timeout);
      clearInterval(interval);
    };
  }, []);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      onSend();
    }
  };

  const images = attachments.filter((attachment) =>
    attachment.type.startsWith("image/"),
  );
  const pdfs = attachments.filter((attachment) =>
    attachment.type.startsWith("application/pdf"),
  );

  const handleVoiceRecording = async () => {
    if (isRecording) {
      stopRecording();
      // For manual stop, we need to handle transcription ourselves
      processRecording();
    } else {
      startRecording();
    }
  };

  const processRecording = async () => {
    // Wait for the recording to finish
    setTimeout(async () => {
      const audioBlob = getAudioBlob();
      if (!audioBlob) {
        toast.info("No audio recorded", {
          description: "Please try again.",
          duration: 2500,
          closeButton: false,
        });
        resetRecording();
        return;
      }
      setIsTranscribing(true);

      try {
        const formData = new FormData();
        formData.append("audio", audioBlob, "recording.webm");
        const { text = "" } = await ky
          .post("/api/transcribe", {
            body: formData,
          })
          .json<{ text: string }>();
        if (text.trim()) {
          onChange({
            target: { value: `${value} ${text}` },
          } as ChangeEvent<HTMLTextAreaElement>);
        }
      } catch (error) {
        toast.error("Failed to transcribe audio", {
          description: "Please try again.",
          duration: 2500,
          closeButton: false,
        });
      } finally {
        setIsTranscribing(false);
        resetRecording();
      }
    }, 100);
  };

  const placeholderTexts =
    messagesCount > 2
      ? [
          "Tell Maya what to do next...",
          "Continue the conversation...",
          "Ask me anything...",
          "Tell Maya what to do next...",
          "Continue the conversation...",
          "Ask me anything...",
        ]
      : [
          "Take me to my work...",
          "Show me the current sprint...",
          "Open my objectives...",
          "Navigate to the roadmap...",
          "Find stories in progress...",
          "Show my key results...",
        ];

  return (
    <Box className="sticky bottom-0 px-6 pb-3">
      {recordingState !== "idle" && (
        <Flex
          align="center"
          className="mb-2 px-1 pt-2"
          gap={2}
          justify="between"
        >
          <Text color="muted">
            {`${recordingState.replace("r", "l").charAt(0).toUpperCase() + recordingState.replace("recording", "listening").slice(1)}...`}
          </Text>
          <Flex align="center" gap={1}>
            <Text
              className={cn({
                "text-success dark:text-success": recordingDuration < 40,
                "text-warning dark:text-warning":
                  recordingDuration >= 40 && recordingDuration < 50,
                "text-danger dark:text-danger": recordingDuration >= 50,
              })}
            >
              {formatDuration(recordingDuration)}
            </Text>
            <Text color="muted">/ {formatDuration(MAX_RECORDING_TIME)}</Text>
          </Flex>
        </Flex>
      )}
      <Box className="border-border rounded-[1.25rem] border py-2">
        {images.length > 0 && (
          <Box className="mt-2.5 grid grid-cols-3 gap-3 px-4">
            {images.map((attachment, idx) => (
              <StoryAttachmentPreview
                file={{
                  id: attachment.name,
                  filename: attachment.name,
                  size: attachment.size,
                  mimeType: attachment.type,
                  url: URL.createObjectURL(attachment),
                  createdAt: new Date().toISOString(),
                  uploadedBy: "me",
                }}
                isInChat
                key={idx}
                onDelete={() => {
                  handleRemoveAttachment(attachments.indexOf(attachment));
                }}
              />
            ))}
          </Box>
        )}
        {pdfs.length > 0 && (
          <Box className="mt-2.5 grid grid-cols-1 gap-3 px-4">
            {pdfs.map((attachment, idx) => (
              <StoryAttachmentPreview
                file={{
                  id: attachment.name,
                  filename: attachment.name,
                  size: attachment.size,
                  mimeType: attachment.type,
                  url: URL.createObjectURL(attachment),
                  createdAt: new Date().toISOString(),
                  uploadedBy: "me",
                }}
                isInChat
                key={idx}
                onDelete={() => {
                  handleRemoveAttachment(attachments.indexOf(attachment));
                }}
              />
            ))}
          </Box>
        )}

        <Box className="relative dark:antialiased">
          <textarea
            aria-label="Chat message"
            autoFocus={Boolean(isOnPage)}
            className={cn(
              "focus-visible:ring-border max-h-40 min-h-12 w-full flex-1 resize-none border-none bg-transparent px-5 py-2 text-[1.1rem] shadow-none focus-visible:ring-2 focus-visible:outline-none dark:text-white",
              {
                "md:min-h-[3.7rem]": isOnPage,
              },
            )}
            name="message"
            onChange={onChange}
            onKeyDown={handleKeyDown}
            placeholder=""
            ref={textareaRef}
            rows={1}
            value={value}
          />
          {!value && (
            <Box
              aria-hidden="true"
              className={cn(
                "text-text-muted pointer-events-none absolute top-2 left-5 text-[1.1rem] transition-[opacity,transform] duration-200 ease-in-out motion-reduce:transition-none",
                {
                  "translate-y-1 opacity-0 first-letter:uppercase": isAnimating,
                },
              )}
            >
              {recordingState === "idle"
                ? placeholderTexts[currentPlaceholderIndex]
                : `${
                    recordingState.charAt(0).toUpperCase() +
                    recordingState.slice(1)
                  }...`}
            </Box>
          )}
        </Box>
        <Flex align="center" className="mb-1 px-3" gap={2} justify="between">
          <Flex align="center" gap={2}>
            <Tooltip side="bottom" title="Add files (max 5 files, 5MB each)">
              <Button
                className="gap-1"
                color="tertiary"
                onClick={open}
                rounded="full"
                variant="naked"
              >
                <PlusIcon /> Attach files
              </Button>
            </Tooltip>
          </Flex>

          <Flex align="center" gap={2}>
            <Button
              className="gap-1"
              color="tertiary"
              leftIcon={
                isTranscribing ? (
                  <LoadingIcon className="animate-spin" />
                ) : isRecording ? (
                  <CloseIcon />
                ) : (
                  <MicrophoneIcon />
                )
              }
              onClick={() => {
                if (isRecording) {
                  stopRecording();
                  resetRecording();
                } else {
                  handleVoiceRecording();
                }
              }}
              rounded="full"
              variant="naked"
            >
              {isTranscribing
                ? "Transcribing..."
                : isRecording
                  ? "Cancel"
                  : "Talk"}
            </Button>
            <Button
              asIcon
              aria-label={
                isRecording
                  ? "Send recording"
                  : status === "submitted" || status === "streaming"
                    ? "Stop response"
                    : "Send message"
              }
              className=""
              color="invert"
              onClick={() => {
                if (isRecording) {
                  handleVoiceRecording();
                } else if (status === "submitted") {
                  onStop();
                } else {
                  onSend();
                }
              }}
              rounded="full"
            >
              {isRecording ? (
                <CheckIcon className="text-current dark:text-current" />
              ) : status === "submitted" || status === "streaming" ? (
                <StopIcon className="text-current dark:text-current" />
              ) : (
                <SendIcon />
              )}
            </Button>
          </Flex>
        </Flex>
      </Box>
      <Text align="center" className="pt-2 opacity-90" color="muted">
        Maya can make mistakes, so double-check important info.
      </Text>
    </Box>
  );
};
