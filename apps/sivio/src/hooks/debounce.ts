import { useRef, useEffect, useCallback } from "react";

/**
 * Debounce a callback function
 * @param callback - The function to debounce
 * @param delay - The delay in milliseconds
 * @returns The debounced function
 */
export const useDebounce = <T, R>(
  callback: (data: T) => Promise<R>,
  delay = 500,
) => {
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  const debouncedCallback = useCallback(
    (data: T): Promise<R> => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }

      return new Promise((resolve) => {
        timeoutRef.current = setTimeout(() => {
          resolve(callback(data));
        }, delay);
      });
    },
    [callback, delay],
  );

  return debouncedCallback;
};
