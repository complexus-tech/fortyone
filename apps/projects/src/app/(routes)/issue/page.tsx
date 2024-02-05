import { TbClipboardCopy, TbLink, TbMoodPlus, TbTrash } from "react-icons/tb";
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
              { name: "COM-12" },
            ]}
          />
          <NewIssueButton />
        </Flex>
      </HeaderContainer>
      <BodyContainer className="grid grid-cols-[auto_400px]">
        <Box className="border-r border-gray-100 dark:border-dark-100">
          <Container className="pt-6">
            <Text fontSize="3xl" className="mb-6">
              This should be an issue title
            </Text>
            <Text color="muted" fontSize="lg" className="leading-7">
              This will hold the description of the issue. It will be a long one
              so we can test the overflow of the container. This will hold the
              description of the issue. It will be a long one. The quick brown
              fox jumped over the lazy dog. This will hold the description of
              the issue. It will be a long one. The quick brown fox jumped over
              the lazy dog. This will hold the description of the issue.
            </Text>
            <Flex gap={1} className="mt-4">
              <Badge
                rounded="full"
                size="lg"
                color="tertiary"
                variant="outline"
              >
                👍🏼 1
              </Badge>
              <Badge
                rounded="full"
                size="lg"
                color="tertiary"
                variant="outline"
              >
                🇿🇼 3
              </Badge>
              <Badge
                rounded="full"
                size="lg"
                color="tertiary"
                variant="outline"
              >
                👌 2
              </Badge>
              <Badge
                className="px-2"
                color="tertiary"
                rounded="full"
                size="lg"
                variant="outline"
              >
                <TbMoodPlus className="h-5 w-auto opacity-80" />
              </Badge>
            </Flex>
          </Container>
        </Box>
        <Box className="bg-gray-50/20 dark:bg-dark-200/40">
          <Box className="flex h-16 items-center border-b border-gray-100 dark:border-dark-100">
            <Container className="flex w-full items-center justify-between">
              <Text fontWeight="medium" color="muted">
                COMP-13
              </Text>
              <Flex gap={2}>
                <Tooltip title="Copy issue link">
                  <Button
                    color="tertiary"
                    leftIcon={<TbLink className="h-5 w-auto" />}
                    variant="naked"
                  >
                    <span className="sr-only">Copy issue link</span>
                  </Button>
                </Tooltip>
                <Tooltip title="Copy issue id">
                  <Button
                    color="tertiary"
                    leftIcon={<TbClipboardCopy className="h-5 w-auto" />}
                    variant="naked"
                  >
                    <span className="sr-only">Copy issue id</span>
                  </Button>
                </Tooltip>
                <Tooltip title="Delete issue">
                  <Button
                    variant="naked"
                    color="danger"
                    leftIcon={<TbTrash className="h-5 w-auto" />}
                  >
                    <span className="sr-only">Delete issue</span>
                  </Button>
                </Tooltip>
              </Flex>
            </Container>
          </Box>
          <Container className="pt-6">
            <Box className="my-4 grid grid-cols-[8rem_auto] items-center gap-3">
              <Text className="truncate" fontWeight="medium" color="muted">
                Status
              </Text>
              <Button className="px-4" color="tertiary" variant="naked">
                <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
                  <IssueStatusIcon status="In Progress" />
                  <Text>In Progress</Text>
                </Box>
              </Button>
            </Box>
            <Box className="my-4 grid grid-cols-[8rem_auto] items-center gap-3">
              <Text className="truncate" fontWeight="medium" color="muted">
                Priority
              </Text>
              <Button className="px-4" color="tertiary" variant="naked">
                <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
                  <PriorityIcon priority="High" />
                  <Text>High</Text>
                </Box>
              </Button>
            </Box>

            <Box className="my-4 grid grid-cols-[8rem_auto] items-center gap-3">
              <Text className="truncate" fontWeight="medium" color="muted">
                Assignee
              </Text>
              <Button className="px-4" color="tertiary" variant="naked">
                <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
                  <IssueStatusIcon status="Backlog" />
                  <Text>Backlog</Text>
                </Box>
              </Button>
            </Box>
            <Box className="my-4 grid grid-cols-[8rem_auto] items-center gap-3">
              <Text className="truncate" fontWeight="medium" color="muted">
                Assignee
              </Text>
              <Button className="px-4" color="tertiary" variant="naked">
                <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
                  <IssueStatusIcon status="Done" />
                  <Text>Backlog</Text>
                </Box>
              </Button>
            </Box>
            <Divider />
            <Box className="my-4 grid grid-cols-[8rem_auto] items-center gap-3">
              <Text className="truncate" fontWeight="medium" color="muted">
                Labels
              </Text>
              <Button className="px-4" color="tertiary" variant="naked">
                <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
                  <IssueStatusIcon status="Backlog" />
                  <Text>Backlog</Text>
                </Box>
              </Button>
            </Box>
          </Container>
        </Box>
      </BodyContainer>
    </>
  );
}
