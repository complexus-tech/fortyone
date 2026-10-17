"use client";

import type { Dispatch, SetStateAction } from "react";
import { useCallback, useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import { useEditor } from "@tiptap/react";
import Collaboration from "@tiptap/extension-collaboration";
import CollaborationCaret from "@tiptap/extension-collaboration-caret";
import { cn } from "lib";
import { toast } from "sonner";
import {
  Box,
  Divider,
  Flex,
  Text,
  TextEditor,
  type BubbleMenuCreateAction,
} from "ui";
import { ObjectiveIcon, StoryIcon } from "icons";
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
import { documentKeys } from "@/shared/documents/keys";
import { useDesktop } from "@/shell/desktop/desktop-provider";
import { createRichTextExtensions } from "@/lib/tiptap/rich-text-extensions";
import {
  getPersistableRichTextContent,
  RICH_TEXT_MEDIA_ACCEPT,
} from "@/lib/tiptap/rich-text-media";
import { RichTextTableMenu } from "@/lib/tiptap/rich-text-table-menu";
import { CommentHighlights } from "@/lib/tiptap/comment-highlights";
import { GoogleDriveFileSection } from "@/modules/google-drive/public/files";
import { useGoogleDriveDescriptionPaste } from "@/modules/google-drive/public/editor";
import {
  DocumentPersonReference,
  DocumentWorkReference,
} from "./document-inline-references";
import { useDocumentMentions } from "./use-document-mentions";
import { DocumentMentionSuggestions } from "./document-mentions";
import indexStyles from "./document-index.module.css";
import { DocumentIndex } from "./document-index";
import {
  collaborationColor,
  useDocumentCollaboration,
} from "./use-document-collaboration";
import { DocumentHistory } from "./document-history";
import { DocumentInsertMenu } from "./document-insert-menu";
import { DocumentCommentsPanel } from "./document-comments";
import {
  DocumentCommentsOverlay,
  DocumentCommentSelectionPopover,
} from "./document-comment-overlays";
import {
  useArchiveDocument,
  useDeleteDocument,
  useDocument,
  useDocumentComments,
  useDuplicateDocument,
} from "./hooks";
import {
  DocumentRelationshipControl,
  RelatedWorkPanel,
} from "./related-work-panel";
import {
  DocumentPageHeader,
  DocumentPageSkeleton,
} from "./document-page-header";
import { useDocumentCommentInteractions } from "./use-document-comment-interactions";
import { useDocumentMedia } from "./use-document-media";
import { useDocumentAutosave } from "./use-document-autosave";
import { DocumentSyncNotice } from "./document-sync-notice";
import styles from "./document-page.module.css";

const DOCUMENT_MEDIA_INPUT_ID = "document-media-upload";

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

export const DocumentPage = ({ documentId }: { documentId: string }) => (
  <DocumentPageContent documentId={documentId} key={documentId} />
);

const DocumentPageContent = ({ documentId }: { documentId: string }) => {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { data: session } = useSession();
  const features = useFeatures();
  const { getTermDisplay } = useTerminology();
  const { userRole } = useUserRole();
  const { withWorkspace, workspaceSlug } = useWorkspacePath();
  const { data: document, isPending } = useDocument(documentId);
  const { data: commentThreads = [] } = useDocumentComments(documentId);
  const collaboration = useDocumentCollaboration(
    document?.id ?? "",
    workspaceSlug,
    session?.user,
    document,
  );
  const {
    updateDocument,
    persist,
    saveTitle,
    flushTitle,
    saveContent,
    flushContent,
    initializeRevision,
  } = useDocumentAutosave(
    documentId,
    !collaboration.configured && Boolean(document),
  );
  const [recoveryOpen, setRecoveryOpen] = useState(false);
  const [openingRestored, setOpeningRestored] = useState(false);
  const [historyOpen, setHistoryOpen] = useState(false);
  const canEditContent = Boolean(
    document?.canEdit &&
      (collaboration.configured
        ? collaboration.ready &&
          collaboration.writable &&
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
  const isDesktop = useMediaQuery("(min-width: 768px)");
  const { isDesktop: isDesktopApp } = useDesktop();
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
  const handleMediaFiles = useDocumentMedia(documentId, workspaceSlug);

  const { callback: refreshDocumentMetadata } = useDebouncedCallback(() => {
    void queryClient.invalidateQueries({
      queryKey: documentKeys.detail(workspaceSlug, documentId),
      exact: true,
    });
  }, 1_000);

  const editor = useEditor(
    {
      extensions: [
        CommentHighlights,
        DocumentMentionSuggestions,
        ...createRichTextExtensions({
          collaborative: Boolean(collaboration.connection),
          mention: DocumentPersonReference,
          reference: DocumentWorkReference.configure({ documentId }),
          onMediaFiles: handleMediaFiles,
          onMediaRequest: () => {
            window.document.getElementById(DOCUMENT_MEDIA_INPUT_ID)?.click();
          },
          placeholder: "Type / for commands or @ to mention people and work",
        }),
        ...(collaboration.connection
          ? [
              Collaboration.configure({
                document: collaboration.connection.document,
              }),
              ...(collaboration.connection.provider
                ? [
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
            ]
          : []),
      ],
      content: "",
      editable: false,
      immediatelyRender: false,
      onUpdate: ({ editor: currentEditor }) => {
        if (!collaboration.configured)
          saveContent(getPersistableRichTextContent(currentEditor));
        else refreshDocumentMetadata(undefined);
      },
      onBlur: flushContent,
    },
    [collaboration.connection],
  );
  useDocumentMentions(editor, documentId, "body");
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
      initializeRevision(document.revision);
      setTitleDraft({ documentId: document.id, value: document.title });
    }
  }, [
    document,
    editor,
    canEditContent,
    collaboration.connection,
    initializeRevision,
  ]);

  const {
    isCommentsOpen,
    setIsCommentsOpen,
    commentDraft,
    setCommentDraft,
    isCommentComposerOpen,
    setIsCommentComposerOpen,
    activeCommentThreadId,
    commentThreadScope,
    setActiveCommentThreadId,
    closeComments,
    openComments,
    startComment,
  } = useDocumentCommentInteractions({
    editor,
    commentThreads,
    documentRevision: document?.revision,
    collaborationReady: collaboration.ready,
    closeRelatedWork,
  });

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
  if (collaboration.localStatus) saveStatus = collaboration.localStatus;
  const canManageDocument =
    session?.user.id === document.createdBy && document.canEdit;
  const canComment = userRole === "admin" || userRole === "member";
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

  return (
    <Flex className="h-full min-h-0 min-w-0" direction="column">
      <Box className="relative min-h-0 flex-1">
        <BoardDividedPanel autoSaveId="workspace:documents:related-work:divided-panel">
          <BoardDividedPanel.MainPanel>
            <Box className={cn("relative h-full min-w-0", indexStyles.host)}>
              <DocumentIndex
                contentSelector=".ProseMirror"
                readingOffset={112}
                scrollContainer={scrollContainer}
              />
              <DocumentPageHeader
                canDuplicate={Boolean(
                  canEditContent &&
                    collaboration.status !== "saving" &&
                    !duplicateDocument.isPending,
                )}
                canManageDocument={canManageDocument}
                commentCount={commentThreads.length}
                document={document}
                isDuplicating={duplicateDocument.isPending}
                onArchive={handleArchive}
                onBack={() => {
                  router.push(withWorkspace("/docs"));
                }}
                onCopyLink={handleCopyLink}
                onDelete={() => {
                  setIsDeleteDialogOpen(true);
                }}
                onDuplicate={handleDuplicate}
                onShowComments={openComments}
                onShowHistory={() => {
                  setHistoryOpen(true);
                }}
                peers={collaboration.peers}
                saveStatus={saveStatus}
                title={title}
              />
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
                    <DocumentSyncNotice
                      collaboration={collaboration}
                      collaborative={document.collaborative}
                      editor={editor}
                      hasSaveError={updateDocument.isError}
                      isDesktopApp={isDesktopApp}
                      onOpenRestored={() => {
                        setRecoveryOpen(true);
                      }}
                    />
                    <DocumentRelationshipControl
                      document={document}
                      onShowRelationships={() => {
                        setIsCommentsOpen(false);
                        setIsCommentComposerOpen(false);
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
                      allowFormatting={canEditContent}
                      bubbleMenuAppendTo={() => window.document.body}
                      bubbleMenuClassName="z-40 max-w-[calc(100vw-2rem)] [&>div>div]:max-w-full [&>div>div]:flex-wrap"
                      bubbleMenuCommentAction={
                        canComment ? startComment : undefined
                      }
                      bubbleMenuCreateActions={bubbleMenuCreateActions}
                      bubbleMenuOptions={{
                        placement: "bottom-start",
                        strategy: "fixed",
                        scrollTarget: scrollContainer ?? undefined,
                        flip: {
                          boundary: scrollContainer ?? "clippingAncestors",
                          padding: { top: 88, right: 16, bottom: 16, left: 16 },
                        },
                        shift: {
                          boundary: scrollContainer ?? "clippingAncestors",
                          padding: { top: 88, right: 16, bottom: 16, left: 16 },
                        },
                        size: {
                          boundary: scrollContainer ?? "clippingAncestors",
                          padding: 16,
                          apply: ({ availableWidth, elements }) => {
                            elements.floating.style.maxWidth = `${Math.max(0, availableWidth)}px`;
                          },
                        },
                      }}
                      bubbleMenuShouldShow={shouldShowDocumentTextMenu}
                      className={cn(
                        "rich-document-editor min-h-[55dvh] text-[1.1rem] leading-7",
                        styles.collaborativeEditor,
                      )}
                      editor={editor}
                      hideBubbleMenu={isCommentComposerOpen}
                      onPaste={handleGoogleDrivePaste}
                    />
                    {googleDrivePastePicker}
                    {editor && canEditContent ? (
                      <DocumentInsertMenu
                        editor={editor}
                        scrollContainer={scrollContainer}
                      />
                    ) : null}
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
            isExpanded={Boolean(isDesktop && isRelatedWorkOpen)}
          >
            <RelatedWorkPanel document={document} onClose={closeRelatedWork} />
          </BoardDividedPanel.SideBar>
        </BoardDividedPanel>
        {isCommentsOpen ? (
          <DocumentCommentsOverlay
            isDesktop={isDesktop}
            onClose={closeComments}
          >
            <DocumentCommentsPanel
              activeThreadId={activeCommentThreadId}
              canComment={canComment}
              documentId={document.id}
              draft={commentDraft}
              editor={editor}
              onClose={closeComments}
              onDraftChange={setCommentDraft}
              onThreadSelect={(thread) => {
                setActiveCommentThreadId(thread.id);
                closeComments();
              }}
              threadScope={commentThreadScope}
            />
          </DocumentCommentsOverlay>
        ) : null}
        {commentDraft && editor && isCommentComposerOpen ? (
          <DocumentCommentSelectionPopover
            canComment={
              canComment ? commentDraft.to > commentDraft.from : false
            }
            documentId={document.id}
            draft={commentDraft}
            editor={editor}
            isDesktop={isDesktop}
            onClose={() => {
              setIsCommentComposerOpen(false);
            }}
            onCreated={(thread) => {
              setCommentDraft(null);
              setIsCommentComposerOpen(false);
              setActiveCommentThreadId(thread.id);
              setIsCommentsOpen(true);
            }}
          />
        ) : null}
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
      {recoveryOpen ? (
        <ConfirmDialog
          confirmText="Open restored version"
          description="This removes the local editing copy. Your unsynced changes remain in the recovery file you exported."
          isLoading={openingRestored}
          isOpen
          loadingText="Opening…"
          onClose={() => {
            setRecoveryOpen(false);
          }}
          onConfirm={() => {
            setOpeningRestored(true);
            void collaboration
              .openRestored?.()
              .catch((reason: unknown) => toast.error(String(reason)))
              .finally(() => {
                setOpeningRestored(false);
              });
          }}
          title="Open the restored document?"
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
