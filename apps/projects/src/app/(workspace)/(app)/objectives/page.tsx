import type { Metadata } from "next";
import { ObjectivesList } from "@/modules/objectives";

export const metadata: Metadata = {
  title: "Objectives",
};

export default function Page() {
  return <ObjectivesList />;
}
