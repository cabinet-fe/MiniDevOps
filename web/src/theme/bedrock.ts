import { heroDarkTheme } from "@veltra/styles/theme";

/**
 * Forge Console theme — ink / verdigris / bone aligned with the login palette.
 * All chrome colors flow through Veltra tokens via loadTheme; do not hardcode
 * business colors for theme matching elsewhere.
 */
export const bedrockTheme = heroDarkTheme.new({
  color: {
    primary: "#4a7a64",
  },
  bg: {
    color: {
      bottom: "#070a08",
      middle: "#0e1411",
      top: "#141b17",
      hover: "#1c2621",
    },
  },
  "text-color": {
    title: "#d4cfc3",
    main: "#d4cfc3",
    second: "#a8b0a8",
    assist: "#a8b0a8",
    placeholder: "#9aa49a",
  },
  border: {
    color: "#2a483c",
    mutedColor: "#1e332b",
  },
  radius: {
    small: 4,
    default: 8,
    large: 14,
  },
  shadow: {
    color: "rgba(0, 0, 0, 0.45)",
    x: 0,
    y: 8,
    blur: 28,
    spread: -4,
  },
});
