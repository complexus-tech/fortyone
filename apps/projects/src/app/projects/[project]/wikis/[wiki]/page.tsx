"use client";
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
import { Container, TextEditor } from "ui";
import { BodyContainer } from "@/components/layout";
import { Header } from "./components";

export default function Page(): JSX.Element {
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
      Placeholder.configure({ placeholder: "Issue description" }),
    ],
    content,
    editable: true,
  });

  const titleEditor = useEditor({
    extensions: [
      Document,
      Paragraph,
      Text,
      Placeholder.configure({ placeholder: "Enter Title..." }),
    ],
    content: "",
    editable: true,
  });
  return (
    <>
      <Header />
      <BodyContainer className="overflow-y-hidden">
        <Container className="pt-6">
          <TextEditor
            asTitle
            className="relative -left-1 text-3xl font-medium"
            editor={titleEditor}
          />
          <TextEditor editor={editor} />
        </Container>
      </BodyContainer>
    </>
  );
}
