"use client";

import type { ComponentType } from "react";
import { useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Box, Button, Flex, Skeleton, Text } from "ui";
import {
  CalendarIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  DocsIcon,
  ObjectiveIcon,
  PlusIcon,
  TeamIcon,
  UserIcon,
} from "icons";
import { cn } from "lib";
import { ExpandableSearchHeader } from "@/components/shared/expandable-search-header";
import { MobileMenuButton } from "@/components/shared/mobile-menu";
import { DocumentsEmptyIllustration } from "@/components/ui/illustrations/empty-state-illustrations";
import { useUserRole } from "@/hooks/role";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSession } from "@/lib/auth/client";
import { DocumentRow } from "./document-row";
import { DocumentFilters, DocumentSortMenu } from "./documents-home-controls";
import {
  documentTemplates,
  type DocumentTemplate,
  type DocumentTemplateIcon,
} from "./document-templates";
import {
  filterAndSortDocumentSummaries,
  getDocumentPageAfterArchive,
  getDocumentListState,
  paginateDocumentSummaries,
} from "./document-list-state";
import { useArchiveDocument, useCreateDocument, useDocuments } from "./hooks";
import type { DocumentScope } from "./types";

const scopeCopy: Record<
  DocumentScope,
  { description: string; heading: string }
> = {
  all: {
    heading: "Recent documents",
    description: "Pick up where your workspace most recently left off.",
  },
  mine: {
    heading: "My documents",
    description: "Documents you created and can continue shaping.",
  },
  shared: {
    heading: "Shared with me",
    description: "Documents other workspace members shared directly with you.",
  },
  templates: {
    heading: "Document templates",
    description: "Start with a useful structure and adapt it to your work.",
  },
};

const templateIcons: Record<
  DocumentTemplateIcon,
  ComponentType<{ className?: string }>
> = {
  blank: DocsIcon,
  meeting: TeamIcon,
  project: ObjectiveIcon,
  "one-to-one": UserIcon,
  update: CalendarIcon,
};

const templateIconStyles: Record<DocumentTemplateIcon, string> = {
  blank:
    "border-border/70 bg-surface-elevated text-foreground dark:border-border-strong/80",
  meeting: "border-info/25 bg-info/10 text-info",
  project: "border-secondary/25 bg-secondary/10 text-secondary",
  "one-to-one": "border-success/25 bg-success/10 text-success",
  update: "border-primary/25 bg-primary/10 text-primary",
};

const isDocumentScope = (value: string | null): value is DocumentScope =>
  value === "all" ||
  value === "mine" ||
  value === "shared" ||
  value === "templates";

const mobileScopes: { label: string; value: DocumentScope }[] = [
  { label: "Recent", value: "all" },
  { label: "My documents", value: "mine" },
  { label: "Shared", value: "shared" },
  { label: "Templates", value: "templates" },
];

const TemplateCard = ({
  disabled,
  isCreating,
  onCreate,
  template,
}: {
  disabled: boolean;
  isCreating: boolean;
  onCreate: (template: DocumentTemplate) => void;
  template: DocumentTemplate;
}) => {
  const Icon = templateIcons[template.icon];

  return (
    <button
      className={cn(
        "border-border/80 bg-surface/60 hover:border-border-strong focus-visible:border-border-strong focus-visible:ring-border-strong/50 dark:border-border/80 dark:bg-surface/70 dark:hover:border-border-strong dark:focus-visible:border-border-strong flex min-h-20 min-w-[13.5rem] items-center gap-2.5 rounded-xl border-[0.5px] px-4 py-3 text-left transition-[border-color] outline-none focus-visible:ring-1 disabled:opacity-60",
        isCreating ? "disabled:cursor-progress" : "disabled:cursor-not-allowed",
      )}
      disabled={disabled}
      onClick={() => {
        onCreate(template);
      }}
      type="button"
    >
      <Flex
        align="center"
        className={cn(
          "size-12 shrink-0 rounded-xl border-[0.5px] [&_svg]:!text-current",
          templateIconStyles[template.icon],
        )}
        justify="center"
      >
        <Icon className="size-8" />
      </Flex>
      <Text as="span" fontWeight="medium">
        {isCreating ? "Creating…" : template.label}
      </Text>
    </button>
  );
};

const DocumentsTableSkeleton = () => (
  <Box>
    {Array.from({ length: 5 }).map((_, index) => (
      <Flex
        align="center"
        className="border-border/80 border-b-[0.5px] py-[0.655rem]"
        justify="between"
        key={index}
      >
        <Skeleton className="h-5 w-52" />
        <Skeleton className="h-5 w-28" />
      </Flex>
    ))}
  </Box>
);

const getEmptyDescription = (
  isFiltered: boolean,
  scope: DocumentScope,
  canCreateDocuments: boolean,
) => {
  if (isFiltered) return "Try adjusting your search or filters.";
  if (scope === "shared") {
    return "Documents shared directly with you will appear here.";
  }
  if (!canCreateDocuments) {
    return "Workspace members can create documents here.";
  }
  return "Create a blank document or start from a template.";
};

