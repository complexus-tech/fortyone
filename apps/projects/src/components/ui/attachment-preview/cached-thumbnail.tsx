"use client";

import type { CSSProperties } from "react";
import { useEffect, useRef, useState, useSyncExternalStore } from "react";
import { cn } from "lib";
import {
  createThumbnailCache,
  THUMBNAIL_ACCESS_LEASE_MS,
} from "./thumbnail-cache";
import { useAttachmentThumbnailCache } from "./thumbnail-cache-context";

type CachedAttachmentThumbnailProps = {
  url: string;
  fileName: string;
  className?: string;
  imageClassName?: string;
  style?: CSSProperties;
  refreshKey?: number;
  onImageLoad?: (width: number, height: number) => void;
};

const Thumbnail = ({
  url,
  fileName,
  className,
  imageClassName,
  style,
  refreshKey = 0,
  onImageLoad,
}: CachedAttachmentThumbnailProps) => {
  const shared = useAttachmentThumbnailCache();
  const [isolated] = useState(createThumbnailCache);
  const cache = shared ?? isolated;
  const revision = useSyncExternalStore(
    cache.subscribe,
    cache.revision,
    cache.revision,
  );
  const [state, setState] = useState<{
    status: "loading" | "ready" | "error";
    source?: string;
  }>({ status: "loading" });
  const container = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const element = container.current;
    if (!element) return;
    const controller = new AbortController();
    let objectURL: string | undefined;
    let timer: ReturnType<typeof setTimeout> | undefined;
    let expiryTimer: ReturnType<typeof setTimeout> | undefined;
    let active = false;
    let visible = false;
    let refreshed = false;
    const discard = () => {
      if (objectURL) URL.revokeObjectURL(objectURL);
      objectURL = undefined;
    };
    const load = () => {
      if (active || !visible || controller.signal.aborted) return;
      active = true;
      if (!objectURL) setState({ status: "loading" });
      void cache
        .load(
          url,
          controller.signal,
          refreshKey > 0 && !refreshed,
          Boolean(objectURL),
        )
        .then(
          (result) => {
            if (controller.signal.aborted) return;
            refreshed = true;
            clearTimeout(expiryTimer);
            discard();
            objectURL = URL.createObjectURL(result.blob);
            setState({ status: "ready", source: objectURL });
            expiryTimer = setTimeout(
              () => {
                discard();
                setState({ status: "loading" });
              },
              Math.max(
                1,
                result.checkedAt + THUMBNAIL_ACCESS_LEASE_MS - Date.now(),
              ),
            );
            timer = setTimeout(
              () => {
                active = false;
                if (visible) load();
                else {
                  discard();
                  setState({ status: "loading" });
                }
              },
              Math.max(
                1,
                result.checkedAt +
                  THUMBNAIL_ACCESS_LEASE_MS -
                  5_000 -
                  Date.now(),
              ),
            );
          },
          () => {
            if (controller.signal.aborted) return;
            clearTimeout(expiryTimer);
            discard();
            setState({ status: "error" });
          },
        );
    };
    const observer =
      typeof IntersectionObserver === "undefined"
        ? null
        : new IntersectionObserver(
            (entries) => {
              visible = entries.some((entry) => entry.isIntersecting);
              if (visible) load();
            },
            { rootMargin: "240px" },
          );
    if (observer) observer.observe(element);
    else {
      visible = true;
      load();
    }
    return () => {
      observer?.disconnect();
      controller.abort();
      clearTimeout(timer);
      clearTimeout(expiryTimer);
      discard();
    };
  }, [cache, refreshKey, revision, url]);

  useEffect(
    () => () => {
      isolated.clear();
    },
    [isolated],
  );
  return (
    <div
      aria-busy={state.status === "loading"}
      aria-label={
        state.status === "error"
          ? `Preview unavailable for ${fileName}`
          : `Preview of ${fileName}`
      }
      className={cn("relative aspect-[8/5] w-full overflow-hidden", className)}
      ref={container}
      role="img"
      style={style}
    >
      {state.source ? (
        <img
          alt=""
          aria-hidden
          className={cn(
            "absolute inset-0 h-full w-full object-cover object-top",
            imageClassName,
          )}
          decoding="async"
          onError={() => {
            setState({ status: "error" });
          }}
          onLoad={(event) => {
            onImageLoad?.(
              event.currentTarget.naturalWidth,
              event.currentTarget.naturalHeight,
            );
          }}
          src={state.source}
        />
      ) : (
        <span className="bg-surface-muted text-text-muted absolute inset-0 flex items-center justify-center px-4 text-center">
          {state.status === "error"
            ? "Preview unavailable · Open to view"
            : "Loading preview…"}
        </span>
      )}
    </div>
  );
};

export const CachedAttachmentThumbnail = (
  props: CachedAttachmentThumbnailProps,
) => <Thumbnail key={props.url} {...props} />;
