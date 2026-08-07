"use client";

import type { ComponentType } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { Box, Button, Flex, Skeleton, Text } from "ui";
import {
  BookIcon,
  ClockIcon,
  DocsIcon,
  LockKeyholeIcon,
  PlusIcon,
  ShareIcon,
  UserIcon,
} from "icons";
import { cn } from "lib";
import { ExpandableSearchHeader } from "@/components/shared";
import { useUserRole, useWorkspacePath } from "@/hooks";
import { DOCUMENTS_SIDEBAR_RECENT_LIMIT } from "./constants";
import { useCreateDocument, useDocuments } from "./hooks";
import { formatDocumentRelativeTime } from "./relative-time";
import type { DocumentScope } from "./types";

const scopeOptions: {
  icon: ComponentType<{ className?: string; strokeWidth?: number }>;
  label: string;
  value: DocumentScope;
}[] = [
  { icon: ClockIcon, label: "Recent documents", value: "all" },
  { icon: UserIcon, label: "My documents", value: "mine" },
  { icon: ShareIcon, label: "Shared with me", value: "shared" },
  { icon: BookIcon, label: "Templates", value: "templates" },
];

const isDocumentScope = (value: string | null): value is DocumentScope =>
  value === "all" ||
  value === "mine" ||
  value === "shared" ||
  value === "templates";

export const DocumentsList = () => {
  const pathname = usePathname();
  const router = useRouter();
  const searchParams = useSearchParams();
  const { userRole } = useUserRole();
  const { withWorkspace } = useWorkspacePath();
  const createDocument = useCreateDocument();
  const { data: recentDocuments = [], isPending: isRecentPending } =
    useDocuments("", "all", DOCUMENTS_SIDEBAR_RECENT_LIMIT);
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
      className="border-border/70 bg-sidebar/35 h-dvh min-w-0 border-r"
      direction="column"
    >
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
            <DocsIcon
              className="text-text-muted size-5 shrink-0"
              strokeWidth={2}
            />
            <Text className="truncate" fontSize="lg" fontWeight="semibold">
              Documents
            </Text>
          </Flex>
        }
        onSubmit={(nextSearch) => {
          router.push(
            getDocumentsHref(scope === "templates" ? "all" : scope, nextSearch),
          );
        }}
        placeholder="Search documents..."
      />

      <Box className="px-3 py-3">
        {scopeOptions.map(({ icon: Icon, label, value }) => {
          const isActive = isDocumentsHome && scope === value;
          return (
            <button
              aria-current={isActive ? "page" : undefined}
              className={cn(
                "text-text-muted hover:bg-state-hover hover:text-foreground flex w-full items-center gap-2.5 rounded-lg px-3 py-2.5 text-left transition-colors",
                { "bg-state-active text-foreground": isActive },
              )}
              key={value}
              onClick={() => {
                router.push(
                  getDocumentsHref(value, value === "templates" ? "" : search),
                );
              }}
              type="button"
            >
              <Icon className="size-[1.1rem] shrink-0" strokeWidth={2} />
              <span className="truncate">{label}</span>
            </button>
          );
        })}
      </Box>

      <Box className="border-border/70 min-h-0 flex-1 overflow-y-auto border-t px-3 py-4">
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
        {!isRecentPending && recentDocuments.length === 0 ? (
          <Text className="px-3 py-3" color="muted">
            No documents yet
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
                  <span className="text-text-muted shrink-0">
                    {formatDocumentRelativeTime(document.updatedAt)}
                  </span>
                </Link>
              );
            })}
      </Box>
    </Flex>
  );
};