const EmptyDocuments = ({
  canCreateDocuments,
  isFiltered,
  scope,
}: {
  canCreateDocuments: boolean;
  isFiltered: boolean;
  scope: DocumentScope;
}) => (
  <Flex
    align="center"
    className="px-8 pt-20 pb-12 text-center"
    direction="column"
    justify="center"
  >
    <DocumentsEmptyIllustration className="w-52" />
    <Text className="mt-4" fontWeight="semibold">
      {isFiltered ? "No matching documents" : "No documents here yet"}
    </Text>
    <Text className="mt-1 max-w-md" color="muted">
      {getEmptyDescription(isFiltered, scope, canCreateDocuments)}
    </Text>
  </Flex>
);

export const DocumentsHome = () => {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { data: session } = useSession();
  const { userRole } = useUserRole();
  const { withWorkspace } = useWorkspacePath();
  const requestedScope = searchParams.get("scope");
  const scope = isDocumentScope(requestedScope) ? requestedScope : "all";
  const search = searchParams.get("search")?.trim() ?? "";
  const rawSearchParams = searchParams.toString();
  const listState = getDocumentListState(new URLSearchParams(rawSearchParams));
  const apiScope = scope === "templates" ? "all" : scope;
  const { data: documents = [], isPending } = useDocuments(search, apiScope);
  const createDocument = useCreateDocument();
  const archiveDocument = useArchiveDocument();
  const [creatingTemplateId, setCreatingTemplateId] = useState<string | null>(
    null,
  );
  const canCreateDocuments = userRole === "admin" || userRole === "member";
  const showTemplateGallery = scope === "templates";
  const listScope = scope === "templates" ? "all" : scope;
  const filteredDocuments = filterAndSortDocumentSummaries({
    currentUserId: session?.user.id,
    documents,
    scope: listScope,
    state: listState,
  });
  const pagination = paginateDocumentSummaries(
    filteredDocuments,
    listState.page,
  );

  const getDocumentsPath = (params: URLSearchParams) => {
    const query = params.toString();
    return withWorkspace(`/docs${query ? `?${query}` : ""}`);
  };

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
    return getDocumentsPath(params);
  };

  const updateListControls = (updates: Record<string, string | null>) => {
    const params = new URLSearchParams(rawSearchParams);
    Object.entries(updates).forEach(([key, value]) => {
      if (value === null) {
        params.delete(key);
      } else {
        params.set(key, value);
      }
    });
    params.delete("page");
    router.push(getDocumentsPath(params), { scroll: false });
  };

  const goToPage = (page: number) => {
    const params = new URLSearchParams(rawSearchParams);
    if (page <= 1) {
      params.delete("page");
    } else {
      params.set("page", String(page));
    }
    router.push(getDocumentsPath(params), { scroll: false });
  };

  const archiveDocumentAndMaintainPage = (documentId: string) => {
    const nextPage = getDocumentPageAfterArchive(
      pagination.page,
      pagination.items.length,
    );
    archiveDocument.mutate(documentId, {
      onSuccess: () => {
        if (nextPage !== pagination.page) {
          goToPage(nextPage);
        }
      },
    });
  };

  const handleCreate = (template: DocumentTemplate) => {
    if (!canCreateDocuments) return;
    setCreatingTemplateId(template.id);
    createDocument.mutate(
      {
        title: template.title,
        contentHtml: template.contentHtml,
        contentText: template.contentText,
        visibility: "workspace",
      },
      {
        onSuccess: (response) => {
          if (response.data) {
            router.push(withWorkspace(`/docs/${response.data.id}`));
          }
        },
        onSettled: () => {
          setCreatingTemplateId(null);
        },
      },
    );
  };

  const visibleTemplates = showTemplateGallery
    ? documentTemplates
    : documentTemplates.slice(0, 5);
  const copy = scopeCopy[scope];
  const activeFilterCount = [
    listState.access !== "all",
    scope !== "mine" && listState.owner !== "all",
    listState.updated !== "all",
  ].filter(Boolean).length;

  return (
    <Box className="h-full min-h-0 min-w-0 overflow-y-auto">
      <Box className="md:hidden">
        <ExpandableSearchHeader
          actions={
            <Button
              aria-label="Create document"
              asIcon
              className="aspect-square"
              color="tertiary"
              disabled={!canCreateDocuments || createDocument.isPending}
              onClick={() => {
                handleCreate(documentTemplates[0]);
              }}
              size="sm"
            >
              <PlusIcon />
            </Button>
          }
          initialValue={search}
          key={search}
          label="Search documents"
          leading={
            <Flex align="center" className="min-w-0" gap={2}>
              <MobileMenuButton />
              <DocsIcon className="text-text-muted size-5 shrink-0" />
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
        <Flex className="border-border/60 overflow-x-auto border-b px-3 py-2">
          {mobileScopes.map((option) => (
            <Link
              aria-current={scope === option.value ? "page" : undefined}
              className={cn(
                "text-foreground hover:bg-state-hover flex h-[2.1rem] w-max shrink-0 items-center rounded-xl px-2 transition-colors",
                { "bg-state-selected": scope === option.value },
              )}
              href={getDocumentsHref(
                option.value,
                option.value === "templates" ? "" : search,
              )}
              key={option.value}
            >
              {option.label}
            </Link>
          ))}
        </Flex>
      </Box>
      <Box className="mx-auto w-full max-w-[90rem] px-6 pt-7 pb-32 sm:px-8 lg:px-10 lg:pt-8">
        <Box className="mb-6">
          <Text as="h1" fontSize="2xl" fontWeight="semibold">
            {search ? `Results for “${search}”` : copy.heading}
          </Text>
          <Text className="mt-0.5" color="muted">
            {search
              ? "Search results across the selected document view."
              : copy.description}
          </Text>
        </Box>

        <Box className="mb-8">
          <Flex align="center" className="mb-3" gap={3}>
            <Text fontWeight="semibold">
              {showTemplateGallery
                ? "Choose a template"
                : "Start a new document"}
            </Text>
          </Flex>
          <Box
            className={
              showTemplateGallery
                ? "grid gap-3 sm:grid-cols-2 xl:grid-cols-3"
                : "grid auto-cols-[minmax(13.5rem,1fr)] grid-flow-col gap-3 overflow-x-auto pb-2"
            }
          >
            {visibleTemplates.map((template) => (
              <TemplateCard
                disabled={!canCreateDocuments || createDocument.isPending}
                isCreating={creatingTemplateId === template.id}
                key={template.id}
                onCreate={handleCreate}
                template={template}
              />
            ))}
          </Box>
        </Box>

        {!showTemplateGallery ? (
          <Box>
            <Flex
              align="center"
              className="border-border/70 border-b-[0.5px] py-3"
              gap={3}
              justify="between"
              wrap
            >
              <DocumentFilters
                access={listState.access}
                activeCount={activeFilterCount}
                onAccessChange={(access) => {
                  updateListControls({
                    access: access === "all" ? null : access,
                  });
                }}
                onClear={() => {
                  updateListControls({
                    access: null,
                    owner: null,
                    updated: null,
                  });
                }}
                onOwnerChange={(owner) => {
                  updateListControls({
                    owner: owner === "all" ? null : owner,
                  });
                }}
                onUpdatedChange={(updated) => {
                  updateListControls({
                    updated: updated === "all" ? null : updated,
                  });
                }}
                owner={listState.owner}
                showOwner={scope !== "mine"}
                updated={listState.updated}
              />
              <Flex align="center" gap={3}>
                <Flex align="center" gap={1}>
                  <Button
                    aria-label="Previous page"
                    asIcon
                    className="h-9"
                    color="tertiary"
                    disabled={pagination.page <= 1}
                    onClick={() => {
                      goToPage(pagination.page - 1);
                    }}
                    rounded="lg"
                    size="sm"
                    variant="naked"
                  >
                    <ChevronLeftIcon className="h-5 w-auto" strokeWidth={2.5} />
                  </Button>
                  <Text
                    className="min-w-24 text-center tabular-nums"
                    color="muted"
                  >
                    Page {pagination.page} of {pagination.pageCount}
                  </Text>
                  <Button
                    aria-label="Next page"
                    asIcon
                    className="h-9"
                    color="tertiary"
                    disabled={pagination.page >= pagination.pageCount}
                    onClick={() => {
                      goToPage(pagination.page + 1);
                    }}
                    rounded="lg"
                    size="sm"
                    variant="naked"
                  >
                    <ChevronRightIcon
                      className="h-5 w-auto"
                      strokeWidth={2.5}
                    />
                  </Button>
                </Flex>
                <DocumentSortMenu
                  direction={listState.direction}
                  field={listState.sort}
                  onChange={(option) => {
                    const isDefault =
                      option.field === "updated" && option.direction === "desc";
                    updateListControls({
                      direction: isDefault ? null : option.direction,
                      sort: isDefault ? null : option.field,
                    });
                  }}
                />
              </Flex>
            </Flex>
            {isPending ? <DocumentsTableSkeleton /> : null}
            {!isPending && filteredDocuments.length === 0 ? (
              <EmptyDocuments
                canCreateDocuments={canCreateDocuments}
                isFiltered={Boolean(search) || activeFilterCount > 0}
                scope={scope}
              />
            ) : null}
            {!isPending && filteredDocuments.length > 0 ? (
              <Box>
                {pagination.items.map((document) => (
                  <DocumentRow
                    document={document}
                    isArchiving={
                      archiveDocument.isPending
                        ? archiveDocument.variables === document.id
                        : false
                    }
                    isOwner={
                      session?.user.id === document.createdBy &&
                      document.canEdit
                    }
                    key={document.id}
                    onArchive={() => {
                      archiveDocumentAndMaintainPage(document.id);
                    }}
                    withWorkspace={withWorkspace}
                  />
                ))}
              </Box>
            ) : null}
          </Box>
        ) : null}
      </Box>
    </Box>
  );
};
