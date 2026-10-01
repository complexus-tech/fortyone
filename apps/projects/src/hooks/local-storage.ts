import { useCallback, useEffect, useState, useSyncExternalStore } from "react";

type SetStoredValue<T> = T | ((value: T) => T);

type LocalStorageOptions = {
  /** Use the initial value for SSR and hydration, then restore browser storage. */
  initializeWithValue?: boolean;
};

const subscribeToHydration = () => () => {};
const getClientSnapshot = () => true;
const getServerSnapshot = () => false;

const readValue = <T>(key: string, initialValue: T): T => {
  const item = typeof window !== "undefined" ? localStorage.getItem(key) : null;
  return item ? (JSON.parse(item) as T) : initialValue;
};

export const useLocalStorage = <T>(
  key: string,
  initialValue: T,
  { initializeWithValue = true }: LocalStorageOptions = {},
): [T, (value: SetStoredValue<T>) => void] => {
  const isHydrated = useSyncExternalStore(
    subscribeToHydration,
    getClientSnapshot,
    getServerSnapshot,
  );
  const canReadStorage = initializeWithValue || isHydrated;
  const [stored, setStored] = useState(() => ({
    key,
    initialized: canReadStorage,
    shouldPersist: false,
    value: canReadStorage ? readValue(key, initialValue) : initialValue,
  }));

  // Restore the new scope before children can persist options from the old one.
  // Track initialization separately so object defaults cannot reset restored data.
  let current = stored;
  if (stored.key !== key || (!stored.initialized && canReadStorage)) {
    current = {
      key,
      initialized: canReadStorage,
      shouldPersist: false,
      value: canReadStorage ? readValue(key, initialValue) : initialValue,
    };
    setStored(current);
  }

  const setValue = useCallback(
    (value: SetStoredValue<T>) => {
      setStored((currentValue) => {
        const nextValue =
          value instanceof Function ? value(currentValue.value) : value;

        if (
          currentValue.key === key &&
          currentValue.shouldPersist &&
          Object.is(currentValue.value, nextValue)
        ) {
          return currentValue;
        }

        return {
          key,
          initialized: true,
          shouldPersist: true,
          value: nextValue,
        };
      });
    },
    [key],
  );

  // Persist committed user changes; React may replay functional updaters.
  useEffect(() => {
    if (stored.key === key && stored.shouldPersist) {
      localStorage.setItem(key, JSON.stringify(stored.value));
    }
  }, [key, stored]);

  return [current.value, setValue];
};
