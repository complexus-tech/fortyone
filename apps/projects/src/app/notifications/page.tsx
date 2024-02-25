import { BodyContainer } from "@/components/layout";
import { ListNotifications } from "@/components/notifications/list";

type Notification = {
  id: number;
  title: string;
  description: string;
  date: string;
  read: boolean;
};

export default function Page(): JSX.Element {
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
    {
      id: 3,
      title: "New comment on issue",
      description: "This is a new comment on the issue.",
      date: "08:56",
      read: false,
    },
    {
      id: 4,
      title: "Subscribed to your project",
      description:
        "The quick brown fox jumps over the lazy dog. This is a test message.",
      date: "08:56",
      read: false,
    },
  ];

  return (
    <BodyContainer className="h-screen">
      <ListNotifications notifications={notifications} />
    </BodyContainer>
  );
}
