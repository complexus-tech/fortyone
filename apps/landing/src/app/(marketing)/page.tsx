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
} from "@/modules/home";

export default function Page() {
  return (
    <>
      <Hero />
      <HeroCards />
      <SampleClients />
      <Maya />
      <Features />
      <Integrations />
      {/* <Testimonials /> */}
      <Faqs />
      <CallToAction />
    </>
  );
}
