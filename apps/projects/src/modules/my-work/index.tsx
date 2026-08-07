"use client";
import { useEffect } from "react";
import dynamic from "next/dynamic";
import { useSearchParams } from "next/navigation";
import { toast } from "sonner";
import { Box, Skeleton } from "ui";
import type { StoriesLayout } from "@/components/ui";
import { useLocalStorage, useMediaQuery } from "@/hooks";
import { Header } from "./components/header";
import { ListMyWork } from "./components/list-my-work";
import { MyWorkProvider } from "./components/provider";
import type { MyWorkLayout } from "./types";

const MyWorkCalendar = dynamic(
  () => import("./components/calendar").then((module) => module.MyWorkCalendar),
  {
    loading: () => (
      <Box className="bg-surface h-[calc(100dvh-4rem)] px-6 py-5">
        <Skeleton className="mb-4 h-9 w-72" />
        <Skeleton className="h-[calc(100%-3.25rem)] w-full" />
      </Box>
    ),
    ssr: false,
  },
);

export const ListMyStories = () => {
  const searchParams = useSearchParams();
  const isMobile = useMediaQuery("(max-width: 768px)");
  const [layout, setLayout] = useLocalStorage<MyWorkLayout>(
    "my-stories:stories:layout",
    isMobile ? "list" : "kanban",
  );
  const requestedCalendar = searchParams.get("layout") === "calendar";
  let activeLayout: MyWorkLayout = layout;
  if (isMobile && layout === "calendar") {
    activeLayout = "list";
  } else if (!isMobile && requestedCalendar) {
    activeLayout = "calendar";
  }
  const storiesLayout: StoriesLayout =
    activeLayout === "calendar" ? "list" : activeLayout;

  useEffect(() => {
    if (isMobile || !requestedCalendar) {
      return;
    }

    setLayout("calendar");
    const url = new URL(window.location.href);
    url.searchParams.delete("layout");
    window.history.replaceState({}, "", url.toString());
  }, [isMobile, layout, requestedCalendar, searchParams, setLayout]);

  useEffect(() => {
    if (
      searchParams.get("session_id") &&
      !sessionStorage.getItem("stripeSession")
    ) {
      toast.success("Payment successful", {
        description: "Payment for your subscription has been successful",
      });
      sessionStorage.setItem("stripeSession", searchParams.get("session_id")!);
    }
  }, [searchParams]);

  return (
    <MyWorkProvider layout={storiesLayout}>
      <Header
        layout={activeLayout}
        setLayout={setLayout}
        showCalendar={!isMobile}
      />
      {activeLayout === "calendar" ? (
        <MyWorkCalendar />
      ) : (
        <ListMyWork layout={activeLayout} />
      )}
    </MyWorkProvider>
  );
};
