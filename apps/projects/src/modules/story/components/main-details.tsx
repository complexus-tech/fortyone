"use client";
import { Container, Divider, TextEditor } from "ui";
import { useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Underline from "@tiptap/extension-underline";
import TaskItem from "@tiptap/extension-task-item";
import TaskList from "@tiptap/extension-task-list";
import Link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import Text from "@tiptap/extension-text";
import { BodyContainer } from "@/components/shared";
import {
  Header,
  Activities,
  Attachments,
  Reactions,
  SubstoriesButton,
} from ".";
import { DetailedStory } from "../types";
import { updateStoryAction } from "@/modules/story/actions/update-story";
import { toast } from "sonner";
import { useCallback, useEffect, useRef } from "react";

const DEBOUNCE_DELAY = 500; // 500ms delay

// Custom debounce hook
const useDebounce = (callback: (...args: any[]) => void, delay: number) => {
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  const debouncedCallback = useCallback(
    (...args: any[]) => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }

      timeoutRef.current = setTimeout(() => {
        callback(...args);
      }, delay);
    },
    [callback, delay],
  );

  return debouncedCallback;
};

export const MainDetails = ({ story }: { story: DetailedStory }) => {
  const {
    id: storyId,
    title,
    descriptionHTML,
    description,
    sequenceId,
    teamId,
    deletedAt,
  } = story;
  const isDeleted = !!deletedAt;

  const handleUpdate = async (data: Partial<DetailedStory>) => {
    try {
      await updateStoryAction(storyId, data);
    } catch (error) {
      console.log(error);
      toast.error("Error updating story", {
        description: "Failed to update story, reverted changes.",
      });
    }
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
      Placeholder.configure({ placeholder: "Story description" }),
    ],
    content: descriptionHTML || description,
    editable: !isDeleted,
    onUpdate: ({ editor }) => {
      debouncedHandleUpdate({
        descriptionHTML: editor.getHTML(),
        description: editor.getText(),
      });
    },
  });

  const titleEditor = useEditor({
    extensions: [
      Document,
      Paragraph,
      Text,
      Placeholder.configure({ placeholder: "Enter title..." }),
    ],
    content: title,
    editable: !isDeleted,
    onUpdate: ({ editor }) => {
      debouncedHandleUpdate({
        title: editor.getText(),
      });
    },
  });

  return (
    <>
      <Header
        sequenceId={sequenceId}
        teamId={teamId}
        isDeleted={isDeleted}
        storyId={storyId}
      />
      <BodyContainer className="overflow-y-auto pb-8">
        <Container className="pt-7 md:px-16">
          <TextEditor
            asTitle
            className="relative -left-1 text-4xl font-medium"
            editor={titleEditor}
          />
          <TextEditor className="mt-8" editor={descriptionEditor} />
          <Reactions />
          <SubstoriesButton />
          <Divider className="my-4" />
          <Attachments />
          <Divider className="my-6" />
          <Activities />
        </Container>
      </BodyContainer>
    </>
  );
};
