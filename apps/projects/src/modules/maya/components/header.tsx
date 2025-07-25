import { HistoryIcon, PlusIcon } from "icons";
import { Flex, Button } from "ui";
import { useState } from "react";
import { HistoryDialog } from "@/components/ui/chat/history-dialog";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";

type MayaHeaderProps = {
  currentChatId: string;
  handleNewChat: () => void;
  handleChatSelect: (chatId: string) => void;
};

export const Header = ({
  currentChatId,
  handleNewChat,
  handleChatSelect,
}: MayaHeaderProps) => {
  const [isHistoryOpen, setIsHistoryOpen] = useState(false);

  return (
    <>
      <HeaderContainer className="justify-between border-b-0">
        <Flex align="center" gap={2}>
          <MobileMenuButton />
          <Button
            color="tertiary"
            leftIcon={<PlusIcon strokeWidth={2.8} />}
            onClick={handleNewChat}
            variant="naked"
          >
            New chat
          </Button>
        </Flex>
        <Button
          className="gap-2"
          color="tertiary"
          leftIcon={<HistoryIcon className="h-[1.15rem]" strokeWidth={2.6} />}
          onClick={() => {
            setIsHistoryOpen(true);
          }}
          variant="naked"
        >
          History
        </Button>
      </HeaderContainer>
      <HistoryDialog
        currentChatId={currentChatId}
        handleChatSelect={handleChatSelect}
        handleNewChat={handleNewChat}
        isOpen={isHistoryOpen}
        setIsOpen={setIsHistoryOpen}
      />
    </>
  );
};
