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
import { AssigneesMenu } from "@/components/ui/story/assignees-menu";
import { StoryStatusIcon, TableCheckbox } from "@/components/ui";

export const EpicRow = ({ title }: { title: string }) => {
  return (
    <RowWrapper>
      <Flex align="center" className="relative select-none" gap={2}>
        <TableCheckbox />
        <Link className="flex items-center gap-1" href="/teams/web/stories">
          <Tooltip title="Objective code: WEB">
            <Text className="w-[55px] truncate text-left" color="muted">
              WEB-01
            </Text>
          </Tooltip>
          <Text className="w-[250px] truncate hover:opacity-90">{title}</Text>
        </Link>
      </Flex>
      <Flex align="center" gap={5}>
        <Flex align="center" className="w-32" gap={2}>
          <ProgressBar className="h-1.5 w-24" progress={20} />
          <Text color="muted" fontWeight="medium">
            20%
          </Text>
        </Flex>
        <Text className="w-30 flex items-center gap-1" color="muted">
          <CalendarIcon className="h-5 w-auto" strokeWidth={2} />
          Sep 27, 2024
        </Text>
        <Text className="flex w-32 items-center gap-1" color="muted">
          <CalendarIcon
            className="h-[1.15rem] w-auto text-primary"
            strokeWidth={2}
          />
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
            <AssigneesMenu.Items onAssigneeSelected={(assigneeId) => {}} />
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
              <Menu.Group>
                <Menu.Item className="py-2">
                  <SettingsIcon className="h-5 w-auto" />
                  Objective settings
                </Menu.Item>
                <Menu.Separator />
                <Menu.Item className="py-2">
                  <StoryStatusIcon className="h-5 w-auto" />
                  Status
                </Menu.Item>
                <Menu.Item className="py-2">
                  <Avatar
                    className="h-5 w-auto"
                    color="naked"
                    name="Joseph Mukorivo"
                    size="sm"
                    src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
                  />
                  Lead
                </Menu.Item>
                <Menu.Separator />
                <Menu.Item className="py-2">
                  <CalendarIcon className="h-5 w-auto" />
                  Start date
                </Menu.Item>
                <Menu.Item className="py-2">
                  <CalendarIcon className="h-5 w-auto" />
                  Due date
                </Menu.Item>
                <Menu.Item className="py-2">
                  <StarIcon className="h-5 w-auto" />
                  Favourite
                </Menu.Item>
                <Menu.Item className="py-2">
                  <DeleteIcon className="h-5 w-auto text-danger" />
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
