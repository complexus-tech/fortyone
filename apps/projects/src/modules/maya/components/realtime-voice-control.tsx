"use client";

import { Box, Button, Flex, Text, Tooltip } from "ui";
import { LoadingIcon, MicrophoneIcon, StopIcon, VolumeIcon } from "icons";
import { cn } from "lib";
import { useMayaRealtimeVoice } from "../hooks/use-maya-realtime-voice";

type RealtimeVoiceControlProps = {
  disabled?: boolean;
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
    return <MicrophoneIcon className="h-4 w-auto text-current" />;
  })();

  const buttonIcon = (() => {
    if (isConnecting) {
      return <LoadingIcon className="h-4 w-auto animate-spin" />;
    }
    if (isConnected) {
      return <StopIcon className="h-4 w-auto text-current dark:text-current" />;
    }
    return <MicrophoneIcon className="h-4 w-auto" />;
  })();

  return (
    <Box className="px-6 pb-2">
      <Flex
        align="center"
        className="border-border bg-surface/80 rounded-xl border px-3 py-2 shadow-xs backdrop-blur"
        gap={3}
        justify="between"
      >
        <Flex align="center" className="min-w-0" gap={2}>
          <Box
            aria-hidden="true"
            className={cn(
              "flex h-8 w-8 shrink-0 items-center justify-center rounded-full",
              {
                "bg-success/12 text-success": isConnected,
                "bg-primary/12 text-primary": isConnecting,
                "bg-muted text-text-muted": !isConnected && !isConnecting,
              },
            )}
          >
            {statusIcon}
          </Box>
          <Box className="min-w-0">
            <Text className="text-sm leading-tight font-medium">
              {statusLabel}
            </Text>
            {error ? (
              <Text className="text-danger dark:text-danger truncate text-xs">
                {error}
              </Text>
            ) : (
              <Text className="text-text-muted truncate text-xs">
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
            color={isConnected ? "danger" : "tertiary"}
            disabled={disabled || isConnecting}
            leftIcon={buttonIcon}
            onClick={() => {
              if (isConnected) {
                disconnect();
              } else {
                void connect();
              }
            }}
            rounded="full"
            size="sm"
            variant={isConnected ? "solid" : "outline"}
          >
            {label}
          </Button>
        </Tooltip>
      </Flex>
    </Box>
  );
};
