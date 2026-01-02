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
        "border-border group flex items-center justify-between border-b-[0.5px] px-5 py-[0.655rem] outline-none hover:bg-state-hover focus-visible:bg-state-active md:px-12",
        {
          "pl-0 pr-0 md:pl-7": pathname.startsWith("/story"),
        },
        className,
      )}
      tabIndex={0}
    >
      {children}
    </Box>
  );
};
