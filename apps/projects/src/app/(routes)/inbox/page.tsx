import {
  TbAdjustmentsHorizontal,
  TbLayoutDashboard,
  TbMail,
  TbPinned,
  TbTrash,
} from "react-icons/tb";
import { Box, BreadCrumbs, Button, Flex, Text } from "ui";
import { HiOutlineInboxArrowDown } from "react-icons/hi2";
import { BodyContainer, HeaderContainer } from "@/components/shared";
import { NewIssueButton, RowWrapper } from "@/components/ui";

type Notification = {
  id: number;
  title: string;
  description: string;
  date: string;
  read: boolean;
};

export default function Page(): JSX.Element {
  // generate 10 random notifications
  const notifications: Notification[] = [
    {
      id: 1,
      title: "Subscribed to your project",
      description:
        "The quick brown fox jumps over the lazy dog. This is a test message.",
      date: "08:56",
      read: false,
    },
    {
      id: 2,
      title: "Updated issue status",
      description: "I have changed the status of the issue to 'In Progress'.",
      date: "08:56",
      read: false,
    },
    {
      id: 3,
      title: "Added new comment",
      description:
        "The quick brown fox jumps over the lazy dog. This is a test message.",
      date: "08:56",
      read: false,
    },
    {
      id: 4,
      title: "Created new issue",
      description:
        "The quick brown fox jumps over the lazy dog. This is a test message.",
      date: "08:56",
      read: false,
    },
    {
      id: 5,
      title: "Created new issue",
      description:
        "The quick brown fox jumps over the lazy dog. This is a test message.",
      date: "08:56",
      read: false,
    },
    {
      id: 6,
      title: "Created new issue",
      description:
        "The quick brown fox jumps over the lazy dog. This is a test message.",
      date: "08:56",
      read: true,
    },
    {
      id: 7,
      title: "Created new issue",
      description:
        "The quick brown fox jumps over the lazy dog. This is a test message.",
      date: "08:56",
      read: true,
    },
    {
      id: 8,
      title: "Created new issue",
      description:
        "The quick brown fox jumps over the lazy dog. This is a test message.",
      date: "08:56",
      read: true,
    },
    {
      id: 9,
      title: "Created new issue",
      description:
        "The quick brown fox jumps over the lazy dog. This is a test message.",
      date: "08:56",
      read: true,
    },
    {
      id: 10,
      title: "Created new issue",
      description:
        "The quick brown fox jumps over the lazy dog. This is a test message.",
      date: "08:56",
      read: true,
    },
  ];

  return (
    <>
      <HeaderContainer className="justify-between">
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "Inbox",
              icon: <TbLayoutDashboard className="h-5 w-auto" />,
            },
          ]}
        />
        <Flex gap={2}>
          <Button
            color="tertiary"
            leftIcon={<TbAdjustmentsHorizontal className="h-5 w-auto" />}
            size="sm"
            variant="outline"
          >
            Display
          </Button>
          <NewIssueButton />
        </Flex>
      </HeaderContainer>
      <BodyContainer>
        <Box className="grid h-full grid-cols-[350px_auto]">
          <Box className="h-full border-r border-gray-100 dark:border-dark-100">
            {notifications.map(({ id, title, description, date }) => (
              <RowWrapper
                className="group border-l-4 px-6 pt-2 dark:bg-dark-100/10"
                key={id}
              >
                <Box className="w-full">
                  <Flex align="center" justify="between">
                    <Text
                      className="truncate opacity-80"
                      fontSize="lg"
                      fontWeight="medium"
                    >
                      Joseph Mukorivo
                    </Text>
                    <Flex
                      align="center"
                      className="opacity-70 transition group-hover:opacity-100"
                      gap={3}
                    >
                      <TbPinned className="h-5 w-auto dark:text-gray" />
                      <TbMail className="h-5 w-auto dark:text-gray" />
                      <TbTrash className="h-5 w-auto dark:text-gray" />
                    </Flex>
                  </Flex>

                  <Flex align="center" className="my-[2px]" justify="between">
                    <Text className="opacity-80">{title}</Text>
                    <Text fontSize="sm">{date}</Text>
                  </Flex>

                  <Text className="truncate" color="muted" fontSize="sm">
                    {description}
                  </Text>
                </Box>
              </RowWrapper>
            ))}
          </Box>
          <Flex align="center" className="h-full" justify="center">
            <Flex align="center" direction="column">
              <HiOutlineInboxArrowDown
                className="h-28 w-auto dark:text-gray"
                strokeWidth={0.8}
              />
              <Text color="muted">Select a notification to read.</Text>
            </Flex>
          </Flex>
        </Box>
      </BodyContainer>
    </>
  );
}
