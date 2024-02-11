import { Box, Button, Container, Divider, Flex, Text, Tooltip } from "ui";
import {
  Link2,
  Clipboard,
  Trash2,
  TimerReset,
  Calendar,
  Plus,
  Layers3,
  ShieldCheck,
  ShieldEllipsis,
  Network,
  Disc,
  BarChart,
  User,
  Tags,
  GitCompareArrows,
} from "lucide-react";
import { IssueStatusIcon, PriorityIcon } from "@/components/ui";

export const Options = () => {
  return (
    <Box className="h-full overflow-y-auto bg-gray-50/20 pb-6 dark:bg-dark-200/40">
      <Box className="flex h-16 items-center border-b border-gray-50 dark:border-dark-200">
        <Container className="flex w-full items-center justify-between">
          <Text color="muted" fontWeight="medium">
            COMP-13
          </Text>
          <Flex gap={2}>
            <Tooltip title="Copy issue link">
              <Button
                color="tertiary"
                leftIcon={<Link2 className="h-5 w-auto" strokeWidth={2.5} />}
                variant="naked"
              >
                <span className="sr-only">Copy issue link</span>
              </Button>
            </Tooltip>
            <Tooltip title="Copy issue id">
              <Button
                color="tertiary"
                leftIcon={<Clipboard className="h-5 w-auto" />}
                variant="naked"
              >
                <span className="sr-only">Copy issue id</span>
              </Button>
            </Tooltip>
            <Tooltip title="Delete issue">
              <Button
                color="danger"
                leftIcon={<Trash2 className="h-5 w-auto" />}
                variant="naked"
              >
                <span className="sr-only">Delete issue</span>
              </Button>
            </Tooltip>
          </Flex>
        </Container>
      </Box>
      <Container className="pt-6 text-gray-300/90">
        <Text fontWeight="medium">Properties</Text>
        <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
          <Text
            className="flex items-center gap-1 truncate"
            color="muted"
            fontWeight="medium"
          >
            <Disc className="h-5 w-auto" />
            Status
          </Text>
          <Button className="px-4" color="tertiary" variant="naked">
            <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
              <IssueStatusIcon status="In Progress" />
              <Text>In Progress</Text>
            </Box>
          </Button>
        </Box>
        <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
          <Text
            className="flex items-center gap-1 truncate"
            color="muted"
            fontWeight="medium"
          >
            <BarChart className="h-5 w-auto" />
            Priority
          </Text>
          <Button className="px-4" color="tertiary" variant="naked">
            <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
              <PriorityIcon priority="High" />
              <Text>High</Text>
            </Box>
          </Button>
        </Box>

        <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
          <Text
            className="flex items-center gap-1 truncate"
            color="muted"
            fontWeight="medium"
          >
            <User className="h-5 w-auto" />
            Assignee
          </Text>
          <Button className="px-4" color="tertiary" variant="naked">
            <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
              <IssueStatusIcon status="Done" />
              <Text>Backlog</Text>
            </Box>
          </Button>
        </Box>

        <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
          <Text
            className="flex items-center gap-1 truncate"
            color="muted"
            fontWeight="medium"
          >
            <Tags className="h-5 w-auto" />
            Labels
          </Text>
          <Button className="px-4" color="tertiary" variant="naked">
            <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
              <IssueStatusIcon status="Backlog" />
              <Text>Backlog</Text>
            </Box>
          </Button>
        </Box>
        <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
          <Text
            className="flex items-center gap-1 truncate"
            color="muted"
            fontWeight="medium"
          >
            <Calendar className="h-5 w-auto" />
            Start date
          </Text>
          <Button className="px-4" color="tertiary" variant="naked">
            <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
              <IssueStatusIcon status="Backlog" />
              <Text>Backlog</Text>
            </Box>
          </Button>
        </Box>
        <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
          <Text
            className="flex items-center gap-1 truncate"
            color="muted"
            fontWeight="medium"
          >
            <Calendar className="h-5 w-auto" />
            Due date
          </Text>
          <Button className="px-4" color="tertiary" variant="naked">
            <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
              <IssueStatusIcon status="Backlog" />
              <Text>Backlog</Text>
            </Box>
          </Button>
        </Box>
        <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
          <Text
            className="flex items-center gap-1 truncate"
            color="muted"
            fontWeight="medium"
          >
            <TimerReset className="h-5 w-auto" />
            Sprint
          </Text>
          <Button className="px-4" color="tertiary" variant="naked">
            <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
              <IssueStatusIcon status="Backlog" />
              <Text>Backlog</Text>
            </Box>
          </Button>
        </Box>
        <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
          <Text
            className="flex items-center gap-1 truncate"
            color="muted"
            fontWeight="medium"
          >
            <Layers3 className="h-5 w-auto" />
            Modules
          </Text>
          <Button className="px-4" color="tertiary" variant="naked">
            <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
              <IssueStatusIcon status="Backlog" />
              <Text>Backlog</Text>
            </Box>
          </Button>
        </Box>
        <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
          <Text
            className="flex items-center gap-1 truncate"
            color="muted"
            fontWeight="medium"
          >
            <Network className="h-5 w-auto" />
            Parent
          </Text>
          <Button className="px-4" color="tertiary" variant="naked">
            <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
              <IssueStatusIcon status="Backlog" />
              <Text>Backlog</Text>
            </Box>
          </Button>
        </Box>
        <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
          <Text
            className="flex items-center gap-1 truncate"
            color="muted"
            fontWeight="medium"
          >
            <ShieldCheck className="h-5 w-auto" />
            Blocking
          </Text>
          <Button className="px-4" color="tertiary" variant="naked">
            <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
              <IssueStatusIcon status="Backlog" />
              <Text>Backlog</Text>
            </Box>
          </Button>
        </Box>
        <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
          <Text
            className="flex items-center gap-1 truncate"
            color="muted"
            fontWeight="medium"
          >
            <ShieldEllipsis className="h-5 w-auto" />
            Blocked by
          </Text>
          <Button className="px-4" color="tertiary" variant="naked">
            <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
              <IssueStatusIcon status="Backlog" />
              <Text>Backlog</Text>
            </Box>
          </Button>
        </Box>
        <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
          <Text
            className="flex items-center gap-1 truncate"
            color="muted"
            fontWeight="medium"
          >
            <GitCompareArrows className="h-5 w-auto" />
            Related to
          </Text>
          <Button className="px-4" color="tertiary" variant="naked">
            <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
              <IssueStatusIcon status="Backlog" />
              <Text>Backlog</Text>
            </Box>
          </Button>
        </Box>

        <Divider className="my-4" />
        <Flex align="center">
          <Text fontWeight="medium">Links</Text>
          <Button
            className="ml-auto"
            color="tertiary"
            leftIcon={<Plus className="h-5 w-auto" strokeWidth={2} />}
            size="sm"
            variant="outline"
          >
            Add
          </Button>
        </Flex>
      </Container>
    </Box>
  );
};
