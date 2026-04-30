import type { Metadata } from "next";
import { SelectIntegrationRequestMessage } from "@/modules/integration-requests/message";

export const metadata: Metadata = {
  title: "Requests",
};

export default function Page() {
  return <SelectIntegrationRequestMessage />;
}
