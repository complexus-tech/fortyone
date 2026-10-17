"use client";
import { Box, Container, Divider, TextEditor } from "ui";
import { useEditor } from "@tiptap/react";
import Placeholder from "@tiptap/extension-placeholder";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import TextExtension from "@tiptap/extension-text";
import { cn } from "lib";
import type { ClipboardEvent, ReactNode } from "react";
import { useCallback, useEffect, useRef } from "react";
import { useLocalStorage, useUserRole } from "@/hooks";
import { useDebouncedCallback } from "@/hooks/debounce";
import { BodyContainer } from "@/components/shared";
import { useLinks } from "@/lib/hooks/links";
import { createRichTextExtensions } from "@/lib/tiptap/rich-text-extensions";
import { getRichTextContentType } from "@/lib/tiptap/markdown";
import {
  getPersistableRichTextContent,
  hasPendingRichTextMedia,
  RICH_TEXT_MEDIA_ACCEPT,
} from "@/lib/tiptap/rich-text-media";
import {
  createRichTextSnapshot,
  type RichTextSnapshot,
} from "@/lib/tiptap/rich-text-snapshot";
import { RichTextTableMenu } from "@/lib/tiptap/rich-text-table-menu";
import { useUpdateStoryMutation } from "@/modules/story/hooks/update-mutation";
import { useStoryDescriptionMedia } from "@/modules/story/hooks/use-story-description-media";
import { useIsAdminOrOwner } from "@/hooks/owner";
import { RelatedDocuments } from "@/modules/documents/related-documents";
import { GoogleDriveFileSection } from "@/modules/google-drive/public/files";
import { useGoogleDriveDescriptionPaste } from "@/modules/google-drive/public/editor";
import { useDesktopDraft } from "@/shell/desktop/use-desktop-draft";
import { DesktopDraftNotice } from "@/shell/desktop/desktop-draft-notice";
import { useStoryById } from "../hooks/story";
import {
  useStoryEditorSaving,
  useStoryEditorSync,
} from "../hooks/use-story-editor-sync";
import type { StoryUpdate } from "../types";
import { Activities } from "./activities";
import { Associations } from "./associations";
import { Attachments } from "./attachments";
import { Links } from "./links";
import { SubStories } from "./sub-stories";
import { StoryChecklist } from "./checklist";
import { StoryBanners } from "./story-banners";
import { LinksSkeleton } from "./links-skeleton";
import { FigmaSection } from "./figma-section";
import { OptionsHeader } from "./options-header";
import { Options } from "./options";
import { useFigmaDescriptionPaste } from "./use-figma-description-paste";
import { StoryDescription } from "./story-description";

const DEBOUNCE_DELAY = 1000; // 1000ms delay

type StoryEditorDraft = {
  title: string;
  description: string;
  descriptionHTML: string;
  baseVersion: string;
};

type QueuedStoryUpdate = {
  payload: StoryUpdate;
  storyId: string;
};

type QueuedDescriptionUpdate = {
  snapshot: RichTextSnapshot;
  storyId: string;
};

