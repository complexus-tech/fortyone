"use client";

import { useEffect, useState } from "react";

/**
 * Returns the current time only after hydration so server HTML and the
 * browser's first render use the same time-independent presentation.
 */
export const useHydratedNow = () => {
  const [now, setNow] = useState<Date | null>(null);

  useEffect(() => {
    setNow(new Date());
  }, []);

  return now;
};
