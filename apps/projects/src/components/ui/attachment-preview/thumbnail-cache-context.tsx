"use client";

import type { ReactNode } from "react";
import { createContext, useContext, useEffect, useState } from "react";
import { createThumbnailCache } from "./thumbnail-cache";
import type { ThumbnailCache } from "./thumbnail-cache";

const CacheContext = createContext<ThumbnailCache | null>(null);

const ScopedThumbnailCache = ({ children }: { children: ReactNode }) => {
  const [cache] = useState(createThumbnailCache);
  useEffect(
    () => () => {
      cache.clear();
    },
    [cache],
  );
  return (
    <CacheContext.Provider value={cache}>{children}</CacheContext.Provider>
  );
};

/** Scope must include user and workspace identity. Switching either destroys
 * all cached private pixels; no thumbnails are persisted to disk. */
export const AttachmentThumbnailCacheProvider = ({
  scope,
  children,
}: {
  scope: string;
  children: ReactNode;
}) => <ScopedThumbnailCache key={scope}>{children}</ScopedThumbnailCache>;

export const useAttachmentThumbnailCache = () => useContext(CacheContext);
