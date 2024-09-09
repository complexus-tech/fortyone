import { Hero, Story } from "@/components/pages/about";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "About | ImageConveta",
};

export default function Page(): JSX.Element {
  return (
    <>
      <Hero />
      <Story />
    </>
  );
}
