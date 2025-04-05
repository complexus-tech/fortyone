import {
  Hero,
  HeroCards,
  SampleClients,
  Features,
  Integrations,
  Testimonials,
} from "@/modules/home";

export default function Page() {
  return (
    <>
      <Hero />
      <HeroCards />
      <SampleClients />
      <Testimonials />
      <Features />
      <Integrations />
    </>
  );
}
