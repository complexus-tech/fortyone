import { BodyContainer } from "@/components/shared";
import { ListNotifications } from "@/modules/notifications/list";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Notifications",
};

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
      title: "Subscribed to your objective",
      description:
        "The quick brown fox jumps over the lazy dog. This is a test message.",
      date: "08:56",
      read: false,
    },
    {
      id: 2,
      title: "Updated story status",
      description: "I have changed the status of the story to 'In Progress'.",
      date: "08:56",
      read: true,
    },
    {
      id: 3,
      title: "New comment on story",
      description: "This is a new comment on the story.",
      date: "08:56",
      read: false,
    },
    {
      id: 4,
      title: "Subscribed to your objective",
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
