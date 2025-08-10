import { Englebert, Inter } from "next/font/google";

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
