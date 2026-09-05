import { useCallback, useState } from "react";

type SetStoredValue<T> = T | ((value: T) => T);

const readValue = <T>(key: string, initialValue: T): T => {
  const item = typeof window !== "undefined" ? localStorage.getItem(key) : null;
  return item ? (JSON.parse(item) as T) : initialValue;
};

export const useLocalStorage = <T>(
  key: string,
  initialValue: T,
): [T, (value: SetStoredValue<T>) => void] => {
  const [stored, setStored] = useState(() => ({
    key,
    value: readValue(key, initialValue),
  }));

  // Restore the new scope before children can persist options from the old one.
  let current = stored;
  if (stored.key !== key) {
    current = { key, value: readValue(key, initialValue) };
    setStored(current);
  }

  const setValue = useCallback(
    (value: SetStoredValue<T>) => {
      setStored((currentValue) => {
        const nextValue =
          value instanceof Function ? value(currentValue.value) : value;

        if (typeof window !== "undefined") {
          localStorage.setItem(key, JSON.stringify(nextValue));
        }

        if (
          currentValue.key === key &&
          Object.is(currentValue.value, nextValue)
        ) {
          return currentValue;
        }

        return { key, value: nextValue };
      });
    },
    [key],
  );

  return [current.value, setValue];
};
