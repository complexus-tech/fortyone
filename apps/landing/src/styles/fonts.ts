import { Inconsolata, Inter, Kalam, Newsreader } from "next/font/google";

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

export const handwritten = Kalam({
  variable: "--font-kalam",
  subsets: ["latin"],
  display: "swap",
  weight: "700",
});

export const serif = Newsreader({
  variable: "--font-newsreader",
  subsets: ["latin"],
  display: "swap",
  style: ["normal", "italic"],
  weight: "400",
});
