"use client";

import { cn } from "lib";
import type { ReactNode } from "react";
import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import {
  ArrowUpRightIcon,
  CopyIcon,
  DocsIcon,
  ObjectiveIcon,
  LinkIcon,
} from "icons";
import { Button } from "ui";
import { toast } from "sonner";
import { Logo } from "@/components/ui/logo";
import { useCopyToClipboard } from "@/hooks/clipboard";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { getStoryReference } from "@/shared/routing/story";
import { NativeObjectiveFields, NativeStoryFields } from "./native-fields";
import { useNativeLink } from "./native-query";
import type { NativeLinkResource } from "./native-query";
import type { NativeLinkTarget } from "./native-target";
import { openNativeLink, useNativeLinkNavigation } from "./native-navigation";
import { useNativePreviewTerminology } from "./native-terminology";
import { PreviewActions } from "./preview-actions";

export { parseNativeLink } from "./native-target";
export type { NativeLinkTarget } from "./native-target";

const KIND_ICONS = {
  story: Logo,
  objective: ObjectiveIcon,
  document: DocsIcon,
  unsupported: LinkIcon,
};

const getSummary = (resource: NativeLinkResource) => {
  switch (resource.kind) {
    case "story":
      return {
        title: resource.value.title,
        reference: getStoryReference(resource.value),
        description: null,
      };
    case "objective":
      return {
        title: resource.value.name,
        reference: `OBJ-${resource.value.sequenceId}`,
        description: resource.value.shortSummary,
      };
    case "document":
      return {
        title: resource.value.title,
        reference: null,
        description: resource.value.contentText,
      };
  }
};

const resourcePath = (
  target: Exclude<NativeLinkTarget, { kind: "unsupported" }>,
) => {
  switch (target.kind) {
    case "story":
      return `/work/${encodeURIComponent(target.id)}`;
    case "objective":
      return `/teams/${target.teamId}/objectives/${target.id}`;
    case "document":
      return `/docs/${target.id}`;
  }
};

const OpenLink = ({
  target,
  className,
  label,
  children,
}: {
  target: NativeLinkTarget;
  className: string;
  label?: string;
  children: ReactNode;
}) => {
  const { workspaceSlug, withWorkspace } = useWorkspacePath();
  const navigate = useNativeLinkNavigation();
  const sameWorkspace = workspaceSlug.toLowerCase() === target.workspaceSlug;
  if (sameWorkspace && target.kind !== "unsupported") {
    const path = resourcePath(target);
    const url = new URL(target.href);
    return (
      <Link
        aria-label={label}
        className={className}
        href={`${withWorkspace(path)}${url.search}${url.hash}`}
        onClick={(event) => {
          openNativeLink(event, target, navigate);
        }}
        prefetch={false}
      >
        {children}
      </Link>
    );
  }
  return (
    <a
      aria-label={label}
      className={className}
      href={target.href}
      onClick={(event) => {
        openNativeLink(event, target, navigate);
      }}
      rel="noopener noreferrer"
      target="_blank"
    >
      {children}
    </a>
  );
};

