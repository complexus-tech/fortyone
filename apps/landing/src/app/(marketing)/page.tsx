import { CallToAction } from "@/components/shared";
import { Pricing } from "@/components/ui";
import { Faqs } from "@/components/ui/faqs";
import {
  Hero,
  HeroCards,
  SampleClients,
  Features,
  Integrations,
  Maya,
  RunEverything,
} from "@/modules/home";

export default function Page() {
  return (
    <>
      <Hero />
      <HeroCards />
      <SampleClients />
      <Features />
      <Maya />
      <Integrations />
      <RunEverything />
      <Faqs />
      <Pricing className="md:pb-16 md:pt-0" hideDescription />
      <CallToAction />
    </>
  );
}
