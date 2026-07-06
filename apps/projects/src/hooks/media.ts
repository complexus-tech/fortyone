import { useSyncExternalStore } from "react";

export const useMediaQuery = (query: string) => {
  return useSyncExternalStore(
    (onStoreChange) => {
      if (typeof window === "undefined") return () => {};

      const mediaQuery = window.matchMedia(query);
      const handler = () => {
        onStoreChange();
      };

      mediaQuery.addEventListener("change", handler);

      return () => {
        mediaQuery.removeEventListener("change", handler);
      };
    },
    () =>
      typeof window === "undefined" ? false : window.matchMedia(query).matches,
    () => false,
  );
};
