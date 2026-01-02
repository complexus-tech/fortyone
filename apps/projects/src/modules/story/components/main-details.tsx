"use client";
import { Box, Container, Divider, TextEditor } from "ui";
import { useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Underline from "@tiptap/extension-underline";
import TaskItem from "@tiptap/extension-task-item";
import TaskList from "@tiptap/extension-task-list";
import Link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import TextExtension from "@tiptap/extension-text";
import { cn } from "lib";
import type { ReactNode } from "react";
import { useEffect } from "react";
import Table from "@tiptap/extension-table";
import TableCell from "@tiptap/extension-table-cell";
import TableHeader from "@tiptap/extension-table-header";
import TableRow from "@tiptap/extension-table-row";
import { useDebounce, useLocalStorage, useUserRole } from "@/hooks";
import { BodyContainer } from "@/components/shared";
import { useLinks } from "@/lib/hooks/links";
import { useUpdateStoryMutation } from "@/modules/story/hooks/update-mutation";
import {
  OptionsHeader,
  Activities,
  Attachments,
  Associations,
} from "@/modules/story/components";
import { useIsAdminOrOwner } from "@/hooks/owner";
import { useStoryById } from "../hooks/story";
import type { DetailedStory } from "../types";
import { Links } from "./links";
import { SubStories } from "./sub-stories";
import { LinksSkeleton } from "./links-skeleton";
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

  const handleUpdate = (data: Partial<DetailedStory>) => {
    updateStory({ storyId, payload: data });
  };

  const debouncedHandleUpdate = useDebounce(handleUpdate, DEBOUNCE_DELAY);

  const descriptionEditor = useEditor({
    extensions: [
      StarterKit,
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
      debouncedHandleUpdate({
        descriptionHTML: editor.getHTML(),
        description: editor.getText(),
      });
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
      debouncedHandleUpdate({
        title: editor.getText(),
      });
    },
    immediatelyRender: false,
  });

  // Sync title editor content when title changes from cache updates
  useEffect(() => {
    if (titleEditor && title && titleEditor.getText() !== title) {
      titleEditor.commands.setContent(title);
    }
    if (
      descriptionEditor &&
      descriptionHTML &&
      descriptionEditor.getHTML() !== descriptionHTML
    ) {
      descriptionEditor.commands.setContent(descriptionHTML);
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
        className={cn("pt-4 md:pt-7", {
          "md:pt-2": isDialog,
        })}
      >
        <TextEditor
          asTitle
          className="mb-8 text-3xl font-medium md:text-4xl"
          editor={titleEditor}
        />
        <TextEditor editor={descriptionEditor} />
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
        <Box className="md:hidden">
          <Divider />
          <Options isNotifications={isNotifications} storyId={storyId} />
        </Box>

        <Attachments
          className={cn(
            "mt-2.5 border-t border-border pt-2.5 d",
            {
              "mt-2 border-0":
                (isSubStoriesOpen || isLinksOpen) &&
                (subStories.length > 0 || links.length > 0),
            },
          )}
          storyId={storyId}
        />
        <Divider className="my-6" />
        <Activities storyId={storyId} teamId={teamId} />
      </Container>
    </BodyContainer>
  );
};
