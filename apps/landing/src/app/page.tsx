import {
  Hero,
  HeroCards,
  SampleClients,
  Features,
  Integrations,
  Reviews,
} from "@/components/pages/home";

export default function Page(): JSX.Element {
  return (
    <>
      <Hero />
      <HeroCards />
      <SampleClients />
      <Features />
      <Reviews />
      <Integrations />
    </>
  );
}
