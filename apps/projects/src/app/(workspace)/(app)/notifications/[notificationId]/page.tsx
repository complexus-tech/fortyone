import { redirect } from "next/navigation";
import { NotificationDetails } from "@/modules/notifications/details";

export default async function Page({
  params,
  searchParams,
}: {
  params: Promise<{ notificationId: string }>;
  searchParams: Promise<{
    entityId?: string;
    entityType?: "story" | "objective";
  }>;
}) {
  const { notificationId } = await params;
  const { entityId, entityType } = await searchParams;

  if (!entityId || !entityType) {
    return redirect("/notifications");
  }

  return (
    <NotificationDetails
      entityId={entityId}
      entityType={entityType}
      notificationId={notificationId}
    />
  );
}
