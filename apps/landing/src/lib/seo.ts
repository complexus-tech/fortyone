export const SITE_URL = "https://www.fortyone.app";

export const DEFAULT_SOCIAL_IMAGE = {
  url: "/opengraph-image.png",
  width: 1682,
  height: 1006,
  alt: "FortyOne customer feedback and project management platform",
};

export const DEFAULT_TWITTER_IMAGE = {
  url: "/twitter-image.png",
  width: 1682,
  height: 1006,
  alt: "FortyOne customer feedback and project management platform",
};

export const getCanonicalUrl = (pathname: string) =>
  new URL(pathname, SITE_URL).toString();
