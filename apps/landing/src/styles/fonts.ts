import {
  Englebert,
  Inter_Tight as Inter,
  Inconsolata,
  Figtree,
} from "next/font/google";

export const body = Inter({
  variable: "--font-body",
  subsets: ["latin"],
  display: "swap",
  weight: "variable",
});

export const heading = Englebert({
  variable: "--font-heading",
  display: "swap",
  subsets: ["latin"],
  weight: "400",
  style: "normal",
});

export const mono = Inconsolata({
  variable: "--font-mono",
  display: "swap",
  subsets: ["latin"],
});

export const figtree = Figtree({
  variable: "--font-figtree",
  display: "swap",
  subsets: ["latin"],
});
