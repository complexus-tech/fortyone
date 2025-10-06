import { useState, useEffect, useCallback } from "react";
import AsyncStorage from "@react-native-async-storage/async-storage";
import type { StoriesViewOptions } from "@/types/stories-view-options";

const defaultViewOptions: StoriesViewOptions = {
  groupBy: "status",
  orderBy: "created",
  orderDirection: "desc",
};

export const useTeamViewOptions = (teamId: string) => {
  const STORAGE_KEY = `team-${teamId}:view-options`;

  const [viewOptions, setViewOptionsState] =
    useState<StoriesViewOptions>(defaultViewOptions);
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    const loadViewOptions = async () => {
      try {
        const stored = await AsyncStorage.getItem(STORAGE_KEY);
        if (stored) {
          const parsed = JSON.parse(stored);
          setViewOptionsState(parsed);
        }
      } catch (error) {
        console.error("Failed to load team view options:", error);
      } finally {
        setIsLoaded(true);
      }
    };

    loadViewOptions();
  }, [STORAGE_KEY]);

  const setViewOptions = useCallback(
    async (newOptions: Partial<StoriesViewOptions>) => {
      const updated = { ...viewOptions, ...newOptions };
      setViewOptionsState(updated);

      try {
        await AsyncStorage.setItem(STORAGE_KEY, JSON.stringify(updated));
      } catch (error) {
        console.error("Failed to save team view options:", error);
      }
    },
    [viewOptions, STORAGE_KEY]
  );

  const resetViewOptions = useCallback(async () => {
    setViewOptionsState(defaultViewOptions);
    try {
      await AsyncStorage.removeItem(STORAGE_KEY);
    } catch (error) {
      console.error("Failed to reset team view options:", error);
    }
  }, [STORAGE_KEY]);

  return {
    viewOptions,
    setViewOptions,
    resetViewOptions,
    isLoaded,
  };
};
