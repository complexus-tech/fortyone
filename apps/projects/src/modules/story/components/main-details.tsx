"use client";
import { Container, Divider, Flex, TextEditor, Text, Badge, Button } from "ui";
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
import { BodyContainer } from "@/components/shared";
import { Header, Activities, Attachments, Reactions } from ".";
import { DetailedStory } from "../types";
import { updateStoryAction } from "@/modules/story/actions/update-story";
import { toast } from "sonner";
import { useCallback, useEffect, useRef, useState } from "react";
import { useParams } from "next/navigation";
import { useStoryById } from "../hooks/story";
import { NewSubStory } from "@/components/ui/new-sub-story";
import { StoriesBoard, StoriesList } from "@/components/ui";
import { ArrowDownIcon, ArrowUpIcon, PlusIcon } from "icons";
import { useLocalStorage } from "@/hooks";
import { cn } from "lib";

const DEBOUNCE_DELAY = 500; // 500ms delay

// Custom debounce hook
const useDebounce = (callback: (...args: any[]) => void, delay = 500) => {
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

export const MainDetails = () => {
  const params = useParams<{ storyId: string }>();
  const { data } = useStoryById(params.storyId);
  const [isSubStoriesOpen, setIsSubStoriesOpen] = useLocalStorage(
    "isSubStoriesOpen",
    true,
  );
  const [isCreateSubStoryOpen, setIsCreateSubStoryOpen] = useState(false);
  const {
    id: storyId,
    title,
    descriptionHTML,
    description,
    sequenceId,
    teamId,
    deletedAt,
    subStories,
  } = data!;
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
      TextExtension,
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
          <Flex align="center" justify="between">
            <Flex align="center" gap={2}>
              <Button
                color="tertiary"
                variant="naked"
                size="sm"
                onClick={() => {
                  setIsSubStoriesOpen(!isSubStoriesOpen);
                }}
                rightIcon={
                  isSubStoriesOpen ? (
                    <ArrowDownIcon className="h-4 w-auto" />
                  ) : (
                    <ArrowUpIcon className="h-4 w-auto" />
                  )
                }
              >
                Sub stories
              </Button>
              <Badge color="tertiary" rounded="full" className="px-1.5">
                1/{subStories.length} Done
              </Badge>
            </Flex>
            <Button
              color="tertiary"
              leftIcon={<PlusIcon className="h-5 w-auto" />}
              size="sm"
              variant="naked"
              onClick={() => setIsCreateSubStoryOpen(true)}
            >
              Add Sub Story
            </Button>
          </Flex>
          <NewSubStory
            teamId={teamId}
            parentId={storyId}
            isOpen={isCreateSubStoryOpen}
            setIsOpen={setIsCreateSubStoryOpen}
          />
          {isSubStoriesOpen && subStories.length > 0 && (
            <StoriesBoard
              layout="list"
              stories={subStories}
              className="mt-2 h-auto border-t border-gray-100/60 pb-0 dark:border-dark-100/80"
              viewOptions={{
                groupBy: "None",
                orderBy: "Priority",
                showEmptyGroups: false,
                displayColumns: ["ID", "Status", "Assignee"],
              }}
            />
          )}

          <Attachments
            className={cn(
              "mt-2.5 border-t border-gray-100/60 pt-2.5 dark:border-dark-100/80",
              {
                "mt-0 border-0": isSubStoriesOpen && subStories.length > 0,
              },
            )}
          />
          <Divider className="my-6" />
          <Activities />
        </Container>
      </BodyContainer>
    </>
  );
};
