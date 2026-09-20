"use client";

import type { Dispatch, SetStateAction } from "react";
import { useCallback, useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { useEditor } from "@tiptap/react";
import Collaboration from "@tiptap/extension-collaboration";
import CollaborationCaret from "@tiptap/extension-collaboration-caret";
import { cn } from "lib";
import { toast } from "sonner";
import {
  Avatar,
  Box,
  Button,
  Divider,
  Flex,
  Menu,
  Skeleton,
  Text,
  TextEditor,
  type BubbleMenuCreateAction,
} from "ui";
import {
  ArchiveIcon,
  HistoryIcon,
  ArrowLeftIcon,
  CopyIcon,
  DeleteIcon,
  DuplicateIcon,
  Comment01Icon,
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
import {
  CommentHighlights,
  updateCommentHighlights,
} from "@/lib/tiptap/comment-highlights";
import { GoogleDriveFileSection } from "@/modules/google-drive/public/files";
import { useGoogleDriveDescriptionPaste } from "@/modules/google-drive/public/editor";
import indexStyles from "./document-index.module.css";
import { DocumentIndex } from "./document-index";
import {
  collaborationColor,
  useDocumentCollaboration,
} from "./use-document-collaboration";
import { DocumentHistory } from "./document-history";
import { DocumentCommentsPanel } from "./document-comments";
import { DocumentAccessMenu } from "./document-access-menu";
import {
  deleteDocumentMediaAction,
  uploadDocumentMediaAction,
} from "./actions";
import {
  useArchiveDocument,
  useDeleteDocument,
  useDocument,
  useDocumentComments,
  useDuplicateDocument,
  useUpdateDocument,
} from "./hooks";
import {
  DocumentRelationshipControl,
  RelatedWorkPanel,
} from "./related-work-panel";
import type { DocumentCommentSelection, DocumentUpdate } from "./types";
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

export const DocumentPage = ({ documentId }: { documentId: string }) => (
  <DocumentPageContent documentId={documentId} key={documentId} />
);

const DocumentPageContent = ({ documentId }: { documentId: string }) => {
  const router = useRouter();
  const { data: session } = useSession();
  const features = useFeatures();
  const { getTermDisplay } = useTerminology();
  const { userRole } = useUserRole();
  const { withWorkspace, workspaceSlug } = useWorkspacePath();
  const { data: document, isPending } = useDocument(documentId);
  const updateDocument = useUpdateDocument(documentId);
  const { data: commentThreads = [] } = useDocumentComments(documentId);
  const collaboration = useDocumentCollaboration(
    document?.id ?? "",
    workspaceSlug,
    session?.user,
  );
  const [historyOpen, setHistoryOpen] = useState(false);
  const canEditContent = Boolean(
    document?.canEdit &&
      (collaboration.configured
        ? collaboration.ready &&
          collaboration.writable &&
          collaboration.status !== "offline" &&
          collaboration.status !== "blocked"
        : !document.collaborative && !updateDocument.isError),
  );
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
  const [isCommentsOpen, setIsCommentsOpen] = useState(false);
  const [commentDraft, setCommentDraft] =
    useState<DocumentCommentSelection | null>(null);
  const isDesktop = useMediaQuery("(min-width: 768px)");
  const [scrollContainer, setScrollContainer] = useState<HTMLDivElement | null>(
    null,
  );
  const titleRef = useRef<HTMLTextAreaElement>(null);
  const loadedDocumentIdRef = useRef<string | null>(null);
  const editingRevision = useRef(0);
  const saveQueue = useRef<Promise<void> | null>(null);
  const saveFailed = useRef(false);
  const closeRelatedWork = useCallback(() => {
    setIsRelatedWorkOpen(false);
  }, [setIsRelatedWorkOpen]);
  const closeComments = useCallback(() => {
    setIsCommentsOpen(false);
    setCommentDraft(null);
  }, []);
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

  const persist = (payload: Omit<DocumentUpdate, "expectedRevision">) => {
    if (collaboration.configured || !document) return Promise.resolve();
    const pending = (saveQueue.current ?? Promise.resolve()).then(async () => {
      if (saveFailed.current) throw new Error("Reload before saving again");
      const response = await updateDocument.mutateAsync({
        ...payload,
        expectedRevision: editingRevision.current,
      });
      if (!response.data) throw new Error("Document was not saved");
      editingRevision.current = response.data.revision;
    });
    saveQueue.current = pending.catch(() => {
      saveFailed.current = true;
    });
    return pending;
  };
  const { callback: saveTitle, flush: flushTitle } = useDebouncedCallback(
    (nextTitle: string) => {
      void persist({ title: nextTitle }).catch(() => {});
    },
    700,
    { flushOnUnmount: true },
  );
  const { callback: saveContent, flush: flushContent } = useDebouncedCallback(
    (content: Pick<DocumentUpdate, "contentHtml" | "contentText">) => {
      void persist(content).catch(() => {});
    },
    700,
    { flushOnUnmount: true },
  );

  const editor = useEditor(
    {
      extensions: [
        CommentHighlights,
        ...createRichTextExtensions({
          collaborative: Boolean(collaboration.connection),
          onMediaFiles: handleMediaFiles,
          onMediaRequest: () => {
            window.document.getElementById(DOCUMENT_MEDIA_INPUT_ID)?.click();
          },
          placeholder: "Type / for commands",
        }),
        ...(collaboration.connection
          ? [
              Collaboration.configure({
                document: collaboration.connection.document,
              }),
              CollaborationCaret.configure({
                provider: collaboration.connection.provider,
                user: {
                  id: session?.user.id,
                  name: session?.user.name || "Teammate",
                  color: collaborationColor(session?.user.id ?? ""),
                },
              }),
            ]
          : []),
      ],
      content: "",
      editable: false,
      immediatelyRender: false,
      onUpdate: ({ editor: currentEditor }) => {
        if (!collaboration.configured)
          saveContent(getPersistableRichTextContent(currentEditor));
      },
      onBlur: flushContent,
    },
    [collaboration.connection],
  );
  const { onPaste: handleGoogleDrivePaste, picker: googleDrivePastePicker } =
    useGoogleDriveDescriptionPaste({
      editor,
      target: { id: documentId, type: "document" },
    });

  useEffect(() => {
    if (!document) return;
    editor?.setEditable(canEditContent, false);
    if (
      !collaboration.connection &&
      editor &&
      loadedDocumentIdRef.current !== document.id
    ) {
      editor.commands.setContent(document.contentHtml, { emitUpdate: false });
      loadedDocumentIdRef.current = document.id;
      editingRevision.current = document.revision;
      setTitleDraft({ documentId: document.id, value: document.title });
    }
  }, [document, editor, canEditContent, collaboration.connection]);

  useEffect(() => {
    if (!editor) return;
    updateCommentHighlights(
      editor,
      commentThreads
        .filter((thread) => !thread.resolvedAt)
        .map((thread) => ({
          from: thread.anchorStart,
          id: thread.id,
          to: thread.anchorEnd,
        })),
    );
  }, [commentThreads, editor, document?.revision, collaboration.ready]);

  useEffect(() => {
    const element = titleRef.current;
    if (!element) return;
    element.style.height = "0px";
    element.style.height = `${element.scrollHeight}px`;
  }, [document?.title, titleDraft, collaboration.title]);

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

  let title =
    titleDraft.documentId === documentId ? titleDraft.value : document.title;
  if (collaboration.ready) title = collaboration.title;
  let saveStatus = "Saved";
  if (updateDocument.isPending) saveStatus = "Saving…";
  if (updateDocument.isError) saveStatus = "Not saved";
  if (collaboration.configured)
    saveStatus = {
      connecting: "Connecting…",
      saved: "Saved",
      saving: "Saving…",
      offline: "Reconnecting…",
      blocked: "Connection needs attention",
    }[collaboration.status];
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
      if (!collaboration.configured) {
        flushTitle();
        flushContent();
        await persist({ title, ...content });
      }
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
                  disabled={
                    !canEditContent ||
                    collaboration.status === "saving" ||
                    duplicateDocument.isPending
                  }
                  onSelect={() => void handleDuplicate()}
                >
                  <DuplicateIcon />
                  {duplicateDocument.isPending
                    ? "Duplicating..."
                    : "Duplicate document"}
                </Menu.Item>
              </Menu.Group>
              <Menu.Group>
                <Menu.Item
                  onSelect={() => {
                    setHistoryOpen(true);
                  }}
                >
                  <HistoryIcon />
                  Version history
                </Menu.Item>
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
          <span aria-live="polite" className="text-text-muted hidden md:inline">
            {saveStatus}
          </span>
          <Flex className="-space-x-2">
            {collaboration.peers.slice(0, 3).map((peer) => (
              <span key={peer.id} title={peer.name}>
                <Avatar name={peer.name} size="xs" />
              </span>
            ))}
          </Flex>
          <Button
            aria-label={`${commentThreads.length} ${commentThreads.length === 1 ? "comment" : "comments"}`}
            className="gap-1.5"
            color="tertiary"
            leftIcon={<Comment01Icon className="size-4" strokeWidth={2} />}
            onClick={() => {
              setIsRelatedWorkOpen(false);
              setIsCommentsOpen(true);
            }}
            size="sm"
            variant="naked"
          >
            {commentThreads.length > 0 ? commentThreads.length : "Comments"}
          </Button>
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
            <Box className={cn("relative h-full min-w-0", indexStyles.host)}>
              <DocumentIndex
                contentSelector=".ProseMirror"
                readingOffset={112}
                scrollContainer={scrollContainer}
              />
              {documentHeader}
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
                <div className={indexStyles.contentPadding}>
                  <Box className="mx-auto w-full max-w-5xl px-8 pt-30 pb-32 sm:px-10 lg:px-12 lg:pt-34">
                    {(collaboration.configured &&
                      (collaboration.status === "blocked" ||
                        collaboration.status === "offline")) ||
                    updateDocument.isError ||
                    (!collaboration.configured && document.collaborative) ? (
                      <Box
                        className="bg-surface-elevated border-border mb-6 rounded-xl border p-4"
                        role="status"
                      >
                        <Text>
                          Editing is paused. Your connection or document access
                          changed. Copy any unsaved text before reloading.
                        </Text>
                        <Button
                          className="mt-3"
                          color="tertiary"
                          onClick={() => {
                            window.location.reload();
                          }}
                          size="sm"
                        >
                          Reload document
                        </Button>
                      </Box>
                    ) : null}
                    <DocumentRelationshipControl
                      document={document}
                      onShowRelationships={() => {
                        setIsCommentsOpen(false);
                        setIsRelatedWorkOpen(true);
                      }}
                    />
                    <textarea
                      aria-label="Document title"
                      className="text-foreground placeholder:text-text-muted mb-6 block min-h-14 w-full resize-none overflow-hidden bg-transparent text-4xl leading-tight font-semibold outline-none md:text-5xl"
                      disabled={!canEditContent}
                      maxLength={255}
                      onBlur={flushTitle}
                      onChange={(event) => {
                        if (collaboration.configured) {
                          collaboration.setTitle(event.target.value);
                          return;
                        }
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
                      bubbleMenuCommentAction={
                        canEditContent && userRole !== "guest"
                          ? (selection) => {
                              setCommentDraft(selection);
                              setIsRelatedWorkOpen(false);
                              setIsCommentsOpen(true);
                            }
                          : undefined
                      }
                      bubbleMenuCreateActions={bubbleMenuCreateActions}
                      bubbleMenuShouldShow={shouldShowDocumentTextMenu}
                      className={cn(
                        "rich-document-editor min-h-[55dvh] text-[1.1rem] leading-7",
                        styles.collaborativeEditor,
                      )}
                      editor={editor}
                      onPaste={handleGoogleDrivePaste}
                    />
                    {googleDrivePastePicker}
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
              </div>
            </Box>
          </BoardDividedPanel.MainPanel>
          <BoardDividedPanel.SideBar
            className="h-full!"
            isExpanded={isDesktop ? isRelatedWorkOpen || isCommentsOpen : false}
          >
            {isCommentsOpen ? (
              <DocumentCommentsPanel
                documentId={document.id}
                draft={commentDraft}
                editor={editor}
                onClose={closeComments}
                onDraftChange={setCommentDraft}
              />
            ) : (
              <RelatedWorkPanel
                document={document}
                onClose={closeRelatedWork}
              />
            )}
          </BoardDividedPanel.SideBar>
        </BoardDividedPanel>
      </Box>
      {historyOpen ? (
        <DocumentHistory
          canRestore={Boolean(
            canEditContent &&
              collaboration.status !== "saving" &&
              !updateDocument.isPending,
          )}
          document={document}
          isOpen
          onClose={() => {
            setHistoryOpen(false);
          }}
        />
      ) : null}
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
