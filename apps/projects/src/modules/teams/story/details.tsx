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
} from "./components";
import { DetailedStory } from "./types";

export const MainDetails = ({ story }: { story: DetailedStory }) => {
  const { title, descriptionHTML, sequenceId } = story;

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
    content: descriptionHTML,
    editable: true,
    // onUpdate: ({ editor }) => {
    //   console.log(editor.getHTML());
    // },
  });

  const titleEditor = useEditor({
    extensions: [
      Document,
      Paragraph,
      Text,
      Placeholder.configure({ placeholder: "Enter title..." }),
    ],
    content: title,
    editable: true,
    // onUpdate: ({ editor }) => {
    //   console.log(editor.getText());
    // },
  });

  return (
    <>
      <Header sequenceId={sequenceId} />
      <BodyContainer className="overflow-y-auto pb-8">
        <Container className="pt-7">
          <TextEditor
            asTitle
            className="relative -left-1 text-3xl font-medium"
            editor={titleEditor}
          />
          <TextEditor editor={descriptionEditor} />
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
