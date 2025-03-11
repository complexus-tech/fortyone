"use client";
import type { StoriesLayout } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { Header } from "./components/header";
import { ListMyWork } from "./components/list-my-work";
import { MyWorkProvider } from "./components/provider";

export const ListMyStories = () => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "my-stories:stories:layout",
    "kanban",
  );

  return (
    <MyWorkProvider>
      <Header layout={layout} setLayout={setLayout} />
      <ListMyWork layout={layout} />
    </MyWorkProvider>
  );
};
