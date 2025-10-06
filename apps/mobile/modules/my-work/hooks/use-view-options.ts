import { useState, useEffect, useCallback } from "react";
import AsyncStorage from "@react-native-async-storage/async-storage";
import type { StoriesViewOptions } from "@/types/stories-view-options";

const STORAGE_KEY = "my-work:view-options";

const defaultViewOptions: StoriesViewOptions = {
  groupBy: "status",
  orderBy: "created",
  orderDirection: "desc",
};

export const useViewOptions = () => {
  const [viewOptions, setViewOptionsState] =
    useState<StoriesViewOptions>(defaultViewOptions);
  const [isLoaded, setIsLoaded] = useState(false);

  // Load from AsyncStorage on mount
  useEffect(() => {
    const loadViewOptions = async () => {
      try {
        const stored = await AsyncStorage.getItem(STORAGE_KEY);
        if (stored) {
          const parsed = JSON.parse(stored);
          setViewOptionsState(parsed);
        }
      } catch (error) {
        console.error("Failed to load view options:", error);
      } finally {
        setIsLoaded(true);
      }
    };

    loadViewOptions();
  }, []);

  // Save to AsyncStorage whenever options change
  const setViewOptions = useCallback(
    async (newOptions: Partial<StoriesViewOptions>) => {
      const updated = { ...viewOptions, ...newOptions };
      setViewOptionsState(updated);

      try {
        await AsyncStorage.setItem(STORAGE_KEY, JSON.stringify(updated));
      } catch (error) {
        console.error("Failed to save view options:", error);
      }
    },
    [viewOptions]
  );

  const resetViewOptions = useCallback(async () => {
    setViewOptionsState(defaultViewOptions);
    try {
      await AsyncStorage.removeItem(STORAGE_KEY);
    } catch (error) {
      console.error("Failed to reset view options:", error);
    }
  }, []);

  return {
    viewOptions,
    setViewOptions,
    resetViewOptions,
    isLoaded,
  };
};
