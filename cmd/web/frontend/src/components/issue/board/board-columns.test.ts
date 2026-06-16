import { describe, expect, it } from "vitest";
import { boardColumns } from "./board-columns";

describe("boardColumns", () => {
  it("groups visible issue statuses by board column", () => {
    expect(boardColumns).toEqual([
      {
        key: "draft",
        titleKey: "issues.board.draft",
        statuses: ["backlog", "blocked"],
      },
      { key: "todo", titleKey: "issues.board.todo", statuses: ["ready"] },
      {
        key: "inProgress",
        titleKey: "issues.board.inProgress",
        statuses: ["in_progress"],
      },
      {
        key: "inReview",
        titleKey: "issues.board.inReview",
        statuses: ["review", "failed"],
      },
    ]);
  });

  it("omits hidden issue statuses from board columns", () => {
    const visibleStatuses = boardColumns.flatMap((column) => column.statuses);

    expect(visibleStatuses).not.toContain("cancelled");
    expect(visibleStatuses).not.toContain("done");
    expect(visibleStatuses).not.toContain("duplicate");
  });
});
