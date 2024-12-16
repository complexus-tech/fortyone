"use client";

import { useTheme } from "next-themes";
import { ComponentProps } from "react";
import { Toaster as Sonner } from "sonner";
import { toasterIcons } from "@/app/toaster-icons";

type ToasterProps = ComponentProps<typeof Sonner>;

export const Toaster = ({ ...props }: ToasterProps) => {
  const { theme = "system" } = useTheme();

  return (
    <Sonner
      theme={theme as ToasterProps["theme"]}
      closeButton
      position="bottom-right"
      duration={10000}
      toastOptions={{
        className: "w-full rounded-lg p-4 flex items-center gap-3 shadow-lg",
        classNames: {
          toast:
            "bg-white/90 dark:bg-dark-100/90 backdrop-blur border border-gray-100/60 dark:border-dark-50",
          closeButton: "bg-white/90 dark:bg-dark-100/90 dark:border-dark-50",
        },
      }}
      icons={toasterIcons}
    />
  );
};
