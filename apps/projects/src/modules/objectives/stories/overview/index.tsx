import { Container, Box, Divider, TextEditor, Menu, Button, Flex } from "ui";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";
import { useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Underline from "@tiptap/extension-underline";
import Link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import TextExtension from "@tiptap/extension-text";
import { DeleteIcon, ArrowDownIcon } from "icons";
import { BoardDividedPanel, ConfirmDialog } from "@/components/ui";
import { useDebounce } from "@/hooks";
import { useIsAdminOrOwner } from "@/hooks/owner";
import {
  useDeleteObjectiveMutation,
  useObjective,
  useUpdateObjectiveMutation,
} from "../../hooks";
import { Sidebar } from "../sidebar";
import type { ObjectiveUpdate } from "../../types";
import { Activity } from "./activity";
import { Properties } from "./properties";
import { KeyResults } from "./key-results";

const DEBOUNCE_DELAY = 700; // 700ms delay

export const Overview = () => {
  const { objectiveId } = useParams<{
    objectiveId: string;
  }>();
  const { data: objective } = useObjective(objectiveId);
  const { isAdminOrOwner } = useIsAdminOrOwner(objective?.createdBy);
  const [isOpen, setIsOpen] = useState(false);
  const updateMutation = useUpdateObjectiveMutation();
  const deleteMutation = useDeleteObjectiveMutation();

  const handleUpdate = (data: ObjectiveUpdate) => {
    updateMutation.mutate({
      objectiveId,
      data,
    });
  };

  const handleDelete = async () => {
    await deleteMutation.mutateAsync(objectiveId);
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
    editable: isAdminOrOwner,
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
    editable: isAdminOrOwner,
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
            <Flex align="center" gap={6} justify="between">
              <TextEditor
                asTitle
                className="text-4xl font-medium"
                editor={nameEditor}
              />
              {isAdminOrOwner ? (
                <Menu>
                  <Menu.Button>
                    <Button
                      className="gap-1 pl-2.5"
                      color="tertiary"
                      rightIcon={<ArrowDownIcon className="h-4" />}
                      size="sm"
                    >
                      More
                    </Button>
                  </Menu.Button>
                  <Menu.Items align="end" className="w-36">
                    <Menu.Group>
                      <Menu.Item
                        onSelect={() => {
                          setIsOpen(true);
                        }}
                      >
                        <DeleteIcon /> Delete...
                      </Menu.Item>
                    </Menu.Group>
                  </Menu.Items>
                </Menu>
              ) : null}
            </Flex>

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

      <ConfirmDialog
        confirmText={deleteMutation.isPending ? "Deleting..." : "Yes, Delete"}
        description="Are you sure you want to delete this objective? This action is irreversible."
        isOpen={isOpen}
        onClose={() => {
          setIsOpen(false);
        }}
        onConfirm={handleDelete}
        title="Delete objective"
      />
    </BoardDividedPanel>
  );
};
