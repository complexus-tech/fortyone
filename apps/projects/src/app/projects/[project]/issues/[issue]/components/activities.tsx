import { Avatar, Box, Button, Flex, Tabs, Text, TextEditor } from "ui";
import { Paperclip, History, MessageSquareText, Hourglass } from "lucide-react";
import { useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Underline from "@tiptap/extension-underline";
import TaskItem from "@tiptap/extension-task-item";
import TaskList from "@tiptap/extension-task-list";
import Link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import { Activity } from "@/components/ui";
import type { ActivityProps } from "@/components/ui";

export const Activities = () => {
  const activites: ActivityProps[] = [
    {
      id: 1,
      user: "josemukorivo",
      action: "changed status from",
      prevValue: "Todo",
      newValue: "In Progress",
      timestamp: "23 hours ago",
    },
    {
      id: 2,
      user: "janedoe",
      action: "created task",
      prevValue: "Issue 1",
      newValue: "Todo",
      timestamp: "2 days ago",
    },
    {
      id: 3,
      user: "johnsmith",
      action: "assigned task to",
      prevValue: "jackdoe",
      newValue: "janedoe",
      timestamp: "1 week ago",
    },
    {
      id: 4,
      user: "johndoe",
      action: "changed status from",
      prevValue: "In Progress",
      newValue: "Done",
      timestamp: "1 hour ago",
    },
  ];

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
  return (
    <Box>
      <Text
        as="h4"
        className="mb-6 flex items-center gap-1"
        fontWeight="medium"
      >
        <History className="h-5 w-auto" />
        Activity feed
      </Text>

      <Tabs defaultValue="all">
        <Tabs.List className="mx-0 mb-5">
          <Tabs.Tab
            className="text-[0.95rem] font-medium"
            leftIcon={
              <History className="h-[1.1rem] w-auto" strokeWidth={2.2} />
            }
            value="all"
          >
            All Activities
          </Tabs.Tab>
          <Tabs.Tab
            className="text-[0.95rem] font-medium"
            leftIcon={<Hourglass className="h-4 w-auto" strokeWidth={2.8} />}
            value="updates"
          >
            Updates
          </Tabs.Tab>
          <Tabs.Tab
            className="text-[0.95rem] font-medium"
            leftIcon={
              <MessageSquareText
                className="h-[1.1rem] w-auto"
                strokeWidth={2.2}
              />
            }
            value="comments"
          >
            Comments
          </Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="all">
          <Flex className="relative" direction="column" gap={4}>
            <Box className="pointer-events-none absolute left-4 top-0 z-0 h-[95%] border-l-[1.5px] border-gray-100 dark:border-dark-200" />
            {activites.map((activity) => (
              <Activity key={activity.id} {...activity} />
            ))}
            <Flex align="start" className="relative z-[2]">
              <Box className="pointer-events-none absolute bottom-0 left-4 h-[calc(100%-3rem)] w-1 bg-white dark:bg-dark" />
              <Box className="z-[1] mt-4 flex aspect-square items-center bg-white p-[0.3rem] dark:bg-dark">
                <Avatar name="Joseph Mukorivo" size="sm" />
              </Box>
              <Flex
                className="ml-1 mt-2 min-h-[6rem] w-full rounded-lg border border-gray-50 px-4 pb-4 text-[0.95rem] shadow-sm transition-shadow duration-200 ease-linear focus-within:shadow-lg dark:border-dark-200/80 dark:bg-dark-200/50 dark:shadow-dark-200/50"
                direction="column"
                gap={2}
                justify="between"
              >
                <TextEditor className="prose-base" editor={editor} />
                <Flex gap={1} justify="end">
                  <Button
                    className="px-3"
                    color="tertiary"
                    leftIcon={<Paperclip className="h-4 w-auto" />}
                    size="sm"
                    variant="naked"
                  >
                    <span className="sr-only">Attach files</span>
                  </Button>
                  <Button
                    className="px-3"
                    color="tertiary"
                    size="sm"
                    variant="outline"
                  >
                    Comment
                  </Button>
                </Flex>
              </Flex>
            </Flex>
          </Flex>
        </Tabs.Panel>
        <Tabs.Panel value="updates">
          <Flex className="relative" direction="column" gap={4}>
            <Box className="pointer-events-none absolute left-4 top-0 z-0 h-full border-l-[1.5px] border-gray-100 dark:border-dark-200" />
            {activites.map((activity) => (
              <Activity key={activity.id} {...activity} />
            ))}
          </Flex>
        </Tabs.Panel>
        <Tabs.Panel value="comments">
          <Text>No comments available</Text>
        </Tabs.Panel>
      </Tabs>
    </Box>
  );
};
