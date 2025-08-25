import {
  Container,
  Box,
  Divider,
  TextEditor,
  Menu,
  Button,
  Flex,
  Tooltip,
} from "ui";
import { useParams } from "next/navigation";
import { useEditor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Underline from "@tiptap/extension-underline";
import Link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import { DeleteIcon, ArrowDownIcon, AiIcon } from "icons";
import { useState } from "react";
import TaskItem from "@tiptap/extension-task-item";
import TaskList from "@tiptap/extension-task-list";
import { BoardDividedPanel, ConfirmDialog } from "@/components/ui";
import { useDebounce, useFeatures, useTerminology, useUserRole } from "@/hooks";
import { useIsAdminOrOwner } from "@/hooks/owner";
import { useChatContext } from "@/context/chat-context";
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

const DEBOUNCE_DELAY = 1000; // 1000ms delay

export const Overview = () => {
  const { objectiveId } = useParams<{
    objectiveId: string;
  }>();
  const features = useFeatures();
  const { getTermDisplay } = useTerminology();
  const { data: objective } = useObjective(objectiveId);
  const { isAdminOrOwner } = useIsAdminOrOwner(objective?.createdBy);
  const [isOpen, setIsOpen] = useState(false);
  const updateMutation = useUpdateObjectiveMutation();
  const deleteMutation = useDeleteObjectiveMutation();
  const { userRole } = useUserRole();
  const { openChat } = useChatContext();

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
      StarterKit,
      Underline,
      TaskList,
      TaskItem.configure({
        nested: true,
      }),
      Link.configure({
        autolink: true,
      }),
      Placeholder.configure({
        placeholder: `${getTermDisplay("objectiveTerm", { capitalize: true })} name...`,
      }),
    ],
    content: objective?.name || "",
    editable: isAdminOrOwner,
    immediatelyRender: true,
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
      Placeholder.configure({
        placeholder: `${getTermDisplay("objectiveTerm", { capitalize: true })} description...`,
      }),
    ],
    content: objective?.description || "",
    editable: isAdminOrOwner,
    immediatelyRender: true,
    onUpdate: ({ editor }) => {
      debouncedHandleUpdate({
        description: editor.getHTML(),
      });
    },
  });

  return (
    <>
      <Box className="hidden md:block">
        <BoardDividedPanel autoSaveId="teams:objectives:stories:divided-panel">
          <BoardDividedPanel.MainPanel>
            <Container className="h-[calc(100dvh-7.7rem)] overflow-y-auto pt-6">
              <Box>
                <Flex align="start" gap={6} justify="between">
                  <TextEditor
                    asTitle
                    className="text-4xl font-semibold antialiased"
                    editor={nameEditor}
                  />
                  <Flex align="center" gap={2}>
                    {userRole !== "guest" && (
                      <Tooltip
                        className="max-w-60"
                        title="Intelligently add work items to the objective"
                      >
                        <Button
                          color="tertiary"
                          leftIcon={
                            <AiIcon className="text-primary dark:text-primary" />
                          }
                          onClick={() => {
                            openChat(
                              `Suggest work items that can be added to the objective "${objective?.name}" from the team`,
                            );
                          }}
                          size="sm"
                          variant="naked"
                        >
                          Smart fill
                        </Button>
                      </Tooltip>
                    )}
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
                </Flex>

                <TextEditor
                  className="text-gray antialiased dark:text-gray-300"
                  editor={descriptionEditor}
                />
              </Box>
              <Properties />
              <Divider className="my-8" />
              <Activity />
              {features.keyResultEnabled ? <KeyResults /> : null}
            </Container>
          </BoardDividedPanel.MainPanel>
          <BoardDividedPanel.SideBar
            className="h-[calc(100dvh-7.7rem)]"
            isExpanded
          >
            <Sidebar className="h-[calc(100dvh-7.7rem)] overflow-y-auto" />
          </BoardDividedPanel.SideBar>
        </BoardDividedPanel>
      </Box>
      <Box className="md:hidden">
        <Container className="h-[calc(100dvh-7.7rem)] overflow-y-auto pt-6">
          <Box>
            <Flex align="center" gap={6} justify="between">
              <TextEditor
                asTitle
                className="text-2xl font-medium md:text-4xl"
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
          <Divider className="my-6 md:my-8" />
          <Activity />
          {features.keyResultEnabled ? <KeyResults /> : null}
        </Container>
      </Box>

      <ConfirmDialog
        confirmText={deleteMutation.isPending ? "Deleting..." : "Yes, Delete"}
        description={`Are you sure you want to delete this ${getTermDisplay("objectiveTerm")}? This action is irreversible.`}
        isOpen={isOpen}
        onClose={() => {
          setIsOpen(false);
        }}
        onConfirm={handleDelete}
        title={`Delete ${getTermDisplay("objectiveTerm")}`}
      />
    </>
  );
};
