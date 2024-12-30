import { cn } from "lib";
import React from "react";

export const TeamColor = ({
  color = "red",
  className,
}: {
  color?: string;
  className?: string;
}) => {
  return (
    <div
      className={cn("size-3.5 rounded", className)}
      style={{ backgroundColor: color }}
    />
  );
};
