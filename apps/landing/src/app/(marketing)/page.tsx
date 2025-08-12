import { CallToAction } from "@/components/shared";
import { Faqs } from "@/components/ui/faqs";
import {
  Hero,
  HeroCards,
  SampleClients,
  Features,
  Integrations,
  Testimonials,
  Transform,
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
      <RunEverything />
      <Integrations />

      {/* <Transform /> */}

      <Testimonials />
      <Faqs />
      <CallToAction />
    </>
  );
}
