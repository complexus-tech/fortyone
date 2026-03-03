"use client";

import { Button } from "ui";
import { logOut } from "@/components/shared/sidebar/actions";
import { useAnalytics } from "@/hooks";

export const Back = () => {
  const { analytics } = useAnalytics();
  const handleLogout = async () => {
    try {
      await logOut();
      analytics.logout(true);
    } catch {
      // continue with redirect
    }
    window.location.href = "/?signedOut=true";
  };

  return (
    <Button className="gap-1 pl-2" color="tertiary" onClick={handleLogout}>
      Back to Home
    </Button>
  );
};
