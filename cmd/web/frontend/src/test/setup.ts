import "@testing-library/jest-dom/vitest";
import { afterEach, beforeEach } from "vitest";
import { cleanup } from "@testing-library/react";
import "@/lib/i18n";
import { i18n } from "@/lib/i18n";

beforeEach(async () => {
  window.localStorage.clear();
  await i18n.changeLanguage("en");
});

afterEach(() => {
  cleanup();
});
