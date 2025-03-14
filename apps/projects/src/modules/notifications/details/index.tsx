import { StoryPage } from "@/modules/story";

export const NotificationDetails = ({
  entityId,
}: {
  notificationId: string;
  entityId: string;
  entityType: "story" | "objective";
}) => {
  return (
    <div>
      <StoryPage storyId={entityId} />
    </div>
  );
};
