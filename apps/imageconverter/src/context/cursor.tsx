"use client";
import type { ReactNode } from "react";
import { createContext, useState, useEffect } from "react";
import MouseFollower from "mouse-follower";
import { gsap } from "gsap";

export const CursorContext = createContext({});

export const CursorProvider = ({ children }: { children: ReactNode }) => {
  const [cursor, setCursor] = useState<MouseFollower>({} as MouseFollower);

  useEffect(() => {
    MouseFollower.registerGSAP(gsap);
    const newCursor = new MouseFollower({
      speed: 1,
    });
    setCursor(newCursor);
    return () => {
      newCursor.destroy();
    };
  }, []);

  return (
    <CursorContext.Provider value={cursor}>{children}</CursorContext.Provider>
  );
};
