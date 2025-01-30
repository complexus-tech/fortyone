"use client";

import { useEffect, useRef } from "react";
import { toast } from "sonner";
import { useOnlineStatus } from "@/hooks";

export const OnlineStatusMonitor = () => {
  const { isOnline, wasOffline, setWasOffline } = useOnlineStatus();
  const toastIdsRef = useRef<{
    offline: string | number;
    online: string | number;
  }>({
    offline: "",
    online: "",
  });

  useEffect(() => {
    if (!isOnline) {
      if (toastIdsRef.current.offline) {
        toast.dismiss(toastIdsRef.current.offline);
      }
      toastIdsRef.current.offline = toast.error("You're offline", {
        description: "Please check your internet connection",
        duration: Infinity,
        id: "offline-toast",
      });
    } else {
      if (toastIdsRef.current.offline) {
        toast.dismiss(toastIdsRef.current.offline);
        toastIdsRef.current.offline = "";
      }

      if (wasOffline) {
        if (toastIdsRef.current.online) {
          toast.dismiss(toastIdsRef.current.online);
        }
        toastIdsRef.current.online = toast.success("Back online", {
          description: "Your internet connection has been restored",
          id: "online-toast",
          duration: 5000,
        });
        setWasOffline(false);
      }
    }
  }, [isOnline, wasOffline, setWasOffline]);

  return null;
};
