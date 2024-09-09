import { Hero, Support } from "@/components/pages/contact";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Contact | ImageConveta",
};

export default function Page(): JSX.Element {
  return (
    <>
      <Hero />
      <Support />
    </>
  );
}
