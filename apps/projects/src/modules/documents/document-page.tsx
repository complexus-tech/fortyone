"use client";

import type { Dispatch, SetStateAction } from "react";
import { useCallback, useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { useEditor } from "@tiptap/react";
import { cn } from "lib";
import { toast } from "sonner";
import {
  Box,
  Button,
  Divider,
  Flex,
  Menu,
  Skeleton,
  Text,
  TextEditor,
  Tooltip,
  type BubbleMenuCreateAction,
} from "ui";
import {
  ArchiveIcon,
  ArrowLeftIcon,
  CopyIcon,
  DeleteIcon,
  DuplicateIcon,
  LinkIcon,
  LockKeyholeIcon,
  MoreHorizontalIcon,
  ObjectiveIcon,
  StoryIcon,
  UserMultiple02Icon,
} from "icons";
import {
  useCopyToClipboard,
  useFeatures,
  useLocalStorage,
  useMediaQuery,
  useTerminology,
  useUserRole,
  useWorkspacePath,
} from "@/hooks";
import {
  BoardDividedPanel,
  ConfirmDialog,
  NewObjectiveDialog,
  NewStoryDialog,
} from "@/components/ui";
import { useDebouncedCallback } from "@/hooks/debounce";
import { useSession } from "@/lib/auth/client";
import { createRichTextExtensions } from "@/lib/tiptap/rich-text-extensions";
import {
  getPersistableRichTextContent,
  RICH_TEXT_MEDIA_ACCEPT,
  uploadRichTextMediaFiles,
} from "@/lib/tiptap/rich-text-media";
import { RichTextTableMenu } from "@/lib/tiptap/rich-text-table-menu";
import { GoogleDriveFileSection } from "@/modules/google-drive";
import { DocumentAccessMenu } from "./document-access-menu";
import {
  deleteDocumentMediaAction,
  uploadDocumentMediaAction,
} from "./actions";
import {
  useArchiveDocument,
  useDeleteDocument,
  useDocument,
  useDuplicateDocument,
  useUpdateDocument,
} from "./hooks";
import { RelatedWorkPanel } from "./related-work-panel";
import type { DocumentUpdate } from "./types";
import styles from "./document-page.module.css";

const documentAccessLabels = {
  private: "Private",
  restricted: "Shared",
  workspace: "Workspace",
} as const;

const DOCUMENT_MEDIA_INPUT_ID = "document-media-upload";
const DOCUMENT_HEADER_BACKDROP_CLASS_NAME =
  "pointer-events-none absolute inset-x-0 top-0 z-20 h-18";

type DocumentCreationDraft = {
  description: string;
  kind: "story" | "objective";
};

const shouldShowDocumentTextMenu = ({
  editor,
}: {
  editor: NonNullable<ReturnType<typeof useEditor>>;
}) =>
  !editor.isActive("image") &&
  !editor.isActive("documentVideo") &&
  !editor.isActive("table");

const DocumentPageSkeleton = () => (
  <Box className="relative h-full min-h-0">
    <Box
      className={cn(DOCUMENT_HEADER_BACKDROP_CLASS_NAME, styles.headerBackdrop)}
    >
      <Flex
        align="center"
        className="pointer-events-auto relative z-10 h-18 px-5"
        justify="between"
      >
        <Skeleton className="h-5 w-44" />
        <Skeleton className="h-8 w-24" />
      </Flex>
    </Box>
    <Box className="mx-auto w-full max-w-5xl px-8 pt-34 pb-16 sm:px-10 lg:px-12 lg:pt-34">
      <Skeleton className="mb-8 h-12 w-3/4" />
      <Skeleton className="mb-3 h-6 w-full" />
      <Skeleton className="mb-3 h-6 w-5/6" />
      <Skeleton className="h-6 w-2/3" />
    </Box>
  </Box>
);

export const DocumentPage = ({ documentId }: { documentId: string }) => {
  const router = useRouter();
  const { data: session } = useSession();
  const features = useFeatures();
  const { getTermDisplay } = useTerminology();
  const { userRole } = useUserRole();
  const { withWorkspace, workspaceSlug } = useWorkspacePath();
  const { data: document, isPending } = useDocument(documentId);
  const updateDocument = useUpdateDocument(documentId);
  const archiveDocument = useArchiveDocument();
  const duplicateDocument = useDuplicateDocument();
  const deleteDocument = useDeleteDocument();
  const [, copyToClipboard] = useCopyToClipboard();
  const [creationDraft, setCreationDraft] =
    useState<DocumentCreationDraft | null>(null);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [titleDraft, setTitleDraft] = useState({
    documentId: "",
    value: "",
  });
  const [isRelatedWorkOpen, setIsRelatedWorkOpen] = useLocalStorage(
    "workspace:documents:related-work:isExpanded",
    false,
  );
  const isDesktop = useMediaQuery("(min-width: 768px)");
  const [scrollContainer, setScrollContainer] = useState<HTMLDivElement | null>(
    null,
  );
  const titleRef = useRef<HTMLTextAreaElement>(null);
  const loadedDocumentIdRef = useRef<string | null>(null);
  const closeRelatedWork = useCallback(() => {
    setIsRelatedWorkOpen(false);
  }, [setIsRelatedWorkOpen]);
  const setCreationDialogOpen: Dispatch<SetStateAction<boolean>> = (
    nextOpen,
  ) => {
    setCreationDraft((current) => {
      const isOpen =
        typeof nextOpen === "function" ? nextOpen(current !== null) : nextOpen;
      return isOpen ? current : null;
    });
  };
  const handleMediaFiles = useCallback(
    (
      currentEditor: NonNullable<ReturnType<typeof useEditor>>,
      files: File[],
      position?: number,
    ) => {
      void uploadRichTextMediaFiles({
        cleanup: async (media) => {
          const response = await deleteDocumentMediaAction(
            documentId,
            media.id,
            workspaceSlug,
          );
          if (response.error) {
            throw new Error(
              response.error.message || "Could not clean up uploaded media.",
            );
          }
        },
        editor: currentEditor,
        files,
        position,
        upload: async (file) => {
          const response = await uploadDocumentMediaAction(
            documentId,
            file,
            workspaceSlug,
          );
          if (response.error || !response.data) {
            throw new Error(
              response.error?.message || "Could not upload this media file.",
            );
          }
          return response.data;
        },
        onError: (_file, error) => {
          toast.error(
            error instanceof Error
              ? error.message
              : "Could not upload this media file.",
          );
        },
      });
    },
    [documentId, workspaceSlug],
  );

  const persist = (payload: DocumentUpdate) => {
    updateDocument.mutate(payload);
  };
  const { callback: saveTitle, flush: flushTitle } = useDebouncedCallback(
    (nextTitle: string) => {
      persist({ title: nextTitle });
    },
    700,
    { flushOnUnmount: true },
  );
  const { callback: saveContent, flush: flushContent } = useDebouncedCallback(
    (content: Pick<DocumentUpdate, "contentHtml" | "contentText">) => {
      persist(content);
    },
    700,
    { flushOnUnmount: true },
  );

  const editor = useEditor({
    extensions: createRichTextExtensions({
      onMediaFiles: handleMediaFiles,
      onMediaRequest: () => {
        window.document.getElementById(DOCUMENT_MEDIA_INPUT_ID)?.click();
      },
      placeholder: "Type / for commands",
    }),
    content: "",
    editable: false,
    immediatelyRender: false,
    onUpdate: ({ editor: currentEditor }) => {
      saveContent(getPersistableRichTextContent(currentEditor));
    },
    onBlur: flushContent,
  });

  useEffect(() => {
    if (!document) return;
    editor?.setEditable(document.canEdit);
    if (editor && loadedDocumentIdRef.current !== document.id) {
      editor.commands.setContent(document.contentHtml, { emitUpdate: false });
      loadedDocumentIdRef.current = document.id;
    }
  }, [document, editor]);

  useEffect(() => {
    const element = titleRef.current;
    if (!element) return;
    element.style.height = "0px";
    element.style.height = `${element.scrollHeight}px`;
  }, [document?.title, titleDraft]);

  if (isPending) return <DocumentPageSkeleton />;

  if (!document) {
    return (
      <Flex align="center" className="h-full px-8" justify="center">
        <Box className="max-w-md text-center">
          <Text className="mb-2" fontSize="xl" fontWeight="semibold">
            Document unavailable
          </Text>
          <Text color="muted">
            It may have been archived or you may no longer have access.
          </Text>
        </Box>
      </Flex>
    );
  }

  const title =
    titleDraft.documentId === documentId ? titleDraft.value : document.title;
  const canManageDocument =
    session?.user.id === document.createdBy && document.canEdit;
  const accessLabel = documentAccessLabels[document.visibility];
  const AccessIcon =
    document.visibility === "private" ? LockKeyholeIcon : UserMultiple02Icon;
  const canCreateWork =
    document.canEdit && userRole !== undefined && userRole !== "guest";
  const canEditGoogleDriveFiles = Boolean(
    document.canEdit && userRole !== "guest",
  );
  const bubbleMenuCreateActions: BubbleMenuCreateAction[] = canCreateWork
    ? [
        {
          id: "story",
          icon: <StoryIcon className="h-4 w-auto" strokeWidth={2} />,
          label: getTermDisplay("storyTerm", { capitalize: true }),
          onSelect: (description) => {
            setCreationDraft({ description, kind: "story" });
          },
        },
        ...(features.objectiveEnabled
          ? [
              {
                id: "objective",
                icon: <ObjectiveIcon className="h-4 w-auto" strokeWidth={2} />,
                label: getTermDisplay("objectiveTerm", { capitalize: true }),
                onSelect: (description: string) => {
                  setCreationDraft({ description, kind: "objective" });
                },
              },
            ]
          : []),
      ]
    : [];

  const handleArchive = () => {
    archiveDocument.mutate(document.id, {
      onSuccess: (response) => {
        if (!response.error) router.push(withWorkspace("/docs"));
      },
    });
  };

  const handleCopyLink = async () => {
    const copied = await copyToClipboard(window.location.href);
    if (copied) {
      toast.success("Document link copied");
      return;
    }
    toast.error("Could not copy the document link");
  };

  const handleDuplicate = async () => {
    const content = editor
      ? getPersistableRichTextContent(editor)
      : {
          contentHtml: document.contentHtml,
          contentText: document.contentText,
        };
    try {
      await updateDocument.mutateAsync({ title, ...content });
      duplicateDocument.mutate(document.id, {
        onSuccess: (response) => {
          if (response.data) {
            router.push(withWorkspace(`/docs/${response.data.id}`));
          }
        },
      });
    } catch {
      // The update mutation surfaces a save error and prevents a stale copy.
    }
  };

  const handleDelete = () => {
    deleteDocument.mutate(document.id, {
      onSuccess: (response) => {
        if (!response.error) {
          setIsDeleteDialogOpen(false);
          router.push(withWorkspace("/docs"));
        }
      },
    });
  };

  const documentHeader = (
    <Box
      className={cn(DOCUMENT_HEADER_BACKDROP_CLASS_NAME, styles.headerBackdrop)}
    >
      <Flex
        align="center"
        className="pointer-events-auto relative z-10 h-18 px-4 md:px-5"
        justify="between"
      >
        <Flex align="center" className="min-w-0" gap={2}>
          <Button
            aria-label="Back to documents"
            asIcon
            className="md:hidden"
            color="tertiary"
            onClick={() => {
              router.push(withWorkspace("/docs"));
            }}
            size="sm"
            variant="naked"
          >
            <ArrowLeftIcon />
          </Button>
          <Text
            className="max-w-80 truncate"
            fontSize="lg"
            fontWeight="semibold"
          >
            {title || "Untitled document"}
          </Text>
          <Menu>
            <Menu.Button>
              <Button
                aria-label="Document actions"
                asIcon
                color="tertiary"
                size="sm"
                variant="naked"
              >
                <MoreHorizontalIcon />
              </Button>
            </Menu.Button>
            <Menu.Items align="start" className="min-w-52">
              <Menu.Group>
                <Menu.Item onSelect={() => void handleCopyLink()}>
                  <CopyIcon />
                  Copy link
                </Menu.Item>
                <Menu.Item
                  disabled={!document.canEdit || duplicateDocument.isPending}
                  onSelect={() => void handleDuplicate()}
                >
                  <DuplicateIcon />
                  {duplicateDocument.isPending
                    ? "Duplicating..."
                    : "Duplicate document"}
                </Menu.Item>
              </Menu.Group>
              <Menu.Group>
                {canManageDocument ? (
                  <Menu.Item onSelect={handleArchive}>
                    <ArchiveIcon />
                    Archive document
                  </Menu.Item>
                ) : (
                  <Menu.Item disabled>Archive unavailable</Menu.Item>
                )}
              </Menu.Group>
              <Menu.Separator />
              <Menu.Group>
                {canManageDocument ? (
                  <Menu.Item
                    className="text-danger dark:text-danger!"
                    onSelect={() => {
                      setIsDeleteDialogOpen(true);
                    }}
                  >
                    <DeleteIcon className="text-danger dark:text-danger!" />
                    Delete permanently...
                  </Menu.Item>
                ) : (
                  <Menu.Item disabled>Delete unavailable</Menu.Item>
                )}
              </Menu.Group>
            </Menu.Items>
          </Menu>
        </Flex>
        <Flex align="center" gap={2}>
          {canManageDocument ? (
            <DocumentAccessMenu document={document} />
          ) : (
            <Button
              aria-label={`Document access: ${accessLabel}`}
              asIcon
              color="tertiary"
              disabled
              size="sm"
              variant="outline"
            >
              <AccessIcon />
            </Button>
          )}
        </Flex>
      </Flex>
    </Box>
  );

  return (
    <Flex className="h-full min-h-0 min-w-0" direction="column">
      <Box className="min-h-0 flex-1">
        <BoardDividedPanel autoSaveId="workspace:documents:related-work:divided-panel">
          <BoardDividedPanel.MainPanel>
            <Box className="relative h-full min-w-0">
              {documentHeader}
              {isDesktop && !isRelatedWorkOpen ? (
                <Box className="absolute top-24 right-5 z-30">
                  <Tooltip title="Show related work">
                    <Button
                      aria-expanded="false"
                      aria-label="Show related work"
                      className="border-border/70 bg-surface-elevated/90 shadow-shadow hover:border-border-strong hover:bg-surface-elevated dark:border-border-strong/80 dark:bg-surface-elevated/90 dark:hover:bg-surface-elevated gap-1.5 border-[0.5px] px-3 shadow-lg backdrop-blur-xl transition-[background-color,border-color,box-shadow,transform] hover:-translate-y-0.5"
                      color="tertiary"
                      leftIcon={<LinkIcon className="size-4" />}
                      onClick={() => {
                        setIsRelatedWorkOpen(true);
                      }}
                      rounded="full"
                      size="sm"
                      variant="outline"
                    >
                      Related work
                    </Button>
                  </Tooltip>
                </Box>
              ) : null}
              <input
                accept={RICH_TEXT_MEDIA_ACCEPT}
                aria-label="Upload document media"
                className="sr-only"
                id={DOCUMENT_MEDIA_INPUT_ID}
                multiple
                onChange={(event) => {
                  const files = Array.from(event.target.files ?? []);
                  event.target.value = "";
                  if (editor && files.length > 0) {
                    handleMediaFiles(editor, files);
                  }
                }}
                type="file"
              />
              <div
                className="h-full min-w-0 overflow-y-auto"
                ref={setScrollContainer}
              >
                <Box className="mx-auto w-full max-w-5xl px-8 pt-30 pb-32 sm:px-10 lg:px-12 lg:pt-34">
                  <textarea
                    aria-label="Document title"
                    className="text-foreground placeholder:text-text-muted mb-6 block min-h-14 w-full resize-none overflow-hidden bg-transparent text-4xl leading-tight font-semibold outline-none md:text-5xl"
                    disabled={!document.canEdit}
                    onBlur={flushTitle}
                    onChange={(event) => {
                      setTitleDraft({
                        documentId,
                        value: event.target.value,
                      });
                      saveTitle(event.target.value);
                    }}
                    placeholder="Untitled document"
                    ref={titleRef}
                    rows={1}
                    value={title}
                  />
                  <Divider className="mb-8" />
                  <TextEditor
                    bubbleMenuCreateActions={bubbleMenuCreateActions}
                    bubbleMenuShouldShow={shouldShowDocumentTextMenu}
                    className="rich-document-editor min-h-[55dvh] text-[1.1rem] leading-7"
                    editor={editor}
                  />
                  <RichTextTableMenu
                    editor={editor}
                    scrollTarget={scrollContainer}
                  />
                  <GoogleDriveFileSection
                    canEdit={canEditGoogleDriveFiles}
                    className="mt-10"
                    suggestedTitle={title}
                    target={{ id: documentId, type: "document" }}
                  />
                </Box>
              </div>
            </Box>
          </BoardDividedPanel.MainPanel>
          <BoardDividedPanel.SideBar
            className="h-full!"
            isExpanded={isDesktop ? isRelatedWorkOpen : false}
          >
            <RelatedWorkPanel document={document} onClose={closeRelatedWork} />
          </BoardDividedPanel.SideBar>
        </BoardDividedPanel>
      </Box>
      {isDeleteDialogOpen ? (
        <ConfirmDialog
          confirmPhrase="delete"
          confirmText="Delete permanently"
          description={`This will permanently delete “${title || "Untitled document"}” and cannot be undone.`}
          isLoading={deleteDocument.isPending}
          isOpen
          loadingText="Deleting..."
          onClose={() => {
            if (!deleteDocument.isPending) setIsDeleteDialogOpen(false);
          }}
          onConfirm={handleDelete}
          title="Delete document?"
        />
      ) : null}
      {creationDraft?.kind === "story" ? (
        <NewStoryDialog
          description={creationDraft.description}
          isOpen
          setIsOpen={setCreationDialogOpen}
        />
      ) : null}
      {creationDraft?.kind === "objective" ? (
        <NewObjectiveDialog
          description={creationDraft.description}
          isOpen
          setIsOpen={setCreationDialogOpen}
        />
      ) : null}
    </Flex>
  );
};
