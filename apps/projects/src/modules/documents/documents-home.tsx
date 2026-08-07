"use client";

import type { ComponentType } from "react";
import { useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Box, Button, Flex, Menu, Skeleton, Text } from "ui";
import {
  ArchiveIcon,
  CalendarIcon,
  DocsIcon,
  MoreHorizontalIcon,
  ObjectiveIcon,
  PlusIcon,
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
import { useArchiveDocument, useCreateDocument, useDocuments } from "./hooks";
import { formatDocumentRelativeTime } from "./relative-time";
import type { DocumentScope, WorkspaceDocument } from "./types";

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
  search: string,
  scope: DocumentScope,
  canCreateDocuments: boolean,
) => {
  if (search) return "Try a different title or phrase.";
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
  search,
  scope,
}: {
  canCreateDocuments: boolean;
  search: string;
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
      {search ? "No matching documents" : "No documents here yet"}
    </Text>
    <Text className="mt-1 max-w-md" color="muted">
      {getEmptyDescription(search, scope, canCreateDocuments)}
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
  const apiScope = scope === "templates" ? "all" : scope;
  const { data: documents = [], isPending } = useDocuments(search, apiScope);
  const createDocument = useCreateDocument();
  const archiveDocument = useArchiveDocument();
  const [creatingTemplateId, setCreatingTemplateId] = useState<string | null>(
    null,
  );
  const canCreateDocuments = userRole === "admin" || userRole === "member";

  const getDocumentsHref = (nextScope: DocumentScope, nextSearch = search) => {
    const params = new URLSearchParams();
    if (nextScope !== "all") params.set("scope", nextScope);
    if (nextSearch.trim() && nextScope !== "templates") {
      params.set("search", nextSearch.trim());
    }
    const query = params.toString();
    return withWorkspace(`/docs${query ? `?${query}` : ""}`);
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

  const showTemplateGallery = scope === "templates";
  const visibleTemplates = showTemplateGallery
    ? documentTemplates
    : documentTemplates.slice(0, 5);
  const copy = scopeCopy[scope];

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
      <Box className="mx-auto w-full max-w-[90rem] px-6 py-7 sm:px-8 lg:px-10 lg:py-8">
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
                {documents.length}{" "}
                {documents.length === 1 ? "document" : "documents"}
              </Text>
            </Flex>
            {isPending ? <DocumentsTableSkeleton /> : null}
            {!isPending && documents.length === 0 ? (
              <EmptyDocuments
                canCreateDocuments={canCreateDocuments}
                scope={scope}
                search={search}
              />
            ) : null}
            {!isPending && documents.length > 0 ? (
              <Box>
                {documents.map((document) => (
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
  document: WorkspaceDocument;
  isArchiving: boolean;
  isOwner: boolean;
  onArchive: () => void;
  withWorkspace: (path: string) => string;
}) => {
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
        <DocsIcon className="text-text-muted size-[1.1rem] shrink-0" />
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
              <Menu.Item className="text-danger" onSelect={onArchive}>
                <ArchiveIcon className="size-4" />
                Archive document
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      ) : null}
    </Flex>
  );
};
