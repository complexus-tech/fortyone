"use client";

import { useTheme } from "next-themes";
import type { ComponentProps } from "react";
import { Toaster as Sonner } from "sonner";
import { toasterIcons } from "@/app/toaster-icons";

type ToasterProps = ComponentProps<typeof Sonner>;

export const Toaster = (_: ToasterProps) => {
  const { resolvedTheme } = useTheme();

  return (
    <Sonner
      closeButton
      duration={8000}
      icons={toasterIcons}
      theme={resolvedTheme as ToasterProps["theme"]}
      toastOptions={{
        className:
          "!w-full !rounded-2xl !p-4 !flex !items-center !gap-3 !shadow-lg",
        classNames: {
          toast:
            "!bg-white/80 dark:!bg-dark-100/80 backdrop-blur !border !border-gray-100/60 dark:!border-dark-50",
          closeButton: "!bg-white/90 dark:!bg-dark-100/90 dark:!border-dark-50",
          description: "!text-gray dark:!text-gray-300",
        },
      }}
    />
  );
};
