"use client";

import type { ReactNode } from "react";
import { useSquircleFallback } from "ui/squircle-fallback";
import "tailwind-config/squircle-fallback.css";

export function LandingCompatibility({ children }: { children: ReactNode }) {
  useSquircleFallback();

  return children;
}
