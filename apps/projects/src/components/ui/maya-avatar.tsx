import { AiIcon } from "icons";
import { cn } from "lib";

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
  size?: keyof typeof sizeClasses;
};

export const MayaAvatar = ({ className, size = "sm" }: MayaAvatarProps) => (
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
