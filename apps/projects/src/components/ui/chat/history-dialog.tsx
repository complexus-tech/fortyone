import { Button, Dialog } from "ui";
import { CloseIcon } from "icons";
import { History } from "./history";

type HistoryDialogProps = {
  isOpen: boolean;
  setIsOpen: (isOpen: boolean) => void;
  currentChatId: string;
  handleChatSelect: (chatId: string) => void;
  handleNewChat: () => void;
};

export const HistoryDialog = ({
  isOpen,
  setIsOpen,
  currentChatId,
  handleChatSelect,
  handleNewChat,
}: HistoryDialogProps) => {
  const handleChatSelectAndClose = (chatId: string) => {
    handleChatSelect(chatId);
    setIsOpen(false);
  };

  return (
    <Dialog onOpenChange={setIsOpen} open={isOpen}>
      <Dialog.Content
        className="mr-0 max-w-[32rem] rounded-none border-0 border-l-[0.5px] font-medium outline-none dark:bg-dark-300 md:mt-auto"
        hideClose
        overlayClassName="justify-end backdrop-blur-sm"
      >
        <Dialog.Header className="flex h-16 items-center border-b-[0.5px] border-gray-100 px-6 dark:border-dark-50">
          <Dialog.Title className="flex w-full items-center gap-2 text-lg font-semibold">
            <Button
              asIcon
              color="tertiary"
              onClick={() => {
                setIsOpen(false);
              }}
              size="sm"
              variant="naked"
            >
              <CloseIcon
                className="text-dark dark:text-gray-200"
                strokeWidth={2.8}
              />
            </Button>
            History
          </Dialog.Title>
        </Dialog.Header>
        <Dialog.Description className="sr-only">
          View and manage your chat history.
        </Dialog.Description>
        <Dialog.Body className="h-[calc(100dvh-4rem)] max-h-[calc(100dvh-4rem)] p-0 pt-6">
          <History
            currentChatId={currentChatId}
            handleChatSelect={handleChatSelectAndClose}
            handleNewChat={handleNewChat}
          />
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
};
