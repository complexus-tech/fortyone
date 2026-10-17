"use client";

import type { MouseEvent } from "react";
import {
  getPreviewHref,
  linkLabel,
  previewTitle,
  PreviewImage,
  useVisibleMetadata,
} from "@/components/ui/link-preview/external-preview";
import {
  LinkPreview as ChatLinkPreview,
  useNativeLinkTarget,
} from "@/components/ui/link-preview/link-preview";
import type { ChatLink, ChatResource } from "./types";
import { ResourceSharedBy } from "./resource-presentation";

export { ChatLinkPreview, getPreviewHref };
const canonicalPreviewUrl = (href: string) => {
  const url = new URL(href);
  url.hash = "";
  return url.href;
};

export const MessageLinkPreviews = ({
  links,
  hasAttachments = false,
}: {
  links: ChatLink[];
  hasAttachments?: boolean;
}) => {
  const unique = new Map<string, ChatLink>();
  for (const link of links) {
    const href = getPreviewHref(link.url);
    if (href && !unique.has(canonicalPreviewUrl(href)))
      unique.set(canonicalPreviewUrl(href), link);
    if (unique.size === 4) break;
  }
  if (!unique.size) return null;
  return (
    <div
      aria-label="Link previews"
      className="mt-2 grid w-full grid-cols-[repeat(auto-fit,minmax(min(100%,13rem),1fr))] items-start gap-2"
      style={{
        maxWidth:
          unique.size === 1 && !hasAttachments
            ? 480
            : unique.size * 300 + (unique.size - 1) * 8,
      }}
    >
      {[...unique].map(([href, link]) => (
        <ChatLinkPreview
          key={href}
          title={link.title}
          url={link.url}
          variant={unique.size === 1 && !hasAttachments ? "full" : "compact"}
        />
      ))}
    </div>
  );
};

const ExternalLinkResourceRow = ({
  resource,
  onPreview,
}: {
  resource: ChatResource;
  onPreview: (
    event: MouseEvent<HTMLButtonElement>,
    resource: ChatResource,
  ) => void;
}) => {
  const { ref, metadata } = useVisibleMetadata(resource.url);
  const title = previewTitle(resource.url, resource.title, metadata?.title);
  return (
    <div ref={ref}>
      <button
        aria-label={`Preview ${title}`}
        className="hover:bg-state-hover/50 focus-visible:bg-state-hover focus-visible:ring-ring flex w-full items-center gap-3 px-4 py-3.5 text-left outline-none focus-visible:ring-2 focus-visible:ring-inset"
        onClick={(event) => {
          onPreview(event, resource);
        }}
        type="button"
      >
        <PreviewImage
          className="size-10 shrink-0 rounded-md"
          icon
          src={metadata?.image}
        />
        <div className="min-w-0 flex-1 space-y-1">
          <p className="truncate font-semibold">{title}</p>
          <p className="text-text-muted truncate">{linkLabel(resource.url)}</p>
          <ResourceSharedBy resource={resource} />
        </div>
        <PreviewImage
          className="hidden h-16 w-28 shrink-0 rounded-md @md:block"
          src={metadata?.previewImage}
        />
      </button>
    </div>
  );
};

export const LinkResourceRow = (props: {
  resource: ChatResource;
  onPreview: (
    event: MouseEvent<HTMLButtonElement>,
    resource: ChatResource,
  ) => void;
}) => {
  const target = useNativeLinkTarget(props.resource.url);
  if (!target) return <ExternalLinkResourceRow {...props} />;
  return (
    <div className="px-4 py-3.5">
      <ChatLinkPreview url={props.resource.url} variant="compact" />
      <div className="mt-2 flex flex-wrap items-center gap-3">
        <ResourceSharedBy resource={props.resource} />
        <button
          className="text-primary rounded-md underline-offset-4 hover:underline"
          onClick={(event) => { props.onPreview(event, props.resource); }}
          type="button"
        >
          View shared link
        </button>
      </div>
    </div>
  );
};
