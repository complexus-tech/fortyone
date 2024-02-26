import {
  Flex,
  Text,
  Tooltip,
  Avatar,
  Button,
  Menu,
  ProgressBar,
  Box,
} from "ui";
import Link from "next/link";
import {
  CalendarIcon,
  DeleteIcon,
  MoreHorizontalIcon,
  SettingsIcon,
  StarIcon,
} from "icons";
import { RowWrapper } from "@/components/ui/row-wrapper";
import { AssigneesMenu } from "@/components/ui/issue/assignees-menu";
import { IssueStatusIcon, TableCheckbox } from "@/components/ui";

export type Project = {
  id: number;
  code: string;
  lead: string;
  name: string;
  description: string;
  date: string;
};

export const ProjectCard = ({ name }: { name: string }) => {
  return (
    <RowWrapper>
      <Flex align="center" className="relative select-none" gap={2}>
        <TableCheckbox />
        <Link className="flex items-center gap-1" href="/projects/web/issues">
          <Tooltip title="Project code: WEB">
            <Text className="w-[55px] truncate text-left" color="muted">
              WEB-01
            </Text>
          </Tooltip>
          <Text className="w-[250px] truncate hover:opacity-90">{name}</Text>
        </Link>
      </Flex>
      <Flex align="center" gap={5}>
        <Flex align="center" className="w-32" gap={2}>
          <ProgressBar className="h-2.5 w-24" progress={20} />
          <Text color="muted" fontWeight="medium">
            20%
          </Text>
        </Flex>
        <Text className="w-30 flex items-center gap-1" color="muted">
          <CalendarIcon className="h-5 w-auto" strokeWidth={2} />
          Sep 27, 2024
        </Text>
        <Text className="flex w-32 items-center gap-1" color="muted">
          <CalendarIcon className="h-5 w-auto text-primary" strokeWidth={2} />
          Sep 27, 2024
        </Text>
        <Box className="w-12">
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
        </Box>
        <Text className="w-28 text-left" color="muted">
          Sep 27, 2024
        </Text>
        <Box className="w-8">
          <Menu>
            <Menu.Button>
              <Button
                color="tertiary"
                leftIcon={<MoreHorizontalIcon className="h-5 w-auto" />}
                size="sm"
                variant="naked"
              >
                <span className="sr-only">More options</span>
              </Button>
            </Menu.Button>
            <Menu.Items align="end" className="w-64">
              <Menu.Group className="mb-3 mt-1 px-4">
                <Text color="muted" textOverflow="truncate">
                  Manage project
                </Text>
              </Menu.Group>
              <Menu.Separator className="mb-1.5" />
              <Menu.Group>
                <Menu.Item>
                  <SettingsIcon className="h-5 w-auto" />
                  Settings
                </Menu.Item>
                <Menu.Item>
                  <IssueStatusIcon className="h-[1.2rem] w-auto" />
                  Status
                </Menu.Item>
                <Menu.Item>
                  <Avatar
                    className="h-5 w-auto"
                    color="naked"
                    name="Joseph Mukorivo"
                    size="sm"
                    src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
                  />
                  Lead
                </Menu.Item>
                <Menu.Item>
                  <CalendarIcon className="h-5 w-auto" />
                  Start date
                </Menu.Item>
                <Menu.Item>
                  <CalendarIcon className="h-5 w-auto" />
                  Due date
                </Menu.Item>
                <Menu.Item>
                  <StarIcon className="h-[1.2rem] w-auto" />
                  Favourite
                </Menu.Item>
                <Menu.Item>
                  <DeleteIcon className="h-[1.2rem] w-auto" />
                  Delete
                </Menu.Item>
              </Menu.Group>
            </Menu.Items>
          </Menu>
        </Box>
      </Flex>
    </RowWrapper>
  );
};
