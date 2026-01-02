import { cn } from "lib";
import { Box } from "./box";

export const ProgressBar = ({
  progress,
  className,
}: {
  progress: number;
  className?: string;
}) => {
  return (
    <Box className={cn("h-1.5 w-full rounded bg-surface-muted", className)}>
      <Box
        className={cn("h-full rounded bg-danger", {
          "bg-warning": progress >= 25 && progress < 50,
          "bg-info": progress >= 50 && progress < 75,
          "bg-success": progress >= 75,
        })}
        style={{ width: `${progress}%` }}
      />
    </Box>
  );
};
