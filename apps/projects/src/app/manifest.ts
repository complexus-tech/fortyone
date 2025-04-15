import type { MetadataRoute } from "next";

export default function manifest(): MetadataRoute.Manifest {
  return {
    name: "Complexus",
    short_name: "Complexus",
    description: "A Product Management Platform",
    start_url: "/",
    display: "standalone",
    theme_color: "#002F61",
    background_color: "#EA6060",
    icons: [
      {
        src: "/icon-192x192.png",
        sizes: "192x192",
        type: "image/png",
      },
      {
        src: "/icon-512x512.png",
        sizes: "512x512",
        type: "image/png",
      },
    ],
  };
}
