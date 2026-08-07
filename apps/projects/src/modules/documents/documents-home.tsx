"use client";

import type { ComponentType } from "react";
import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Box, Button, Flex, Menu, Skeleton, Text } from "ui";
import {
  ArchiveIcon,
  ArrowDown2Icon,
  ArrowUpDownIcon,
  CalendarIcon,
  CheckIcon,
  DocsIcon,
  LockKeyholeIcon,
  MoreHorizontalIcon,
  ObjectiveIcon,
  PlusIcon,
  ShareIcon,
  TeamIcon,
  UserIcon,
} from "icons";
import { cn } from "lib";
import { ExpandableSearchHeader, MobileMenuButton } from "@/components/shared";
import { useUserRole, useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import {
  documentTemplates,
  type DocumentTemplate,
  type DocumentTemplateIcon,
} from "./document-templates";
import {
  filterAndSortDocumentSummaries,
  getDocumentListState,
  paginateDocumentSummaries,
} from "./document-list-state";
import { useArchiveDocument, useCreateDocument, useDocuments } from "./hooks";
import { formatDocumentRelativeTime } from "./relative-time";
import type {
  DocumentAccessFilter,
  DocumentOwnerFilter,
  DocumentScope,
  DocumentSortDirection,
  DocumentSortField,
  DocumentUpdatedFilter,
  WorkspaceDocumentSummary,
} from "./types";

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

type FilterOption<T extends string> = {
  label: string;
  value: T;
};

const accessOptions: FilterOption<DocumentAccessFilter>[] = [
  { label: "All access", value: "all" },
  { label: "Workspace", value: "workspace" },
  { label: "Shared", value: "restricted" },
  { label: "Private", value: "private" },
];

const ownerOptions: FilterOption<DocumentOwnerFilter>[] = [
  { label: "Anyone", value: "all" },
  { label: "Owned by me", value: "mine" },
  { label: "Owned by others", value: "others" },
];

const updatedOptions: FilterOption<DocumentUpdatedFilter>[] = [
  { label: "Any time", value: "all" },
  { label: "Today", value: "today" },
  { label: "Past 7 days", value: "7d" },
  { label: "Past 30 days", value: "30d" },
  { label: "Past 90 days", value: "90d" },
];

type DocumentSortOption = {
  direction: DocumentSortDirection;
  field: DocumentSortField;
  label: string;
};

const sortOptions: DocumentSortOption[] = [
  { direction: "desc", field: "updated", label: "Updated: newest" },
  { direction: "asc", field: "updated", label: "Updated: oldest" },
  { direction: "desc", field: "created", label: "Created: newest" },
  { direction: "asc", field: "created", label: "Created: oldest" },
  { direction: "asc", field: "title", label: "Title: A–Z" },
  { direction: "desc", field: "title", label: "Title: Z–A" },
];

function DocumentControlMenu<T extends string>({
  active,
  icon: Icon,
  label,
  onChange,
  options,
  value,
}: {
  active?: boolean;
  icon: ComponentType<{ className?: string; strokeWidth?: number }>;
  label: string;
  onChange: (value: T) => void;
  options: FilterOption<T>[];
  value: T;
}) {
  return (
    <Menu>
      <Menu.Button>
        <Button
          className={cn("gap-1.5 whitespace-nowrap", {
            "bg-state-active": active,
          })}
          color="tertiary"
          leftIcon={<Icon className="h-4 w-auto" strokeWidth={2} />}
          rightIcon={
            <ArrowDown2Icon className="h-3.5 w-auto" strokeWidth={2} />
          }
          size="sm"
          variant="outline"
        >
          {label}
        </Button>
      </Menu.Button>
      <Menu.Items align="start" className="min-w-48 p-1">
        {options.map((option) => (
          <Menu.Item
            active={option.value === value}
            className="justify-between gap-3"
            key={option.value}
            onSelect={() => {
              onChange(option.value);
            }}
          >
            <span>{option.label}</span>
            {option.value === value ? (
              <CheckIcon className="h-4 w-auto" strokeWidth={2} />
            ) : null}
          </Menu.Item>
        ))}
      </Menu.Items>
    </Menu>
  );
}

const DocumentSortMenu = ({
  direction,
  field,
  onChange,
}: {
  direction: DocumentSortDirection;
  field: DocumentSortField;
  onChange: (option: DocumentSortOption) => void;
}) => {
  const selectedOption =
    sortOptions.find(
      (option) => option.field === field && option.direction === direction,
    ) ?? sortOptions[0];

  return (
    <Menu>
      <Menu.Button>
        <Button
          className="gap-1.5 whitespace-nowrap"
          color="tertiary"
          leftIcon={<ArrowUpDownIcon className="h-4 w-auto" strokeWidth={2} />}
          rightIcon={
            <ArrowDown2Icon className="h-3.5 w-auto" strokeWidth={2} />
          }
          size="sm"
          variant="outline"
        >
          {selectedOption.label}
        </Button>
      </Menu.Button>
      <Menu.Items align="start" className="min-w-52 p-1">
        {sortOptions.map((option) => {
          const isActive =
            option.field === field && option.direction === direction;
          return (
            <Menu.Item
              active={isActive}
              className="justify-between gap-3"
              key={`${option.field}:${option.direction}`}
              onSelect={() => {
                onChange(option);
              }}
            >
              <span>{option.label}</span>
              {isActive ? (
                <CheckIcon className="h-4 w-auto" strokeWidth={2} />
              ) : null}
            </Menu.Item>
          );
        })}
      </Menu.Items>
    </Menu>
  );
};

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
        "border-border/60 bg-surface/60 hover:border-border-strong focus-visible:border-border-strong focus-visible:ring-border-strong/50 dark:border-border/75 dark:bg-surface/70 dark:hover:border-border-strong dark:focus-visible:border-border-strong flex min-h-20 min-w-[13.5rem] items-center gap-2.5 rounded-2xl border-[0.5px] px-4 py-3 text-left transition-[border-color] outline-none focus-visible:ring-1 disabled:opacity-60",
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
        className="border-border/70 bg-surface-elevated text-foreground dark:border-border-strong/80 size-12 shrink-0 rounded-xl border-[0.5px]"
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
    <DocsIcon className="text-text-muted h-16 w-auto" strokeWidth={1.5} />
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
  const selectedAccessLabel =
    accessOptions.find((option) => option.value === listState.access)?.label ??
    "All access";
  const selectedOwnerLabel =
    ownerOptions.find((option) => option.value === listState.owner)?.label ??
    "Anyone";
  const selectedUpdatedLabel =
    updatedOptions.find((option) => option.value === listState.updated)
      ?.label ?? "Any time";
  const hasCustomListControls =
    listState.access !== "all" ||
    (scope !== "mine" && listState.owner !== "all") ||
    listState.updated !== "all" ||
    listState.sort !== "updated" ||
    listState.direction !== "desc";

  useEffect(() => {
    if (
      isPending ||
      showTemplateGallery ||
      listState.page === pagination.page
    ) {
      return;
    }
    const params = new URLSearchParams(rawSearchParams);
    if (pagination.page === 1) {
      params.delete("page");
    } else {
      params.set("page", String(pagination.page));
    }
    const query = params.toString();
    router.replace(withWorkspace(`/docs${query ? `?${query}` : ""}`), {
      scroll: false,
    });
  }, [
    isPending,
    listState.page,
    pagination.page,
    rawSearchParams,
    router,
    showTemplateGallery,
    withWorkspace,
  ]);

  return (
    <Box className="h-dvh min-w-0 overflow-y-auto">
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
              className="border-border border-b-[0.5px] pb-3"
              justify="between"
            >
              <Text fontSize="lg" fontWeight="semibold">
                {search ? "Matching documents" : copy.heading}
              </Text>
              <Text color="muted">
                {filteredDocuments.length}{" "}
                {filteredDocuments.length === 1 ? "document" : "documents"}
              </Text>
            </Flex>
            <Flex
              align="center"
              className="border-border/70 border-b-[0.5px] py-3"
              gap={2}
              wrap
            >
              <DocumentControlMenu
                active={listState.access !== "all"}
                icon={LockKeyholeIcon}
                label={
                  listState.access === "all"
                    ? "Access"
                    : `Access: ${selectedAccessLabel}`
                }
                onChange={(access) => {
                  updateListControls({
                    access: access === "all" ? null : access,
                  });
                }}
                options={accessOptions}
                value={listState.access}
              />
              {scope !== "mine" ? (
                <DocumentControlMenu
                  active={listState.owner !== "all"}
                  icon={UserIcon}
                  label={
                    listState.owner === "all" ? "Owner" : selectedOwnerLabel
                  }
                  onChange={(owner) => {
                    updateListControls({
                      owner: owner === "all" ? null : owner,
                    });
                  }}
                  options={ownerOptions}
                  value={listState.owner}
                />
              ) : null}
              <DocumentControlMenu
                active={listState.updated !== "all"}
                icon={CalendarIcon}
                label={
                  listState.updated === "all" ? "Updated" : selectedUpdatedLabel
                }
                onChange={(updated) => {
                  updateListControls({
                    updated: updated === "all" ? null : updated,
                  });
                }}
                options={updatedOptions}
                value={listState.updated}
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
              {hasCustomListControls ? (
                <Button
                  color="tertiary"
                  onClick={() => {
                    updateListControls({
                      access: null,
                      direction: null,
                      owner: null,
                      sort: null,
                      updated: null,
                    });
                  }}
                  size="sm"
                  variant="naked"
                >
                  Clear
                </Button>
              ) : null}
            </Flex>
            {isPending ? <DocumentsTableSkeleton /> : null}
            {!isPending && filteredDocuments.length === 0 ? (
              <EmptyDocuments
                canCreateDocuments={canCreateDocuments}
                isFiltered={Boolean(search) || hasCustomListControls}
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
                      archiveDocument.mutate(document.id);
                    }}
                    withWorkspace={withWorkspace}
                  />
                ))}
                <Flex
                  align="center"
                  className="border-border/70 py-3"
                  justify="between"
                >
                  <Text color="muted">
                    {pagination.start}–{pagination.end} of {pagination.total}
                  </Text>
                  <Flex align="center" gap={2}>
                    <Button
                      color="tertiary"
                      disabled={pagination.page <= 1}
                      onClick={() => {
                        goToPage(pagination.page - 1);
                      }}
                      size="sm"
                      variant="outline"
                    >
                      Previous
                    </Button>
                    <Button
                      color="tertiary"
                      disabled={pagination.page >= pagination.pageCount}
                      onClick={() => {
                        goToPage(pagination.page + 1);
                      }}
                      size="sm"
                      variant="outline"
                    >
                      Next
                    </Button>
                  </Flex>
                </Flex>
              </Box>
            ) : null}
          </Box>
        ) : null}
      </Box>
    </Box>
  );
};

