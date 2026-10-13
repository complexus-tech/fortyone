import { Button, Menu } from "ui";
import { DeleteIcon, EditIcon, MoreHorizontalIcon, StoryIcon } from "icons";

export const ChecklistItemActions = ({
  title,
  isConverted,
  isPending,
  storyTerm,
  onEdit,
  onConvert,
  onDelete,
}: {
  title: string;
  isConverted: boolean;
  isPending: boolean;
  storyTerm: string;
  onEdit: () => void;
  onConvert: () => void;
  onDelete: () => void;
}) => (
  <Menu>
    <Menu.Button>
      <Button
        aria-label={`Actions for ${title}`}
        asIcon
        color="tertiary"
        disabled={isPending}
        size="sm"
        variant="naked"
      >
        <MoreHorizontalIcon />
      </Button>
    </Menu.Button>
    <Menu.Items align="end">
      <Menu.Group>
        {!isConverted ? (
          <>
            <Menu.Item onSelect={onEdit}>
              <EditIcon /> Edit item
            </Menu.Item>
            <Menu.Item onSelect={onConvert}>
              <StoryIcon /> Convert to {storyTerm}
            </Menu.Item>
            <Menu.Separator />
          </>
        ) : null}
        <Menu.Item className="text-danger" onSelect={onDelete}>
          <DeleteIcon /> Delete item
        </Menu.Item>
      </Menu.Group>
    </Menu.Items>
  </Menu>
);
