import type { ReactNode } from "react";
import { OnboardingLayout } from "@/components/layouts/onboarding-layout";

export default function Layout({ children }: { children: ReactNode }) {
  return <OnboardingLayout>{children}</OnboardingLayout>;
}
