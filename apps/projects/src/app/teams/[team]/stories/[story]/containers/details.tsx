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
import {
  Header,
  Activities,
  Attachments,
  Reactions,
  SubstoriesButton,
} from "@/app/teams/[team]/stories/[story]/components";
import { BodyContainer } from "@/components/shared";

export const MainDetails = () => {
  const content = `
  <h3>We need the ability to do the following.</h3>
  <p> - Allow submission , editing and updating of third party bank details and to be displayed at Finance before disbursement</p>
   <p> - Ability to Reject and Approve documents at the BackOffice (Credit Process) </p>
    <p> - Ability to view the documents and details submitted by the third party </p>
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
      Placeholder.configure({ placeholder: "Story description" }),
    ],
    content,
    editable: true,
  });

  const titleEditor = useEditor({
    extensions: [
      Document,
      Paragraph,
      Text,
      Placeholder.configure({ placeholder: "Enter title..." }),
    ],
    content: "BackOffice - Third Party Bank Details and Documents Validations",
    editable: true,
  });
  return (
    <>
      <Header />
      <BodyContainer className="overflow-y-auto pb-8">
        <Container className="pt-7">
          <TextEditor
            asTitle
            className="relative -left-1 text-3xl font-medium"
            editor={titleEditor}
          />
          <TextEditor editor={editor} />
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
