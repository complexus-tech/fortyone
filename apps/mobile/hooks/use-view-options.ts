import { useState, useEffect, useCallback } from "react";
import AsyncStorage from "@react-native-async-storage/async-storage";
import type { StoriesViewOptions } from "@/types/stories-view-options";

const defaultViewOptions: StoriesViewOptions = {
  groupBy: "status",
  orderBy: "created",
  orderDirection: "desc",
  displayColumns: ["Status", "Assignee", "Priority"],
};

export const useViewOptions = (storageKey: string) => {
  const [viewOptions, setViewOptionsState] =
    useState<StoriesViewOptions>(defaultViewOptions);
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    const loadViewOptions = async () => {
      try {
        const stored = await AsyncStorage.getItem(storageKey);
        if (stored) {
          const parsed = JSON.parse(stored);
          setViewOptionsState(parsed);
        }
      } catch (error) {
        console.error(`Failed to load view options for ${storageKey}:`, error);
      } finally {
        setIsLoaded(true);
      }
    };

    loadViewOptions();
  }, [storageKey]);

  const setViewOptions = useCallback(
    async (newOptions: Partial<StoriesViewOptions>) => {
      const updated = { ...viewOptions, ...newOptions };
      setViewOptionsState(updated);

      try {
        await AsyncStorage.setItem(storageKey, JSON.stringify(updated));
      } catch (error) {
        console.error(`Failed to save view options for ${storageKey}:`, error);
      }
    },
    [viewOptions, storageKey]
  );

  const resetViewOptions = useCallback(async () => {
    setViewOptionsState(defaultViewOptions);
    try {
      await AsyncStorage.removeItem(storageKey);
    } catch (error) {
      console.error(`Failed to reset view options for ${storageKey}:`, error);
    }
  }, [storageKey]);

  return {
    viewOptions,
    setViewOptions,
    resetViewOptions,
    isLoaded,
  };
};
