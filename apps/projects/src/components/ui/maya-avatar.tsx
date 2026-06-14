import { AiIcon } from "icons";
import { cn } from "lib";
import { Avatar } from "ui";

const sizeClasses = {
  xs: "size-5",
  sm: "size-6",
  md: "size-8",
} as const;

const iconSizeClasses = {
  xs: "size-3",
  sm: "size-3.5",
  md: "size-4.5",
} as const;

type MayaAvatarProps = {
  className?: string;
  name?: string;
  size?: keyof typeof sizeClasses;
  src?: string | null;
};

export const MayaAvatar = ({
  className,
  name = "Maya",
  size = "sm",
  src,
}: MayaAvatarProps) => {
  if (src) {
    return <Avatar className={className} name={name} size={size} src={src} />;
  }

  return (
    <span
      className={cn(
        "inline-flex shrink-0 items-center justify-center rounded-full bg-black text-white dark:bg-white dark:text-black",
        sizeClasses[size],
        className,
      )}
    >
      <AiIcon className={cn("text-current", iconSizeClasses[size])} />
    </span>
  );
};
