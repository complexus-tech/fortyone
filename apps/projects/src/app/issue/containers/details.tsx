import { Box, Container, Divider, Text, TextEditor } from "ui";
import { useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Underline from "@tiptap/extension-underline";
import TaskItem from "@tiptap/extension-task-item";
import TaskList from "@tiptap/extension-task-list";
import Link from "@tiptap/extension-link";
import {
  Activities,
  Attachments,
  Reactions,
  SubissuesButton,
} from "../components";

export const MainDetails = () => {
  const content = `
          <h4>Jira Issue Description Example</h4>
          <p>This is a sample HTML text for a Jira issue description. It can include various elements such as headings, paragraphs, lists, and more.</p>
          <h4>Steps to Reproduce:</h4>
          <ol>
              <li>Open the application.</li>
              <li>Go to the settings page.</li>
              <li>Change the language to French.</li>
              <li>Save the settings.</li>
          </ol>
            `;

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
    ],
    content,
    editable: true,
  });
  return (
    <Box className="h-full overflow-y-auto border-r border-gray-50 pb-8 dark:border-dark-200">
      <Container className="pt-6">
        <Text className="mb-6" fontSize="3xl" fontWeight="medium">
          Change the color of the button to red
        </Text>
        <TextEditor editor={editor} placeholder="Test" />
        <Reactions />
        <SubissuesButton />
        <Divider className="my-4" />
        <Attachments />
        <Divider className="mb-6 mt-8" />
        <Activities />
      </Container>
    </Box>
  );
};
