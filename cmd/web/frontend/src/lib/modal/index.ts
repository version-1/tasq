import { useCallback, useState } from "react";

export const modalIDs = {
  addIssue: "addIssue",
} as const;

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

export function useModalState(): ModalController {
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
