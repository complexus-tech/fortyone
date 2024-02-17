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
import { Box, Container, TextEditor, Text } from "ui";
import { BodyContainer } from "@/components/layout";
import { Header, Toolbar } from "./components";

export default function Page(): JSX.Element {
  const content = `
  <div class="tiptap ProseMirror" contenteditable="true" tabindex="0" translate="no" spellcheck="false"><h4>Jira Issue Description Example</h4><p>This is a sample HTML text for a Jira issue description. It can include various elements such as headings, paragraphs, lists, and more.</p><h4>Steps to Reproduce:</h4><ol><li><p>Open the application.</p></li><li><p>Go to the settings page.</p></li><li><p>Change the language to French.</p></li><li><p>Save the settings.</p></li></ol><p><strong>This is my List</strong></p><ul data-type="taskList"><li data-checked="false"><label contenteditable="false"><input type="checkbox"><span></span></label><div><p>Item 1</p></div></li><li data-checked="true"><label contenteditable="false"><input type="checkbox" checked="checked"><span></span></label><div><p>Item 2</p></div></li><li data-checked="false"><label contenteditable="false"><input type="checkbox"><span></span></label><div><p>Item 3</p></div></li></ul></div>
            `;

  const titleEditor = useEditor({
    extensions: [
      Document,
      Paragraph,
      TextExt,
      Placeholder.configure({ placeholder: "Wiki Title..." }),
    ],
    content: "Complexus app docs",
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
      Placeholder.configure({ placeholder: "Issue description" }),
    ],
    content,
    editable: true,
    autofocus: true,
  });

  return (
    <>
      <Header />
      <BodyContainer className="overflow-y-hidden">
        <Toolbar editor={editor} />
        <Box className="grid grid-cols-4">
          <Container className="col-span-3 pt-6">
            <TextEditor
              asTitle
              className="relative -left-1 text-3xl font-medium"
              editor={titleEditor}
            />
            <TextEditor editor={editor} hideBubbleMenu />
          </Container>
          <Container className="w-full pl-6 pt-8">
            <Text
              className="mb-10"
              color="muted"
              fontWeight="medium"
              transform="uppercase"
            >
              Table of contents
            </Text>

            <Text color="muted">
              The table of contents will be generated based on the headings in
              the document.
            </Text>
          </Container>
        </Box>
      </BodyContainer>
    </>
  );
}
