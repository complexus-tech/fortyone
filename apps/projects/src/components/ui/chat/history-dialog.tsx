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
        className="mb-8 mr-6 max-w-148 rounded-3xl border border-border font-medium shadow-2xl shadow-shadow outline-none bg-surface-elevated md:mt-auto"
        hideClose
        overlayClassName="justify-end bg-black/5"
      >
        <Dialog.Header className="flex h-16 items-center border-b border-border px-6">
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
                className="text-foreground"
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
