import { Container, Box, Divider, TextEditor } from "ui";
import { useParams } from "next/navigation";
import { useEffect } from "react";
import { useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Underline from "@tiptap/extension-underline";
import Link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import TextExtension from "@tiptap/extension-text";
import { BoardDividedPanel } from "@/components/ui";
import { useDebounce } from "@/hooks";
import { useObjective, useUpdateObjectiveMutation } from "../../hooks";
import { Sidebar } from "../sidebar";
import type { ObjectiveUpdate } from "../../types";
import { Activity } from "./activity";
import { Properties } from "./properties";
import { KeyResults } from "./key-results";

const DEBOUNCE_DELAY = 700; // 700ms delay

export const Overview = () => {
  const { objectiveId } = useParams<{ objectiveId: string }>();
  const { data: objective } = useObjective(objectiveId);
  const updateMutation = useUpdateObjectiveMutation();

  const handleUpdate = (data: ObjectiveUpdate) => {
    updateMutation.mutate({
      objectiveId,
      data,
    });
  };

  const debouncedHandleUpdate = useDebounce<ObjectiveUpdate>(
    handleUpdate,
    DEBOUNCE_DELAY,
  );

  const nameEditor = useEditor({
    extensions: [
      Document,
      Paragraph,
      TextExtension,
      Placeholder.configure({ placeholder: "Objective name..." }),
    ],
    content: objective?.name || "",
    editable: true,
    onUpdate: ({ editor }) => {
      debouncedHandleUpdate({
        name: editor.getText(),
      });
    },
  });

  const descriptionEditor = useEditor({
    extensions: [
      StarterKit,
      Underline,
      Link.configure({
        autolink: true,
      }),
      Placeholder.configure({ placeholder: "Objective description..." }),
    ],
    content: objective?.description || "",
    editable: true,
    onUpdate: ({ editor }) => {
      debouncedHandleUpdate({
        description: editor.getHTML(),
      });
    },
  });

  useEffect(() => {
    if (objective) {
      nameEditor?.commands.setContent(objective.name);
      descriptionEditor?.commands.setContent(objective.description);
    }
  }, [objective, nameEditor, descriptionEditor]);

  return (
    <BoardDividedPanel autoSaveId="teams:objectives:stories:divided-panel">
      <BoardDividedPanel.MainPanel>
        <Container className="h-[calc(100vh-7.7rem)] overflow-y-auto pt-6">
          <Box>
            <TextEditor
              asTitle
              className="text-4xl font-medium"
              editor={nameEditor}
            />
            <TextEditor
              className="text-gray antialiased dark:text-gray-300"
              editor={descriptionEditor}
            />
          </Box>
          <Properties />
          <Divider className="my-8" />
          <Activity />
          <KeyResults />
        </Container>
      </BoardDividedPanel.MainPanel>
      <BoardDividedPanel.SideBar className="h-[calc(100vh-7.7rem)]" isExpanded>
        <Sidebar className="h-[calc(100vh-7.7rem)] overflow-y-auto" />
      </BoardDividedPanel.SideBar>
    </BoardDividedPanel>
  );
};
