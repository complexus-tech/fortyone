import { CallToAction } from "@/components/shared";
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
      <Transform />
      <Integrations />
      <Testimonials />
      <CallToAction />
    </>
  );
}
