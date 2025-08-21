import { CallToAction } from "@/components/shared";
import { Pricing } from "@/components/ui";
import { Faqs } from "@/components/ui/faqs";
import {
  Hero,
  HeroCards,
  SampleClients,
  // Features,
  Integrations,
  Maya,
  CoreValues,
} from "@/modules/home";

export default function Page() {
  return (
    <>
      <Hero />
      <HeroCards />
      <SampleClients />
      <CoreValues />
      {/* <Features /> */}
      <Maya />
      <CoreValues />
      <Integrations />
      <Pricing className="md:pb-16 md:pt-0" hideDescription />
      <Faqs />
      <CallToAction />
    </>
  );
}
