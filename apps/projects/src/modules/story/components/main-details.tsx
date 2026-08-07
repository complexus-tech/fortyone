"use client";
import { Box, Container, Divider, TextEditor } from "ui";
import { useEditor } from "@tiptap/react";
import Placeholder from "@tiptap/extension-placeholder";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import TextExtension from "@tiptap/extension-text";
import { cn } from "lib";
import type { ReactNode } from "react";
import { useCallback, useEffect } from "react";
import { useLocalStorage, useUserRole } from "@/hooks";
import { useDebouncedCallback } from "@/hooks/debounce";
import { BodyContainer } from "@/components/shared";
import { useLinks } from "@/lib/hooks/links";
import { createRichTextExtensions } from "@/lib/tiptap/rich-text-extensions";
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
import { useStoryById } from "../hooks/story";
import type { StoryUpdate } from "../types";
import { Activities } from "./activities";
import { Associations } from "./associations";
import { Attachments } from "./attachments";
import { FeedbackSection } from "./feedback-section";
import { GitHubSection } from "./github-section";
import { Links } from "./links";
import { SubStories } from "./sub-stories";
import { LinksSkeleton } from "./links-skeleton";
import { OptionsHeader } from "./options-header";
import { Options } from "./options";

const DEBOUNCE_DELAY = 1000; // 1000ms delay

type QueuedStoryUpdate = {
  payload: StoryUpdate;
  storyId: string;
};

export const MainDetails = ({
  storyId,
  isNotifications,
  isDialog,
  mainHeader,
}: {
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
    editable: !isDeleted && userRole !== "guest",
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
      className={cn("h-dvh overflow-y-auto pb-8", {
        "h-[84.99999dvh] pb-48": isDialog,
      })}
    >
      {mainHeader}
      <Box className="md:hidden">
        <OptionsHeader isAdminOrOwner={isAdminOrOwner} storyId={storyId} />
      </Box>

      <Container
        className={cn("max-w-6xl pt-4 md:pt-7", {
          "md:pt-2": isDialog,
        })}
      >
        {isNotifications ? (
          <Box className="notification-story-top-options-header relative -top-4.5 -mb-2 hidden [&>div]:h-auto [&>div]:px-0 [&>div]:pt-0">
            <OptionsHeader isAdminOrOwner={isAdminOrOwner} storyId={storyId} />
          </Box>
        ) : null}
        <GitHubSection.Banner storyId={storyId} />
        <FeedbackSection.Banner storyId={storyId} />
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
          className="rich-document-editor text-foreground/90 mb-10 text-[1.1rem]"
          editor={descriptionEditor}
        />
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
        <Box
          className={cn("md:hidden", {
            "mt-4": isNotifications && isLinksOpen && links.length > 0,
            "notification-story-inline-options": isNotifications,
          })}
        >
          <Options
            isNotifications={isNotifications}
            storyId={storyId}
            variant={isNotifications ? "inline" : "sidebar"}
          />
        </Box>

        <Attachments className="mt-4" storyId={storyId} />
        <Divider className="my-6" />
        <Activities isDialog={isDialog} storyId={storyId} teamId={teamId} />
      </Container>
    </BodyContainer>
  );
};
