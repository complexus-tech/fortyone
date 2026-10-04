import type { Metadata } from "next";
import { ViewsClient } from "./views-client";

export const metadata: Metadata = { title: "Views" };

export default function Page() {
  return <ViewsClient />;
}
