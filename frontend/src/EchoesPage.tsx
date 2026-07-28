import { useEffect, useState, type FormEvent } from "react";
import { deleteOwnedEcho, getEcho, listEchoes, listOwnedEchoes, listSonatas, saveOwnedEcho } from "./lib/backend";
import type { Echo, EchoFilter, OwnedEcho, Sonata } from "./types";

const initialFilter: EchoFilter = { query: "", cost: 0, sonataId: 0, ownedOnly: false, sort: "name" };

export function EchoesPage({ mode = "echoes", onError }: { mode?: "echoes" | "sonata"; onError: (message: string) => void }) {
  const [echoes, setEchoes] = useState<Echo[]>([]);
  const [sonatas, setSonatas] = useState<Sonata[]>([]);
  const [owned, setOwned] = useState<OwnedEcho[]>([]);
  const [filter, setFilter] = useState<EchoFilter>(initialFilter);
  const [tab, setTab] = useState<"catalog" | "inventory">("catalog");
  const [detail, setDetail] = useState<Echo>();
  const [draft, setDraft] = useState<OwnedEcho>();
  const [loading, setLoading] = useState(true);

  async function load(nextFilter = filter) {
    setLoading(true);
    try {
      const [nextEchoes, nextSonatas, nextOwned] = await Promise.all([
        listEchoes(nextFilter), listSonatas(), listOwnedEchoes()
      ]);
      setEchoes(nextEchoes);
      setSonatas(nextSonatas);
      setOwned(nextOwned);
      onError("");
    } catch (cause) {
      onError(messageFrom(cause));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    const timer = window.setTimeout(() => void load(filter), 150);
    return () => window.clearTimeout(timer);
  }, [filter.query, filter.cost, filter.sonataId, filter.ownedOnly, filter.sort]);

  async function openEcho(id: number) {
    try { setDetail(await getEcho(id)); }
    catch (cause) { onError(messageFrom(cause)); }
  }

  function addToInventory(echo: Echo) {
    setDraft({
      id: 0, echoId: echo.id, echoName: echo.name, iconPath: echo.iconPath, cost: echo.cost,
      mainStat: "", substatsJson: "[]", level: 0, sonataId: firstSonata(echo),
      sonataName: "", characterName: "", locked: false, favorite: false, note: ""
    });
  }

  async function save(event: FormEvent) {
    event.preventDefault();
    if (!draft) return;
    try {
      await saveOwnedEcho(draft);
      setDraft(undefined);
      setDetail(undefined);
      await load();
      setTab("inventory");
    } catch (cause) { onError(messageFrom(cause)); }
  }

  async function remove(item: OwnedEcho) {
    if (!window.confirm(`Remover ${item.echoName} do inventário?`)) return;
    try { await deleteOwnedEcho(item.id); await load(); }
    catch (cause) { onError(item.locked ? "Desbloqueie o Echo antes de removê-lo." : messageFrom(cause)); }
  }

  if (mode === "sonata") {
    return <div><div className="pageIntro"><div><span className="eyebrow">EFEITOS DE CONJUNTO</span><h1>SONATA EFFECTS</h1><p>Consulte bônus de duas e cinco peças disponíveis na versão sincronizada.</p></div></div>
      {loading ? <Loading /> : <div className="sonataGrid">{sonatas.map((set) => <article className="sonataCard" key={set.id}>
        <div className="sonataTitle">{image(set.iconPath, set.name)}<div><span>SONATA #{set.id}</span><h2>{set.name}</h2></div></div>
        <div><strong>2 PEÇAS</strong><p>{set.twoPiece || "Efeito não informado."}</p></div>
        <div><strong>5 PEÇAS</strong><p>{set.fivePiece || "Efeito não informado."}</p></div>
      </article>)}</div>}
    </div>;
  }

  const visibleOwned = filter.query
    ? owned.filter((item) => item.echoName.toLowerCase().includes(filter.query.toLowerCase()))
    : owned;

  return <div>
    <div className="pageIntro"><div><span className="eyebrow">BANCO DE DADOS DE TACET DISCORDS</span><h1>ECHOES</h1><p>Explore o catálogo e registre cada peça do seu inventário local.</p></div>
      <div className="catalogTabs"><button className={tab === "catalog" ? "active" : ""} onClick={() => setTab("catalog")}>Catálogo</button><button className={tab === "inventory" ? "active" : ""} onClick={() => setTab("inventory")}>Meu inventário <span>{owned.length}</span></button></div>
    </div>
    <div className="toolbar">
      <label className="search"><span aria-hidden="true">⌕</span><span className="srOnly">Pesquisar Echoes</span><input value={filter.query} onChange={(event) => setFilter({ ...filter, query: event.target.value })} placeholder="Buscar Echo por nome…" /></label>
      <select value={filter.sort} onChange={(event) => setFilter({ ...filter, sort: event.target.value as EchoFilter["sort"] })}><option value="name">Nome A–Z</option><option value="cost">Maior custo</option><option value="id">ID</option></select>
    </div>
    <div className="filterRow echoFilters">
      <div className="chips">{[0, 1, 3, 4].map((cost) => <button key={cost} className={filter.cost === cost ? "chip active" : "chip"} onClick={() => setFilter({ ...filter, cost })}>{cost ? `Custo ${cost}` : "Todos"}</button>)}</div>
      <select aria-label="Filtrar por Sonata" value={filter.sonataId} onChange={(event) => setFilter({ ...filter, sonataId: Number(event.target.value) })}><option value={0}>Todas as Sonatas</option>{sonatas.map((set) => <option value={set.id} key={set.id}>{set.name}</option>)}</select>
    </div>
    {loading ? <Loading /> : tab === "catalog" ? <div className="echoGrid">{echoes.map((echo) => <article className="echoCard" key={echo.id} onClick={() => void openEcho(echo.id)}>
      <div className={`echoImage cost-${echo.cost}`}>{image(echo.iconPath, echo.name)}<span>COST {echo.cost}</span></div>
      <div className="echoBody"><small>{echo.class || "Echo"}</small><h2>{echo.name}</h2><p>{sonataNames(echo, sonatas)}</p><footer><span>{echo.ownedCount ? `${echo.ownedCount} no inventário` : "Não registrado"}</span><b>›</b></footer></div>
    </article>)}</div> : visibleOwned.length ? <div className="ownedEchoList">{visibleOwned.map((item) => <article key={item.id}>
      {image(item.iconPath, item.echoName)}<div><small>CUSTO {item.cost} · NV. +{item.level}</small><h2>{item.echoName}</h2><p>{item.mainStat || "Atributo principal não informado"}{item.sonataName ? ` · ${item.sonataName}` : ""}</p></div>
      <button onClick={() => setDraft(item)}>Editar</button><button onClick={() => void remove(item)} disabled={item.locked}>×</button>
    </article>)}</div> : <div className="emptyState"><div className="emptyGlyph">◇</div><h2>Inventário vazio</h2><p>Abra um Echo do catálogo e registre uma peça com seus atributos.</p></div>}
    {detail && <EchoDetail echo={detail} sonatas={sonatas} onClose={() => setDetail(undefined)} onAdd={() => addToInventory(detail)} />}
    {draft && <OwnedEchoModal draft={draft} setDraft={setDraft} sonatas={sonatas} onSubmit={save} />}
  </div>;
}

