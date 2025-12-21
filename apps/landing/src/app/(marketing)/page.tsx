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
      <SampleClients />
      <HeroCards />
      <CoreValues />
      {/* <Features /> */}
      <Maya />
      <Integrations />
      <Pricing className="md:pb-16 md:pt-0" hideDescription />
      <Faqs />
      <CallToAction />
    </>
  );
}
