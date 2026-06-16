import { describe, expect, it } from "vitest";
import { boardColumns } from "./board-columns";

describe("boardColumns", () => {
  it("omits done issues from board columns", () => {
    const visibleStatuses = boardColumns.flatMap((column) => column.statuses);

    expect(visibleStatuses).not.toContain("done");
  });
});
