import { Badge, Button, Flex, Menu } from "ui";
import {
  ArrowDown2Icon,
  ArrowUp2Icon,
  DeleteIcon,
  MoreHorizontalIcon,
} from "icons";

export const ChecklistHeader = ({
  id,
  isOpen,
  itemCount,
  completedCount,
  canEdit,
  isPending,
  onToggle,
  onDelete,
}: {
  id: string;
  isOpen: boolean;
  itemCount: number;
  completedCount: number;
  canEdit: boolean;
  isPending: boolean;
  onToggle: () => void;
  onDelete: () => void;
}) => (
  <Flex
    align="center"
    className="border-border border-b-[0.5px] pb-1.5"
    justify="between"
  >
    <Flex align="center" gap={2}>
      <h2 id={id}>
        <Button
          aria-expanded={isOpen}
          className="font-semibold"
          color="tertiary"
          onClick={() => {
            onToggle();
          }}
          rightIcon={
            isOpen ? (
              <ArrowDown2Icon className="h-4 w-auto" />
            ) : (
              <ArrowUp2Icon className="h-4 w-auto" />
            )
          }
          size="sm"
          variant="naked"
        >
          Checklist{itemCount ? ` (${itemCount})` : ""}
        </Button>
      </h2>
      {itemCount ? (
        <Badge
          aria-label={`${completedCount} of ${itemCount} checklist items complete`}
          className="px-1.5"
          color="tertiary"
        >
          {completedCount}/{itemCount} Done
        </Badge>
      ) : null}
    </Flex>
    {canEdit && itemCount ? (
      <Menu>
        <Menu.Button>
          <Button
            aria-label="Checklist actions"
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
            <Menu.Item
              className="text-danger"
              onSelect={() => {
                onDelete();
              }}
            >
              <DeleteIcon /> Delete checklist
            </Menu.Item>
          </Menu.Group>
        </Menu.Items>
      </Menu>
    ) : null}
  </Flex>
);
