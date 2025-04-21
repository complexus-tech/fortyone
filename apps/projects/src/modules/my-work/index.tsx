"use client";
import type { StoriesLayout } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { Header } from "./components/header";
import { ListMyWork } from "./components/list-my-work";
import { MyWorkProvider } from "./components/provider";
import { MyWorkSkeleton } from "./components/my-work-skeleton";
import { useMyStories } from "./hooks/my-stories";

export const ListMyStories = () => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "my-stories:stories:layout",
    "list",
  );
  const { isPending } = useMyStories();

  if (isPending) return <MyWorkSkeleton layout={layout} />;

  return (
    <MyWorkProvider>
      <Header layout={layout} setLayout={setLayout} />
      <ListMyWork layout={layout} />
    </MyWorkProvider>
  );
};
