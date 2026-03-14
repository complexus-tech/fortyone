import {
  Inter_Tight as InterTight,
  Inconsolata,
  Manrope,
} from "next/font/google";

export const body = Manrope({
  variable: "--font-body",
  subsets: ["latin"],
  display: "swap",
  weight: "variable",
});

export const mono = Inconsolata({
  variable: "--font-mono",
  display: "swap",
  subsets: ["latin"],
});

export const heading = InterTight({
  variable: "--font-heading",
  display: "swap",
  subsets: ["latin"],
  weight: "variable",
});
