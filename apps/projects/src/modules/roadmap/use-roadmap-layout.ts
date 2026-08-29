"use client";

import { useCallback, useEffect } from "react";
import { parseAsStringLiteral, useQueryState } from "nuqs";
import { useLocalStorage } from "@/hooks";
import {
  ROADMAP_LAYOUTS,
  type RoadmapLayoutType,
} from "@/modules/roadmap/types";

const DEFAULT_ROADMAP_LAYOUT: RoadmapLayoutType = "gantt";

export const useRoadmapLayout = () => {
  const [storedLayout, setStoredLayout] = useLocalStorage<RoadmapLayoutType>(
    "objectivesLayout",
    DEFAULT_ROADMAP_LAYOUT,
  );
  const [queryLayout, setQueryLayout] = useQueryState(
    "view",
    parseAsStringLiteral(ROADMAP_LAYOUTS).withOptions({
      clearOnDefault: false,
      history: "push",
    }),
  );
  const layout = queryLayout ?? storedLayout;

  useEffect(() => {
    if (queryLayout === null) {
      void setQueryLayout(storedLayout, { history: "replace" });
      return;
    }

    if (queryLayout !== storedLayout) {
      setStoredLayout(queryLayout);
    }
  }, [queryLayout, setQueryLayout, setStoredLayout, storedLayout]);

  const setLayout = useCallback(
    (nextLayout: RoadmapLayoutType) => {
      setStoredLayout(nextLayout);
      void setQueryLayout(nextLayout);
    },
    [setQueryLayout, setStoredLayout],
  );

  return { layout, setLayout };
};
