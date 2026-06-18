"use client";

import { Box, Button, Flex, Text, Tooltip, Wrapper } from "ui";
import { LoadingIcon, StopIcon, VoiceIcon, VolumeIcon } from "icons";
import { cn } from "lib";
import { useMayaRealtimeVoice } from "../hooks/use-maya-realtime-voice";

type RealtimeVoiceControlProps = {
  disabled?: boolean;
  isOnPage?: boolean;
};

const getButtonLabel = (
  status: ReturnType<typeof useMayaRealtimeVoice>["status"],
) => {
  if (status === "connecting") return "Connecting";
  if (status === "connected") return "End voice";
  return "Live voice";
};

const getStatusLabel = ({
  isListening,
  isSpeaking,
  status,
}: Pick<
  ReturnType<typeof useMayaRealtimeVoice>,
  "isListening" | "isSpeaking" | "status"
>) => {
  if (status === "connecting") return "Starting voice session";
  if (isSpeaking) return "Maya is speaking";
  if (isListening) return "Listening";
  return "Ready for realtime voice";
};

export const RealtimeVoiceControl = ({
  disabled = false,
  isOnPage = false,
}: RealtimeVoiceControlProps) => {
  const { connect, disconnect, error, isListening, isSpeaking, status } =
    useMayaRealtimeVoice();
  const isConnected = status === "connected";
  const isConnecting = status === "connecting";
  const label = getButtonLabel(status);
  const statusLabel = getStatusLabel({ isListening, isSpeaking, status });

  const statusIcon = (() => {
    if (isConnecting) {
      return <LoadingIcon className="h-4 w-auto animate-spin text-current" />;
    }
    if (isSpeaking) {
      return <VolumeIcon className="h-4 w-auto text-current" />;
    }
    return <VoiceIcon className="h-6 w-auto text-current" />;
  })();

  const buttonIcon = (() => {
    if (isConnecting) {
      return <LoadingIcon className="h-5 w-auto animate-spin text-current" />;
    }
    if (isConnected) {
      return <StopIcon className="h-5 w-auto text-current dark:text-current" />;
    }
    return <VoiceIcon className="h-5 w-auto text-current" />;
  })();

  return (
    <Box className="px-6 pb-2">
      <Wrapper
        className={cn(
          "dark:bg-surface/60 flex items-center justify-between gap-3 shadow-xs",
          {
            "md:px-4 md:py-3": isOnPage,
          },
        )}
      >
        <Flex align="center" className="min-w-0" gap={2}>
          <Box
            aria-hidden="true"
            className={cn(
              "flex size-11 shrink-0 items-center justify-center rounded-lg",
              {
                "bg-success/12 text-success": isConnected,
                "bg-primary/12 text-primary": isConnecting,
                "bg-muted text-foreground": !isConnected && !isConnecting,
              },
            )}
          >
            {statusIcon}
          </Box>
          <Box className="min-w-0">
            <Text
              className={cn("leading-tight font-semibold", {
                "md:text-lg": isOnPage,
              })}
            >
              {statusLabel}
            </Text>
            {error ? (
              <Text
                className={cn(
                  "text-danger dark:text-danger truncate text-[0.95rem]",
                  {
                    "md:mt-0.5 md:text-base md:leading-[1.3rem]": isOnPage,
                  },
                )}
              >
                {error}
              </Text>
            ) : (
              <Text
                className={cn("text-text-muted truncate text-[0.95rem]", {
                  "md:mt-0.5 md:text-base md:leading-[1.3rem]": isOnPage,
                })}
              >
                {isConnected ? "Realtime session active" : "Realtime voice"}
              </Text>
            )}
          </Box>
        </Flex>

        <Tooltip
          side="top"
          title={
            isConnected
              ? "End the realtime voice session"
              : "Start a realtime voice conversation"
          }
        >
          <Button
            aria-label={label}
            className="shrink-0 gap-1.5"
            color="invert"
            disabled={disabled || isConnecting}
            leftIcon={buttonIcon}
            onClick={() => {
              if (isConnected) {
                disconnect();
              } else {
                void connect();
              }
            }}
          >
            {label}
          </Button>
        </Tooltip>
      </Wrapper>
    </Box>
  );
};
