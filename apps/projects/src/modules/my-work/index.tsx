"use client";
import { Suspense } from "react";
import type { StoriesLayout } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { Header } from "./components/header";
import { ListMyWork } from "./components/list-my-work";
import { MyWorkProvider } from "./components/provider";
import { MyWorkSkeleton } from "./components/my-work-skeleton";

const ListMyStories = () => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "my-stories:stories:layout",
    "kanban",
  );

  return (
    <MyWorkProvider>
      <Header layout={layout} setLayout={setLayout} />
      <Suspense fallback={<MyWorkSkeleton layout={layout} />}>
        <ListMyWork layout={layout} />
      </Suspense>
    </MyWorkProvider>
  );
};

export default ListMyStories;
