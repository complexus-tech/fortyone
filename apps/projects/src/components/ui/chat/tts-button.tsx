import type { ButtonProps } from "ui";
import { Button } from "ui";
import { cn } from "lib";
import { VolumeIcon, LoadingIcon, StopIcon } from "icons";
import { useTTS } from "@/hooks/use-tts";

export const TTSButton = ({
  text,
  size = "sm",
  className,
}: ButtonProps & { text: string }) => {
  const { playText, stop, currentAudio, isPlaying, isLoading } = useTTS();

  const isCurrentText = currentAudio?.text === text;
  const isCurrentPlaying = isCurrentText && isPlaying;
  const isAnyLoading = isLoading;

  const handleClick = async () => {
    if (isCurrentPlaying) {
      stop();
    } else {
      stop();
      await playText(text);
    }
  };

  const getButtonIcon = () => {
    if (isAnyLoading) {
      return <LoadingIcon className="animate-spin" />;
    }
    if (isCurrentPlaying) {
      return <StopIcon />;
    }
    return <VolumeIcon />;
  };

  return (
    <Button
      asIcon
      className={cn("transition-all duration-200", className)}
      color="tertiary"
      onClick={handleClick}
      size={size}
      variant="naked"
    >
      {getButtonIcon()}
    </Button>
  );
};
