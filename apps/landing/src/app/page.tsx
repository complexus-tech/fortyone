import {
  Hero,
  HeroCards,
  SampleClients,
  Features,
  Integrations,
  ProductDemo,
  // Reviews,
} from "@/components/pages/home";
import { Pricing } from "@/components/ui";

export default function Page(): JSX.Element {
  return (
    <>
      <Hero />
      <HeroCards />
      <SampleClients />
      <ProductDemo />
      <Features />
      {/* <Reviews /> */}
      <Integrations />
      <Pricing />
    </>
  );
}
