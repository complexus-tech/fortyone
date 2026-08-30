"use client";

import { useRef } from "react";
import { useHotkeys } from "react-hotkeys-hook";

export const useOptionHotkey = (shortcut: string, enabled: boolean) => {
  const buttonRef = useRef<HTMLButtonElement>(null);

  useHotkeys(shortcut, (event) => {
    event.preventDefault();

    if (enabled) {
      buttonRef.current?.click();
    }
  });

  return buttonRef;
};