function EchoDetail({ echo, sonatas, onClose, onAdd }: { echo: Echo; sonatas: Sonata[]; onClose: () => void; onAdd: () => void }) {
  return <div className="modalBackdrop" onMouseDown={onClose}><div className="echoDetailModal" onMouseDown={(event) => event.stopPropagation()}>
    <div className="modalHeader"><div><span className="sectionLabel">{echo.class} · CUSTO {echo.cost}</span><h2>{echo.name}</h2></div><button onClick={onClose}>×</button></div>
    <div className="echoDetailHead">{image(echo.iconPath, echo.name)}<div><strong>Sonata Effects</strong><p>{sonataNames(echo, sonatas)}</p><strong>Localização</strong><p>{echo.place || "Não informada"}</p></div></div>
    <div className="echoSkill"><span className="sectionLabel">HABILIDADE DO ECHO</span><p>{echo.skill || "Detalhes da habilidade indisponíveis."}</p></div>
    <div className="modalActions"><button onClick={onClose}>Fechar</button><button className="primaryButton" onClick={onAdd}>＋ Registrar no inventário</button></div>
  </div></div>;
}

function OwnedEchoModal({ draft, setDraft, sonatas, onSubmit }: { draft: OwnedEcho; setDraft: (value?: OwnedEcho) => void; sonatas: Sonata[]; onSubmit: (event: FormEvent) => void }) {
  const substats = parseSubstats(draft.substatsJson);
  return <div className="modalBackdrop" onMouseDown={() => setDraft(undefined)}><form className="buildModal" onSubmit={onSubmit} onMouseDown={(event) => event.stopPropagation()}>
    <div className="modalHeader"><div><span className="sectionLabel">PEÇA DO INVENTÁRIO</span><h2>{draft.echoName}</h2></div><button type="button" onClick={() => setDraft(undefined)}>×</button></div>
    <div className="buildFormGrid">
      <label className="buildField">Nível<input type="number" min={0} max={25} value={draft.level} onChange={(event) => setDraft({ ...draft, level: Number(event.target.value) })} /></label>
      <label className="buildField">Atributo principal<input value={draft.mainStat} placeholder="Ex.: CRIT Rate 22%" onChange={(event) => setDraft({ ...draft, mainStat: event.target.value })} /></label>
      <label className="buildField wide">Sonata<select value={draft.sonataId ?? 0} onChange={(event) => setDraft({ ...draft, sonataId: Number(event.target.value) || undefined })}><option value={0}>Não definida</option>{sonatas.map((set) => <option key={set.id} value={set.id}>{set.name}</option>)}</select></label>
      {Array.from({ length: 5 }, (_, index) => <label className="buildField" key={index}>Subatributo {index + 1}<input value={substats[index] ?? ""} placeholder="Ex.: CRIT DMG 15%" onChange={(event) => { const next = [...substats]; next[index] = event.target.value; setDraft({ ...draft, substatsJson: JSON.stringify(next.filter(Boolean)) }); }} /></label>)}
      <label className="buildField wide">Notas<textarea rows={3} value={draft.note} onChange={(event) => setDraft({ ...draft, note: event.target.value })} /></label>
    </div>
    <div className="buildChecks"><label><input type="checkbox" checked={draft.favorite} onChange={(event) => setDraft({ ...draft, favorite: event.target.checked })} />Favorito</label><label><input type="checkbox" checked={draft.locked} onChange={(event) => setDraft({ ...draft, locked: event.target.checked })} />Bloqueado</label></div>
    <div className="modalActions"><button type="button" onClick={() => setDraft(undefined)}>Cancelar</button><button className="primaryButton">Salvar Echo</button></div>
  </form></div>;
}

function image(path: string, name: string) {
  return path.startsWith("/cache/") ? <img src={path} alt="" loading="lazy" /> : <span className="echoFallback">{name.slice(0, 1)}</span>;
}
function firstSonata(echo: Echo) { try { return (JSON.parse(echo.sonataIdsJson) as number[])[0]; } catch { return undefined; } }
function sonataNames(echo: Echo, sonatas: Sonata[]) { try { const ids = JSON.parse(echo.sonataIdsJson) as number[]; return ids.map((id) => sonatas.find((set) => set.id === id)?.name).filter(Boolean).join(" · ") || "Sem Sonata"; } catch { return "Sem Sonata"; } }
function parseSubstats(value: string) { try { const parsed = JSON.parse(value); return Array.isArray(parsed) ? parsed.map(String) : []; } catch { return []; } }
function Loading() { return <div className="skeletonGrid">{Array.from({ length: 8 }, (_, index) => <div className="skeleton weaponSkeleton" key={index} />)}</div>; }
function messageFrom(cause: unknown) { return cause instanceof Error ? cause.message : String(cause); }
