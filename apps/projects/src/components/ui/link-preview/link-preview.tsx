"use client";

import { lazy, Suspense } from "react";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { ExternalLinkPreview, usePreviewVisibility } from "./external-preview";
import { parseNativeLink } from "./native-target";
import type { NativeLinkTarget } from "./native-target";

const NativeLinkPreview = lazy(() =>
  import("./native-preview").then((module) => ({
    default: module.NativeLinkPreview,
  })),
);

const NativePreviewPlaceholder = () => (
  <div
    className="border-border text-text-muted rounded-lg border px-4 py-3"
    role="status"
  >
    Loading preview…
  </div>
);

export const useNativeLinkTarget = (url: string) => {
  const { workspaceSlug } = useWorkspacePath();
  return parseNativeLink(url, {
    workspaceSlug,
    origin: typeof window === "undefined" ? undefined : window.location.origin,
  });
};

const VisibleNativePreview = ({
  target,
  compact,
}: {
  target: NativeLinkTarget;
  compact: boolean;
}) => {
  const { ref, visible } = usePreviewVisibility();
  return (
    <div className="min-w-0" ref={ref}>
      {visible ? (
        <Suspense fallback={<NativePreviewPlaceholder />}>
          <NativeLinkPreview compact={compact} target={target} />
        </Suspense>
      ) : (
        <NativePreviewPlaceholder />
      )}
    </div>
  );
};

export const LinkPreview = ({
  url,
  title,
  variant = "full",
}: {
  url: string;
  title?: string;
  variant?: "full" | "compact";
}) => {
  const target = useNativeLinkTarget(url);
  return target ? (
    <VisibleNativePreview compact={variant === "compact"} target={target} />
  ) : (
    <ExternalLinkPreview title={title} url={url} variant={variant} />
  );
};
