"use client";

import { getCalApi } from "@calcom/embed-react";
import { cn } from "lib";
import { useEffect } from "react";
import { buttonVariants } from "ui";

export const RequestDemo = () => {
  useEffect(() => {
    void getCalApi({ namespace: "30min" }).then((cal) => {
      cal("ui", { hideEventTypeDetails: false, layout: "month_view" });
    });
  }, []);
  return (
    <button
      className={cn(
        buttonVariants({ color: "tertiary", variant: "naked", rounded: "lg" }),
        "hidden text-[0.93rem] md:flex",
      )}
      data-cal-config='{"layout":"month_view"}'
      data-cal-link="complexus/30min"
      data-cal-namespace="30min"
      type="button"
    >
      Book a demo
    </button>
  );
};
