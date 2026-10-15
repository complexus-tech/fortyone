"use client";

import { Button, Tooltip } from "ui";
import { cn } from "lib";
import { StopIcon, VoiceIcon } from "icons";
import type { useMayaRealtimeVoice } from "../hooks/use-maya-realtime-voice";

type RealtimeVoiceControlProps = {
  color?: "invert" | "tertiary";
  disabled?: boolean;
  disabledReason?: string;
  isHighlighted?: boolean;
  voice: ReturnType<typeof useMayaRealtimeVoice>;
};

const getButtonLabel = (voice: ReturnType<typeof useMayaRealtimeVoice>) => {
  const { isSpeaking, remainingSeconds, status } = voice;
  if (status === "connecting") return "Connecting";
  if (status === "disconnecting") return "Ending";
  if (status === "connected") {
    const seconds = Math.max(0, remainingSeconds ?? 0);
    const time = `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(
      2,
      "0",
    )}`;
    return `${isSpeaking ? "Speaking" : "Listening"} · ${time}`;
  }
  return "Live voice";
};

export const RealtimeVoiceControl = ({
  color = "invert",
  disabled = false,
  disabledReason,
  isHighlighted = false,
  voice,
}: RealtimeVoiceControlProps) => {
  const { connect, disconnect, status } = voice;
  const isConnected = status === "connected";
  const isConnecting = status === "connecting";
  const isDisconnecting = status === "disconnecting";
  const isActive = status !== "idle";
  const label = getButtonLabel(voice);
  let tooltip = disabledReason ?? "Start a realtime voice conversation";
  if (isActive) tooltip = "End the realtime voice session";
  if (isConnecting) tooltip = "Cancel voice connection";

  const buttonIcon = (() => {
    if (isConnected) {
      return <StopIcon className="h-5 w-auto text-current dark:text-current" />;
    }
    return <VoiceIcon className="h-5 w-auto text-current" />;
  })();

  return (
    <Tooltip side="bottom" title={tooltip}>
      <span className="inline-flex">
        <Button
          aria-label={label}
          className={cn("shrink-0 gap-1.5", {
            "bg-state-hover dark:bg-state-hover": isHighlighted,
          })}
          color={color}
          disabled={(disabled && !isActive) || isDisconnecting}
          leftIcon={buttonIcon}
          onClick={() => {
            if (isActive) {
              disconnect();
            } else {
              void connect();
            }
          }}
          rounded="md"
        >
          {label}
        </Button>
      </span>
    </Tooltip>
  );
};
