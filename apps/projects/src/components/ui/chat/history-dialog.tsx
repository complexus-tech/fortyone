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
        className="mb-8 mr-6 max-w-[37rem] rounded-3xl border border-gray-100 font-medium shadow-2xl shadow-gray-200 outline-none dark:border-dark-100 dark:bg-dark-300 dark:shadow-none md:mt-auto"
        hideClose
        overlayClassName="justify-end bg-dark/5 dark:bg-dark/5"
      >
        <Dialog.Header className="flex h-16 items-center border-b border-gray-100/60 px-6 dark:border-dark-100">
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
        <Dialog.Body className="h-[calc(100dvh-8rem)] max-h-[calc(100dvh-8rem)] p-0 pt-6">
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
