import { useState, useEffect } from "react";
import AsyncStorage from "@react-native-async-storage/async-storage";

export type ObjectiveViewOptions = {
  groupBy: "status" | "priority" | "assignee";
  orderBy: "created" | "updated" | "deadline" | "priority";
  orderDirection: "asc" | "desc";
};

const defaultViewOptions: ObjectiveViewOptions = {
  groupBy: "status",
  orderBy: "created",
  orderDirection: "desc",
};

export const useObjectiveViewOptions = (objectiveId: string) => {
  const STORAGE_KEY = `objective-${objectiveId}:view-options`;
  const [viewOptions, setViewOptionsState] =
    useState<ObjectiveViewOptions>(defaultViewOptions);
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    const loadViewOptions = async () => {
      try {
        const storedOptions = await AsyncStorage.getItem(STORAGE_KEY);
        if (storedOptions) {
          setViewOptionsState(JSON.parse(storedOptions));
        }
      } catch (error) {
        console.error(
          "Failed to load objective view options from storage",
          error
        );
      } finally {
        setIsLoaded(true);
      }
    };

    loadViewOptions();
  }, [STORAGE_KEY]);

  const setViewOptions = async (newOptions: ObjectiveViewOptions) => {
    try {
      setViewOptionsState(newOptions);
      await AsyncStorage.setItem(STORAGE_KEY, JSON.stringify(newOptions));
    } catch (error) {
      console.error("Failed to save objective view options to storage", error);
    }
  };

  const resetViewOptions = async () => {
    try {
      setViewOptionsState(defaultViewOptions);
      await AsyncStorage.removeItem(STORAGE_KEY);
    } catch (error) {
      console.error("Failed to reset objective view options in storage", error);
    }
  };

  return { viewOptions, setViewOptions, resetViewOptions, isLoaded };
};
