import { Inconsolata, Inter } from "next/font/google";

export const sans = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
  display: "swap",
  weight: "variable",
});

export const mono = Inconsolata({
  variable: "--font-mono",
  display: "swap",
  subsets: ["latin"],
});
