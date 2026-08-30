"use client";

import { useCallback, useMemo, useState, useSyncExternalStore } from "react";
import {
  parseStoredStrategyNodePositions,
  type StrategyNodePositions,
} from "./strategy-map-layout";

type StrategyMapStorageObjective = {
  id: string;
  keyResultCount: number;
};

export const LAYOUT_STORAGE_VERSION = 2;
export const KEY_RESULT_OWNER_STORAGE_VERSION = 1;
export const EXPANSION_STORAGE_VERSION = 1;
export const EXPANSION_STORAGE_EVENT = "strategy-map-expansion-change";

const subscribeToStaticStorage = () => () => undefined;

const subscribeToExpansionStorage = (onStoreChange: () => void) => {
  window.addEventListener("storage", onStoreChange);
  window.addEventListener(EXPANSION_STORAGE_EVENT, onStoreChange);
  return () => {
    window.removeEventListener("storage", onStoreChange);
    window.removeEventListener(EXPANSION_STORAGE_EVENT, onStoreChange);
  };
};

const parseCollapsedObjectiveIds = (value: string | null) => {
  try {
    const parsed = value ? (JSON.parse(value) as unknown) : [];
    return new Set(
      Array.isArray(parsed)
        ? parsed.filter((item): item is string => typeof item === "string")
        : [],
    );
  } catch {
    return new Set<string>();
  }
};

const parseStoredKeyResultOwnerIds = (value: string | null) => {
  try {
    const parsed = value ? (JSON.parse(value) as unknown) : {};
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
      return new Map<string, string>();
    }

    const ownerIds = new Map<string, string>();
    Object.entries(parsed).forEach(([nodeId, objectiveId]) => {
      if (nodeId.startsWith("key-result:") && typeof objectiveId === "string") {
        ownerIds.set(nodeId, objectiveId);
      }
    });
    return ownerIds;
  } catch {
    return new Map<string, string>();
  }
};

export const useStrategyMapCanvasStorage = ({
  objectives,
  workspaceSlug,
}: {
  objectives: readonly StrategyMapStorageObjective[];
  workspaceSlug: string;
}) => {
  const expansionStorageKey = `strategy-map-expansion:v${EXPANSION_STORAGE_VERSION}:${workspaceSlug}`;
  const getStoredExpansionSnapshot = useCallback(
    () => window.localStorage.getItem(expansionStorageKey),
    [expansionStorageKey],
  );
  const storedExpansionValue = useSyncExternalStore(
    subscribeToExpansionStorage,
    getStoredExpansionSnapshot,
    () => null,
  );
  const collapsedObjectiveIds = useMemo(
    () => parseCollapsedObjectiveIds(storedExpansionValue),
    [storedExpansionValue],
  );
  const expandedObjectiveIds = useMemo(() => {
    const result = new Set<string>();
    objectives.forEach(({ id, keyResultCount }) => {
      if (keyResultCount > 0 && !collapsedObjectiveIds.has(id)) {
        result.add(id);
      }
    });
    return result;
  }, [collapsedObjectiveIds, objectives]);
  const storageKey = `strategy-map-layout:v${LAYOUT_STORAGE_VERSION}:${workspaceSlug}`;
  const keyResultOwnerStorageKey = `strategy-map-key-result-owner:v${KEY_RESULT_OWNER_STORAGE_VERSION}:${workspaceSlug}`;
  const getStoredLayoutSnapshot = useCallback(
    () => window.localStorage.getItem(storageKey),
    [storageKey],
  );
  const storedLayoutValue = useSyncExternalStore(
    subscribeToStaticStorage,
    getStoredLayoutSnapshot,
    () => null,
  );
  const storedPositions = useMemo(
    () => parseStoredStrategyNodePositions(storedLayoutValue),
    [storedLayoutValue],
  );
  const getStoredKeyResultOwnerSnapshot = useCallback(
    () => window.localStorage.getItem(keyResultOwnerStorageKey),
    [keyResultOwnerStorageKey],
  );
  const storedKeyResultOwnerValue = useSyncExternalStore(
    subscribeToStaticStorage,
    getStoredKeyResultOwnerSnapshot,
    () => null,
  );
  const storedKeyResultOwnerIds = useMemo(
    () => parseStoredKeyResultOwnerIds(storedKeyResultOwnerValue),
    [storedKeyResultOwnerValue],
  );
  const [transientLayout, setTransientLayout] = useState<{
    positions: StrategyNodePositions;
    storageKey: string;
  }>({ positions: {}, storageKey });
  const transientPositions =
    transientLayout.storageKey === storageKey ? transientLayout.positions : {};
  const setTransientPositions = useCallback(
    (positions: StrategyNodePositions) => {
      setTransientLayout({ positions, storageKey });
    },
    [storageKey],
  );
  const toggleObjectiveKeyResults = useCallback(
    (objectiveId: string) => {
      const next = new Set(collapsedObjectiveIds);
      if (next.has(objectiveId)) next.delete(objectiveId);
      else next.add(objectiveId);

      try {
        window.localStorage.setItem(
          expansionStorageKey,
          JSON.stringify(Array.from(next)),
        );
        window.dispatchEvent(new Event(EXPANSION_STORAGE_EVENT));
      } catch {
        // Expansion remains at its last persisted state if storage is unavailable.
      }
    },
    [collapsedObjectiveIds, expansionStorageKey],
  );

  return {
    expandedObjectiveIds,
    keyResultOwnerStorageKey,
    setTransientPositions,
    storageKey,
    storedKeyResultOwnerIds,
    storedPositions,
    toggleObjectiveKeyResults,
    transientPositions,
  };
};

export const persistStrategyMapCanvasLayout = ({
  keyResultOwnerIds,
  keyResultOwnerStorageKey,
  positions,
  storageKey,
}: {
  keyResultOwnerIds: ReadonlyMap<string, string>;
  keyResultOwnerStorageKey: string;
  positions: StrategyNodePositions;
  storageKey: string;
}) => {
  try {
    window.localStorage.setItem(storageKey, JSON.stringify(positions));
    window.localStorage.setItem(
      keyResultOwnerStorageKey,
      JSON.stringify(Object.fromEntries(keyResultOwnerIds)),
    );
  } catch {
    // The canvas remains usable if storage is unavailable or full.
  }
};

export const persistStrategyMapKeyResultOwners = ({
  keyResultOwnerIds,
  keyResultOwnerStorageKey,
}: {
  keyResultOwnerIds: ReadonlyMap<string, string>;
  keyResultOwnerStorageKey: string;
}) => {
  try {
    window.localStorage.setItem(
      keyResultOwnerStorageKey,
      JSON.stringify(Object.fromEntries(keyResultOwnerIds)),
    );
  } catch {
    // Existing node positions remain usable when storage is unavailable or full.
  }
};
