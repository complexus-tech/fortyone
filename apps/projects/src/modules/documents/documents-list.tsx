"use client";

import type { ComponentType } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { Box, Button, Flex, Skeleton, Text } from "ui";
import {
  ArchiveIcon,
  BookIcon,
  ClockIcon,
  DocsIcon,
  LockKeyholeIcon,
  PlusIcon,
  ShareIcon,
  StarIcon,
  UserIcon,
} from "icons";
import { cn } from "lib";
import { ExpandableSearchHeader } from "@/components/shared";
import { useUserRole, useWorkspacePath } from "@/hooks";
import { useFavorites } from "@/shared/favorites";
import { walkthroughTargets } from "@/shared/walkthrough/targets";
import { DOCUMENTS_SIDEBAR_RECENT_LIMIT } from "./constants";
import { useCreateDocument, useDocuments } from "./hooks";
import { isDocumentScope } from "./document-list-state";
import type { DocumentScope } from "./types";

const scopeOptions: {
  icon: ComponentType<{ className?: string; strokeWidth?: number }>;
  label: string;
  value: DocumentScope;
}[] = [
  { icon: ClockIcon, label: "Recent documents", value: "all" },
  { icon: UserIcon, label: "My documents", value: "mine" },
  { icon: ShareIcon, label: "Shared with me", value: "shared" },
  { icon: StarIcon, label: "Favorites", value: "favorites" },
  { icon: BookIcon, label: "Templates", value: "templates" },
  { icon: ArchiveIcon, label: "Archived", value: "archived" },
];

