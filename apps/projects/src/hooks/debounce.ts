import { useCallback, useEffect, useRef } from "react";

type DebounceOptions = {
  flushOnUnmount?: boolean;
};

type DebouncedCallbackControls<T> = {
  callback: (data: T) => void;
  cancel: () => void;
  flush: () => void;
};

export const useDebouncedCallback = <T>(
  callback: (data: T) => void,
  delay = 500,
  { flushOnUnmount = false }: DebounceOptions = {},
): DebouncedCallbackControls<T> => {
  const callbackRef = useRef(callback);
  const pendingDataRef = useRef<T | undefined>(undefined);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    callbackRef.current = callback;
  }, [callback]);

  const cancel = useCallback(() => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
    pendingDataRef.current = undefined;
  }, []);

  const flush = useCallback(() => {
    if (!timeoutRef.current) return;

    clearTimeout(timeoutRef.current);
    timeoutRef.current = null;

    const pendingData = pendingDataRef.current;
    pendingDataRef.current = undefined;
    if (pendingData !== undefined) {
      callbackRef.current(pendingData);
    }
  }, []);

  const debouncedCallback = useCallback(
    (data: T) => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }

      pendingDataRef.current = data;
      timeoutRef.current = setTimeout(() => {
        timeoutRef.current = null;
        pendingDataRef.current = undefined;
        callbackRef.current(data);
      }, delay);
    },
    [delay],
  );

  useEffect(
    () => () => {
      if (flushOnUnmount) {
        flush();
      } else {
        cancel();
      }
    },
    [cancel, flush, flushOnUnmount],
  );

  return { callback: debouncedCallback, cancel, flush };
};

export const useDebounce = <T>(callback: (data: T) => void, delay = 500) => {
  const { callback: debouncedCallback } = useDebouncedCallback(callback, delay);
  return debouncedCallback;
};
