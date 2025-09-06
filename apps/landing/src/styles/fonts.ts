import { Inter_Tight as Inter, Inconsolata, Figtree } from "next/font/google";

export const body = Inter({
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

export const heading = Figtree({
  variable: "--font-heading",
  display: "swap",
  subsets: ["latin"],
});
