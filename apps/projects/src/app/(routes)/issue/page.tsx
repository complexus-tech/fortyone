import { TbMoodPlus } from "react-icons/tb";
import {
  Badge,
  Box,
  BreadCrumbs,
  Button,
  Container,
  Divider,
  Flex,
  Text,
  Tooltip,
} from "ui";
import {
  Bell,
  Link2,
  Clipboard,
  Trash2,
  ChevronUp,
  ChevronDown,
  Star,
  TimerReset,
  Calendar,
  Plus,
} from "lucide-react";
import { BodyContainer, HeaderContainer } from "../../../components/shared";
import {
  IssueStatusIcon,
  NewIssueButton,
  PriorityIcon,
} from "../../../components/ui";

export default function Page(): JSX.Element {
  return (
    <>
      <HeaderContainer>
        <Flex align="center" className="w-full" justify="between">
          <BreadCrumbs
            breadCrumbs={[
              { name: "Complexus" },
              { name: "Web design" },
              {
                name: "COM-12",
                icon: <Link2 className="h-5 w-auto" strokeWidth={2} />,
              },
            ]}
          />
          <Flex align="center" gap={2} justify="between">
            <Text className="mr-2">
              2 /{" "}
              <Text as="span" color="muted">
                8
              </Text>
            </Text>
            <Button color="tertiary" size="sm">
              <ChevronUp className="h-4 w-auto" />
            </Button>
            <Button className="mr-2" color="tertiary" disabled size="sm">
              <ChevronDown className="h-4 w-auto" />
            </Button>
            <Button
              className="px-3"
              color="tertiary"
              leftIcon={<Star className="h-4 w-auto" />}
              size="sm"
            >
              Favourite
            </Button>
            <Button
              className="px-2"
              leftIcon={<Bell className="h-4 w-auto" />}
              size="sm"
            >
              Subscribe
            </Button>
          </Flex>
        </Flex>
      </HeaderContainer>
      <BodyContainer className="grid grid-cols-[auto_380px]">
        <Box className="border-r border-gray-100 dark:border-dark-100">
          <Container className="pt-6">
            <Text
              as="h3"
              className="flex items-center gap-2"
              color="muted"
              fontSize="xl"
              fontWeight="medium"
            >
              <Link2 className="h-5 w-auto" strokeWidth={2.5} />
              COMP-13
            </Text>
            <Text className="mb-6 mt-3" fontSize="3xl">
              Change the color of the button to red
            </Text>
            <Text className="leading-7" color="muted" fontSize="lg">
              Change the color of the button to red. This will hold the
              description of the issue. It will be a long one so we can test the
              overflow of the container.
              <br />
              <br />
              Use the color palette from the design system to change the color
              of the button. The pallette is available in the design system
              documentation. Use the color palette from the design system to
              change the color of the button. The pallette is available in the
              design system documentation.
            </Text>

            <Flex className="mb-4 mt-6" gap={1}>
              <Badge color="tertiary" rounded="lg" size="lg" variant="outline">
                👍🏼 1
              </Badge>
              <Badge color="tertiary" rounded="lg" size="lg" variant="outline">
                🇿🇼 3
              </Badge>
              <Badge color="tertiary" rounded="lg" size="lg" variant="outline">
                👌 2
              </Badge>
              <Badge
                className="px-2"
                color="tertiary"
                rounded="lg"
                size="lg"
                variant="outline"
              >
                <TbMoodPlus className="h-5 w-auto opacity-80" />
              </Badge>
            </Flex>
            <Flex justify="end">
              <Button
                color="tertiary"
                leftIcon={<Plus className="h-5 w-auto" />}
                size="sm"
                variant="naked"
              >
                Add sub issue
              </Button>
            </Flex>

            <Divider className="my-4" />

            <Box>
              <Text as="h4" fontWeight="medium">
                Attachements
              </Text>
              <Box className="mb-4 mt-3 flex h-24 cursor-pointer items-center justify-center rounded-xl border-[1.5px] border-dashed bg-gray-100 dark:border-dark-50 dark:bg-dark-200/50">
                <Text color="muted">Click or drag files here</Text>
              </Box>

              <Box className="grid grid-cols-5 gap-4">
                <Box className="h-28 rounded-xl bg-gray-100 dark:bg-dark-200" />
                <Box className="h-28 rounded-xl bg-gray-100 dark:bg-dark-200" />
                <Box className="h-28 rounded-xl bg-gray-100 dark:bg-dark-200" />

                <Box className="h-28 rounded-xl bg-gray-100 dark:bg-dark-200" />
                <Box className="h-28 rounded-xl bg-gray-100 dark:bg-dark-200" />
                <Box className="h-28 rounded-xl bg-gray-100 dark:bg-dark-200" />
              </Box>
            </Box>
          </Container>
        </Box>
        <Box className="bg-gray-50/20 dark:bg-dark-200/40">
          <Box className="flex h-16 items-center border-b border-gray-100 dark:border-dark-100">
            <Container className="flex w-full items-center justify-between">
              <Text color="muted" fontWeight="medium">
                COMP-13
              </Text>
              <Flex gap={2}>
                <Tooltip title="Copy issue link">
                  <Button
                    color="tertiary"
                    leftIcon={
                      <Link2 className="h-5 w-auto" strokeWidth={2.5} />
                    }
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
          <Container className="pt-4">
            <Text fontWeight="medium">Properties</Text>
            <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
              <Text className="truncate" color="muted" fontWeight="medium">
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
              <Text className="truncate" color="muted" fontWeight="medium">
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
              <Text className="truncate" color="muted" fontWeight="medium">
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
              <Text className="truncate" color="muted" fontWeight="medium">
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
              <Text className="truncate" color="muted" fontWeight="medium">
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
              <Text className="truncate" color="muted" fontWeight="medium">
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
              <Text className="truncate" color="muted" fontWeight="medium">
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
              <Text className="truncate" color="muted" fontWeight="medium">
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
              <Text className="truncate" color="muted" fontWeight="medium">
                Relates to
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
              >
                Add
              </Button>
            </Flex>
          </Container>
        </Box>
      </BodyContainer>
    </>
  );
}
