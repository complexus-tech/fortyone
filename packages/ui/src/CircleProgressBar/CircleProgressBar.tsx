import React from "react";
import { cn } from "lib";
import { Box } from "../Box/Box";

export const CircleProgressBar = ({
  progress,
  size = 24,
  strokeWidth = 2,
  className,
  children,
  invertColors = false,
}: {
  progress: number;
  size?: number;
  strokeWidth?: number;
  className?: string;
  children?: React.ReactNode;
  invertColors?: boolean;
}) => {
  const radius = (size - strokeWidth) / 2;
  const circumference = 2 * Math.PI * radius;
  const strokeDashoffset = circumference - (progress / 100) * circumference;

  return (
    <Box
      className={cn(
        "relative inline-flex items-center justify-center",
        className
      )}
    >
      <svg
        width={size}
        height={size}
        viewBox={`0 0 ${size} ${size}`}
        className="rotate-[-90deg]"
      >
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          className="fill-none stroke-gray-100/50 dark:stroke-dark-100"
          strokeWidth={strokeWidth}
        />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          className={cn("fill-none transition-all duration-300 ease-in-out", {
            ...(invertColors
              ? {
                  "stroke-success": progress < 25,
                  "stroke-info": progress >= 25 && progress < 50,
                  "stroke-warning": progress >= 50 && progress < 75,
                  "stroke-danger": progress >= 75,
                }
              : {
                  "stroke-danger": progress < 25,
                  "stroke-warning": progress >= 25 && progress < 50,
                  "stroke-info": progress >= 50 && progress < 75,
                  "stroke-success": progress >= 75,
                }),
          })}
          strokeWidth={strokeWidth}
          strokeDasharray={circumference}
          strokeDashoffset={strokeDashoffset}
          strokeLinecap="round"
        />
      </svg>
      {children && (
        <Box className="absolute inset-0 flex items-center justify-center">
          {children}
        </Box>
      )}
    </Box>
  );
};
