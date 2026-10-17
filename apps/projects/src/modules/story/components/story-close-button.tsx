import { CloseIcon } from "icons";
import { Button, Tooltip } from "ui";

export const StoryCloseButton = ({
  onClose,
  term,
}: {
  onClose: () => void;
  term: string;
}) => (
  <Tooltip title={`Close ${term}`}>
    <Button
      aria-label={`Close ${term}`}
      asIcon
      autoFocus
      color="tertiary"
      onClick={onClose}
      variant="naked"
    >
      <CloseIcon className="size-4.5" />
    </Button>
  </Tooltip>
);
