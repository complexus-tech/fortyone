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
} from "icons";
import type { FileRejection } from "react-dropzone";
import { useDropzone } from "react-dropzone";
import { toast } from "sonner";
import ky from "ky";
import { StoryAttachmentPreview } from "@/modules/story/components/story-attachment-preview";
import { useVoiceRecording } from "@/hooks/use-voice-recording";

type ChatInputProps = {
  value: string;
  onChange: (e: ChangeEvent<HTMLTextAreaElement>) => void;
  onSend: () => void;
  onStop: () => void;
  isLoading: boolean;
  attachments: File[];
  onAttachmentsChange: (files: File[]) => void;
};

const placeholderTexts = [
  "Show me my stories...",
  "Take me to my work...",
  "Show me the current sprint...",
  "Open my objectives...",
  "Show team analytics...",
  "Navigate to the roadmap...",
  "Find stories in progress...",
  "Show my key results...",
];

const SendIcon = () => {
  return (
    <svg
      className="h-6 w-auto text-white dark:text-dark"
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

const StopIcon = () => {
  return (
    <svg
      className="h-5 w-auto text-white dark:text-dark"
      fill="none"
      height="24"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M12.0436 3.25C13.6463 3.24999 14.9086 3.24998 15.913 3.35586C16.9399 3.4641 17.7833 3.68971 18.5113 4.19945C19.0129 4.55072 19.4493 4.98706 19.8005 5.48872C20.3103 6.21671 20.5359 7.06008 20.6441 8.08697C20.75 9.0914 20.75 10.3537 20.75 11.9564V12.0436C20.75 13.6463 20.75 14.9086 20.6441 15.913C20.5359 16.9399 20.3103 17.7833 19.8005 18.5113C19.4493 19.0129 19.0129 19.4493 18.5113 19.8005C17.7833 20.3103 16.9399 20.5359 15.913 20.6441C14.9086 20.75 13.6463 20.75 12.0436 20.75H11.9564C10.3537 20.75 9.0914 20.75 8.08697 20.6441C7.06008 20.5359 6.21671 20.3103 5.48872 19.8005C4.98706 19.4493 4.55072 19.0129 4.19945 18.5113C3.68971 17.7833 3.4641 16.9399 3.35586 15.913C3.24998 14.9086 3.24999 13.6463 3.25 12.0436V11.9564C3.24999 10.3537 3.24998 9.0914 3.35586 8.08697C3.4641 7.06008 3.68971 6.21671 4.19945 5.48872C4.55072 4.98706 4.98706 4.55072 5.48872 4.19945C6.21671 3.68971 7.06008 3.4641 8.08697 3.35586C9.0914 3.24998 10.3537 3.24999 11.9564 3.25H12.0436Z"
        fill="currentColor"
      />
    </svg>
  );
};

export const ChatInput = ({
  value,
  onChange,
  onSend,
  isLoading,
  onStop,
  attachments,
  onAttachmentsChange,
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
  } = useVoiceRecording();

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
    onDrop: (acceptedFiles) => {
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
    } else {
      startRecording();
    }
  };

  return (
    <Box className="sticky bottom-0 px-6 pb-3">
      {recordingState !== "idle" && (
        <Flex align="center" className="mb-2 px-1" gap={2} justify="between">
          <Text className="text-sm" color="muted">
            {`${recordingState.charAt(0).toUpperCase() + recordingState.slice(1)}...`}
          </Text>
          <Flex align="center" gap={1}>
            <Text
              className={cn("text-sm", {
                "text-success": recordingDuration < 40,
                "text-warning":
                  recordingDuration >= 40 && recordingDuration < 50,
                "text-danger": recordingDuration >= 50,
              })}
            >
              {formatDuration(recordingDuration)}
            </Text>
            <Text className="text-sm" color="muted">
              / {formatDuration(MAX_RECORDING_TIME)}
            </Text>
          </Flex>
        </Flex>
      )}
      <Box className="rounded-[1.25rem] border border-gray-100 bg-gray-50/80 py-2 backdrop-blur-lg dark:border-dark-50/80 dark:bg-dark-200/70">
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

        <Box className="relative">
          <textarea
            autoFocus
            className="max-h-40 min-h-9 w-full flex-1 resize-none border-none bg-transparent px-5 py-2 text-[1.1rem] shadow-none focus:outline-none focus:ring-0 dark:text-white"
            onChange={onChange}
            onKeyDown={handleKeyDown}
            placeholder=""
            ref={textareaRef}
            rows={1}
            value={value}
          />
          {!value && (
            <Box
              className={cn(
                "pointer-events-none absolute left-5 top-2 text-[1.1rem] text-gray transition-all duration-200 ease-in-out dark:text-gray-200/60",
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
        <Flex align="center" className="px-3" gap={2} justify="between">
          <Tooltip side="bottom" title="Add files (max 5 files, 5MB each)">
            <Button
              asIcon
              className="mb-0.5 dark:hover:bg-dark-50 md:h-10"
              color="tertiary"
              onClick={open}
              rounded="full"
            >
              <PlusIcon />
            </Button>
          </Tooltip>

          <Flex align="center" gap={2}>
            <Tooltip
              side="bottom"
              title={isRecording ? "Stop dictation" : "Dictate"}
            >
              <Button
                asIcon
                className="mb-0.5 border-0 bg-gray-100/60 md:h-[2.6rem]"
                color={isRecording ? "primary" : "tertiary"}
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
              >
                <span className="sr-only">
                  {isRecording ? "Stop dictation" : "Dictate"}
                </span>
              </Button>
            </Tooltip>
            <Button
              asIcon
              className="mb-0.5 md:h-10"
              color="invert"
              onClick={() => {
                if (isRecording) {
                  handleVoiceRecording();
                } else if (isLoading) {
                  onStop();
                } else {
                  onSend();
                }
              }}
              rounded="full"
            >
              {isRecording ? (
                <CheckIcon className="text-white dark:text-dark" />
              ) : isLoading ? (
                <StopIcon />
              ) : (
                <SendIcon />
              )}
            </Button>
          </Flex>
        </Flex>
      </Box>
      <Text
        align="center"
        className="mt-2 antialiased opacity-90"
        color="muted"
      >
        Maya can make mistakes, so double-check important info.
      </Text>
    </Box>
  );
};
