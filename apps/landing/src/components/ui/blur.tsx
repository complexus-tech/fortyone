import { Box } from "ui";
import { cn } from "lib";

export const Blur = ({ className }: { className?: string }) => {
  return (
    <Box
      className={cn(
        "pointer-events-none h-[300px] w-[300px] rounded-full blur-3xl",
        className,
      )}
    />
  );
};
