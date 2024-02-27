import { useState } from "react";

type LocalStorageValue<T> = T | null;

export const useLocalStorage = <T>(
  key: string,
  initialValue: T,
): [LocalStorageValue<T>, (value: T) => void] => {
  const [storedValue, setStoredValue] = useState<LocalStorageValue<T>>(() => {
    let item = null;
    if (typeof window !== "undefined") {
      item = localStorage.getItem(key);
    }
    return item ? (JSON.parse(item) as T) : initialValue;
  });

  const setValue = (value: T) => {
    setStoredValue(value);
    if (typeof window !== "undefined") {
      localStorage.setItem(key, JSON.stringify(value));
    }
  };

  return [storedValue, setValue];
};
