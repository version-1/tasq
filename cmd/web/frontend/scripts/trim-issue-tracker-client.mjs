import { readFile, writeFile } from "node:fs/promises";

// Keep Orval's generated client free of trailing whitespace, as in the
// committed client, without hand-editing the generated contract.
const path = "src/lib/generated/issue-tracker.ts";
const source = await readFile(path, "utf8");
await writeFile(path, source.replace(/[\t ]+$/gm, ""));
