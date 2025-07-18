import {
  Box,
  Button,
  Dialog,
  Flex,
  Input,
  Menu,
  Skeleton,
  Text,
  Tooltip,
} from "ui";
import { ChatIcon, DeleteIcon, EditIcon, MoreHorizontalIcon } from "icons";
import { useState } from "react";
import { useAiChats } from "@/modules/ai-chats/hooks/use-ai-chats";
import type { AiChatSession } from "@/modules/ai-chats/types";
import { useDeleteAiChat } from "@/modules/ai-chats/hooks/delete-mutation";
import { useUpdateAiChat } from "@/modules/ai-chats/hooks/update-mutation";
import { RowWrapper } from "../row-wrapper";
import { ConfirmDialog } from "../confirm-dialog";

const Row = ({
  chat,
  currentChatId,
  handleNewChat,
}: {
  chat: AiChatSession;
  currentChatId: string;
  handleNewChat: () => void;
}) => {
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [isRenameOpen, setIsRenameOpen] = useState(false);
  const [title, setTitle] = useState(chat.title);
  const deleteMutation = useDeleteAiChat();
  const updateMutation = useUpdateAiChat();
  return (
    <>
      <RowWrapper className="px-2 py-2 first-of-type:border-t-[0.5px] md:px-1">
        <Tooltip
          title={
            chat.createdAt ? new Date(chat.createdAt).toLocaleString() : null
          }
        >
          <Flex align="center" className="flex-1 cursor-pointer" gap={2}>
            <ChatIcon className="h-4 shrink-0" />
            <Text className="line-clamp-1">{chat.title}</Text>
          </Flex>
        </Tooltip>
        <Menu>
          <Menu.Button>
            <Button
              asIcon
              color="tertiary"
              leftIcon={<MoreHorizontalIcon />}
              size="sm"
              variant="naked"
            >
              <span className="sr-only">Delete</span>
            </Button>
          </Menu.Button>
          <Menu.Items align="end" className="w-28">
            <Menu.Group>
              <Menu.Item
                onSelect={() => {
                  setIsRenameOpen(true);
                }}
              >
                <EditIcon />
                Rename
              </Menu.Item>
              <Menu.Item
                className="text-danger"
                onSelect={() => {
                  setIsDeleteOpen(true);
                }}
              >
                <DeleteIcon className="h-4 text-danger dark:text-danger" />
                Delete
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </RowWrapper>

      <ConfirmDialog
        description="Are you sure you want to delete this chat?"
        isOpen={isDeleteOpen}
        onCancel={() => {
          setIsDeleteOpen(false);
        }}
        onConfirm={() => {
          deleteMutation.mutate(chat.id);
          if (currentChatId === chat.id) {
            handleNewChat();
          }
          setIsDeleteOpen(false);
        }}
        title="Delete chat"
      />

      <Dialog
        onOpenChange={() => {
          setIsRenameOpen(false);
        }}
        open={isRenameOpen}
      >
        <Dialog.Content>
          <Dialog.Header>
            <Dialog.Title className="px-6 pt-0.5 text-lg">
              Rename chat
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body>
            <Input
              autoFocus
              className="rounded-[0.6rem]"
              onChange={(e) => {
                setTitle(e.target.value);
              }}
              placeholder="Enter chat title"
              value={title}
            />
          </Dialog.Body>
          <Dialog.Footer className="justify-end gap-3 border-0 pt-2">
            <Button
              className="px-4"
              color="tertiary"
              onClick={() => {
                setIsRenameOpen(false);
              }}
            >
              Cancel
            </Button>
            <Button
              className="px-4"
              onClick={() => {
                updateMutation.mutate({
                  id: chat.id,
                  payload: {
                    title,
                  },
                });
                setIsRenameOpen(false);
              }}
            >
              Rename
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </>
  );
};

export const History = ({
  currentChatId,
  handleNewChat,
}: {
  currentChatId: string;
  handleNewChat: () => void;
}) => {
  const { data: chats = [], isPending } = useAiChats();
  if (isPending)
    return (
      <Box className="px-6">
        <Skeleton className="mb-4 h-4 w-28" />
        {new Array(4).fill(0).map((_, index) => (
          <Flex className="mb-3 gap-2" key={index}>
            <Skeleton className="h-8 w-full" />
            <Skeleton className="h-8 w-8 rounded-full" />
          </Flex>
        ))}

        <Skeleton className="mb-4 mt-10 h-4 w-28" />
        {new Array(6).fill(0).map((_, index) => (
          <Flex className="mb-3 gap-2" key={index}>
            <Skeleton className="h-8 w-full" />
            <Skeleton className="h-8 w-8 rounded-full" />
          </Flex>
        ))}
      </Box>
    );
  return (
    <Box className="px-6">
      <Text className="mb-4 px-2">Today</Text>
      {chats.map((chat) => (
        <Row
          chat={chat}
          currentChatId={currentChatId}
          handleNewChat={handleNewChat}
          key={chat.id}
        />
      ))}
    </Box>
  );
};
