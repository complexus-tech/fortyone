"use client";

import type { MouseEvent } from "react";
import { useRef, useState } from "react";
import { ArrowDown2Icon, DocsIcon, SearchIcon } from "icons";
import { cn } from "lib";
import { Button, Input, Menu } from "ui";
import { useDebounce } from "@/hooks/debounce";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import type {
  ChatConversation,
  ChatResource,
  ChatResourceFilter,
  ChatResourceSort,
} from "./types";
import { useChatResources } from "./queries";
import { getChatAttachmentUrl } from "./attachment-url";
import { getChatLinkHref } from "./chat-rich-text";
import {
  ResourceIcon,
  resourceMetadata,
  ResourceSharedBy,
  ResourceThumbnail,
  resourceTitle,
} from "./resource-presentation";
import { ResourcePreview } from "./resource-preview";
import { ChatAudioAttachment } from "./chat-audio-attachment";
import { LinkResourceRow } from "./link-preview";
import { useChatAttachmentPreview } from "./attachment-preview-context";

const FILTERS: { value: ChatResourceFilter; label: string }[] = [
  { value: "all", label: "All" },
  { value: "files", label: "Files" },
  { value: "media", label: "Media" },
  { value: "links", label: "Links" },
];
const SORT_OPTIONS: { value: ChatResourceSort; label: string }[] = [
  { value: "newest", label: "Newest" },
  { value: "oldest", label: "Oldest" },
];

