import { useState, useEffect } from "react";
import AsyncStorage from "@react-native-async-storage/async-storage";
import type { SprintViewOptions } from "../types";

const defaultViewOptions: SprintViewOptions = {
  groupBy: "status",
  orderBy: "created",
  orderDirection: "desc",
};

export const useSprintViewOptions = (sprintId: string) => {
  const STORAGE_KEY = `sprint-${sprintId}:view-options`;
  const [viewOptions, setViewOptionsState] =
    useState<SprintViewOptions>(defaultViewOptions);
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    const loadViewOptions = async () => {
      try {
        const storedOptions = await AsyncStorage.getItem(STORAGE_KEY);
        if (storedOptions) {
          setViewOptionsState(JSON.parse(storedOptions));
        }
      } catch (error) {
        console.error("Failed to load sprint view options from storage", error);
      } finally {
        setIsLoaded(true);
      }
    };

    loadViewOptions();
  }, [STORAGE_KEY]);

  const setViewOptions = async (newOptions: SprintViewOptions) => {
    try {
      setViewOptionsState(newOptions);
      await AsyncStorage.setItem(STORAGE_KEY, JSON.stringify(newOptions));
    } catch (error) {
      console.error("Failed to save sprint view options to storage", error);
    }
  };

  const resetViewOptions = async () => {
    try {
      setViewOptionsState(defaultViewOptions);
      await AsyncStorage.removeItem(STORAGE_KEY);
    } catch (error) {
      console.error("Failed to reset sprint view options in storage", error);
    }
  };

  return { viewOptions, setViewOptions, resetViewOptions, isLoaded };
};
