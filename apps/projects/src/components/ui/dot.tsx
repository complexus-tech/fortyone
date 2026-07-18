import { cn } from "lib";

export const Dot = ({
  className,
  color,
}: {
  className?: string;
  color?: string;
}) => {
  return (
    <span
      aria-hidden="true"
      className={cn("inline-block size-2 shrink-0 rounded-sm", className)}
      style={{ backgroundColor: color ?? "currentColor" }}
    />
  );
};