export const FilesLinks = ({
  conversation,
  onViewMessage,
}: {
  conversation: ChatConversation;
  onViewMessage: (messageId: string) => void;
}) => {
  const { workspaceSlug } = useWorkspacePath();
  const attachmentPreview = useChatAttachmentPreview();
  const [search, setSearch] = useState("");
  const [query, setQuery] = useState("");
  const updateQuery = useDebounce(setQuery, 250);
  const [filter, setFilter] = useState<ChatResourceFilter>("all");
  const [sort, setSort] = useState<ChatResourceSort>("newest");
  const [selected, setSelected] = useState<ChatResource | null>(null);
  const triggerRef = useRef<HTMLButtonElement | null>(null);
  const returningToMessage = useRef(false);
  const results = useChatResources(conversation.id, filter, query, sort);
  // Find media even when it is older than the overview's first page.
  const mediaResults = useChatResources(
    conversation.id,
    "media",
    query,
    sort,
    filter === "all",
  );
  const resources = results.data?.pages.flatMap((page) => page.items) ?? [];
  const media =
    filter === "media"
      ? resources
      : mediaResults.data?.pages[0]?.items.slice(0, 4) ?? [];
  const rows = filter === "media" ? [] : resources;
  const resourceUrl = (resource: ChatResource) =>
    resource.kind === "file"
      ? getChatAttachmentUrl(
          workspaceSlug,
          conversation.id,
          resource.attachmentId,
        )
      : getChatLinkHref(resource.url);
  const openPreview = (
    event: MouseEvent<HTMLButtonElement>,
    resource: ChatResource,
  ) => {
    if (resource.kind === "file") {
      attachmentPreview.openResource(resource, event.currentTarget);
      return;
    }
    triggerRef.current = event.currentTarget;
    returningToMessage.current = false;
    setSelected(resource);
  };

  return (
    <div className="@container min-h-0 flex-1 overflow-y-auto px-5 py-5 md:px-6">
      <Input
        aria-label="Search files and links"
        className="h-10 bg-transparent dark:bg-transparent"
        leftIcon={<SearchIcon className="size-5" />}
        onChange={(event) => {
          setSearch(event.target.value);
          updateQuery(event.target.value.trim());
        }}
        placeholder="Search files and links…"
        type="search"
        value={search}
      />
      <div className="mt-4 mb-6 flex flex-wrap items-center justify-between gap-3">
        <div
          aria-label="Filter shared items"
          className="flex gap-2"
          role="group"
        >
          {FILTERS.map((item) => (
            <Button
              aria-pressed={filter === item.value}
              className={cn(
                "px-3.5 font-medium",
                filter === item.value &&
                  "border-primary/20 bg-primary/10 text-primary hover:bg-primary/15 dark:bg-primary/10",
              )}
              color="tertiary"
              key={item.value}
              onClick={() => {
                setFilter(item.value);
              }}
              size="sm"
              variant="outline"
            >
              {item.label}
            </Button>
          ))}
        </div>
        <Menu>
          <Menu.Button>
            <Button
              aria-label={`Sort shared items: ${sort === "newest" ? "Newest" : "Oldest"}`}
              className="gap-3 px-3"
              color="tertiary"
              rightIcon={<ArrowDown2Icon className="size-3.5" />}
              size="sm"
              variant="outline"
            >
              {sort === "newest" ? "Newest" : "Oldest"}
            </Button>
          </Menu.Button>
          <Menu.Items align="end" className="min-w-36">
            <Menu.Group>
              {SORT_OPTIONS.map((option) => (
                <Menu.Item
                  active={sort === option.value}
                  key={option.value}
                  onSelect={() => {
                    setSort(option.value);
                  }}
                >
                  {option.label}
                </Menu.Item>
              ))}
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </div>
      {results.error ? (
        <div className="mb-5 flex items-center gap-3" role="alert">
          <p className="text-text-muted">Could not load shared items.</p>
          <Button
            color="tertiary"
            onClick={() => void results.refetch()}
            size="sm"
          >
            Try again
          </Button>
        </div>
      ) : null}
      {results.isPending ? (
        <p className="text-text-muted py-6" role="status">
          Loading shared items…
        </p>
      ) : null}
      {(filter === "all" || filter === "media") && media.length > 0 ? (
        <section aria-label="Photos and videos" className="mb-6">
          <div className="mb-3 flex items-center justify-between">
            <h3 className="font-semibold">Photos and videos</h3>
            {filter === "all" ? (
              <Button
                className="px-1"
                color="primary"
                onClick={() => {
                  setFilter("media");
                }}
                size="sm"
                variant="naked"
              >
                See all
              </Button>
            ) : null}
          </div>
          <div className="grid grid-cols-2 gap-3 @lg:grid-cols-4">
            {media.map((resource) => (
              <button
                aria-label={`Preview ${resourceTitle(resource)}`}
                className="group focus-visible:ring-ring min-w-0 rounded-lg text-left outline-none focus-visible:ring-2"
                key={resource.id}
                onClick={(event) => {
                  openPreview(event, resource);
                }}
                type="button"
              >
                <ResourceThumbnail
                  resource={resource}
                  url={resourceUrl(resource)}
                />
                <p className="group-hover:text-primary mt-2 truncate font-medium">
                  {resourceTitle(resource)}
                </p>
                <p className="text-text-muted mt-0.5 truncate">
                  {resource.author.fullName}
                </p>
              </button>
            ))}
          </div>
        </section>
      ) : null}
      {filter === "all" && mediaResults.error ? (
        <div
          className="text-text-muted mb-5 flex items-center gap-3"
          role="alert"
        >
          <p>Could not load media previews.</p>
          <Button
            color="tertiary"
            onClick={() => void mediaResults.refetch()}
            size="sm"
          >
            Try again
          </Button>
        </div>
      ) : null}
      {rows.length > 0 ? (
        <ul
          aria-label="Shared files and links"
          className="border-border/60 divide-border divide-y overflow-hidden rounded-lg border"
        >
          {rows.map((resource) => (
            <li
              key={resource.id}
              style={{
                contentVisibility: "auto",
                containIntrinsicSize: "auto 72px",
              }}
            >
              {resource.kind === "link" ? (
                <LinkResourceRow onPreview={openPreview} resource={resource} />
              ) : null}
              {resource.kind !== "link" &&
              resource.mediaType.startsWith("audio/") ? (
                <div className="p-3">
                  <ChatAudioAttachment
                    attachment={{
                      id: resource.attachmentId,
                      fileName: resource.fileName,
                      mediaType: resource.mediaType,
                      sizeBytes: resource.sizeBytes,
                      url: resource.url,
                      kind: resource.attachmentKind,
                      durationMs: resource.durationMs,
                    }}
                    conversationId={conversation.id}
                    messageId={resource.messageId}
                  />
                  <div className="mt-2">
                    <ResourceSharedBy resource={resource} />
                  </div>
                </div>
              ) : null}
              {resource.kind !== "link" &&
              !resource.mediaType.startsWith("audio/") ? (
                <button
                  aria-label={`Preview ${resourceTitle(resource)}`}
                  className="hover:bg-state-hover/50 focus-visible:bg-state-hover focus-visible:ring-ring flex w-full items-center gap-3 px-4 py-3.5 text-left outline-none focus-visible:ring-2 focus-visible:ring-inset"
                  onClick={(event) => {
                    openPreview(event, resource);
                  }}
                  type="button"
                >
                  <ResourceIcon resource={resource} />
                  <div className="min-w-0 flex-1 space-y-1">
                    <p className="truncate font-semibold">
                      {resourceTitle(resource)}
                    </p>
                    <ResourceSharedBy resource={resource} />
                  </div>
                  <span className="text-text-muted hidden max-w-40 shrink-0 truncate @xl:block">
                    {resourceMetadata(resource)}
                  </span>
                </button>
              ) : null}
            </li>
          ))}
        </ul>
      ) : null}
      {!results.error && !results.isPending && resources.length === 0 ? (
        <div className="flex flex-col items-center px-6 py-16 text-center">
          <DocsIcon className="text-text-muted mb-3 size-8" />
          <h3 className="font-semibold">
            {query ? "No matching items" : "Nothing shared here yet"}
          </h3>
          <p className="text-text-muted mt-2 max-w-80">
            {query
              ? "Try a different search or filter."
              : "Files, photos, videos, and links shared in this conversation will appear here."}
          </p>
        </div>
      ) : null}
      {results.hasNextPage ? (
        <Button
          className="mx-auto mt-5"
          color="tertiary"
          disabled={results.isFetchingNextPage}
          onClick={() => {
            void results.fetchNextPage({ cancelRefetch: false });
          }}
          size="sm"
        >
          {results.isFetchingNextPage ? "Loading…" : "Load more"}
        </Button>
      ) : null}
      <ResourcePreview
        conversationName={conversation.name}
        onClose={() => {
          setSelected(null);
        }}
        onViewMessage={onViewMessage}
        resource={selected}
        returningToMessage={returningToMessage}
        triggerRef={triggerRef}
        url={selected ? resourceUrl(selected) : null}
      />
    </div>
  );
};
