import { useContext } from "react";
import type MouseFollower from "mouse-follower";
import { CursorContext } from "@/context";

export const useCursor = () => {
  const cursor = useContext(CursorContext);
  return cursor as MouseFollower;
};
