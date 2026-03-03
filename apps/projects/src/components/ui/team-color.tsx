import { cn } from "lib";

export const TeamColor = ({
  color = "red",
  className,
}: {
  color?: string;
  className?: string;
}) => {
  return (
    <div
      className={cn("size-[0.95rem] rounded", className)}
      style={{ backgroundColor: color }}
    />
  );
};
