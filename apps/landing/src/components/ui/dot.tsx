import React from "react";
import { cn } from "lib";

export const Dot = ({
  className,
  color,
}: {
  className?: string;
  color?: string;
}) => {
  return (
    <svg
      className={cn("size-2", className)}
      fill="none"
      height="24"
      style={{ color }}
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <circle cx={12} cy={12} fill="currentColor" r={12} />
    </svg>
  );
};
