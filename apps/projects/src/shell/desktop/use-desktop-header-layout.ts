import { useEffect, useRef } from "react";

export const useDesktopHeaderLayout = (enabled: boolean) => {
  const headerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!enabled) return;
    const header = headerRef.current;
    const actions = header?.querySelector<HTMLElement>(
      "[data-desktop-header-actions]",
    );
    if (!header || !actions) return;

    // Equal side tracks keep search centered even when permissions, labels, or
    // workspace terminology change the width of the right-hand actions.
    const observer = new ResizeObserver(() => {
      header.style.setProperty(
        "--desktop-header-side-width",
        `${Math.ceil(actions.getBoundingClientRect().width) + 28}px`,
      );
    });
    observer.observe(actions);

    return () => {
      observer.disconnect();
      header.style.removeProperty("--desktop-header-side-width");
    };
  }, [enabled]);

  return headerRef;
};
