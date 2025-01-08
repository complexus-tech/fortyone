import { Avatar, Box, Button, Flex, Tabs, Text, TextEditor } from "ui";
import { useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Underline from "@tiptap/extension-underline";
import TaskItem from "@tiptap/extension-task-item";
import TaskList from "@tiptap/extension-task-list";
import Link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import { ClockIcon, CommentIcon } from "icons";
import { Activity } from "@/components/ui";
import { StoryActivity } from "@/modules/stories/types";
import { useSession } from "next-auth/react";
import { useCommentStoryMutation } from "../hooks/comment-mutation";
import { toast } from "sonner";
import { Comment } from "@/types";
import { Comments } from "@/components/ui/comments";

export const Activities = ({
  activities,
  className,
  storyId,
}: {
  activities: StoryActivity[];
  className?: string;
  storyId: string;
}) => {
  const { data: session } = useSession();
  const { mutateAsync } = useCommentStoryMutation();
  const editor = useEditor({
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
      Placeholder.configure({ placeholder: "Leave a comment..." }),
    ],
    content: "",
    editable: true,
  });

  const handleComment = async () => {
    const comment = editor?.getHTML() ?? "";
    if (editor?.isEmpty) {
      toast.error("Comment is required", {
        description: "Please enter a comment before submitting",
      });
      return;
    }
    await mutateAsync({
      storyId,
      payload: { comment, parentId: null },
    }).then(() => {
      editor?.commands?.clearContent();
    });
  };

  return (
    <Box className={className}>
      <Text
        as="h4"
        className="mb-4 flex items-center gap-1"
        fontWeight="medium"
      >
        <ClockIcon className="h-5 w-auto" />
        Activity feed
      </Text>

      <Tabs defaultValue="comments">
        <Tabs.List className="mx-0 mb-5">
          <Tabs.Tab
            className="gap-1 px-2"
            leftIcon={
              <CommentIcon className="h-[1.1rem] w-auto" strokeWidth={2.2} />
            }
            value="comments"
          >
            Comments
          </Tabs.Tab>
          <Tabs.Tab
            className="gap-1 px-2"
            leftIcon={
              <ClockIcon className="h-[1.1rem] w-auto" strokeWidth={2.2} />
            }
            value="updates"
          >
            Updates
          </Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="updates">
          <Flex direction="column">
            {activities.length === 0 ? (
              <Text>No updates available</Text>
            ) : (
              activities.map((activity) => (
                <Activity key={activity.id} {...activity} />
              ))
            )}
          </Flex>
        </Tabs.Panel>
        <Tabs.Panel value="comments">
          <Comments storyId={storyId} />
          <Flex align="start">
            <Box className="z-[1] flex aspect-square items-center rounded-full bg-white p-[0.3rem] dark:bg-dark-300">
              <Avatar
                name={session?.user?.name ?? undefined}
                src={session?.user?.image ?? undefined}
                size="xs"
              />
            </Box>
            <Flex
              className="ml-1 min-h-[6rem] w-full rounded-lg border border-gray-50 bg-gray-50/40 px-4 pb-4 text-[0.95rem] shadow-sm transition-shadow duration-200 ease-linear focus-within:shadow-lg dark:border-dark-200/80 dark:bg-dark-200/50 dark:shadow-dark-200/50"
              direction="column"
              gap={2}
              justify="between"
            >
              <TextEditor
                className="prose-base font-normal antialiased"
                editor={editor}
              />
              <Flex justify="end">
                <Button
                  className="px-3"
                  color="tertiary"
                  disabled={editor?.isEmpty}
                  size="sm"
                  variant="outline"
                  onClick={handleComment}
                >
                  Comment
                </Button>
              </Flex>
            </Flex>
          </Flex>
        </Tabs.Panel>
      </Tabs>
    </Box>
  );
};
