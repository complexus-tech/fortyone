"use client";
import { useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import UnderlineExt from "@tiptap/extension-underline";
import TaskItem from "@tiptap/extension-task-item";
import TaskList from "@tiptap/extension-task-list";
import LinkExt from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import Heading from "@tiptap/extension-heading";
import TextExt from "@tiptap/extension-text";
import Table from "@tiptap/extension-table";
import TableCell from "@tiptap/extension-table-cell";
import TableHeader from "@tiptap/extension-table-header";
import TableRow from "@tiptap/extension-table-row";
import { Container, TextEditor, Divider, Flex, Button } from "ui";
import { ChatIcon, DocsIcon, EpicsIcon, SprintsIcon } from "icons";
import { BodyContainer } from "@/components/shared";
import { Header } from "./components";

export default function Page(): JSX.Element {
  const content = `
 <p>Use the Retrospective template to run a retro with your team at the end of an Iteration or a project and record action items for follow up.</p><h4>Steps to Reproduce:</h4><ol><li><p>Open the application.</p></li><li><p>Go to the settings page.</p></li><li><p>Change the language to French.</p></li><li><p>Save the settings.</p></li></ol><p><strong>This is my List</strong></p><ul data-type="taskList"><li data-checked="false"><label contenteditable="false"><input type="checkbox"><span></span></label><div><p>Item 1</p></div></li><li data-checked="true"><label contenteditable="false"><input type="checkbox" checked="checked"><span></span></label><div><p>Item 2</p></div></li><li data-checked="false"><label contenteditable="false"><input type="checkbox"><span></span></label><div><p>Item 3</p></div></li></ul></div>

  <table>
    <thead>
      <tr>
        <th>Column 1</th>
        <th>Column 2</th>
        <th>Column 3</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td>Row 1, Column 1</td>
        <td>Row 1, Column 2</td>
        <td>Row 1, Column 3</td>
      </tr>
      <tr>
        <td>Row 2, Column 1</td>
        <td>Row 2, Column 2</td>
        <td>Row 2, Column 3</td>
      </tr>
      <tr>
        <td>Row 3, Column 1</td>
        <td>Row 3, Column 2</td>
        <td>Row 3, Column 3</td>
      </tr>
    </tbody>
  </table>
            `;

  const titleEditor = useEditor({
    extensions: [
      Document,
      Paragraph,
      TextExt,
      Placeholder.configure({ placeholder: "Enter Title..." }),
    ],
    content: "Release Notes: v1.0.0",
    editable: true,
  });

  const editor = useEditor({
    extensions: [
      StarterKit,
      UnderlineExt,
      TaskList,
      Heading.configure({
        levels: [1, 2, 3, 4, 5],
      }),
      TaskItem.configure({
        nested: true,
      }),
      LinkExt.configure({
        autolink: true,
      }),
      Placeholder.configure({ placeholder: "Write here..." }),
      Table.configure({
        resizable: true,
      }),
      TableRow,
      TableHeader,
      TableCell,
    ],
    content,
    editable: true,
    autofocus: true,
  });

  return (
    <>
      <Header />
      <BodyContainer>
        <Container className="max-w-6xl pt-16">
          <Flex align="center" gap={3}>
            <DocsIcon
              className="relative -top-[1px] h-10 w-auto text-gray dark:text-gray-200"
              strokeWidth={1.8}
            />
            <TextEditor
              asTitle
              className="relative -left-1 text-[2.5rem] text-gray dark:text-gray-200"
              editor={titleEditor}
            />
          </Flex>
          <Flex className="mt-4" gap={2}>
            <Button
              color="tertiary"
              leftIcon={<SprintsIcon className="h-4 w-auto" />}
              size="sm"
            >
              Add to Sprint
            </Button>
            <Button
              color="tertiary"
              leftIcon={<EpicsIcon className="h-4 w-auto" />}
              size="sm"
            >
              Add to Epic
            </Button>
            <Button
              color="tertiary"
              leftIcon={<ChatIcon className="h-4 w-auto" />}
              size="sm"
            >
              Add comment
            </Button>
          </Flex>
          <Divider className="mt-6" />
          <TextEditor
            className="text-[1.15rem] leading-normal text-gray dark:text-gray-200/90"
            editor={editor}
          />
        </Container>
      </BodyContainer>
    </>
  );
}
