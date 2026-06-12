"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { usePathname } from "next/navigation";
import type { StoriesFilter } from "./stories-filter-types";
import { DEFAULT_STORIES_FILTER } from "./stories-filter-types";
import {
  getStoriesFilterStorageKey,
  mergeStoriesFilterDefaults,
} from "./stories-filter-storage";

const readStoredFilters = (storageKey: string): StoriesFilter => {
  if (typeof window === "undefined") {
    return DEFAULT_STORIES_FILTER;
  }

  try {
    const storedValue = window.localStorage.getItem(storageKey);
    if (!storedValue) {
      return DEFAULT_STORIES_FILTER;
    }

    return mergeStoriesFilterDefaults(
      JSON.parse(storedValue) as Partial<StoriesFilter>,
    );
  } catch {
    return DEFAULT_STORIES_FILTER;
  }
};

const writeStoredFilters = (storageKey: string, filters: StoriesFilter) => {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.setItem(storageKey, JSON.stringify(filters));
};

const clearStoredFilters = (storageKey: string) => {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.removeItem(storageKey);
};

export const useStoriesFilters = () => {
  const pathname = usePathname();
  const storageKey = useMemo(
    () => getStoriesFilterStorageKey(pathname),
    [pathname],
  );
  const [storedFilters, setStoredFilters] = useState<StoriesFilter>(() =>
    readStoredFilters(storageKey),
  );

  useEffect(() => {
    setStoredFilters(readStoredFilters(storageKey));
  }, [storageKey]);

  const setFilters = useCallback(
    (value: StoriesFilter) => {
      const nextFilters = mergeStoriesFilterDefaults(value);
      setStoredFilters(nextFilters);
      writeStoredFilters(storageKey, nextFilters);
    },
    [storageKey],
  );

  const resetFilters = useCallback(() => {
    setStoredFilters(DEFAULT_STORIES_FILTER);
    clearStoredFilters(storageKey);
  }, [storageKey]);

  return {
    filters: storedFilters,
    setFilters,
    resetFilters,
  };
};
