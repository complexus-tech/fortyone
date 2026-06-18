"use client";

import { Button, Tooltip } from "ui";
import { LoadingIcon, StopIcon, VoiceIcon } from "icons";
import type { useMayaRealtimeVoice } from "../hooks/use-maya-realtime-voice";

type RealtimeVoiceControlProps = {
  disabled?: boolean;
  voice: ReturnType<typeof useMayaRealtimeVoice>;
};

const getButtonLabel = (
  status: ReturnType<typeof useMayaRealtimeVoice>["status"],
) => {
  if (status === "connecting") return "Connecting";
  if (status === "connected") return "End voice";
  return "Live voice";
};

export const RealtimeVoiceControl = ({
  disabled = false,
  voice,
}: RealtimeVoiceControlProps) => {
  const { connect, disconnect, status } = voice;
  const isConnected = status === "connected";
  const isConnecting = status === "connecting";
  const label = getButtonLabel(status);

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
    <Tooltip
      side="bottom"
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
        rounded="full"
      >
        {label}
      </Button>
    </Tooltip>
  );
};
