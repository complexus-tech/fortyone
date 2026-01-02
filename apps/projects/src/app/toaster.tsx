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
          "w-full! rounded-2xl! p-4! flex! items-center! gap-3! shadow-lg!",
        classNames: {
          toast:
            "bg-surface-elevated/80! backdrop-blur border! border-border/60!",
          closeButton: "bg-surface-elevated/90! border-border!",
          description: "text-text-muted!",
        },
      }}
    />
  );
};
