"use client";

import type { ComponentType } from "react";
import { useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Box, Button, Flex, Input, Skeleton, Text } from "ui";
import {
  BookIcon,
  CalendarPlusIcon,
  CheckListIcon,
  SearchIcon,
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
import { useMembers } from "@/lib/hooks/members";
import { useFavorites } from "@/shared/favorites";
import { DocumentRow, DOCUMENT_LIBRARY_COLUMNS } from "./document-row";
import { DocumentFilters, DocumentSortMenu } from "./documents-home-controls";
import {
  documentTemplates,
  documentTemplateCategories,
  type DocumentTemplate,
  type DocumentTemplateIcon,
} from "./document-templates";
import {
  filterAndSortDocumentSummaries,
  getDocumentPageAfterArchive,
  getDocumentListState,
  isDocumentScope,
  paginateDocumentSummaries,
} from "./document-list-state";
import {
  useArchiveDocument,
  useCreateDocument,
  useDocuments,
  useRestoreArchivedDocument,
} from "./hooks";
import type { DocumentScope } from "./types";

const scopeCopy: Record<
  DocumentScope,
  { description: string; heading: string }
> = {
  all: {
    heading: "Documents",
    description: "Find the context behind your work.",
  },
  mine: {
    heading: "My documents",
    description: "Documents you created and can continue shaping.",
  },
  shared: {
    heading: "Shared with me",
    description: "Documents other workspace members shared directly with you.",
  },
  favorites: {
    heading: "Favorite documents",
    description: "Keep the documents you return to close at hand.",
  },
  archived: {
    heading: "Archived documents",
    description:
      "Restore documents you archived. Public links stay off until you share them again.",
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
  guide: BookIcon,
  research: SearchIcon,
  event: CalendarPlusIcon,
  checklist: CheckListIcon,
};

const NEUTRAL_TEMPLATE_ICON_STYLE =
  "border-border/70 bg-surface-elevated text-foreground dark:border-border-strong/80";

const templateIconStyles: Record<DocumentTemplateIcon, string> = {
  blank: NEUTRAL_TEMPLATE_ICON_STYLE,
  meeting: NEUTRAL_TEMPLATE_ICON_STYLE,
  project: "border-secondary/25 bg-secondary/10 text-secondary",
  "one-to-one": "border-success/25 bg-success/10 text-success",
  update: "border-primary/25 bg-primary/10 text-primary",
  guide: "border-secondary/25 bg-secondary/10 text-secondary",
  research: NEUTRAL_TEMPLATE_ICON_STYLE,
  event: "border-primary/25 bg-primary/10 text-primary",
  checklist: "border-success/25 bg-success/10 text-success",
};

const mobileScopes: { label: string; value: DocumentScope }[] = [
  { label: "Recent", value: "all" },
  { label: "My documents", value: "mine" },
  { label: "Shared", value: "shared" },
  { label: "Favorites", value: "favorites" },
  { label: "Templates", value: "templates" },
  { label: "Archived", value: "archived" },
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
  if (scope === "favorites")
    return "Favorite a document using the star beside its title.";
  if (scope === "archived")
    return "Documents you archive will appear here, ready to restore.";
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
  const apiScope =
    scope === "templates" || scope === "favorites" ? "all" : scope;
  const {
    data: documents = [],
    isPending,
    isError,
    refetch,
  } = useDocuments(search, apiScope);
  const { data: members = [] } = useMembers();
  const membersById = new Map(members.map((member) => [member.id, member]));
  const { favorites, error: favoritesError } = useFavorites();
  const favoriteDocumentIds = new Set(
    favorites
      .filter((favorite) => favorite.kind === "document")
      .map((favorite) => favorite.id),
  );
  const createDocument = useCreateDocument();
  const archiveDocument = useArchiveDocument();
  const restoreDocument = useRestoreArchivedDocument();
  const [creatingTemplateId, setCreatingTemplateId] = useState<string | null>(
    null,
  );
  const canCreateDocuments = userRole === "admin" || userRole === "member";
  const showTemplateGallery = scope === "templates";
  const listScope = scope === "templates" ? "all" : scope;
  const filteredDocuments = filterAndSortDocumentSummaries({
    currentUserId: session?.user.id,
    documents,
    favoriteDocumentIds,
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
    if (nextScope === "archived" && params.get("access") === "public") {
      params.delete("access");
    }
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
        visibility: template.id === "one-to-one" ? "private" : "workspace",
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
              <DocsIcon className="size-5 shrink-0" />
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
      <Box className="@container mx-auto w-full max-w-[90rem] px-6 pt-7 pb-32 sm:px-8 lg:px-10 lg:pt-8">
        <Flex
          align="start"
          className="mb-8 flex-col gap-x-6 gap-y-4 @3xl:flex-row"
          justify="between"
          wrap
        >
          <Box className="w-full min-w-0 @3xl:w-auto @3xl:flex-1">
            <Text as="h1" fontSize="2xl" fontWeight="semibold">
              {search ? `Results for “${search}”` : copy.heading}
            </Text>
            <Text className="mt-1 max-w-2xl" color="muted">
              {search
                ? "Search results across the selected document view."
                : copy.description}
            </Text>
          </Box>
          <Flex
            align="center"
            className="w-full @3xl:w-auto @3xl:shrink-0"
            gap={3}
            justify="between"
            wrap
          >
            {!showTemplateGallery ? (
              <Button
                color="tertiary"
                href={getDocumentsHref("templates", "")}
                variant="naked"
              >
                Browse templates
              </Button>
            ) : null}
            <Button
              disabled={!canCreateDocuments || createDocument.isPending}
              leftIcon={<PlusIcon className="size-5" />}
              loading={creatingTemplateId === "blank"}
              loadingText="Creating…"
              onClick={() => {
                handleCreate(documentTemplates[0]);
              }}
            >
              {showTemplateGallery ? "Blank document" : "New document"}
            </Button>
          </Flex>
        </Flex>

        {showTemplateGallery ? (
          <Box className="space-y-8">
            {documentTemplateCategories.map((category) => (
              <section
                aria-labelledby={`template-category-${category.id}`}
                key={category.id}
              >
                <Text
                  as="h2"
                  className="mb-3"
                  fontSize="lg"
                  fontWeight="semibold"
                  id={`template-category-${category.id}`}
                >
                  {category.label}
                </Text>
                <Box className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
                  {documentTemplates
                    .filter((template) => template.category === category.id)
                    .map((template) => (
                      <TemplateCard
                        disabled={
                          !canCreateDocuments || createDocument.isPending
                        }
                        isCreating={creatingTemplateId === template.id}
                        key={template.id}
                        onCreate={handleCreate}
                        template={template}
                      />
                    ))}
                </Box>
              </section>
            ))}
          </Box>
        ) : null}

        {!showTemplateGallery ? (
          <Box>
            <Flex align="center" className="mb-7 gap-3" wrap>
              <form
                className="min-w-48 flex-1"
                key={search}
                onSubmit={(event) => {
                  event.preventDefault();
                  const value = new FormData(event.currentTarget).get("search");
                  router.push(
                    getDocumentsHref(
                      scope,
                      typeof value === "string" ? value : "",
                    ),
                    { scroll: false },
                  );
                }}
                role="search"
              >
                <Input
                  aria-label="Search documents"
                  className="h-10 text-base"
                  defaultValue={search}
                  leftIcon={<SearchIcon className="size-5" />}
                  name="search"
                  placeholder="Search documents…"
                  type="search"
                />
              </form>
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
                showPublic={scope !== "archived"}
                updated={listState.updated}
              />
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
            <div
              aria-hidden
              className={cn(
                DOCUMENT_LIBRARY_COLUMNS,
                "text-text-muted border-border/70 border-b-[0.5px] pb-3",
              )}
            >
              <span>Document</span>
              <span className="hidden @4xl:block">Owner</span>
              <span className="hidden @4xl:block">Related work</span>
              <span className="hidden @2xl:block">Access</span>
              <span className="hidden @2xl:block">
                {scope === "archived" ? "Archived" : "Updated"}
              </span>
              <span />
            </div>
            {favoritesError && scope === "favorites" ? (
              <Text className="py-4" color="danger">
                Favorites could not be loaded in this browser.
              </Text>
            ) : null}
            {isError ? (
              <Flex align="center" className="py-12" direction="column" gap={3}>
                <Text>Documents could not be loaded.</Text>
                <Button
                  color="tertiary"
                  onClick={() => {
                    void refetch();
                  }}
                  variant="outline"
                >
                  Try again
                </Button>
              </Flex>
            ) : null}
            {isPending ? <DocumentsTableSkeleton /> : null}
            {!isPending && !isError && filteredDocuments.length === 0 ? (
              <EmptyDocuments
                canCreateDocuments={canCreateDocuments}
                isFiltered={Boolean(search) || activeFilterCount > 0}
                scope={scope}
              />
            ) : null}
            {!isPending && !isError && filteredDocuments.length > 0 ? (
              <ul aria-label={copy.heading}>
                {pagination.items.map((document) => (
                  <DocumentRow
                    document={document}
                    isArchived={scope === "archived"}
                    isArchiving={
                      archiveDocument.isPending
                        ? archiveDocument.variables === document.id
                        : false
                    }
                    isOwner={
                      session?.user.id === document.createdBy &&
                      document.canEdit
                    }
                    isRestoring={
                      restoreDocument.isPending
                        ? restoreDocument.variables === document.id
                        : false
                    }
                    key={document.id}
                    onArchive={() => {
                      archiveDocumentAndMaintainPage(document.id);
                    }}
                    onRestore={() => {
                      const nextPage = getDocumentPageAfterArchive(
                        pagination.page,
                        pagination.items.length,
                      );
                      restoreDocument.mutate(document.id, {
                        onSuccess: () => {
                          if (nextPage !== pagination.page) goToPage(nextPage);
                        },
                      });
                    }}
                    owner={membersById.get(document.createdBy)}
                    withWorkspace={withWorkspace}
                  />
                ))}
              </ul>
            ) : null}
            {!isPending && !isError && pagination.total > 0 ? (
              <Flex
                align="center"
                className="mt-5 gap-3"
                justify="between"
                wrap
              >
                <Text color="muted">
                  {pagination.start}–{pagination.end} of {pagination.total}{" "}
                  documents
                </Text>
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
                    variant="naked"
                  >
                    <ChevronRightIcon
                      className="h-5 w-auto"
                      strokeWidth={2.5}
                    />
                  </Button>
                </Flex>
              </Flex>
            ) : null}
          </Box>
        ) : null}
      </Box>
    </Box>
  );
};