export const DocumentsList = () => {
  const pathname = usePathname();
  const router = useRouter();
  const searchParams = useSearchParams();
  const { userRole } = useUserRole();
  const { withWorkspace } = useWorkspacePath();
  const createDocument = useCreateDocument();
  const {
    data: documents = [],
    isPending: isRecentPending,
    isError,
  } = useDocuments("", "all");
  const { favorites } = useFavorites();
  const favoriteDocumentIds = new Set(
    favorites
      .filter((favorite) => favorite.kind === "document")
      .map((favorite) => favorite.id),
  );
  const favoriteDocuments = documents.filter((document) =>
    favoriteDocumentIds.has(document.id),
  );
  const recentDocuments = documents.slice(0, DOCUMENTS_SIDEBAR_RECENT_LIMIT);
  const requestedScope = searchParams.get("scope");
  const scope = isDocumentScope(requestedScope) ? requestedScope : "all";
  const search = searchParams.get("search") ?? "";
  const isDocumentsHome = pathname.endsWith("/docs");
  const canCreateDocuments = userRole === "admin" || userRole === "member";
  const rawSearchParams = searchParams.toString();

  const getDocumentsHref = (nextScope: DocumentScope, nextSearch = search) => {
    const params = new URLSearchParams(rawSearchParams);
    params.delete("page");
    if (nextScope === "all") {
      params.delete("scope");
    } else {
      params.set("scope", nextScope);
    }
    if (nextScope === "templates") {
      ["access", "direction", "owner", "search", "sort", "updated"].forEach(
        (key) => {
          params.delete(key);
        },
      );
    } else if (nextSearch.trim()) {
      params.set("search", nextSearch.trim());
    } else {
      params.delete("search");
    }
    if (nextScope === "mine") params.delete("owner");
    if (nextScope === "archived" && params.get("access") === "public") {
      params.delete("access");
    }
    const query = params.toString();
    return withWorkspace(`/docs${query ? `?${query}` : ""}`);
  };

  const handleCreate = () => {
    if (!canCreateDocuments) return;
    createDocument.mutate(undefined, {
      onSuccess: (response) => {
        if (response.data) {
          router.push(withWorkspace(`/docs/${response.data.id}`));
        }
      },
    });
  };

  return (
    <Flex
      className="border-border/70 bg-sidebar/35 h-full min-h-0 min-w-0 overflow-hidden border-r"
      direction="column"
    >
      <Box data-walkthrough-target={walkthroughTargets.documentsSearch}>
        <ExpandableSearchHeader
          actions={
            <Button
              aria-label="Create document"
              asIcon
              className="aspect-square"
              color="tertiary"
              disabled={!canCreateDocuments || createDocument.isPending}
              onClick={handleCreate}
              size="sm"
            >
              <PlusIcon strokeWidth={2} />
            </Button>
          }
          initialValue={search}
          key={search}
          label="Search documents"
          leading={
            <Flex align="center" className="min-w-0" gap={2}>
              <DocsIcon className="size-5 shrink-0" strokeWidth={2} />
              <Text className="truncate" fontSize="lg" fontWeight="semibold">
                Documents
              </Text>
            </Flex>
          }
          onSubmit={(nextSearch) => {
            router.push(
              getDocumentsHref(
                scope === "templates" ? "all" : scope,
                nextSearch,
              ),
            );
          }}
          placeholder="Search documents..."
        />
      </Box>

      <Box
        className="px-3 py-3"
        data-walkthrough-target={walkthroughTargets.documentsNavigation}
      >
        {scopeOptions.map(({ icon: Icon, label, value }) => {
          const isActive = isDocumentsHome && scope === value;
          return (
            <Link
              aria-current={isActive ? "page" : undefined}
              className={cn(
                "text-text-muted hover:bg-state-hover hover:text-foreground flex w-full items-center gap-2.5 rounded-lg px-3 py-2.5 text-left transition-colors",
                { "bg-state-active text-foreground": isActive },
              )}
              href={getDocumentsHref(
                value,
                value === "templates" ? "" : search,
              )}
              key={value}
            >
              <Icon className="size-[1.1rem] shrink-0" strokeWidth={2} />
              <span className="truncate">{label}</span>
            </Link>
          );
        })}
      </Box>

      <Box
        className="border-border/70 min-h-0 flex-1 overflow-y-auto border-t px-3 py-4"
        data-walkthrough-target={walkthroughTargets.documentsRecent}
      >
        {favoriteDocuments.length > 0 ? (
          <Box className="mb-5">
            <Text className="mb-2 px-3" color="muted" fontWeight="semibold">
              Favorites
            </Text>
            {favoriteDocuments.map((document) => {
              const href = withWorkspace(`/docs/${document.id}`);
              const isActive = pathname === href;
              return (
                <Link
                  aria-current={isActive ? "page" : undefined}
                  className={cn(
                    "text-text-muted hover:bg-state-hover hover:text-foreground mb-1 flex min-w-0 items-center gap-2 rounded-lg px-3 py-2.5 transition-colors",
                    { "bg-state-active text-foreground": isActive },
                  )}
                  href={href}
                  key={document.id}
                >
                  <StarIcon className="text-primary size-5 shrink-0" />
                  <span className="min-w-0 truncate">{document.title}</span>
                </Link>
              );
            })}
          </Box>
        ) : null}
        <Text className="mb-2 px-3" color="muted" fontWeight="semibold">
          Recent
        </Text>
        {isRecentPending ? (
          <Box className="space-y-2 px-3 py-2">
            {Array.from({ length: 4 }).map((_, index) => (
              <Skeleton className="h-9 w-full" key={index} />
            ))}
          </Box>
        ) : null}
        {!isRecentPending && !isError && recentDocuments.length === 0 ? (
          <Text className="px-3 py-3" color="muted">
            No documents yet
          </Text>
        ) : null}
        {isError ? (
          <Text className="px-3 py-3" color="muted">
            Recent documents could not be loaded.
          </Text>
        ) : null}
        {isRecentPending
          ? null
          : recentDocuments.map((document) => {
              const href = withWorkspace(`/docs/${document.id}`);
              const isActive = pathname === href;
              return (
                <Link
                  aria-current={isActive ? "page" : undefined}
                  className={cn(
                    "text-text-muted hover:bg-state-hover hover:text-foreground mb-1 flex min-w-0 items-center gap-2 rounded-lg px-3 py-2.5 transition-colors",
                    { "bg-state-active text-foreground": isActive },
                  )}
                  href={href}
                  key={document.id}
                >
                  {document.visibility === "private" ? (
                    <LockKeyholeIcon
                      className="size-[1.1rem] shrink-0"
                      strokeWidth={2}
                    />
                  ) : (
                    <DocsIcon
                      className="size-[1.1rem] shrink-0"
                      strokeWidth={2}
                    />
                  )}
                  <span className="min-w-0 flex-1 truncate">
                    {document.title}
                  </span>
                </Link>
              );
            })}
      </Box>
    </Flex>
  );
};
