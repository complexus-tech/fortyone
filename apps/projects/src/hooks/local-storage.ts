import { useCallback, useState } from "react";

type SetStoredValue<T> = T | ((value: T) => T);

export const useLocalStorage = <T>(
  key: string,
  initialValue: T,
): [T, (value: SetStoredValue<T>) => void] => {
  const [storedValue, setStoredValue] = useState<T>(() => {
    let item = null;
    if (typeof window !== "undefined") {
      item = localStorage.getItem(key);
    }
    return item ? (JSON.parse(item) as T) : initialValue;
  });

  const setValue = useCallback(
    (value: SetStoredValue<T>) => {
      setStoredValue((currentValue) => {
        const nextValue =
          value instanceof Function ? value(currentValue) : value;

        if (typeof window !== "undefined") {
          localStorage.setItem(key, JSON.stringify(nextValue));
        }

        return nextValue;
      });
    },
    [key],
  );

  return [storedValue, setValue];
};
