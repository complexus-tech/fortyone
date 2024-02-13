import type { ReactNode } from "react";
import { Box, Button, Container, Divider, Text, DatePicker } from "ui";
import { Calendar, Plus, CalendarCheck2 } from "lucide-react";
import { PrioritiesMenu, StatusesMenu, AssigneesMenu } from "@/components/ui";
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
    <Box className="h-full overflow-y-auto bg-gray-50/20 pb-6 dark:bg-dark-200/40">
      <OptionsHeader />
      <Container className="px-8 pt-6 text-gray-300/90">
        <Text fontWeight="medium">Properties</Text>
        <Option
          label="Status"
          value={<StatusesMenu asIcon={false} status="In Progress" />}
        />
        <Option
          label="Priority"
          value={<PrioritiesMenu asIcon={false} priority="Urgent" />}
        />
        <Option
          label="Assignee"
          value={<AssigneesMenu asIcon={false} user="josemukorivo" />}
        />
        <Option label="Labels" value={<Labels />} />
        <Option
          label="Start date"
          value={
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  color="tertiary"
                  leftIcon={<Calendar className="h-[1.15rem] w-auto" />}
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
                  leftIcon={<CalendarCheck2 className="h-[1.15rem] w-auto" />}
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
          label="Sprint"
          value={
            <Button
              color="tertiary"
              leftIcon={<Plus className="h-5 w-auto" />}
              variant="naked"
            >
              Add sprint
            </Button>
          }
        />
        <Option
          label="Module"
          value={
            <Button
              color="tertiary"
              leftIcon={<Plus className="h-5 w-auto" />}
              variant="naked"
            >
              Add module
            </Button>
          }
        />
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
