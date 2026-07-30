let counter = 0;

/** Locally-unique id — good enough for idempotency keys and list keys (no crypto needed). */
export function generateId(prefix = 'id'): string {
  counter += 1;
  return `${prefix}_${Date.now().toString(36)}_${counter}_${Math.random().toString(36).slice(2, 10)}`;
}
