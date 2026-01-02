"use client";
import { cn } from "lib";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";
import { Box } from "ui";

export const RowWrapper = ({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) => {
  const pathname = usePathname();
  return (
    <Box
      className={cn(
        "border-border group hover:bg-state-hover/50 focus-visible:bg-state-active flex items-center justify-between border-b-[0.5px] px-5 py-[0.655rem] outline-none md:px-12",
        {
          "pr-0 pl-0 md:pl-7": pathname.startsWith("/story"),
        },
        className,
      )}
      tabIndex={0}
    >
      {children}
    </Box>
  );
};
