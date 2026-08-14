export type ChangeRequestVariant = "continue" | "reject";

export type ChangeRequestShortcut = {
  id: string;
  label: string;
  body: string;
};

export const builtInChangeRequestShortcuts = {
  continue: [{ id: "ok", label: "Ok", body: "Ok" }],
  reject: [
    {
      id: "fix-ci-conflict",
      label: "Fix CI & Conflict",
      body: "Fix CI & Conflict",
    },
  ],
} satisfies Record<ChangeRequestVariant, readonly ChangeRequestShortcut[]>;
