import {
  Hero,
  HeroCards,
  SampleClients,
  Features,
  Integrations,
  Reviews,
  ProductDemo,
} from "@/components/pages/home";

export default function Page(): JSX.Element {
  return (
    <>
      <Hero />
      <HeroCards />
      <SampleClients />
      <ProductDemo />
      <Features />
      <Reviews />
      <Integrations />
    </>
  );
}
