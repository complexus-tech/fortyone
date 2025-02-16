import {
  Hero,
  HeroCards,
  // SampleClients,
  Features,
  Integrations,
  Story,
  // ProductDemo,
  // Reviews,
} from "@/components/pages/home";

export default function Page() {
  return (
    <>
      <Hero />
      <HeroCards />
      {/* <SampleClients /> */}
      {/* <ProductDemo /> */}
      <Story />
      <Features />
      {/* <Reviews /> */}
      <Integrations />
    </>
  );
}
