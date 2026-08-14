import { describe, expect, it } from "vitest";
import { getBuildInfo } from "./build-info";

describe("getBuildInfo", () => {
  it("reads release metadata from the document", () => {
    document.head.innerHTML = `
      <meta name="tasq-version" content="0.1.0" />
      <meta name="tasq-commit" content="abc1234" />
    `;

    expect(getBuildInfo()).toEqual({ version: "v0.1.0", commit: "abc1234" });
  });

  it("falls back to dev and omits an unknown commit", () => {
    document.head.innerHTML = `
      <meta name="tasq-version" content="__TASQ_VERSION__" />
      <meta name="tasq-commit" content="unknown" />
    `;

    expect(getBuildInfo()).toEqual({ version: "dev", commit: null });
  });
});
