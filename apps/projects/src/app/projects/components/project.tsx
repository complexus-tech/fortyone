import { Flex, Text, Tooltip, Avatar, Button, Menu } from "ui";
import Link from "next/link";
import { RowWrapper } from "@/components/ui/row-wrapper";
import { AssigneesMenu } from "@/components/ui/issue/assignees-menu";
import { TableCheckbox } from "@/components/ui";
import { ProjectStatusesMenu } from "./statuses-menu";
import {
  CalendarIcon,
  EditIcon,
  MoreHorizontalIcon,
  ProjectsIcon,
} from "icons";

export const Project = ({
  name,
  description,
}: {
  name: string;
  description: string;
}) => {
  return (
    <RowWrapper>
      <Flex align="center" className="relative select-none" gap={2}>
        <TableCheckbox />
        <ProjectsIcon className="h-[1.2rem] w-auto" />
        <Link className="flex items-center gap-5" href="/projects/web/issues">
          <Text className="w-[215px] truncate hover:opacity-90">{name}</Text>
          <Text className="max-w-lg truncate hover:opacity-90" color="muted">
            {description}
          </Text>
        </Link>
      </Flex>
      <Flex align="center" gap={3}>
        <Text className="flex items-center gap-1" fontWeight="medium">
          <ProjectStatusesMenu asIcon={false} status="Testing" />
        </Text>
        <Tooltip title="Starts on">
          <Text className="flex items-center gap-1" color="muted">
            <CalendarIcon className="h-5 w-auto" strokeWidth={2} />
            Sep 27
          </Text>
        </Tooltip>
        <Tooltip title="Due on">
          <Text className="flex items-center gap-1" color="muted">
            <CalendarIcon className="h-5 w-auto text-primary" strokeWidth={2} />
            Sep 27
          </Text>
        </Tooltip>
        <AssigneesMenu>
          <AssigneesMenu.Trigger>
            <button className="flex" type="button">
              <Avatar
                name="Joseph Mukorivo"
                size="xs"
                src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
              />
            </button>
          </AssigneesMenu.Trigger>
          <AssigneesMenu.Items />
        </AssigneesMenu>
        <Menu>
          <Menu.Button>
            <Button
              color="tertiary"
              leftIcon={<MoreHorizontalIcon className="h-5 w-auto" />}
              variant="naked"
              size="sm"
            >
              <span className="sr-only">More options</span>
            </Button>
          </Menu.Button>
          <Menu.Items align="end" className="w-64">
            <Menu.Group>
              <Menu.Item>Settings</Menu.Item>
              <Menu.Item>Status</Menu.Item>
              <Menu.Item>Lead</Menu.Item>
              <Menu.Item>Start date</Menu.Item>
              <Menu.Item>Due date</Menu.Item>
              <Menu.Item>Favourite</Menu.Item>
              <Menu.Item>Delete</Menu.Item>
              <Menu.Item></Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
    </RowWrapper>
  );
};
