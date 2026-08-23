/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { render, screen } from "@testing-library/react";
import { AVATAR_COLORS, getAvatarColor } from "lib";
import { Avatar } from "ui";

const getRelativeLuminance = (hexColor: string) => {
  const channels = [1, 3, 5].map((offset) => {
    const value = Number.parseInt(hexColor.slice(offset, offset + 2), 16) / 255;
    return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
  });

  return 0.2126 * channels[0] + 0.7152 * channels[1] + 0.0722 * channels[2];
};

describe("avatar colors", () => {
  it("uses the six approved avatar colors", () => {
    expect(AVATAR_COLORS).toHaveLength(6);
  });

  it("keeps initials readable on every generated background", () => {
    for (const { backgroundColor, foregroundColor } of AVATAR_COLORS) {
      const backgroundLuminance = getRelativeLuminance(backgroundColor);
      const foregroundLuminance = getRelativeLuminance(foregroundColor);
      const lighter = Math.max(backgroundLuminance, foregroundLuminance);
      const darker = Math.min(backgroundLuminance, foregroundLuminance);

      expect((lighter + 0.05) / (darker + 0.05)).toBeGreaterThanOrEqual(4.5);
    }
  });

  it("returns the same color for equivalent names", () => {
    expect(getAvatarColor("Joseph Mukorio")).toEqual(
      getAvatarColor("  joseph   mukorio  "),
    );
  });

  it("distributes different names across the palette", () => {
    const backgrounds = new Set(
      [
        "Joseph Mukorio",
        "Leah Mensah",
        "Amara Ndlovu",
        "Daniel Okafor",
        "Ethan Clarke",
        "Maya",
      ].map((name) => getAvatarColor(name).backgroundColor),
    );

    expect(backgrounds.size).toBeGreaterThan(1);
  });

  it("uses a stable fallback for an empty name", () => {
    expect(getAvatarColor("")).toEqual(getAvatarColor("user"));
  });

  it("automatically colors named avatars without an image", () => {
    const name = "Joseph Mukorio";
    const color = getAvatarColor(name);

    render(<Avatar name={name} />);

    expect(screen.getByTitle(name).parentElement).toHaveStyle({
      backgroundColor: color.backgroundColor,
      color: color.foregroundColor,
    });
  });

  it("prioritizes an explicit color prop", () => {
    const name = "Joseph Mukorio";

    render(<Avatar color="primary" name={name} />);

    expect(screen.getByTitle(name).parentElement).not.toHaveAttribute("style");
  });

  it("prioritizes explicit inline colors", () => {
    const name = "Joseph Mukorio";

    render(
      <Avatar
        name={name}
        style={{ backgroundColor: "#FFFFFF", color: "#000000" }}
      />,
    );

    expect(screen.getByTitle(name).parentElement).toHaveStyle({
      backgroundColor: "#FFFFFF",
      color: "#000000",
    });
  });
});