export const NativeLinkPreview = ({
  target,
  compact = false,
  enabled = true,
  className,
}: {
  target: NativeLinkTarget;
  compact?: boolean;
  enabled?: boolean;
  className?: string;
}) => {
  const cardRef = useRef<HTMLElement>(null);
  const [active, setActive] = useState(false);
  useEffect(() => {
    const card = cardRef.current;
    if (!card) return;
    if (typeof IntersectionObserver === "undefined") {
      setActive(true);
      return;
    }
    const observer = new IntersectionObserver(([entry]) => {
      setActive(entry.isIntersecting);
    });
    observer.observe(card);
    return () => {
      observer.disconnect();
    };
  }, []);
  const { workspaceSlug } = useWorkspacePath();
  const {
    data,
    loading,
    unavailable,
    retry,
    refresh,
    refreshing,
    refreshedAt,
  } = useNativeLink(target, enabled, active);
  const { label, term } = useNativePreviewTerminology(target, data);
  const [, copyText] = useCopyToClipboard();
  const sameWorkspace = workspaceSlug.toLowerCase() === target.workspaceSlug;
  const Icon = KIND_ICONS[target.kind];
  let fallbackTitle =
    target.kind === "unsupported" ? "Open in FortyOne" : "Preview unavailable";
  if (loading) fallbackTitle = "Loading preview…";
  const { title, reference, description } = data
    ? getSummary(data)
    : { title: fallbackTitle, reference: null, description: null };
  let content = (
    <p className="line-clamp-2 font-semibold [overflow-wrap:anywhere]">
      {title}
    </p>
  );
  if (!compact && sameWorkspace && data?.kind === "story")
    content = <NativeStoryFields story={data.value} />;
  if (!compact && sameWorkspace && data?.kind === "objective")
    content = (
      <div className="pointer-events-auto">
        <NativeObjectiveFields objective={data.value} />
      </div>
    );
  return (
    <article
      aria-busy={loading || refreshing}
      className={cn(
        "border-border bg-surface relative w-full min-w-0 overflow-hidden rounded-lg border dark:bg-[color-mix(in_oklab,var(--color-background)_65%,var(--color-surface))]",
        compact ? "max-w-[300px]" : "max-w-[380px]",
        className,
      )}
      ref={cardRef}
    >
      <OpenLink
        className="focus-visible:ring-primary/40 absolute inset-0 z-0 rounded-lg outline-none focus-visible:ring-2 focus-visible:ring-inset"
        label={`Open ${data ? title : term}`}
        target={target}
      >
        <span className="sr-only">Open {term}</span>
      </OpenLink>
      <div className="pointer-events-none relative flex items-start gap-3 p-3">
        {target.kind === "story" ? (
          <span
            aria-hidden
            className="flex size-9 shrink-0 items-center justify-center rounded-md bg-black text-white"
          >
            <Logo asIcon className="h-5.5 text-white dark:text-white" />
          </span>
        ) : (
          <Icon
            aria-hidden
            className="text-text-muted mt-0.5 size-7 shrink-0"
          />
        )}
        <div className="min-w-0 flex-1">
          <p className="text-text-muted mb-1 truncate pr-7">
            {label}
            {reference ? ` · ${reference}` : ""}
          </p>
          {content}
          {!compact && description ? (
            <p className="text-text-muted mt-3 line-clamp-1 [overflow-wrap:anywhere]">
              {description.slice(0, 400)}
            </p>
          ) : null}
          {!compact && unavailable ? (
            <p className="text-text-muted mt-2">
              This item may be private, deleted, or temporarily unavailable.
            </p>
          ) : null}
          {!compact && data && !sameWorkspace ? (
            <p className="text-text-muted mt-2">
              Open this workspace to make changes.
            </p>
          ) : null}
          {unavailable ? (
            <Button
              className="pointer-events-auto mt-2"
              color="tertiary"
              onClick={retry}
              size="sm"
              variant="naked"
            >
              Retry
            </Button>
          ) : null}
        </div>
      </div>
      {data ? (
        <div className="absolute top-2 right-2">
          <PreviewActions
            onRefresh={refresh}
            refreshedAt={refreshedAt}
            refreshing={refreshing}
          />
        </div>
      ) : null}
      {!compact ? (
        <div className="border-border relative flex items-center justify-between gap-3 border-t px-3 py-[calc(0.5rem+1.5px)]">
          <OpenLink
            className="text-primary focus-visible:ring-primary/40 inline-flex min-h-8 items-center gap-1.5 rounded-md font-medium underline-offset-4 outline-none hover:underline focus-visible:ring-2"
            target={target}
          >
            View {term}
            <ArrowUpRightIcon aria-hidden className="size-4" />
          </OpenLink>
          <button
            aria-label={`Copy ${term} link`}
            className="text-text-muted hover:text-foreground focus-visible:ring-primary/40 inline-flex min-h-8 items-center gap-1.5 rounded-md transition-colors outline-none focus-visible:ring-2"
            onClick={async () => {
              const copied = await copyText(target.href);
              if (copied) toast.success(`${label} link copied`);
              else toast.error("Could not copy link. Please try again.");
            }}
            type="button"
          >
            <CopyIcon aria-hidden className="size-4" />
            Copy link
          </button>
        </div>
      ) : null}
    </article>
  );
};
