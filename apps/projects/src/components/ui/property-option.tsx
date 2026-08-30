"use client";

import type { ReactNode } from "react";
import { Box, Text } from "ui";
import { cn } from "lib";
import { useMediaQuery } from "@/hooks/media";

export type PropertyOptionProps = {
  label: string;
  value: ReactNode;
  className?: string;
  isCompact?: boolean;
  isNotifications: boolean;
};

export const PropertyOption = ({
  label,
  value,
  className,
  isCompact = false,
  isNotifications,
}: PropertyOptionProps) => {
  const isMobile = useMediaQuery("(max-width: 768px)");

  if (isMobile || isCompact) {
    return value;
  }

  return (
    <Box
      className={cn(
        "my-4 grid grid-cols-[7.875rem_auto] items-center gap-3 md:my-5",
        { "grid-cols-1": isNotifications },
        className,
      )}
    >
      {!isNotifications ? (
        <Text
          className="flex items-center gap-1 truncate"
          color="muted"
          fontWeight="medium"
        >
          {label}
        </Text>
      ) : null}
      {value}
    </Box>
  );
};
