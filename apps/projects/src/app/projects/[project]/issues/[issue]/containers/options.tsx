import { Box, Button, Container, Divider, Text, DatePicker, Avatar } from "ui";
import type { ReactNode } from "react";
import { CalendarIcon } from "icons";
import {
  PrioritiesMenu,
  StatusesMenu,
  AssigneesMenu,
  ModulesMenu,
  SprintsMenu,
  IssueStatusIcon,
  PriorityIcon,
} from "@/components/ui";
import { Labels } from "@/components/ui/issue/labels";
import { AddLinks, OptionsHeader } from "../components";

const Option = ({ label, value }: { label: string; value: ReactNode }) => {
  return (
    <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
      <Text
        className="flex items-center gap-1 truncate"
        color="muted"
        fontWeight="medium"
      >
        {label}
      </Text>
      {value}
    </Box>
  );
};

export const Options = () => {
  return (
    <Box className="h-full overflow-y-auto bg-gray-50/20 pb-6 dark:bg-dark-200/10">
      <OptionsHeader />
      <Container className="px-8 pt-6 text-gray-300/90">
        <Text fontWeight="medium">Properties</Text>
        <Option
          label="Status"
          value={
            <StatusesMenu>
              <StatusesMenu.Trigger>
                <Button
                  color="tertiary"
                  leftIcon={<IssueStatusIcon />}
                  type="button"
                  variant="naked"
                >
                  Backlog
                </Button>
              </StatusesMenu.Trigger>
              <StatusesMenu.Items />
            </StatusesMenu>
          }
        />
        <Option
          label="Priority"
          value={
            <PrioritiesMenu>
              <PrioritiesMenu.Trigger>
                <Button
                  color="tertiary"
                  leftIcon={<PriorityIcon />}
                  type="button"
                  variant="naked"
                >
                  No Priority
                </Button>
              </PrioritiesMenu.Trigger>
              <PrioritiesMenu.Items />
            </PrioritiesMenu>
          }
        />
        <Option
          label="Assignee"
          value={
            <AssigneesMenu>
              <AssigneesMenu.Trigger>
                <Button
                  className="font-medium"
                  color="tertiary"
                  leftIcon={
                    <Avatar
                      name="Joseph Mukorivo"
                      size="xs"
                      src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
                    />
                  }
                  type="button"
                  variant="naked"
                >
                  josemukorivo
                </Button>
              </AssigneesMenu.Trigger>
              <AssigneesMenu.Items />
            </AssigneesMenu>
          }
        />
        <Option label="Labels" value={<Labels />} />
        <Option
          label="Start date"
          value={
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  color="tertiary"
                  leftIcon={<CalendarIcon className="h-[1.15rem] w-auto" />}
                  variant="naked"
                >
                  Sep 27, 2024
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar />
            </DatePicker>
          }
        />
        <Option
          label="Due date"
          value={
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  color="tertiary"
                  leftIcon={<CalendarIcon className="h-[1.15rem] w-auto" />}
                  variant="naked"
                >
                  Sep 27, 2024
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar />
            </DatePicker>
          }
        />
        <Option label="Sprint" value={<SprintsMenu />} />
        <Option label="Module" value={<ModulesMenu />} />
        <Option
          label="Parent"
          value={
            <Button color="tertiary" variant="naked">
              None
            </Button>
          }
        />
        <Option
          label="Blocking"
          value={
            <Button color="tertiary" variant="naked">
              None
            </Button>
          }
        />
        <Option
          label="Blocked by"
          value={
            <Button color="tertiary" variant="naked">
              None
            </Button>
          }
        />
        <Option
          label="Related to"
          value={
            <Button color="tertiary" variant="naked">
              None
            </Button>
          }
        />

        <Divider className="my-4" />
        <AddLinks />
      </Container>
    </Box>
  );
};
