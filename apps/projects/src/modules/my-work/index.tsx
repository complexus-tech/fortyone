"use client";
import type { StoriesLayout } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { Header } from "./components/header";
import { ListMyWork } from "./components/list-my-work";
import { MyWorkProvider } from "./components/provider";
import { useMyStories } from "./hooks/my-stories";
import { MyWorkSkeleton } from "./components/my-work-skeleton";

export const ListMyStories = () => {
  const { isPending } = useMyStories();
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "my-stories:stories:layout",
    "kanban",
  );

  if (isPending) return <MyWorkSkeleton layout={layout} />;

  return (
    <MyWorkProvider>
      <Header layout={layout} setLayout={setLayout} />
      <ListMyWork layout={layout} />
    </MyWorkProvider>
  );
};
