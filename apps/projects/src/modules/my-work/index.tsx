"use client";
import type { StoriesLayout } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { useNotifications } from "../notifications/hooks/notifications";
import { Header } from "./components/header";
import { ListMyWork } from "./components/list-my-work";
import { MyWorkProvider } from "./components/provider";
import { MyWorkSkeleton } from "./components/my-work-skeleton";

export const ListMyStories = () => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "my-stories:stories:layout",
    "kanban",
  );
  const { isPending } = useNotifications();

  if (isPending) return <MyWorkSkeleton layout={layout} />;

  return (
    <MyWorkProvider>
      <Header layout={layout} setLayout={setLayout} />
      <ListMyWork layout={layout} />
    </MyWorkProvider>
  );
};
