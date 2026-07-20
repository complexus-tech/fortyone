"use client";
import { Box, Container, Divider, TextEditor } from "ui";
import { useEditor } from "@tiptap/react";
import Underline from "@tiptap/extension-underline";
import { TaskItem, TaskList } from "@tiptap/extension-list";
import Link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import TextExtension from "@tiptap/extension-text";
import { cn } from "lib";
import type { ReactNode } from "react";
import { useCallback, useEffect } from "react";
import { Table } from "@tiptap/extension-table";
import TableCell from "@tiptap/extension-table-cell";
import TableHeader from "@tiptap/extension-table-header";
import TableRow from "@tiptap/extension-table-row";
import { useLocalStorage, useUserRole } from "@/hooks";
import { useDebouncedCallback } from "@/hooks/debounce";
import { BodyContainer } from "@/components/shared";
import { useLinks } from "@/lib/hooks/links";
import { createRichTextStarterKit } from "@/lib/tiptap/starter-kit";
import { useUpdateStoryMutation } from "@/modules/story/hooks/update-mutation";
import { useIsAdminOrOwner } from "@/hooks/owner";
import { useStoryById } from "../hooks/story";
import type { DetailedStory } from "../types";
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
    subStories,
    reporterId,
    associations = [],
  } = data!;
  const isDeleted = Boolean(deletedAt);
  const { isAdminOrOwner } = useIsAdminOrOwner(reporterId);
  const hasOpenSectionBeforeAttachments =
    (isSubStoriesOpen && subStories.length > 0) ||
    (isAssociationsOpen && associations.length > 0) ||
    (isLinksOpen && links.length > 0);

  const handleUpdate = useCallback(
    (data: Partial<DetailedStory>) => {
      updateStory({ storyId, payload: data });
    },
    [storyId, updateStory],
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
    extensions: [
      createRichTextStarterKit(),
      Underline,
      TaskList,
      TaskItem.configure({
        nested: true,
      }),
      Link.configure({
        autolink: true,
      }),
      Placeholder.configure({ placeholder: "Enter description..." }),
      Table.configure({
        resizable: true,
      }),
      TableRow,
      TableHeader,
      TableCell,
    ],
    content: descriptionHTML || description,
    editable: !isDeleted && userRole !== "guest",
    onUpdate: ({ editor }) => {
      debouncedDescriptionUpdate({
        descriptionHTML: editor.getHTML(),
        description: editor.getText(),
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
        title: editor.getText(),
      });
    },
    onBlur: () => {
      flushTitleUpdate();
    },
    immediatelyRender: false,
  });

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
        <OptionsHeader
          isAdminOrOwner={isAdminOrOwner}
          isNotifications={isNotifications}
          storyId={storyId}
        />
      </Box>

      <Container
        className={cn("max-w-7xl pt-4 md:pt-7", {
          "md:pt-2": isDialog,
        })}
      >
        {isNotifications ? (
          <Box className="notification-story-top-options-header relative -top-4.5 -mb-2 hidden [&>div]:h-auto [&>div]:px-0 [&>div]:pt-0">
            <OptionsHeader
              isAdminOrOwner={isAdminOrOwner}
              isNotifications={isNotifications}
              storyId={storyId}
            />
          </Box>
        ) : null}
        <GitHubSection.Banner storyId={storyId} />
        <FeedbackSection.Banner storyId={storyId} />
        <TextEditor
          asTitle
          className="text-foreground relative -left-px text-3xl font-semibold md:text-4xl"
          editor={titleEditor}
        />
        <TextEditor className="text-lg" editor={descriptionEditor} />
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

        <Attachments
          className={cn("border-border mt-2.5 border-t-[0.5px] pt-2.5", {
            "mt-0 border-0 pt-0": hasOpenSectionBeforeAttachments,
          })}
          compactHeader={hasOpenSectionBeforeAttachments}
          storyId={storyId}
        />
        <Divider className="my-6" />
        <Activities isDialog={isDialog} storyId={storyId} teamId={teamId} />
      </Container>
    </BodyContainer>
  );
};