const DocumentRow = ({
  document,
  isArchiving,
  isOwner,
  onArchive,
  withWorkspace,
}: {
  document: WorkspaceDocumentSummary;
  isArchiving: boolean;
  isOwner: boolean;
  onArchive: () => void;
  withWorkspace: (path: string) => string;
}) => {
  let DocumentIcon = DocsIcon;
  if (document.visibility === "private") DocumentIcon = LockKeyholeIcon;
  if (document.visibility === "restricted") DocumentIcon = ShareIcon;

  return (
    <Flex
      align="center"
      className="border-border/80 hover:bg-state-hover/50 border-b-[0.5px] py-[0.655rem] transition-colors"
      gap={2}
    >
      <Link
        className="focus-visible:ring-ring/40 flex min-w-0 flex-1 items-center gap-2 rounded-sm outline-none focus-visible:ring-1"
        href={withWorkspace(`/docs/${document.id}`)}
      >
        <DocumentIcon
          className="text-text-muted size-[1.1rem] shrink-0"
          strokeWidth={2}
        />
        <span className="min-w-0 flex-1 truncate">{document.title}</span>
        <Text as="span" className="shrink-0" color="muted">
          {formatDocumentRelativeTime(document.updatedAt)}
        </Text>
      </Link>
      {isOwner ? (
        <Menu>
          <Menu.Button>
            <Button
              aria-label={`Actions for ${document.title}`}
              asIcon
              color="tertiary"
              disabled={isArchiving}
              size="sm"
              variant="naked"
            >
              <MoreHorizontalIcon />
            </Button>
          </Menu.Button>
          <Menu.Items align="end" className="min-w-52">
            <Menu.Group>
              <Menu.Item
                className="text-danger dark:!text-danger"
                onSelect={onArchive}
              >
                <ArchiveIcon
                  className="text-danger dark:!text-danger size-4"
                  strokeWidth={2}
                />
                Archive document
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      ) : null}
    </Flex>
  );
};
