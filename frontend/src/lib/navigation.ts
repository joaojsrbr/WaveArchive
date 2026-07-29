export type OpenTarget = { kind: string; id: number; title?: string };

export function readOpenTarget(kind: string): OpenTarget | undefined {
  try {
    const value = JSON.parse(sessionStorage.getItem('wavearchive:open-target') || '') as OpenTarget;
    if (value.kind !== kind || !Number.isFinite(value.id)) return undefined;
    sessionStorage.removeItem('wavearchive:open-target');
    return value;
  } catch {
    return undefined;
  }
}
