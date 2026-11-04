import { afterAll, beforeAll, describe, expect, it } from "vitest";
import { setupServer } from "msw/node";
import { handlers } from "./handlers";

const server = setupServer(...handlers);
const base = `${window.location.origin}/tracker/api/v1`;

beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterAll(() => server.close());

describe("document artifact mock HTTP routes", () => {
  it("lists summaries without document bodies and retrieves the selected document", async () => {
    const listResponse = await fetch(`${base}/issues/1/artifacts/documents`);

    expect(listResponse.status).toBe(200);
    const list = (await listResponse.json()).data;
    expect(list).toEqual(expect.arrayContaining([
      expect.objectContaining({ id: "doc-1", title: "Acceptance criteria", type: "reference_document" }),
      expect.objectContaining({ id: "doc-2", title: "Work report", type: "work_report" }),
    ]));
    expect(list[0]).not.toHaveProperty("body");

    const documentResponse = await fetch(`${base}/issues/1/artifacts/documents/doc-1`);
    expect(documentResponse.status).toBe(200);
    expect((await documentResponse.json()).data).toMatchObject({ id: "doc-1", body: expect.stringContaining("Acceptance criteria") });
  });

  it("keeps document routes scoped to an existing issue", async () => {
    expect((await fetch(`${base}/issues/999999/artifacts/documents`)).status).toBe(404);
    expect((await fetch(`${base}/issues/1/artifacts/documents/unknown`)).status).toBe(404);
  });
});
