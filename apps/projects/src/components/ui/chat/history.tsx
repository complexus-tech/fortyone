import { Box, Button, Flex, Skeleton, Text } from "ui";
import { ChatIcon, MoreHorizontalIcon } from "icons";
import { useAiChats } from "@/modules/ai-chats/hooks/use-ai-chats";
import { RowWrapper } from "../row-wrapper";

export const History = () => {
  const { data: chats = [], isPending } = useAiChats();
  if (isPending)
    return (
      <Box className="px-6">
        <Flex className="gap-2">
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-10 w-10 rounded-full" />
        </Flex>
      </Box>
    );
  return (
    <Box className="space-y-2 px-6">
      <Text className="mb-4 px-2">Today</Text>
      {chats.map((chat) => (
        <RowWrapper
          className="px-2 first-of-type:border-t-[0.5px] md:px-1"
          key={chat.id}
        >
          <Flex align="center" gap={2}>
            <ChatIcon className="h-4 shrink-0" />
            <Text className="line-clamp-1">{chat.title}</Text>
          </Flex>
          <Button
            asIcon
            color="tertiary"
            leftIcon={<MoreHorizontalIcon />}
            size="sm"
            variant="naked"
          >
            <span className="sr-only">Delete</span>
          </Button>
        </RowWrapper>
      ))}
    </Box>
  );
};
