import { useState, type ReactNode } from "react";
import { createPortal } from "react-dom";
import { modalIDs } from "@/constants";
import type { LayoutShellData } from "@/components/layout";
import { AddIssueDialog } from "@/components/layout/add-issue-dialog";

export function ModalSlot({ shellData }: { shellData: LayoutShellData }) {
  const [slotElement, setSlotElement] = useState<HTMLDivElement | null>(null);

  return (
    <>
      <div ref={setSlotElement} />
      {slotElement ? createPortal(renderModal(shellData), slotElement) : null}
    </>
  );
}

function renderModal(shellData: LayoutShellData): ReactNode {
  if (shellData.modal.activeModalID === modalIDs.addIssue) {
    return (
      <AddIssueDialog
        error={shellData.addIssueError}
        initialStatus={shellData.addIssueInitialStatus}
        project={shellData.activeProject}
        onCancel={shellData.onCloseModal}
        onSubmit={shellData.onCreateIssue}
      />
    );
  }

  return null;
}
