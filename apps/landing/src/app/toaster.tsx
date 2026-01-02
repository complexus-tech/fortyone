"use client";

import type { ComponentProps } from "react";
import { Toaster as Sonner } from "sonner";
import { useTheme } from "next-themes";
import { toasterIcons } from "@/app/toaster-icons";

type ToasterProps = ComponentProps<typeof Sonner>;

export const Toaster = (_: ToasterProps) => {
  const { theme } = useTheme();
  return (
    <Sonner
      closeButton
      duration={10000}
      icons={toasterIcons}
      position="bottom-right"
      theme={theme as "light" | "dark" | "system"}
      toastOptions={{
        className:
          "w-full rounded-[0.6rem] p-4 flex items-center gap-3 shadow-lg",
        classNames: {
          toast:
            "bg-white/90 dark:bg-dark-100/90 backdrop-blur border border-border/60 d",
          closeButton: "bg-surface-elevated/90 border-border",
        },
      }}
    />
  );
};
