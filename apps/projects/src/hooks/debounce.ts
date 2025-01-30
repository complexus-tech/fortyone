import { useRef, useEffect, useCallback } from "react";

export const useDebounce = <T>(callback: (data: T) => void, delay = 500) => {
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  const debouncedCallback = useCallback(
    (data: T) => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }

      timeoutRef.current = setTimeout(() => {
        callback(data);
      }, delay);
    },
    [callback, delay],
  );

  return debouncedCallback;
};
