import { Box } from "ui";
import { BodyContainer } from "@/components/layout";
import { Card, Message } from "./components";

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
    <BodyContainer className="h-screen">
      <Box className="grid h-full grid-cols-[350px_auto]">
        <Box className="h-full border-r border-gray-100 dark:border-dark-100">
          {notifications.map(({ id }) => (
            <Card key={id} />
          ))}
        </Box>
        <Message />
      </Box>
    </BodyContainer>
  );
}
