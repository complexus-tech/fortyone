import { HistoryIcon, PlusIcon, AiIcon } from "icons";
import { Flex, Button, BreadCrumbs } from "ui";
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
      <HeaderContainer className="justify-between">
        <Flex align="center" gap={2}>
          <MobileMenuButton />
          <BreadCrumbs
            breadCrumbs={[
              {
                name: "Plan with Maya",
                icon: <AiIcon />,
              },
            ]}
          />
        </Flex>
        <Flex align="center" gap={2}>
          <Button
            color="tertiary"
            leftIcon={<PlusIcon strokeWidth={2.8} />}
            onClick={handleNewChat}
            size="sm"
          >
            New chat
          </Button>
          <Button
            className="gap-2"
            color="tertiary"
            leftIcon={<HistoryIcon className="h-[1.15rem]" strokeWidth={2.6} />}
            onClick={() => {
              setIsHistoryOpen(true);
            }}
            size="sm"
          >
            History
          </Button>
        </Flex>
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