export const MainDetails = ({
  inlineProperties = false,
  storyId,
  isNotifications,
  isDialog,
  mainHeader,
  onClose,
}: {
  inlineProperties?: boolean;
  storyId: string;
  isNotifications: boolean;
  isDialog?: boolean;
  mainHeader?: ReactNode;
  onClose?: () => void;
}) => {
  const { data } = useStoryById(storyId);
  const { data: links = [], isLoading: isLinksLoading } = useLinks(storyId);
  const { mutate: updateStory } = useUpdateStoryMutation();
  const isSaving = useStoryEditorSaving(storyId);
  const { userRole } = useUserRole();
  const {
    handleMediaFiles,
    inputRef: mediaInputRef,
    openMediaPicker,
  } = useStoryDescriptionMedia(storyId);

  const [isSubStoriesOpen, setIsSubStoriesOpen] = useLocalStorage(
    "isSubStoriesOpen",
    true,
  );
  const [isLinksOpen, setIsLinksOpen] = useLocalStorage("isLinksOpen", true);
  const [isAssociationsOpen, setIsAssociationsOpen] = useLocalStorage(
    "isAssociationsOpen",
    true,
  );
  const {
    title,
    descriptionHTML,
    description,
    teamId,
    deletedAt,
    reporterId,
    associations = [],
  } = data!;
  const isDeleted = Boolean(deletedAt);
  const desktopDraft = useDesktopDraft<StoryEditorDraft>(`story:${storyId}`);
  const latestDraft = useRef<{
    storyId: string;
    payload: StoryEditorDraft;
  } | null>(null);
  const saveEditorDraft = useCallback(
    (patch: Partial<StoryEditorDraft>) => {
      if (
        !desktopDraft.supported ||
        !desktopDraft.ready ||
        desktopDraft.recovered
      )
        return Promise.resolve();
      const previous =
        latestDraft.current?.storyId === storyId
          ? latestDraft.current.payload
          : {
              title,
              description,
              descriptionHTML,
              baseVersion: data!.updatedAt,
            };
      const payload = { ...previous, ...patch };
      latestDraft.current = { storyId, payload };
      return desktopDraft.save(payload);
    },
    [data, description, descriptionHTML, desktopDraft, storyId, title],
  );
  const editorEnabled =
    !isDeleted &&
    userRole !== "guest" &&
    desktopDraft.ready &&
    !desktopDraft.recovered;
  const { isAdminOrOwner } = useIsAdminOrOwner(reporterId);
  const handleUpdate = useCallback(
    ({ payload, storyId: targetStoryId }: QueuedStoryUpdate) => {
      updateStory({ storyId: targetStoryId, payload });
    },
    [updateStory],
  );
  const handleDescriptionUpdate = useCallback(
    ({ snapshot, storyId: targetStoryId }: QueuedDescriptionUpdate) => {
      const content = getPersistableRichTextContent(snapshot);
      handleUpdate({
        payload: {
          descriptionHTML: content.contentHtml,
          description: content.contentText,
          reconcileDescriptionMedia: !hasPendingRichTextMedia(snapshot),
        },
        storyId: targetStoryId,
      });
    },
    [handleUpdate],
  );

  // Keep independent queues so editing one field never cancels the other field's save.
  const {
    callback: debouncedDescriptionUpdate,
    flush: flushDescriptionUpdate,
  } = useDebouncedCallback(handleDescriptionUpdate, DEBOUNCE_DELAY, {
    flushOnUnmount: true,
  });
  const { callback: debouncedTitleUpdate, flush: flushTitleUpdate } =
    useDebouncedCallback(handleUpdate, DEBOUNCE_DELAY, {
      flushOnUnmount: true,
    });

  const descriptionEditor = useEditor({
    extensions: createRichTextExtensions({
      onMediaFiles: handleMediaFiles,
      onMediaRequest: openMediaPicker,
      placeholder: "Enter description or type / for commands...",
    }),
    content: descriptionHTML || description,
    contentType: getRichTextContentType(description, descriptionHTML),
    editable: editorEnabled,
    onCreate: ({ editor }) => {
      if (
        getRichTextContentType(description, descriptionHTML) !== "markdown" ||
        isDeleted ||
        userRole === "guest" ||
        !desktopDraft.ready ||
        desktopDraft.recovered
      ) {
        return;
      }

      const content = getPersistableRichTextContent(editor);
      handleUpdate({
        payload: {
          descriptionHTML: content.contentHtml,
          description: content.contentText,
          reconcileDescriptionMedia: true,
        },
        storyId,
      });
    },
    onUpdate: ({ editor }) => {
      if (desktopDraft.supported) {
        const content = getPersistableRichTextContent(editor);
        void saveEditorDraft({
          description: content.contentText,
          descriptionHTML: content.contentHtml,
        });
      }
      debouncedDescriptionUpdate({
        snapshot: createRichTextSnapshot(editor),
        storyId,
      });
    },
    onBlur: () => {
      flushDescriptionUpdate();
    },
    immediatelyRender: false,
  });

  const titleEditor = useEditor({
    extensions: [
      Document,
      Paragraph,
      TextExtension,
      Placeholder.configure({ placeholder: "Enter title..." }),
    ],
    content: title,
    editable: editorEnabled,
    onUpdate: ({ editor }) => {
      void saveEditorDraft({ title: editor.getText() });
      debouncedTitleUpdate({
        payload: { title: editor.getText() },
        storyId,
      });
    },
    onBlur: () => {
      flushTitleUpdate();
    },
    immediatelyRender: false,
  });
  useEffect(() => {
    if (!desktopDraft.supported) return;
    titleEditor?.setEditable(editorEnabled, false);
    descriptionEditor?.setEditable(editorEnabled, false);
  }, [descriptionEditor, desktopDraft.supported, editorEnabled, titleEditor]);

  useEffect(() => {
    const beforeClose = (event: Event) => {
      flushTitleUpdate();
      flushDescriptionUpdate();
      if (
        latestDraft.current?.storyId !== storyId ||
        !titleEditor ||
        !descriptionEditor ||
        !desktopDraft.ready ||
        desktopDraft.recovered
      )
        return;
      const content = getPersistableRichTextContent(descriptionEditor);
      const save = saveEditorDraft({
        title: titleEditor.getText(),
        description: content.contentText,
        descriptionHTML: content.contentHtml,
      });
      (
        event as CustomEvent<{ waitUntil: (write: Promise<void>) => void }>
      ).detail.waitUntil(save);
    };
    window.addEventListener("fortyone:desktop-before-close", beforeClose);
    return () => {
      window.removeEventListener("fortyone:desktop-before-close", beforeClose);
    };
  }, [
    descriptionEditor,
    desktopDraft.ready,
    desktopDraft.recovered,
    flushDescriptionUpdate,
    flushTitleUpdate,
    saveEditorDraft,
    storyId,
    titleEditor,
  ]);

  const restoreEditorDraft = () => {
    const recovered = desktopDraft.recovered;
    if (
      !recovered ||
      typeof recovered.title !== "string" ||
      typeof recovered.description !== "string" ||
      typeof recovered.descriptionHTML !== "string" ||
      typeof recovered.baseVersion !== "string"
    )
      return;
    titleEditor?.commands.setContent(recovered.title, { emitUpdate: false });
    descriptionEditor?.commands.setContent(
      recovered.descriptionHTML || recovered.description,
      { emitUpdate: false },
    );
    markTitleEdited();
    markDescriptionEdited();
    latestDraft.current = { storyId, payload: recovered };
    desktopDraft.acceptRecovered();
    updateStory({
      storyId,
      desktopBaseVersion: recovered.baseVersion,
      payload: {
        title: recovered.title,
        description: recovered.description,
        descriptionHTML: recovered.descriptionHTML,
        reconcileDescriptionMedia: true,
      },
    });
  };

  const handleFigmaDescriptionPaste = useFigmaDescriptionPaste({
    editor: descriptionEditor,
    storyId,
  });
  const {
    onPaste: handleGoogleDriveDescriptionPaste,
    picker: googleDriveDescriptionPicker,
  } = useGoogleDriveDescriptionPaste({
    editor: descriptionEditor,
    target: { id: storyId, type: "story" },
  });
  const handleDescriptionPaste = (event: ClipboardEvent<HTMLDivElement>) => {
    handleGoogleDriveDescriptionPaste(event);
    handleFigmaDescriptionPaste(event);
  };

  useEffect(
    () => () => {
      flushDescriptionUpdate();
      flushTitleUpdate();
    },
    [flushDescriptionUpdate, flushTitleUpdate, storyId],
  );

  const markTitleEdited = useStoryEditorSync({
    editor: titleEditor,
    storyId,
    text: title,
    saving: isSaving,
  });
  const markDescriptionEdited = useStoryEditorSync({
    editor: descriptionEditor,
    storyId,
    text: description,
    html: descriptionHTML,
    saving: isSaving,
  });

  return (
    <BodyContainer
      className={cn("h-full min-h-0 overflow-y-auto pb-8", {
        "pb-48": isDialog,
      })}
    >
      {mainHeader}
      {onClose ? (
        <OptionsHeader
          isAdminOrOwner={isAdminOrOwner}
          onClose={onClose}
          storyId={storyId}
        />
      ) : (
        <Box className="md:hidden">
          <OptionsHeader isAdminOrOwner={isAdminOrOwner} storyId={storyId} />
        </Box>
      )}

      <Container
        className={cn("max-w-6xl pt-4 md:pt-7", {
          "md:px-6 md:pt-2": isDialog,
          "px-5 pt-4 md:px-5 md:pt-4": Boolean(onClose),
        })}
      >
        {isNotifications && !onClose ? (
          <Box className="notification-story-top-options-header relative -top-4.5 -mb-2 hidden [&>div]:h-auto [&>div]:px-0 [&>div]:pt-0">
            <OptionsHeader isAdminOrOwner={isAdminOrOwner} storyId={storyId} />
          </Box>
        ) : null}
        <StoryBanners story={data!} />
        <DesktopDraftNotice
          error={desktopDraft.error}
          onDiscard={() => {
            void desktopDraft.discard();
          }}
          onRestore={restoreEditorDraft}
          onRetry={() => {
            void desktopDraft.retry();
          }}
          recovered={Boolean(desktopDraft.recovered)}
          status={desktopDraft.status}
        />
        <TextEditor
          asTitle
          className="text-foreground relative -left-px mb-4 text-3xl font-semibold md:text-4xl"
          editor={titleEditor}
        />
        <input
          accept={RICH_TEXT_MEDIA_ACCEPT}
          aria-label="Upload story description media"
          className="sr-only"
          multiple
          onChange={(event) => {
            const files = Array.from(event.target.files ?? []);
            event.target.value = "";
            if (descriptionEditor && files.length > 0) {
              handleMediaFiles(descriptionEditor, files);
            }
          }}
          ref={mediaInputRef}
          type="file"
        />
        <StoryDescription key={`description-${storyId}`}>
          <TextEditor
            className="rich-document-editor story-description-editor text-foreground dark:text text-[1.1rem]"
            editor={descriptionEditor}
            onPaste={handleDescriptionPaste}
          />
        </StoryDescription>
        {googleDriveDescriptionPicker}
        <RichTextTableMenu editor={descriptionEditor} scrollTarget={null} />
        <StoryChecklist key={`checklist-${storyId}`} story={data!} />
        <SubStories
          isSubStoriesOpen={isSubStoriesOpen}
          parent={data!}
          setIsSubStoriesOpen={setIsSubStoriesOpen}
        />
        <Associations
          associations={associations}
          isAssociationsOpen={isAssociationsOpen}
          setIsAssociationsOpen={setIsAssociationsOpen}
          storyId={storyId}
        />
        <FigmaSection storyId={storyId} />
        {isLinksLoading ? (
          <LinksSkeleton />
        ) : (
          <Links
            isLinksOpen={isLinksOpen}
            links={links}
            setIsLinksOpen={setIsLinksOpen}
            storyId={storyId}
          />
        )}
        <RelatedDocuments entityId={storyId} entityType="story" />
        <GoogleDriveFileSection
          canEdit={!isDeleted && userRole !== "guest"}
          suggestedTitle={title}
          target={{ id: storyId, type: "story" }}
        />
        <Box
          className={cn("md:hidden", {
            "md:block": inlineProperties,
            "mt-4": isNotifications && isLinksOpen && links.length > 0,
            "notification-story-inline-options": isNotifications,
          })}
        >
          <Options
            isNotifications={isNotifications}
            storyId={storyId}
            variant={isNotifications || inlineProperties ? "inline" : "sidebar"}
          />
        </Box>

        <Attachments className="mt-4" storyId={storyId} />
        <Divider className="my-6" />
        <Activities isDialog={isDialog} storyId={storyId} teamId={teamId} />
      </Container>
    </BodyContainer>
  );
};
