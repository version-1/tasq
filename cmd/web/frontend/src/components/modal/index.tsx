import { useState, type ReactNode } from "react";
import { createPortal } from "react-dom";

export function ModalSlot({ children }: { children: ReactNode }) {
  const [slotElement, setSlotElement] = useState<HTMLDivElement | null>(null);

  return (
    <>
      <div ref={setSlotElement} />
      {slotElement ? createPortal(children, slotElement) : null}
    </>
  );
}
