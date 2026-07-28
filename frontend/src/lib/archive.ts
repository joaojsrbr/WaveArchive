export type ArchiveHeader = { schemaVersion: number; exportedAt: string };
export function validateArchiveHeader(value: unknown): ArchiveHeader {
  if (!value || typeof value !== 'object') throw new Error('Arquivo inválido.');
  const item = value as Partial<ArchiveHeader>;
  if (item.schemaVersion !== 1)
    throw new Error(`Versão de arquivo não suportada: ${String(item.schemaVersion)}`);
  if (typeof item.exportedAt !== 'string' || !item.exportedAt)
    throw new Error('Data de exportação ausente.');
  return item as ArchiveHeader;
}
export function parseArchive(text: string): unknown {
  if (!text.trim()) throw new Error('Arquivo vazio.');
  const value = JSON.parse(text);
  validateArchiveHeader(value);
  return value;
}
