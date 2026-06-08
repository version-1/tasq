export type ToastType = "error" | "success";

export type Toast = {
  readonly id: string;
  readonly message: string;
  readonly type: ToastType;
};

export type ToastInput = {
  readonly message: string;
  readonly type: ToastType;
};

export type ToastListener = (toasts: readonly Toast[]) => void;

const toastTimeoutMs: Record<ToastType, number> = {
  error: 5_000,
  success: 3_000,
};

const listeners = new Set<ToastListener>();
const timers = new Map<string, ReturnType<typeof setTimeout>>();

let nextToastID = 1;
let toasts: Toast[] = [];

export function addToast(input: ToastInput): Toast {
  const toast: Toast = {
    id: createToastID(),
    message: input.message,
    type: input.type,
  };

  toasts = [...toasts, toast];
  timers.set(
    toast.id,
    setTimeout(() => {
      removeToast(toast.id);
    }, toastTimeoutMs[toast.type]),
  );
  notifyListeners();

  return toast;
}

export function subscribe(listener: ToastListener): () => void {
  listeners.add(listener);
  listener(getToastsSnapshot());

  return () => {
    listeners.delete(listener);
  };
}

export function removeToast(id: string): void {
  const timer = timers.get(id);
  if (timer) {
    clearTimeout(timer);
    timers.delete(id);
  }

  const nextToasts = toasts.filter((toast) => toast.id !== id);
  if (nextToasts.length === toasts.length) {
    return;
  }

  toasts = nextToasts;
  notifyListeners();
}

export function clearToasts(): void {
  timers.forEach((timer) => {
    clearTimeout(timer);
  });
  timers.clear();
  toasts = [];
  notifyListeners();
}

export function getToastsSnapshot(): readonly Toast[] {
  return toasts;
}

export const toastStore = {
  subscribe(listener: () => void): () => void {
    return subscribe(listener);
  },
  getSnapshot(): readonly Toast[] {
    return getToastsSnapshot();
  },
  dismiss(id: string): void {
    removeToast(id);
  },
  error(input: { readonly message: string }): Toast {
    return addToast({ type: "error", message: input.message });
  },
  success(input: { readonly message: string }): Toast {
    return addToast({ type: "success", message: input.message });
  },
  clear(): void {
    clearToasts();
  },
};

export const toast = {
  error(input: { readonly message: string }): Toast {
    return toastStore.error(input);
  },
  success(input: { readonly message: string }): Toast {
    return toastStore.success(input);
  },
  dismiss(id: string): void {
    toastStore.dismiss(id);
  },
};

function createToastID(): string {
  const id = `toast-${nextToastID}`;
  nextToastID += 1;
  return id;
}

function notifyListeners(): void {
  const snapshot = getToastsSnapshot();
  listeners.forEach((listener) => listener(snapshot));
}
