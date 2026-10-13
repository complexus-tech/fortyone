import type { MouseEvent } from "react";
import type { DesktopWindowAction } from "./bridge";
import { getDesktopBridge } from "./bridge";

const INTERACTIVE_ELEMENTS = [
  "a",
  "button",
  "input",
  "select",
  "textarea",
  "summary",
  '[role="button"]',
  '[role="link"]',
  '[role="menuitem"]',
  '[role="tab"]',
  '[role="checkbox"]',
  '[role="radio"]',
  '[role="switch"]',
  '[role="slider"]',
  '[role="combobox"]',
  '[role="textbox"]',
  "[tabindex]",
  '[contenteditable]:not([contenteditable="false"])',
  "[data-desktop-no-drag]",
].join(",");

const isBlankRegion = (event: MouseEvent<HTMLElement>) =>
  !event.defaultPrevented &&
  event.button === 0 &&
  event.target instanceof Element &&
  !event.target.closest(INTERACTIVE_ELEMENTS);

const performWindowAction = (
  event: MouseEvent<HTMLElement>,
  action: DesktopWindowAction,
) => {
  const bridge = getDesktopBridge();
  if (!bridge || !isBlankRegion(event)) return;
  event.preventDefault();

  void bridge.windowAction(action).catch((error: unknown) => {
    // eslint-disable-next-line no-console -- Native failures must remain visible in desktop diagnostics.
    console.error("Could not update the desktop window", error);
  });
};

export const DESKTOP_DRAG_HANDLERS = {
  onMouseDown: (event: MouseEvent<HTMLElement>) => {
    if (event.detail === 1) performWindowAction(event, "drag");
  },
  onDoubleClick: (event: MouseEvent<HTMLElement>) => {
    performWindowAction(event, "toggle-maximize");
  },
};

export const EMPTY_DRAG_HANDLERS = {
  onMouseDown: undefined,
  onDoubleClick: undefined,
};
