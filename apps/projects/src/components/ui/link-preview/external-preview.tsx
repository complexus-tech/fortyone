"use client";

import { useEffect, useRef, useState } from "react";
import Image from "next/image";
import { cn } from "lib";
import { FileTypeIcon } from "@/components/ui/file-type-icon";
import { useLinkMetadata } from "@/lib/hooks/link-metadata";
import { PreviewActions } from "./preview-actions";

export const getPreviewHref = (value: string) => {
  try {
    const url = new URL(value);
    if (
      !["https:", "http:"].includes(url.protocol) ||
      url.username ||
      url.password
    )
      return null;
    return url.href;
  } catch {
    return null;
  }
};

const metadataHref = (href: string) => {
  const url = new URL(href);
  url.hash = "";
  return url.href;
};

export const usePreviewVisibility = () => {
  const ref = useRef<HTMLDivElement>(null);
  const [visible, setVisible] = useState(false);
  useEffect(() => {
    const element = ref.current;
    if (!element) return;
    if (typeof IntersectionObserver === "undefined") {
      setVisible(true);
      return;
    }
    const observer = new IntersectionObserver(
      (entries) => {
        if (!entries.some((entry) => entry.isIntersecting)) return;
        setVisible(true);
        observer.disconnect();
      },
      { rootMargin: "160px" },
    );
    observer.observe(element);
    return () => {
      observer.disconnect();
    };
  }, []);
  return { ref, visible };
};

export const useVisibleMetadata = (url: string) => {
  const { ref, visible } = usePreviewVisibility();
  const href = getPreviewHref(url);
  const query = useLinkMetadata(href ? metadataHref(href) : "", {
    enabled: visible && Boolean(href),
  });
  return {
    ref,
    href,
    metadata: query.data,
    refresh: query.refreshPreview,
    refreshing: query.isRefreshing,
  };
};

export const linkLabel = (url: string) => {
  try {
    return new URL(url).hostname.replace(/^www\./, "");
  } catch {
    return url;
  }
};

export const previewTitle = (
  url: string,
  title?: string,
  metadataTitle?: string,
) =>
  title && getPreviewHref(title) !== getPreviewHref(url)
    ? title
    : metadataTitle || linkLabel(url);

// Only images returned by the authenticated metadata proxy are rendered.
const proxyImage = (url?: string) =>
  url?.startsWith("/api/metadata/image?") ? url : undefined;

export const PreviewImage = ({
  src,
  className,
  icon = false,
}: {
  src?: string;
  className?: string;
  icon?: boolean;
}) => {
  const [failed, setFailed] = useState<string | null>(null);
  const safe = proxyImage(src);
  if (!safe || failed === safe)
    return icon ? (
      <FileTypeIcon className={cn("size-10", className)} kind="link" />
    ) : null;
  return (
    <span
      className={cn(
        "bg-surface-muted relative block overflow-hidden",
        className,
      )}
    >
      <Image
        alt=""
        className={icon ? "object-contain p-2" : "object-cover"}
        decoding="async"
        fill
        loading="lazy"
        onError={() => {
          setFailed(safe);
        }}
        referrerPolicy="no-referrer"
        sizes={icon ? "40px" : "(max-width: 640px) 80vw, 300px"}
        src={safe}
        unoptimized
      />
    </span>
  );
};

export const ExternalLinkPreview = ({
  url,
  title,
  variant = "full",
}: {
  url: string;
  title?: string;
  variant?: "full" | "compact";
}) => {
  const { ref, href, metadata, refresh, refreshing } = useVisibleMetadata(url);
  if (!href) return null;
  const name = previewTitle(href, title, metadata?.title);
  const domain = linkLabel(href);
  const compact =
    variant === "compact" ||
    (!metadata?.previewImage && !metadata?.description);
  return (
    <div
      aria-busy={refreshing}
      className="relative w-full max-w-[300px] min-w-0"
      ref={ref}
    >
      <a
        aria-label={`Open ${name}`}
        className="border-border bg-surface hover:border-border-strong focus-visible:ring-primary/40 block overflow-hidden rounded-lg border text-left transition-colors outline-none focus-visible:ring-2 dark:bg-[color-mix(in_oklab,var(--color-background)_65%,var(--color-surface))]"
        href={href}
        rel="noopener noreferrer"
        target="_blank"
      >
        {!compact ? (
          <PreviewImage
            className="aspect-[2/1] w-full"
            src={metadata.previewImage}
          />
        ) : null}
        {compact ? (
          <div className="flex items-center gap-2.5 p-3 pr-10">
            <PreviewImage
              className="size-10 shrink-0 rounded-md"
              icon
              src={metadata?.image}
            />
            <div className="min-w-0 flex-1">
              {name !== domain ? (
                <p className="text-text-muted mb-0.5 truncate">{domain}</p>
              ) : null}
              <p className="line-clamp-2 font-semibold [overflow-wrap:anywhere]">
                {name}
              </p>
            </div>
          </div>
        ) : (
          <div className="p-3">
            <p className="line-clamp-2 pr-7 font-semibold [overflow-wrap:anywhere]">
              {name}
            </p>
            {metadata.description ? (
              <p
                className="text-text-muted mt-1 line-clamp-1 [overflow-wrap:anywhere]"
                title={metadata.description}
              >
                {metadata.description}
              </p>
            ) : null}
            <div className="text-text-muted mt-3 flex items-center gap-2">
              <PreviewImage
                className="size-8 shrink-0 rounded-md"
                icon
                src={metadata.image}
              />
              <span className="min-w-0 truncate">{domain}</span>
            </div>
          </div>
        )}
      </a>
      <div className="absolute top-2 right-2">
        <PreviewActions
          onRefresh={refresh}
          refreshedAt={metadata?.fetchedAt}
          refreshing={refreshing}
        />
      </div>
    </div>
  );
};
