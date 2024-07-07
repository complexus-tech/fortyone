import { Avatar, Button, Flex, Text, Tooltip, Menu } from "ui";
import Link from "next/link";
import {
  DeleteIcon,
  DocsIcon,
  EditIcon,
  LinkIcon,
  MoreHorizontalIcon,
} from "icons";
import { RowWrapper } from "@/components/ui/row-wrapper";
import { TableCheckbox } from "@/components/ui";

export const DocumentRow = ({ title }: { title: string }) => {
  return (
    <RowWrapper>
      <Flex align="center" className="relative w-full select-none" gap={2}>
        <TableCheckbox />
        <DocsIcon className="h-[1.3rem] w-auto shrink-0" />
        <Link
          className="line-clamp-1 items-center gap-5 text-[1.05rem]"
          href="/teams/web/documents/test1"
        >
          {title}
        </Link>
      </Flex>
      <Flex align="center" className="shrink-0" gap={16}>
        <Tooltip title="Last updated">
          <Text color="muted">Sep 27, 2021</Text>
        </Tooltip>
        <Tooltip title="Creaed by">
          <Avatar
            name="John Doe"
            size="sm"
            src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
          />
        </Tooltip>
        <Menu>
          <Menu.Button>
            <Button
              className="aspect-square"
              color="tertiary"
              leftIcon={<MoreHorizontalIcon className="h-5 w-auto" />}
              size="sm"
              variant="naked"
            >
              <span className="sr-only">More actions</span>
            </Button>
          </Menu.Button>
          <Menu.Items align="end" className="w-48">
            <Menu.Group>
              <Menu.Item>
                <EditIcon className="h-4 w-auto" />
                Edit
              </Menu.Item>
              <Menu.Item>
                <DeleteIcon className="h-4 w-auto text-danger" />
                Delete
              </Menu.Item>
              <Menu.Item>
                <LinkIcon className="h-4 w-auto" />
                Copy link
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
    </RowWrapper>
  );
};
