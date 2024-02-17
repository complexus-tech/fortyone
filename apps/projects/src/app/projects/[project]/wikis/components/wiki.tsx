import { Avatar, Button, Flex, Text, Tooltip, Menu } from "ui";
import Link from "next/link";
import {
  LockKeyhole,
  MoreVertical,
  NotepadText,
  Star,
  Pencil,
  Trash2,
  Link2,
} from "lucide-react";
import { RowWrapper } from "@/components/ui/row-wrapper";
import { TableCheckbox } from "@/components/ui";

export const Wiki = ({ name }: { name: string }) => {
  return (
    <RowWrapper>
      <Flex align="center" className="relative select-none" gap={2}>
        <TableCheckbox />
        <NotepadText className="h-[1.2rem] w-auto" />
        <Link
          className="flex items-center gap-5"
          href="/projects/web/wikis/test1"
        >
          <Text className="w-[215px] truncate hover:opacity-90">{name}</Text>
        </Link>
      </Flex>
      <Flex align="center" gap={3}>
        <Tooltip title="Unlock">
          <Button
            className="aspect-square"
            color="tertiary"
            leftIcon={<LockKeyhole className="h-5 w-auto" />}
            size="sm"
            variant="naked"
          >
            <span className="sr-only">Favourite</span>
          </Button>
        </Tooltip>
        <Tooltip title="Favourite">
          <Button
            className="aspect-square"
            color="tertiary"
            leftIcon={<Star className="h-5 w-auto" />}
            size="sm"
            variant="naked"
          >
            <span className="sr-only">Favourite</span>
          </Button>
        </Tooltip>
        <Tooltip title="Last updated">
          <Text color="muted">2 days ago</Text>
        </Tooltip>
        <Tooltip title="Creaed by">
          <Avatar name="John Doe" size="sm" />
        </Tooltip>
        <Menu>
          <Menu.Button>
            <Button
              className="aspect-square"
              color="tertiary"
              leftIcon={<MoreVertical className="h-5 w-auto" />}
              size="sm"
              variant="naked"
            >
              <span className="sr-only">More actions</span>
            </Button>
          </Menu.Button>
          <Menu.Items align="end" className="w-48">
            <Menu.Group>
              <Menu.Item>
                <Pencil className="h-4 w-auto" />
                Edit
              </Menu.Item>
              <Menu.Item>
                <Trash2 className="h-4 w-auto text-danger" />
                Delete
              </Menu.Item>
              <Menu.Item>
                <Link2 className="h-4 w-auto" />
                Copy link
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
    </RowWrapper>
  );
};
