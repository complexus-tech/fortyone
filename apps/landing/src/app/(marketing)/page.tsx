import {
  Hero,
  HeroCards,
  SampleClients,
  Features,
  Integrations,
  Story,
  // Reviews,
  // SampleClients,
} from "@/modules/home";

export default function Page() {
  return (
    <>
      <Hero />
      <HeroCards />
      <SampleClients />
      <Story />
      <Features />
      {/* <Reviews /> */}
      <Integrations />
    </>
  );
}
