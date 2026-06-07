import { createContext, createElement, useCallback, useContext, useState, type ReactNode } from "react";
import { modalIDs } from "@/constants";

export type ModalID = (typeof modalIDs)[keyof typeof modalIDs];

export type ModalState =
  | { kind: "closed" }
  | { kind: "open"; id: ModalID };

export type ModalController = {
  activeModalID: ModalID | null;
  closeModal: () => void;
  isModalOpen: (id: ModalID) => boolean;
  openModal: (id: ModalID) => void;
};

const modalContext = createContext<ModalController | null>(null);

export function ModalProvider({ children }: { children: ReactNode }) {
  const modal = useModalState();

  return createElement(modalContext.Provider, { value: modal }, children);
}

export function useModal(): ModalController {
  const modal = useContext(modalContext);
  if (!modal) {
    throw new Error("useModal must be used within ModalProvider");
  }
  return modal;
}

function useModalState(): ModalController {
  const [modalState, setModalState] = useState<ModalState>({ kind: "closed" });

  const closeModal = useCallback(() => {
    setModalState({ kind: "closed" });
  }, []);

  const openModal = useCallback((id: ModalID) => {
    setModalState({ kind: "open", id });
  }, []);

  const activeModalID = modalState.kind === "open" ? modalState.id : null;

  return {
    activeModalID,
    closeModal,
    isModalOpen: (id) => activeModalID === id,
    openModal,
  };
}
