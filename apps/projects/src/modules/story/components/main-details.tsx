"use client";
import { Box, Container, Divider, TextEditor } from "ui";
import { useEditor } from "@tiptap/react";
import Placeholder from "@tiptap/extension-placeholder";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import TextExtension from "@tiptap/extension-text";
import { cn } from "lib";
import type { ClipboardEvent, ReactNode } from "react";
import { useCallback, useEffect } from "react";
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
import { RichTextTableMenu } from "@/lib/tiptap/rich-text-table-menu";
import { useUpdateStoryMutation } from "@/modules/story/hooks/update-mutation";
import { useStoryDescriptionMedia } from "@/modules/story/hooks/use-story-description-media";
import { useIsAdminOrOwner } from "@/hooks/owner";
import { RelatedDocuments } from "@/modules/documents/related-documents";
import { GoogleDriveFileSection } from "@/modules/google-drive";
import { useStoryById } from "../hooks/story";
import type { StoryUpdate } from "../types";
import { Activities } from "./activities";
import { Associations } from "./associations";
import { Attachments } from "./attachments";
import { Links } from "./links";
import { SubStories } from "./sub-stories";
import { StoryBanners } from "./story-banners";
import { LinksSkeleton } from "./links-skeleton";
import { FigmaSection } from "./figma-section";
import { OptionsHeader } from "./options-header";
import { Options } from "./options";
import { useFigmaDescriptionPaste } from "./use-figma-description-paste";
import { useGoogleDriveDescriptionPaste } from "./use-google-drive-description-paste";

const DEBOUNCE_DELAY = 1000; // 1000ms delay

type QueuedStoryUpdate = {
  payload: StoryUpdate;
  storyId: string;
};

export const MainDetails = ({
  inlineProperties = false,
  storyId,
  isNotifications,
  isDialog,
  mainHeader,
}: {
  inlineProperties?: boolean;
  storyId: string;
  isNotifications: boolean;
  isDialog?: boolean;
  mainHeader?: ReactNode;
}) => {
  const { data } = useStoryById(storyId);
  const { data: links = [], isLoading: isLinksLoading } = useLinks(storyId);
  const { mutate: updateStory } = useUpdateStoryMutation();
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
  const { isAdminOrOwner } = useIsAdminOrOwner(reporterId);
  const handleUpdate = useCallback(
    ({ payload, storyId: targetStoryId }: QueuedStoryUpdate) => {
      updateStory({ storyId: targetStoryId, payload });
    },
    [updateStory],
  );

  // Keep independent queues so editing one field never cancels the other field's save.
  const {
    callback: debouncedDescriptionUpdate,
    flush: flushDescriptionUpdate,
  } = useDebouncedCallback(handleUpdate, DEBOUNCE_DELAY, {
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
    editable: !isDeleted && userRole !== "guest",
    onCreate: ({ editor }) => {
      if (
        getRichTextContentType(description, descriptionHTML) !== "markdown" ||
        isDeleted ||
        userRole === "guest"
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
      const content = getPersistableRichTextContent(editor);
      debouncedDescriptionUpdate({
        payload: {
          descriptionHTML: content.contentHtml,
          description: content.contentText,
          reconcileDescriptionMedia: !hasPendingRichTextMedia(editor),
        },
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
    editable: !isDeleted && userRole !== "guest",
    onUpdate: ({ editor }) => {
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
  const handleFigmaDescriptionPaste = useFigmaDescriptionPaste({
    editor: descriptionEditor,
    storyId,
  });
  const {
    onPaste: handleGoogleDriveDescriptionPaste,
    picker: googleDriveDescriptionPicker,
  } = useGoogleDriveDescriptionPaste({ editor: descriptionEditor, storyId });
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

  // Only apply external updates while the field is idle. Replacing a focused
  // Tiptap document resets its selection and can overwrite a newer local draft.
  useEffect(() => {
    if (
      titleEditor &&
      !titleEditor.isFocused &&
      title &&
      titleEditor.getText() !== title
    ) {
      titleEditor.commands.setContent(title, { emitUpdate: false });
    }
    if (
      descriptionEditor &&
      !descriptionEditor.isFocused &&
      descriptionHTML &&
      descriptionEditor.getHTML() !== descriptionHTML
    ) {
      descriptionEditor.commands.setContent(descriptionHTML, {
        emitUpdate: false,
      });
    }
  }, [title, titleEditor, descriptionEditor, descriptionHTML]);

  return (
    <BodyContainer
      className={cn("h-full min-h-0 overflow-y-auto pb-8", {
        "h-[84.99999dvh] pb-48": isDialog,
      })}
    >
      {mainHeader}
      <Box className="md:hidden">
        <OptionsHeader isAdminOrOwner={isAdminOrOwner} storyId={storyId} />
      </Box>

      <Container
        className={cn("max-w-6xl pt-4 md:pt-7", {
          "md:px-6 md:pt-2": isDialog,
        })}
      >
        {isNotifications ? (
          <Box className="notification-story-top-options-header relative -top-4.5 -mb-2 hidden [&>div]:h-auto [&>div]:px-0 [&>div]:pt-0">
            <OptionsHeader isAdminOrOwner={isAdminOrOwner} storyId={storyId} />
          </Box>
        ) : null}
        <StoryBanners story={data!} />
        <TextEditor
          asTitle
          className="text-foreground relative -left-px mb-6 text-3xl font-semibold md:text-4xl"
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
        <TextEditor
          className="rich-document-editor text-foreground dark:text mb-10 text-[1.1rem]"
          editor={descriptionEditor}
          onPaste={handleDescriptionPaste}
        />
        {googleDriveDescriptionPicker}
        <RichTextTableMenu editor={descriptionEditor} scrollTarget={null} />
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
