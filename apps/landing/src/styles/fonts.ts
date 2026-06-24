import {
  Geist,
  Playfair_Display as Playfair,
  Inconsolata,
} from "next/font/google";

export const body = Geist({
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

export const heading = Playfair({
  variable: "--font-heading",
  display: "swap",
  subsets: ["latin"],
  weight: "variable",
});
