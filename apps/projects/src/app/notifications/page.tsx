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
      read: true,
    },
  ];

  return (
    <BodyContainer className="h-screen">
      <Box className="grid h-full grid-cols-[350px_auto]">
        <Box className="h-full overflow-y-auto border-r border-gray-100 pb-6 dark:border-dark-200">
          {notifications.map(({ id }) => (
            <Card key={id} />
          ))}
        </Box>
        <Message />
      </Box>
    </BodyContainer>
  );
}
