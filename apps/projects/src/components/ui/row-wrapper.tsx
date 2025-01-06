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
        "group flex items-center justify-between border-b-[0.5px] border-gray-100 px-12 py-[0.655rem] outline-none hover:bg-gray-50/50 focus:bg-gray-50/50 dark:border-dark-100/70 hover:dark:bg-dark-200/20 focus:dark:bg-dark-200/50",
        className,
        {
          "pl-7 pr-0": pathname.startsWith("/story"),
        },
      )}
      tabIndex={0}
    >
      {children}
    </Box>
  );
};
