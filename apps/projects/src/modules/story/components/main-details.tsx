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
import TextExtension from "@tiptap/extension-text";
import { cn } from "lib";
import { useDebounce, useLocalStorage, useUserRole } from "@/hooks";
import { BodyContainer } from "@/components/shared";
import type { StoryActivity } from "@/modules/stories/types";
import { useLinks } from "@/lib/hooks/links";
import { useUpdateStoryMutation } from "@/modules/story/hooks/update-mutation";
import { useStoryById } from "../hooks/story";
import type { DetailedStory } from "../types";
import { useStoryActivities } from "../hooks/story-activities";
import { Links } from "./links";
import { SubStories } from "./sub-stories";
import { LinksSkeleton } from "./links-skeleton";
import { ActivitiesSkeleton } from "./activities-skeleton";
import { Activities, Attachments } from ".";

const DEBOUNCE_DELAY = 500; // 500ms delay

export const MainDetails = ({ storyId }: { storyId: string }) => {
  const { data } = useStoryById(storyId);
  const { data: links = [], isLoading: isLinksLoading } = useLinks(storyId);
  const { data: activities = [], isLoading: isActivitiesLoading } =
    useStoryActivities(storyId);
  const { mutate: updateStory } = useUpdateStoryMutation();
  const { userRole } = useUserRole();

  const [isSubStoriesOpen, setIsSubStoriesOpen] = useLocalStorage(
    "isSubStoriesOpen",
    true,
  );
  const [isLinksOpen, setIsLinksOpen] = useLocalStorage("isLinksOpen", true);
  const {
    title,
    descriptionHTML,
    description,
    teamId,
    deletedAt,
    subStories,
    createdAt,
    reporterId,
  } = data!;
  const isDeleted = Boolean(deletedAt);

  const createStoryActivity: StoryActivity = {
    id: "1",
    type: "create",
    createdAt,
    storyId,
    userId: reporterId,
    field: "title",
    currentValue: title,
  };
  const allActivities = reporterId
    ? [createStoryActivity, ...activities]
    : activities;

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
      Placeholder.configure({ placeholder: "Story description" }),
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

  return (
    <BodyContainer className="h-screen overflow-y-auto pb-8">
      <Container className="pt-7">
        <TextEditor
          asTitle
          className="relative -left-1 text-4xl font-medium"
          editor={titleEditor}
        />
        <TextEditor className="mt-8" editor={descriptionEditor} />
        <SubStories
          isSubStoriesOpen={isSubStoriesOpen}
          parentId={storyId}
          setIsSubStoriesOpen={setIsSubStoriesOpen}
          subStories={subStories}
          teamId={teamId}
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
        <Attachments
          className={cn(
            "mt-2.5 border-t border-gray-100/60 pt-2.5 dark:border-dark-100/80",
            {
              "mt-2 border-0":
                (isSubStoriesOpen || isLinksOpen) &&
                (subStories.length > 0 || links.length > 0),
            },
          )}
        />
        <Divider className="my-6" />
        {isActivitiesLoading ? (
          <ActivitiesSkeleton />
        ) : (
          <Activities activities={allActivities} storyId={storyId} />
        )}
      </Container>
    </BodyContainer>
  );
};
