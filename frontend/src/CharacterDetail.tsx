import { useEffect, useMemo, useState, type CSSProperties, type FormEvent } from "react";
import { ChevronRight, CircleDot, Crosshair, GitBranch, Info, Music2, Orbit, Radio, Search, Shield, Sparkles, Swords, X, Zap } from "lucide-react";
import { calculateCharacterProgression } from "./lib/backend";
import { isRoverCharacter } from "./lib/characters";
import type { CharacterAccountUpdate, CharacterProfile, MaterialCost, ProgressionPlan } from "./types";

type DetailTab = "overview" | "kit" | "forte" | "tree" | "chains" | "materials" | "stats" | "lore";

export function CharacterDetail({ profile, onBack, onSaveAccount }: { profile: CharacterProfile; onBack: () => void; onSaveAccount: (update: CharacterAccountUpdate) => Promise<void> }) {
  const [tab, setTab] = useState<DetailTab>("overview");
  const [accountDraft, setAccountDraft] = useState<CharacterAccountUpdate>();
  const [savingAccount, setSavingAccount] = useState(false);
  const character = profile.character;
  const kitFocused = tab === "kit" || tab === "forte" || tab === "tree";

  function openAccountEditor() {
    setAccountDraft({
      characterId: character.id,
      owned: character.owned,
      level: character.level || 1,
      sequence: character.sequence,
      favorite: character.favorite
    });
  }

  async function submitAccount(event: FormEvent) {
    event.preventDefault();
    if (!accountDraft) return;
    setSavingAccount(true);
    try {
      await onSaveAccount(accountDraft);
      setAccountDraft(undefined);
    } finally {
      setSavingAccount(false);
    }
  }

  return <div className={`detailPage${kitFocused ? " kitFocused" : ""}`}>
    <button className="backButton" onClick={onBack}>← Voltar ao catálogo</button>

    <section className={`characterHero element-${character.elementCode}`}>
      {character.backgroundPath.startsWith("/cache/") && <img className="heroBackground" src={character.backgroundPath} alt="" />}
      <div className="heroArt" aria-hidden="true">
        {character.iconPath.startsWith("/cache/") ? <img src={character.iconPath} alt="" /> : initials(character.name)}
      </div>
      <div className="heroContent">
        <span className="eyebrow">{character.element} · {character.weaponType}</span>
        <h1>{character.name}</h1>
        {character.nickname && <p className="heroNickname">{character.nickname}</p>}
        <div className="heroTags">
          <span>{"◆".repeat(character.rarity)} <small>{character.rarity} ESTRELAS</small></span>
          {profile.region && <span>{profile.region}</span>}
          {profile.faction && <span>{profile.faction}</span>}
        </div>
        <p className="heroDescription">{profile.description || "Detalhes ainda não sincronizados para este personagem."}</p>
        <div className="heroActions">
          <button onClick={openAccountEditor}>{character.owned ? "Editar minha conta" : "Registrar na conta"}</button>
          <button className="primaryButton">Criar build</button>
          <button>Perguntar à IA</button>
        </div>
      </div>
      <aside className="heroMeta">
        <span>STATUS DA CONTA</span>
        <strong>{character.owned ? `Possuído · S${character.sequence}` : "Não registrado"}</strong>
        {character.owned && <><span>NÍVEL</span><strong>{character.level}</strong></>}
        {profile.birthday && <><span>ANIVERSÁRIO</span><strong>{profile.birthday}</strong></>}
        {profile.gender && <><span>GÊNERO</span><strong>{profile.gender}</strong></>}
      </aside>
    </section>

    <nav className="detailTabs" aria-label="Seções do personagem">
      <button className={tab === "overview" ? "active" : ""} onClick={() => setTab("overview")}>Visão geral</button>
      <button className={tab === "kit" || tab === "forte" || tab === "tree" ? "active" : ""} onClick={() => setTab("kit")}>Kit & Árvore <span>{profile.skills.length + (profile.extras?.skillTree?.length || 0)}</span></button>
      <button className={tab === "chains" ? "active" : ""} onClick={() => setTab("chains")}>Sequências <span>{profile.chains.length}</span></button>
      <button className={tab === "materials" ? "active" : ""} onClick={() => setTab("materials")}>Materiais <span>{profile.progression?.ascensions?.length || 0}</span></button>
      <button className={tab === "stats" ? "active" : ""} onClick={() => setTab("stats")}>Atributos</button>
      <button className={tab === "lore" ? "active" : ""} onClick={() => setTab("lore")}>Lore</button>
    </nav>

    {tab === "overview" && <Overview profile={profile} />}
    {kitFocused && <ResonanceKitTree profile={profile} />}
    {tab === "chains" && <Chains profile={profile} />}
    {tab === "materials" && <Materials profile={profile} />}
    {tab === "stats" && <Stats profile={profile} />}
    {tab === "lore" && <Lore profile={profile} />}
    {accountDraft && <div className="modalBackdrop" onMouseDown={() => setAccountDraft(undefined)}>
      <form className="accountModal" role="dialog" aria-modal="true" aria-labelledby="account-title" onSubmit={(event) => void submitAccount(event)} onMouseDown={(event) => event.stopPropagation()}>
        <div className="modalHeader"><div><span className="sectionLabel">CONTA LOCAL</span><h2 id="account-title">{character.name}</h2></div><button type="button" onClick={() => setAccountDraft(undefined)} aria-label="Fechar">×</button></div>
        <label className="ownershipToggle"><input type="checkbox" checked={accountDraft.owned} onChange={(event) => setAccountDraft({ ...accountDraft, owned: event.target.checked })} /><span><strong>Personagem possuído</strong><small>Inclui o personagem em builds e alternativas da conta.</small></span></label>
        <div className="accountFields">
          <label>Nível<input type="number" min={1} max={90} disabled={!accountDraft.owned} value={accountDraft.level} onChange={(event) => setAccountDraft({ ...accountDraft, level: Number(event.target.value) })} /></label>
          <label>Sequência<select disabled={!accountDraft.owned} value={accountDraft.sequence} onChange={(event) => setAccountDraft({ ...accountDraft, sequence: Number(event.target.value) })}>{Array.from({ length: 7 }, (_, sequence) => <option key={sequence} value={sequence}>S{sequence}</option>)}</select></label>
        </div>
        <label className="ownershipToggle compact"><input type="checkbox" checked={accountDraft.favorite} onChange={(event) => setAccountDraft({ ...accountDraft, favorite: event.target.checked })} /><span><strong>Favorito</strong><small>Pode ser favorito mesmo sem estar possuído.</small></span></label>
        <div className="modalActions"><button type="button" onClick={() => setAccountDraft(undefined)}>Cancelar</button><button className="primaryButton" disabled={savingAccount}>{savingAccount ? "Salvando…" : "Salvar na conta"}</button></div>
      </form>
    </div>}
  </div>;
}

export function Materials({ profile }: { profile: CharacterProfile }) {
  const skills = useMemo(() => (profile.progression?.skills || []).filter((skill) => skill.maxLevel > 1), [profile]);
  const [currentLevel, setCurrentLevel] = useState(1);
  const [targetLevel, setTargetLevel] = useState(90);
  const [currentSkills, setCurrentSkills] = useState<Record<string, number>>(() => Object.fromEntries(skills.map((skill) => [skill.nodeId, 1])));
  const [targetSkills, setTargetSkills] = useState<Record<string, number>>(() => Object.fromEntries(skills.map((skill) => [skill.nodeId, skill.maxLevel])));
  const [includeUnlocks, setIncludeUnlocks] = useState(true);
  const [plan, setPlan] = useState<ProgressionPlan>();
  const [error, setError] = useState("");
  const [calculating, setCalculating] = useState(false);

  async function calculate() {
    setCalculating(true);
    setError("");
    try {
      setPlan(await calculateCharacterProgression({
        characterId: profile.character.id, currentLevel, targetLevel,
        currentSkills, targetSkills, includeUnlocks
      }));
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : String(reason));
    } finally {
      setCalculating(false);
    }
  }

  useEffect(() => {
    const timer = window.setTimeout(() => void calculate(), 120);
    return () => window.clearTimeout(timer);
  }, [currentLevel, targetLevel, currentSkills, targetSkills, includeUnlocks]);

  if (!profile.progression?.ascensions?.length) return <DetailEmpty text="Materiais ainda não sincronizados." />;
  return <div className="materialsPlanner">
    <header className="materialsTitle">
      <div><h2>Materiais de ascensão</h2><p>Nv.{currentLevel} → Nv.{targetLevel}</p></div>
      {calculating && <span className="calculatingBadge">Atualizando…</span>}
    </header>
    <MaterialCards costs={plan?.ascensions || []} emptyText="Nenhum material de ascensão neste intervalo." />

    <section className="levelPlanner">
      <header><h3>Nível</h3><strong>Nv.{currentLevel} → Nv.{targetLevel}</strong></header>
      <label><span>De {currentLevel}</span><input type="range" min={1} max={90} value={currentLevel} onChange={(event) => setCurrentLevel(Math.min(Number(event.target.value), targetLevel))} /></label>
      <label><span>Até {targetLevel}</span><input type="range" min={1} max={90} value={targetLevel} onChange={(event) => setTargetLevel(Math.max(Number(event.target.value), currentLevel))} /></label>
      <div className="levelTicks">{[1, 20, 30, 40, 50, 60, 70, 80, 90].map((level) => <span key={level}>{level}</span>)}</div>
    </section>

    <section className="skillMaterialSection">
      <header><div><h2>Materiais de habilidade</h2><p>Forte e nós selecionados</p></div><label className="check"><input type="checkbox" checked={includeUnlocks} onChange={(event) => setIncludeUnlocks(event.target.checked)} />Incluir nós passivos</label></header>
      <MaterialCards costs={plan?.skills || []} emptyText="Nenhum material de habilidade neste intervalo." />
      <div className="skillLevelGrid">{skills.map((skill) => <article key={skill.nodeId}>
        <div><small>{skill.type}</small><strong>{skill.name}</strong></div>
        <label><span>De</span><select value={currentSkills[skill.nodeId]} onChange={(event) => setCurrentSkills({ ...currentSkills, [skill.nodeId]: Number(event.target.value) })}>{levels(skill.maxLevel).map((level) => <option key={level}>{level}</option>)}</select></label>
        <span>→</span>
        <label><span>Até</span><select value={targetSkills[skill.nodeId]} onChange={(event) => setTargetSkills({ ...targetSkills, [skill.nodeId]: Number(event.target.value) })}>{levels(skill.maxLevel).map((level) => <option key={level}>{level}</option>)}</select></label>
      </article>)}</div>
    </section>
    {plan && <footer className="materialsTotal"><span>Total combinado</span><strong>{plan.total.length} tipos de material</strong></footer>}
    {error && <p className="progressionError" role="alert">{error}</p>}
  </div>;
}

function MaterialCards({ costs, emptyText }: { costs: MaterialCost[]; emptyText: string }) {
  if (!costs.length) return <div className="materialCardsEmpty">{emptyText}</div>;
  return <div className="materialCards">{costs.map((cost) => <article className={`materialTile rarity-${cost.material.rarity}`} key={cost.material.id} title={cost.material.description}>
    <div>{cost.material.iconPath?.startsWith("/cache/") ? <img src={cost.material.iconPath} alt="" /> : <span>◇</span>}</div>
    <strong>x{formatQuantity(cost.quantity)}</strong>
    <small>{cost.material.name || `Item ${cost.material.id}`}</small>
  </article>)}</div>;
}

function levels(max: number) { return Array.from({ length: max }, (_, index) => index + 1); }
function formatQuantity(value: number) { return new Intl.NumberFormat("pt-BR").format(value); }

function Overview({ profile }: { profile: CharacterProfile }) {
  const rover = isRoverCharacter(profile.character);
  return <div className="overviewGrid">
    {!!profile.extras?.tags?.length && <section className="detailPanel roleTagsPanel">
      <span className="sectionLabel">FUNÇÕES OFICIAIS</span>
      <div className="roleTagGrid">{profile.extras.tags.map((tag) => <article key={tag.id} style={{ "--tag-color": `#${tag.color || "67dce3"}` } as CSSProperties}>
        <i /><div><strong>{tag.name}</strong><small>{tag.description}</small></div>
      </article>)}</div>
    </section>}
    <section className="detailPanel">
      <span className="sectionLabel">IDENTIDADE DO RESSONADOR</span>
      <h2>{profile.talentName || "Visão geral"}</h2>
      <p>{profile.talentDescription || profile.description || "Os detalhes serão preenchidos na próxima sincronização."}</p>
    </section>
    <section className="detailPanel weaponPanel">
      <span className="sectionLabel">{rover ? "ARMA RECOMENDADA" : "ARMA ASSINATURA"}</span>
      {profile.signatureWeapon?.name ? <>
        <div className="weaponHeading"><div className="weaponGlyph" aria-hidden="true">{profile.signatureWeapon.iconPath.startsWith("/cache/") ? <img src={profile.signatureWeapon.iconPath} alt="" /> : "◇"}</div><div><h2>{profile.signatureWeapon.name}</h2><small>{profile.signatureWeapon.type} · {"◆".repeat(profile.signatureWeapon.rarity)}</small></div></div>
        <h3>{profile.signatureWeapon.effectName}</h3>
        <p>{profile.signatureWeapon.effect}</p>
      </> : <p>{rover ? "Nenhuma arma recomendada foi encontrada nos dados sincronizados." : "Nenhuma arma assinatura foi encontrada nos dados sincronizados."}</p>}
    </section>
    <section className="detailPanel quickKit">
      <span className="sectionLabel">KIT EM RESUMO</span>
      {profile.skills.slice(0, 6).map((skill) => <div key={skill.nodeId}><i /><span><strong>{skill.type || "Habilidade"}</strong><small>{skill.name}</small></span></div>)}
    </section>
  </div>;
}

function LegacyUnifiedKitTree({ profile }: { profile: CharacterProfile }) {
  const nodes = profile.extras?.skillTree || [];
  const branches = profile.extras?.skillBranches || [];
  const forte = profile.extras?.forte;
  const byID = useMemo(() => new Map(profile.skills.map((skill) => [skill.nodeId, skill])), [profile.skills]);

  const defaultId = nodes[0]?.nodeId || profile.skills[0]?.nodeId || "1";
  const [selectedNodeId, setSelectedNodeId] = useState<string>(defaultId);
  const [openMultiplier, setOpenMultiplier] = useState<boolean>(true);

  if (!profile.skills.length && !nodes.length) {
    return <DetailEmpty text="Nenhum dado de Kit ou Árvore sincronizado." />;
  }

  const selectedNode = nodes.find((n) => n.nodeId === selectedNodeId);
  const selectedSkill = byID.get(selectedNodeId) || profile.skills.find((s) => s.nodeId === selectedNodeId);
  const progression = profile.progression?.skills?.find((entry) => entry.nodeId === selectedNodeId);
  const isForteCircuit = selectedNode?.nodeType === 1 || selectedSkill?.name.toLowerCase().includes("forte") || selectedSkill?.type?.toLowerCase().includes("forte");

  return <div className="unifiedKitTree">
    {!!branches.length && <section className="branchSection" style={{ marginTop: 0, marginBottom: 24 }}>
      <span className="sectionLabel">RAMIFICAÇÕES ESPECIAIS (MODOS DE RESSONÂNCIA)</span>
      <div>{branches.map((branch) => <article key={branch.id}><small>#{branch.id}</small><h3>{branch.name}</h3><FormattedText text={branch.description} /></article>)}</div>
    </section>}

    <section className="treePage" style={{ marginTop: 12 }}>
      <header>
        <div>
          <span className="sectionLabel">ÁRVORE INTERATIVA DE MAESTRIA</span>
          <h2>Seletor de Nós e Dependências</h2>
        </div>
        <span>{nodes.length || profile.skills.length} nós · Clique em qualquer quadrado para inspecionar</span>
      </header>
      <div className="skillTreeGrid">{nodes.map((node) => {
        const skill = byID.get(node.nodeId);
        const isSelected = node.nodeId === selectedNodeId;
        return <article
          className={`treeNode nodeType-${node.nodeType}`}
          key={node.nodeId}
          onClick={() => { setSelectedNodeId(node.nodeId); setOpenMultiplier(true); }}
          style={{
            cursor: "pointer",
            borderColor: isSelected ? "var(--cyan)" : undefined,
            background: isSelected ? "rgba(103, 220, 227, 0.08)" : undefined,
            boxShadow: isSelected ? "0 0 14px rgba(103, 220, 227, 0.2)" : undefined,
            transition: "all 0.15s ease"
          }}
        >
          <div className="treeNodeIndex" style={{ borderColor: isSelected ? "var(--cyan)" : undefined, background: isSelected ? "var(--cyan)" : undefined, color: isSelected ? "#090e13" : undefined, fontWeight: isSelected ? "bold" : undefined }}>{node.nodeId}</div>
          <div>
            <small>{skill?.type || nodeTypeName(node.nodeType)}</small>
            <strong style={{ color: isSelected ? "var(--cyan)" : undefined }}>{skill?.name || `Nó ${node.nodeId}`}</strong>
            {!!node.parentNodes?.length && <span>Depende de {node.parentNodes.join(", ")}</span>}
          </div>
          {!!node.branchIds?.length && <b>{node.branchIds.length} modos</b>}
        </article>;
      })}</div>
    </section>

    {selectedSkill ? (
      <section className="skillList" style={{ marginTop: 28 }}>
        <header style={{ marginBottom: 14, display: "flex", justifyContent: "space-between", alignItems: "baseline", borderBottom: "1px solid var(--border)", paddingBottom: 8 }}>
          <div>
            <span className="sectionLabel" style={{ color: "var(--cyan)" }}>INSPETOR DE MAESTRIA · NÓ SELECIONADO #{selectedNodeId}</span>
            <h2 style={{ fontSize: 20, margin: "4px 0 0 0" }}>{selectedSkill.name}</h2>
          </div>
          <small style={{ color: "var(--muted)", font: "11px Consolas, monospace" }}>{selectedSkill.type || nodeTypeName(selectedNode?.nodeType || 2)}</small>
        </header>

        <article className={openMultiplier ? "skillCard expanded" : "skillCard"} style={{ borderColor: "rgba(103, 220, 227, 0.4)", background: "rgba(14, 19, 27, 0.7)" }}>
          <div className="skillIndex" style={{ borderColor: "var(--cyan)", color: "var(--cyan)" }}>{selectedNodeId}</div>
          <div>
            <span className="sectionLabel">{selectedSkill.type || "HABILIDADE"}</span>
            <h2>{selectedSkill.name}</h2>
            <FormattedText text={selectedSkill.description} />

            {isForteCircuit && forte && (forte.actions?.length || 0) > 0 && (
              <div className="forteActionsInline" style={{ marginTop: 18, paddingTop: 16, borderTop: "1px solid var(--border)" }}>
                <span className="sectionLabel" style={{ display: "block", marginBottom: 12 }}>ENTRADAS E COMBOS DE ROTAÇÃO (GUIA DO FORTE)</span>
                <div className="forteActions" style={{ gap: 8 }}>
                  {(forte.actions || []).map((action, idx) => (
                    <article key={`${action.name}-${idx}`} style={{ padding: 12 }}>
                      <header><span>{action.name || `Etapa ${idx + 1}`}</span><div>{(action.inputs || []).map((input) => <kbd key={input}>{input}</kbd>)}</div></header>
                      <FormattedText text={action.description} />
                    </article>
                  ))}
                </div>
              </div>
            )}
          </div>

          <button
            disabled={!progression?.values?.length}
            aria-expanded={openMultiplier}
            aria-label={`Ver multiplicadores de ${selectedSkill.name}`}
            onClick={() => setOpenMultiplier(!openMultiplier)}
          >
            MULTIPLICADORES <span>{openMultiplier ? "⌄" : "›"}</span>
          </button>

          {openMultiplier && progression && (
            <div className="skillValues">
              <div className="skillValueHeader"><span>Escala oficial</span>{levels(progression.maxLevel).map((level) => <b key={level}>Nv. {level}</b>)}</div>
              {progression.values.map((row) => (
                <div className="skillValueRow" key={row.name}>
                  <strong>{row.name}</strong>
                  {levels(progression.maxLevel).map((level) => <span key={level}>{row.values[level - 1] || "—"}</span>)}
                </div>
              ))}
            </div>
          )}
        </article>
      </section>
    ) : (
      <div className="emptyState detailEmpty" style={{ marginTop: 28, padding: 32 }}>
        <div className="emptyGlyph">◇</div>
        <h2>Selecione um nó acima</h2>
        <p>Clique em qualquer quadrado do grid da árvore para visualizar a descrição completa e a tabela de multiplicadores.</p>
      </div>
    )}
  </div>;
}

function ResonanceKitTree({ profile }: { profile: CharacterProfile }) {
  const nodes = profile.extras?.skillTree || [];
  const branches = profile.extras?.skillBranches || [];
  const forte = profile.extras?.forte;
  const byNodeID = useMemo(() => new Map(nodes.map((node) => [node.nodeId, node])), [nodes]);
  const defaultSkill = profile.skills.find((skill) => byNodeID.get(skill.nodeId)?.nodeType === 1) || profile.skills[0];
  const [selectedBranchID, setSelectedBranchID] = useState<number | undefined>(branches[0]?.id);
  const [expandedSkillID, setExpandedSkillID] = useState(defaultSkill?.nodeId || "");
  const [query, setQuery] = useState("");
  const [skillLevels, setSkillLevels] = useState<Record<string, number>>({});

  if (!profile.skills.length) {
    return <DetailEmpty text="Nenhum dado de Kit ou Árvore sincronizado." />;
  }

  const normalizedQuery = query.trim().toLocaleLowerCase();
  const visibleSkills = profile.skills.filter((skill) => {
    const node = byNodeID.get(skill.nodeId);
    if (node?.nodeType === 4) return false;
    const matchesBranch = !selectedBranchID || !node?.branchIds.length || node.branchIds.includes(selectedBranchID);
    const matchesSearch = !normalizedQuery || `${skill.name} ${skill.type} ${skill.nodeId}`.toLocaleLowerCase().includes(normalizedQuery);
    return matchesBranch && matchesSearch;
  });

  function progressionFor(nodeID: string) {
    return profile.progression?.skills?.find((entry) => entry.nodeId === nodeID);
  }

  function levelFor(nodeID: string) {
    const maxLevel = Math.max(1, progressionFor(nodeID)?.maxLevel || 1);
    return Math.min(skillLevels[nodeID] || maxLevel, maxLevel);
  }

  function setSkillLevel(nodeID: string, level: number) {
    setSkillLevels((current) => ({ ...current, [nodeID]: level }));
  }

  function bonusesFor(nodeID: string) {
    return nodes
      .filter((node) => node.nodeType === 4 && node.parentNodes.includes(Number(nodeID)))
      .map((node) => profile.skills.find((skill) => skill.nodeId === node.nodeId))
      .filter((skill): skill is CharacterProfile["skills"][number] => Boolean(skill));
  }

  return <div className="nanokaKit">
    <header className="nanokaKitHeader">
      <div>
        <span className="sectionLabel">DADOS OFICIAIS · {profile.character.gameVersion}</span>
        <h2>Progressão de habilidades</h2>
        <p>Níveis, bônus e descrições do perfil sincronizado.</p>
      </div>
      <label className="nanokaKitSearch">
        <Search size={17} aria-hidden="true" />
        <span className="srOnly">Buscar habilidade</span>
        <input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Buscar habilidade…" />
      </label>
    </header>

    {!!branches.length && <div className="nanokaKitModes" role="tablist" aria-label="Modos de ressonância">
      {branches.map((branch) => <button
        role="tab"
        aria-selected={branch.id === selectedBranchID}
        className={branch.id === selectedBranchID ? "active" : ""}
        key={branch.id}
        onClick={() => setSelectedBranchID(branch.id)}
      >
        {branch.iconPath?.startsWith("/cache/") && <img src={branch.iconPath} alt="" />}
        <span><small>#{branch.id}</small><strong>{branch.name}</strong></span>
      </button>)}
    </div>}

    <section className="nanokaProgression" aria-labelledby="skill-progression-title">
      <header>
        <div><span className="sectionLabel">COMBAT SKILLS</span><h3 id="skill-progression-title">Níveis do Kit</h3></div>
        <span>{visibleSkills.length} habilidades</span>
      </header>
      <div className="nanokaProgressionList">
        {visibleSkills.map((skill) => {
          const node = byNodeID.get(skill.nodeId);
          const progression = progressionFor(skill.nodeId);
          const maxLevel = Math.max(1, progression?.maxLevel || 1);
          const level = levelFor(skill.nodeId);
          const bonuses = bonusesFor(skill.nodeId);
          return <article key={skill.nodeId}>
            <button className="nanokaProgressionIdentity" onClick={() => setExpandedSkillID(skill.nodeId)}>
              <span className="nanokaSkillIcon" aria-hidden="true">
                {skill.iconPath?.startsWith("/cache/")
                  ? <img src={skill.iconPath} alt="" />
                  : <SkillNodeIcon type={skill.type} nodeType={node?.nodeType || 0} size={22} />}
              </span>
              <span><strong>{skill.name}</strong><small>{skill.type}</small></span>
              <b>{maxLevel > 1 ? `1 → ${level}` : "Passiva"}</b>
            </button>
            {maxLevel > 1 && <label className="nanokaLevelControl">
              <span>Do nível 1 ao {level}</span>
              <input
                type="range"
                min={1}
                max={maxLevel}
                value={level}
                onChange={(event) => setSkillLevel(skill.nodeId, Number(event.target.value))}
                aria-label={`Nível de ${skill.name}`}
              />
              <small><span>1</span><span>{maxLevel}</span></small>
            </label>}
            {!!bonuses.length && <div className="nanokaBonusChips" aria-label={`Bônus de ${skill.name}`}>
              <span>STAT BONUS</span>
              <div>{bonuses.map((bonus) => <button key={bonus.nodeId} onClick={() => setExpandedSkillID(bonus.nodeId)}>
                <small>{bonus.nodeId}</small><strong>{bonus.name}</strong>
              </button>)}</div>
            </div>}
          </article>;
        })}
        {!visibleSkills.length && <p className="nanokaNoResults">Nenhuma habilidade corresponde aos filtros atuais.</p>}
      </div>
    </section>

    <section className="nanokaSkillLibrary" aria-labelledby="combat-skills-title">
      <header>
        <div><span className="sectionLabel">DETALHES SINCRONIZADOS</span><h3 id="combat-skills-title">Habilidades de combate</h3></div>
        <span>Selecione um card para expandir</span>
      </header>
      <div className="nanokaSkillGrid">
        {visibleSkills.map((skill) => {
          const node = byNodeID.get(skill.nodeId);
          const progression = progressionFor(skill.nodeId);
          const maxLevel = Math.max(1, progression?.maxLevel || 1);
          const level = levelFor(skill.nodeId);
          const expanded = skill.nodeId === expandedSkillID;
          const bonuses = bonusesFor(skill.nodeId);
          const isForte = node?.nodeType === 1 || skill.type.toLocaleLowerCase().includes("forte");
          return <article className={`nanokaSkillCard${expanded ? " expanded" : ""}`} key={skill.nodeId}>
            <button className="nanokaSkillCardHeader" onClick={() => setExpandedSkillID(expanded ? "" : skill.nodeId)} aria-expanded={expanded}>
              <span className="nanokaSkillIcon" aria-hidden="true">
                {skill.iconPath?.startsWith("/cache/")
                  ? <img src={skill.iconPath} alt="" />
                  : <SkillNodeIcon type={skill.type} nodeType={node?.nodeType || 0} size={25} />}
              </span>
              <span><strong>{skill.name}</strong><small>{skill.type}</small></span>
              <ChevronRight size={18} aria-hidden="true" />
            </button>
            {expanded && <div className="nanokaSkillCardBody">
              <section>
                <span className="sectionLabel">DESCRIÇÃO OFICIAL</span>
                {skill.description ? <FormattedText text={skill.description} /> : <p>Não disponível nos dados sincronizados.</p>}
              </section>

              {!!bonuses.length && <section className="nanokaCardBonuses">
                <span className="sectionLabel">STAT BONUS</span>
                <div>{bonuses.map((bonus) => <article key={bonus.nodeId}>
                  <SkillNodeIcon type={bonus.type} nodeType={4} size={19} />
                  <span><strong>{bonus.name}</strong><small>{bonus.description || "Valor não disponível."}</small></span>
                </article>)}</div>
              </section>}

              <section className="nanokaOfficialValues">
                <span className="sectionLabel">SKILL ATTRIBUTES{maxLevel > 1 ? ` · NV.${level}` : ""}</span>
                {progression?.values.length ? <div>{progression.values.map((row) => <article key={row.name}>
                  <span>{row.name}</span><strong>{row.values[level - 1] || "Não disponível"}</strong>
                </article>)}</div> : <p>Valores numéricos não disponíveis nos dados sincronizados.</p>}
              </section>

              {!!node?.branchIds.length && <section className="nanokaLinkedModes">
                <span className="sectionLabel">MODOS VINCULADOS</span>
                <div>{node.branchIds.map((branchID) => {
                  const branch = branches.find((item) => item.id === branchID);
                  return <button key={branchID} onClick={() => setSelectedBranchID(branchID)}>
                    <small>#{branchID}</small><strong>{branch?.name || "Não disponível"}</strong>
                  </button>;
                })}</div>
              </section>}

              {isForte && forte && forte.actions.length > 0 && <section className="nanokaForteActions">
                <span className="sectionLabel">SKILL INPUT</span>
                <div>{forte.actions.map((action, index) => <article key={`${action.name}-${index}`}>
                  <header><strong>{action.name || `Etapa ${index + 1}`}</strong><div>{action.inputs.map((input) => <kbd key={input}>{input}</kbd>)}</div></header>
                  {action.description && <FormattedText text={action.description} />}
                </article>)}</div>
              </section>}
            </div>}
          </article>;
        })}
      </div>
    </section>
  </div>;
}

function LegacyResonanceKitTree({ profile }: { profile: CharacterProfile }) {
  const nodes = profile.extras?.skillTree || [];
  const branches = profile.extras?.skillBranches || [];
  const forte = profile.extras?.forte;
  const byID = useMemo(() => new Map(profile.skills.map((skill) => [skill.nodeId, skill])), [profile.skills]);
  const defaultID = nodes.find((node) => node.nodeType === 1)?.nodeId || nodes[0]?.nodeId || profile.skills[0]?.nodeId || "1";
  const [selectedNodeID, setSelectedNodeID] = useState(defaultID);
  const [selectedBranchID, setSelectedBranchID] = useState<number | undefined>(branches[0]?.id);
  const [selectedLevel, setSelectedLevel] = useState(1);
  const [query, setQuery] = useState("");

  if (!profile.skills.length && !nodes.length) {
    return <DetailEmpty text="Nenhum dado de Kit ou Árvore sincronizado." />;
  }

  const visualNodes = nodes.length ? nodes : profile.skills.map((skill) => ({
    nodeId: skill.nodeId,
    nodeType: 2,
    coordinate: skill.sortOrder,
    parentNodes: [] as number[],
    branchIds: [] as number[],
    unlockCondition: 0
  }));
  const selectedNode = visualNodes.find((node) => node.nodeId === selectedNodeID);
  const selectedSkill = byID.get(selectedNodeID);
  const progression = profile.progression?.skills?.find((entry) => entry.nodeId === selectedNodeID);
  const maxLevel = Math.max(1, progression?.maxLevel || 1);
  const level = Math.min(selectedLevel, maxLevel);
  const normalizedQuery = query.trim().toLocaleLowerCase();
  const filteredNodes = visualNodes.filter((node) => {
    const skill = byID.get(node.nodeId);
    return !normalizedQuery || `${node.nodeId} ${skill?.name || ""} ${skill?.type || nodeTypeName(node.nodeType)}`
      .toLocaleLowerCase().includes(normalizedQuery);
  });
  const activeBranch = branches.find((branch) => branch.id === selectedBranchID);
  const parentNodes = (selectedNode?.parentNodes || []).map((parentID) => {
    const parent = visualNodes.find((node) => Number(node.nodeId) === parentID);
    return { id: String(parentID), name: parent ? byID.get(parent.nodeId)?.name : undefined };
  });
  const isForteCircuit = selectedNode?.nodeType === 1
    || selectedSkill?.name.toLocaleLowerCase().includes("forte")
    || selectedSkill?.type.toLocaleLowerCase().includes("forte");

  const ringGroups = [
    visualNodes.filter((node) => node.nodeType === 2),
    visualNodes.filter((node) => node.nodeType === 1 || node.nodeType === 3),
    visualNodes.filter((node) => node.nodeType === 4)
  ];
  const assignedNodes = new Set(ringGroups.flat().map((node) => node.nodeId));
  const unassignedNodes = visualNodes.filter((node) => !assignedNodes.has(node.nodeId));
  ringGroups[1] = [...ringGroups[1], ...unassignedNodes];

  function selectNode(nodeID: string) {
    setSelectedNodeID(nodeID);
    setSelectedLevel(1);
  }

  return <div className="resonanceWorkspace resonanceWorkspacePremium">
    <aside className="resonanceNavigator" aria-label="Navegador do Kit e Árvore">
      <header><span className="sectionLabel">NAVEGADOR</span><span>{visualNodes.length} nós</span></header>
      <label>
        <Search size={15} aria-hidden="true" />
        <span className="srOnly">Buscar habilidade ou nó</span>
        <input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Buscar habilidade…" />
      </label>
      <div className="resonanceNodeList">
        {filteredNodes.map((node) => {
          const skill = byID.get(node.nodeId);
          return <button className={node.nodeId === selectedNodeID ? "selected" : ""} key={node.nodeId} onClick={() => selectNode(node.nodeId)}>
            <span>{node.nodeId}</span>
            <div><strong>{skill?.name || `Nó ${node.nodeId}`}</strong><small>{skill?.type || nodeTypeName(node.nodeType)}</small></div>
            {!!node.branchIds.length && <b>{node.branchIds.length} {node.branchIds.length === 1 ? "modo" : "modos"}</b>}
          </button>;
        })}
        {!filteredNodes.length && <p>Nenhum nó corresponde à busca.</p>}
      </div>
      <footer>
        <span><i className="ringKey active" />Ativa</span>
        <span><i className="ringKey passive" />Passiva</span>
        <span><i className="ringKey attribute" />Atributo</span>
      </footer>
    </aside>

    <section className="resonanceLensPanel">
      {!!branches.length && <div className="resonanceModes" role="tablist" aria-label="Modos de ressonância">
        {branches.map((branch) => <button
          role="tab"
          aria-selected={branch.id === selectedBranchID}
          className={branch.id === selectedBranchID ? "active" : ""}
          key={branch.id}
          onClick={() => setSelectedBranchID(branch.id)}
          title={branch.description || undefined}
        >
          {branch.iconPath?.startsWith("/cache/") && <img src={branch.iconPath} alt="" />}
          <span><small>#{branch.id}</small><strong>{branch.name}</strong></span>
        </button>)}
      </div>}

      <header className="resonanceLensHeader">
        <div><span className="sectionLabel">MAPA DE RESSONÂNCIA</span><h2>Kit e dependências</h2></div>
        <span>{activeBranch?.name || `${visualNodes.length} nós sincronizados`}</span>
      </header>

      <div className="resonanceLens" aria-label="Mapa radial de habilidades">
        <div className="lensRing lensRingOuter" aria-hidden="true" />
        <div className="lensRing lensRingMiddle" aria-hidden="true" />
        <div className="lensRing lensRingInner" aria-hidden="true" />
        <div className="lensCore" aria-hidden="true"><GitBranch size={22} /></div>
        {ringGroups.map((ring, ringIndex) => ring.map((node, index) => {
          const skill = byID.get(node.nodeId);
          const branchMatch = !selectedBranchID || !node.branchIds.length || node.branchIds.includes(selectedBranchID);
          const angle = -90 + (360 / Math.max(ring.length, 1)) * index;
          return <button
            className={`lensNode lensRingNode-${ringIndex} nodeType-${node.nodeType}${node.nodeId === selectedNodeID ? " selected" : ""}${branchMatch ? "" : " muted"}`}
            style={{ "--node-angle": `${angle}deg` } as CSSProperties}
            key={node.nodeId}
            onClick={() => selectNode(node.nodeId)}
            aria-label={`${skill?.name || `Nó ${node.nodeId}`}${node.nodeId === selectedNodeID ? ", selecionado" : ""}`}
          >
            <span>{node.nodeId}</span>
            <strong>{skill?.name || `Nó ${node.nodeId}`}</strong>
            {!!node.branchIds.length && <small>{node.branchIds.length} {node.branchIds.length === 1 ? "modo" : "modos"}</small>}
          </button>;
        }))}
      </div>
    </section>

    <aside className="resonanceInspector" aria-label="Dados do nó selecionado">
      {selectedSkill ? <>
        <header>
          <div className="inspectorNodeIcon">
            {selectedSkill.iconPath?.startsWith("/cache/")
              ? <img src={selectedSkill.iconPath} alt="" />
              : <span>{selectedNodeID}</span>}
          </div>
          <div><span className="sectionLabel">{selectedSkill.type || nodeTypeName(selectedNode?.nodeType || 2)}</span><h2>{selectedSkill.name}</h2></div>
          {!!selectedNode?.branchIds.length && <b>{selectedNode.branchIds.length} {selectedNode.branchIds.length === 1 ? "modo" : "modos"}</b>}
        </header>

        <section className="inspectorDescription">
          <span className="sectionLabel">DESCRIÇÃO OFICIAL</span>
          {selectedSkill.description
            ? <FormattedText text={selectedSkill.description} />
            : <p>Descrição não disponível nos dados sincronizados.</p>}
        </section>

        <section className="dependencyPath">
          <span className="sectionLabel">DEPENDÊNCIAS</span>
          <div>
            {parentNodes.length ? parentNodes.map((parent) => <button key={parent.id} onClick={() => selectNode(parent.id)}>
              <span>{parent.id}</span><strong>{parent.name || `Nó ${parent.id}`}</strong><ChevronRight size={14} />
            </button>) : <p>Sem dependência registrada para este nó.</p>}
            {!!parentNodes.length && <div className="currentDependency"><span>{selectedNodeID}</span><strong>{selectedSkill.name}</strong></div>}
          </div>
        </section>

        {!!selectedNode?.branchIds.length && <section className="nodeModes">
          <span className="sectionLabel">MODOS VINCULADOS</span>
          <div>{selectedNode.branchIds.map((branchID) => {
            const branch = branches.find((item) => item.id === branchID);
            return <button className={branchID === selectedBranchID ? "active" : ""} key={branchID} onClick={() => setSelectedBranchID(branchID)}>
              <span>#{branchID}</span><strong>{branch?.name || `Modo ${branchID}`}</strong>
            </button>;
          })}</div>
        </section>}

        {maxLevel > 1 && <section className="inspectorLevels">
          <span className="sectionLabel">NÍVEL</span>
          <div>{levels(maxLevel).map((value) => <button className={value === level ? "active" : ""} key={value} onClick={() => setSelectedLevel(value)}>{value}</button>)}</div>
        </section>}

        <section className="officialValues">
          <header><span className="sectionLabel">VALORES OFICIAIS{maxLevel > 1 ? ` · NV. ${level}` : ""}</span><Info size={14} aria-hidden="true" /></header>
          {progression?.values.length ? <div>{progression.values.map((row) => <article key={row.name}>
            <span>{row.name}</span><strong>{row.values[level - 1] || "Não disponível"}</strong>
          </article>)}</div> : <p>Valores numéricos não disponíveis nos dados sincronizados.</p>}
        </section>

        {isForteCircuit && forte && forte.actions.length > 0 && <section className="inspectorForte">
          <span className="sectionLabel">GUIA DO FORTE</span>
          <div>{forte.actions.map((action, index) => <article key={`${action.name}-${index}`}>
            <header><strong>{action.name || `Etapa ${index + 1}`}</strong><div>{action.inputs.map((input) => <kbd key={input}>{input}</kbd>)}</div></header>
            {action.description && <FormattedText text={action.description} />}
          </article>)}</div>
        </section>}
      </> : <div className="inspectorEmpty"><GitBranch size={30} /><strong>Nó não disponível</strong><p>Selecione um nó com dados sincronizados.</p></div>}
    </aside>
  </div>;
}

function ForteMeridianKitTree({ profile }: { profile: CharacterProfile }) {
  const nodes = profile.extras?.skillTree || [];
  const branches = profile.extras?.skillBranches || [];
  const byID = useMemo(() => new Map(profile.skills.map((skill) => [skill.nodeId, skill])), [profile.skills]);
  const defaultID = nodes.find((node) => node.nodeType === 1)?.nodeId || nodes[0]?.nodeId || profile.skills[0]?.nodeId || "1";
  const [selectedNodeID, setSelectedNodeID] = useState(defaultID);
  const [selectedBranchID, setSelectedBranchID] = useState<number | undefined>(branches[0]?.id);
  const [query, setQuery] = useState("");
  const [searchOpen, setSearchOpen] = useState(false);

  if (!profile.skills.length && !nodes.length) {
    return <DetailEmpty text="Nenhum dado de Kit ou Árvore sincronizado." />;
  }

  const visualNodes = nodes.length ? nodes : profile.skills.map((skill) => ({
    nodeId: skill.nodeId,
    nodeType: 2,
    coordinate: skill.sortOrder,
    parentNodes: [] as number[],
    branchIds: [] as number[],
    unlockCondition: 0
  }));
  const selectedNode = visualNodes.find((node) => node.nodeId === selectedNodeID);
  const selectedSkill = byID.get(selectedNodeID);
  const normalizedQuery = query.trim().toLocaleLowerCase();
  const activeRoots = visualNodes
    .filter((node) => node.nodeType === 2)
    .sort((left, right) => left.coordinate - right.coordinate);
  const forteRoot = visualNodes.find((node) => node.nodeType === 1);
  const roots = forteRoot
    ? [...activeRoots.slice(0, 2), forteRoot, ...activeRoots.slice(2)]
    : activeRoots;
  const rootChains = roots.map((root) => {
    const chain = [root];
    const seen = new Set([root.nodeId]);
    let current = root;
    while (chain.length < 4) {
      const next = visualNodes
        .filter((node) => node.parentNodes.includes(Number(current.nodeId)) && !seen.has(node.nodeId))
        .sort((left, right) => left.coordinate - right.coordinate)[0];
      if (!next) break;
      chain.push(next);
      seen.add(next.nodeId);
      current = next;
    }
    return chain.reverse();
  });
  const placedIDs = new Set(rootChains.flat().map((node) => node.nodeId));
  const modeNodes = visualNodes.filter((node) => !placedIDs.has(node.nodeId) && node.branchIds.length > 0);
  modeNodes.forEach((node) => placedIDs.add(node.nodeId));
  const remainingNodes = visualNodes.filter((node) => !placedIDs.has(node.nodeId));

  function selectNode(nodeID: string) {
    setSelectedNodeID(nodeID);
  }

  function matchesQuery(nodeID: string) {
    if (!normalizedQuery) return true;
    const skill = byID.get(nodeID);
    return `${nodeID} ${skill?.name || ""} ${skill?.type || ""}`.toLocaleLowerCase().includes(normalizedQuery);
  }

  return <div className="forteMeridian">
    <aside className="meridianInspector" aria-label="Dados do nó selecionado">
      {selectedSkill ? <>
        <div className="meridianEmblem" aria-hidden="true">
          <SkillNodeIcon type={selectedSkill.type} nodeType={selectedNode?.nodeType || 0} size={38} />
        </div>
        <span className="meridianType">{selectedSkill.type || nodeTypeName(selectedNode?.nodeType || 0)}</span>
        <h2>{selectedSkill.name}</h2>

        <section className="meridianModeDetails">
          <h3>Modos</h3>
          {selectedNode?.branchIds.length ? selectedNode.branchIds.map((branchID) => {
            const branch = branches.find((entry) => entry.id === branchID);
            return <button
              className={branchID === selectedBranchID ? "active" : ""}
              key={branchID}
              onClick={() => setSelectedBranchID(branchID)}
            >
              <Sparkles size={18} aria-hidden="true" />
              <span><small>#{branchID}</small><strong>{branch?.name || "Não disponível"}</strong></span>
            </button>;
          }) : <p>Não disponível</p>}
        </section>

        <section className="meridianDescription">
          <h3>Descrição oficial</h3>
          {selectedSkill.description ? <FormattedText text={selectedSkill.description} /> : <p>Não disponível</p>}
        </section>
      </> : <p>Não disponível</p>}
    </aside>

    <section className="meridianCanvas" aria-label="Árvore de habilidades">
      <header className="meridianToolbar">
        <div className="meridianModes" role="tablist" aria-label="Modos de ressonância">
          {branches.map((branch) => <button
            role="tab"
            aria-selected={branch.id === selectedBranchID}
            className={branch.id === selectedBranchID ? "active" : ""}
            key={branch.id}
            onClick={() => setSelectedBranchID(branch.id)}
          >
            <small>#{branch.id}</small>
            <strong>{branch.name}</strong>
          </button>)}
        </div>
        <div className={`meridianSearch ${searchOpen ? "open" : ""}`}>
          {searchOpen && <input
            autoFocus
            aria-label="Buscar habilidade ou nó"
            placeholder="Buscar nó…"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
          />}
          <button
            aria-label={searchOpen ? "Fechar busca" : "Buscar habilidade ou nó"}
            onClick={() => { setSearchOpen(!searchOpen); if (searchOpen) setQuery(""); }}
          >
            {searchOpen ? <X size={18} /> : <Search size={18} />}
          </button>
          <span>{visualNodes.length} nós</span>
        </div>
      </header>

      <div className="meridianForest">
        {rootChains.map((chain, columnIndex) => <div
          className={`meridianBranch ${chain.some((node) => node.nodeType === 1) ? "forteBranch" : ""}`}
          key={chain[chain.length - 1]?.nodeId || columnIndex}
        >
          {chain.map((node, index) => <ForestNode
            key={node.nodeId}
            node={node}
            skill={byID.get(node.nodeId)}
            selected={node.nodeId === selectedNodeID}
            branchMuted={!!selectedBranchID && node.branchIds.length > 0 && !node.branchIds.includes(selectedBranchID)}
            queryMuted={!matchesQuery(node.nodeId)}
            connected={index < chain.length - 1}
            onSelect={() => selectNode(node.nodeId)}
          />)}
        </div>)}
      </div>

      {!!modeNodes.length && <div className="meridianModeNodes" aria-label="Nós vinculados a modos">
        {modeNodes.map((node) => <ForestNode
          key={node.nodeId}
          node={node}
          skill={byID.get(node.nodeId)}
          selected={node.nodeId === selectedNodeID}
          branchMuted={!!selectedBranchID && !node.branchIds.includes(selectedBranchID)}
          queryMuted={!matchesQuery(node.nodeId)}
          onSelect={() => selectNode(node.nodeId)}
        />)}
      </div>}

      {!!remainingNodes.length && <div className="meridianRemaining" aria-label="Outros nós sincronizados">
        {remainingNodes.map((node) => <ForestNode
          key={node.nodeId}
          node={node}
          skill={byID.get(node.nodeId)}
          selected={node.nodeId === selectedNodeID}
          branchMuted={!!selectedBranchID && node.branchIds.length > 0 && !node.branchIds.includes(selectedBranchID)}
          queryMuted={!matchesQuery(node.nodeId)}
          onSelect={() => selectNode(node.nodeId)}
        />)}
      </div>}
    </section>
  </div>;
}

function ForestNode({
  node,
  skill,
  selected,
  branchMuted,
  queryMuted,
  connected = false,
  onSelect
}: {
  node: CharacterProfile["extras"]["skillTree"][number];
  skill?: CharacterProfile["skills"][number];
  selected: boolean;
  branchMuted: boolean;
  queryMuted: boolean;
  connected?: boolean;
  onSelect: () => void;
}) {
  return <article className={`forestNode${selected ? " selected" : ""}${branchMuted ? " branchMuted" : ""}${queryMuted ? " queryMuted" : ""}${connected ? " connected" : ""}`}>
    <button onClick={onSelect} aria-label={`${skill?.name || `Nó ${node.nodeId}`}${selected ? ", selecionado" : ""}`}>
      <span className="forestDiamond" aria-hidden="true" />
      <b>{node.nodeId}</b>
    </button>
    <strong>{skill?.name || `Nó ${node.nodeId}`}</strong>
    <small>{skill?.type || nodeTypeName(node.nodeType)}</small>
    {!!node.parentNodes.length && <em>Depende de {node.parentNodes.join(", ")}</em>}
    {!!node.branchIds.length && <em>{node.branchIds.length} {node.branchIds.length === 1 ? "modo" : "modos"}</em>}
  </article>;
}

function SkillNodeIcon({ type, nodeType, size }: { type: string; nodeType: number; size: number }) {
  const normalized = type.toLocaleLowerCase();
  if (normalized.includes("normal")) return <Swords size={size} />;
  if (normalized.includes("resonance skill")) return <Radio size={size} />;
  if (normalized.includes("liberation")) return <Sparkles size={size} />;
  if (normalized.includes("intro")) return <Orbit size={size} />;
  if (normalized.includes("forte")) return <Zap size={size} />;
  if (normalized.includes("outro")) return <Shield size={size} />;
  if (normalized.includes("tune")) return <Music2 size={size} />;
  if (nodeType === 4) return <Crosshair size={size} />;
  return <CircleDot size={size} />;
}

function Stats({ profile }: { profile: CharacterProfile }) {
  const [ascension, setAscension] = useState(6);
  const ranges = [[1, 20], [20, 40], [40, 50], [50, 60], [60, 70], [70, 80], [80, 90]];
  const range = ranges[ascension] || ranges[0];
  const [level, setLevel] = useState(90);
  const boundedLevel = Math.max(range[0], Math.min(range[1], level));
  const stat = profile.progression?.stats?.find((entry) => entry.ascension === ascension && entry.level === boundedLevel);
  const weakness = profile.extras?.weakness;
  if (!profile.progression?.stats?.length) return <DetailEmpty text="Atributos ainda não sincronizados." />;
  return <div className="statsPage">
    <section className="statCalculator"><header><div><span className="sectionLabel">ATRIBUTOS EXATOS</span><h2>Nível e ascensão</h2></div><strong>A{ascension} · Nv.{boundedLevel}</strong></header>
      <label>Ascensão <input type="range" min={0} max={6} value={ascension} onChange={(event) => { const next = Number(event.target.value); setAscension(next); setLevel(ranges[next][1]); }} /></label>
      <label>Nível <input type="range" min={range[0]} max={range[1]} value={boundedLevel} onChange={(event) => setLevel(Number(event.target.value))} /></label>
      <div className="statNumbers"><article><small>HP BASE</small><strong>{formatStat(stat?.hp)}</strong></article><article><small>ATK BASE</small><strong>{formatStat(stat?.atk)}</strong></article><article><small>DEF BASE</small><strong>{formatStat(stat?.def)}</strong></article></div>
    </section>
    <section className="weaknessPanel"><span className="sectionLabel">FRAQUEZA E QUEBRA</span><h2>Parâmetros internos</h2>
      <div><StatLine label="Acúmulo" value={weakness?.buildUp} /><StatLine label="Limite de acúmulo" value={weakness?.buildUpMax} /><StatLine label="Bônus total" value={weakness?.totalBonus} percent /><StatLine label="Multiplicador de quebra" value={weakness?.breakRatio} percent /><StatLine label="Maestria" value={weakness?.mastery} /></div>
      <p>Valores oficiais da versão {profile.character.gameVersion}. Percentuais são convertidos da escala interna de 10.000 pontos.</p>
    </section>
  </div>;
}

function Lore({ profile }: { profile: CharacterProfile }) {
  const entries = [...(profile.extras?.stories || []), ...(profile.extras?.goods || [])];
  if (!entries.length) return <DetailEmpty text="Histórias ainda não sincronizadas." />;
  return <div className="lorePage"><header><span className="sectionLabel">ARQUIVO PESSOAL</span><h2>Histórias e objetos</h2><p>{entries.length} registros oficiais de {profile.character.name}</p></header>
    <div>{entries.map((entry, index) => <details key={`${entry.title}-${index}`} open={index === 0}><summary><span>{String(index + 1).padStart(2, "0")}</span><strong>{entry.title}</strong><i>+</i></summary><FormattedText text={entry.content} /></details>)}</div>
  </div>;
}

function StatLine({ label, value = 0, percent = false }: { label: string; value?: number; percent?: boolean }) {
  return <div><span>{label}</span><strong>{percent ? `${formatStat(value / 100)}%` : formatStat(value)}</strong></div>;
}

function formatStat(value = 0) { return new Intl.NumberFormat("pt-BR", { maximumFractionDigits: 2 }).format(value); }
function nodeTypeName(type: number) { return type === 1 ? "Forte Circuit" : type === 2 ? "Habilidade ativa" : type === 3 ? "Habilidade inerente" : type === 4 ? "Atributo" : "Nó"; }

function Chains({ profile }: { profile: CharacterProfile }) {
  if (!profile.chains.length) return <DetailEmpty text="Nenhuma sequência sincronizada." />;
  return <div className="chainTimeline">{profile.chains.map((chain) => <article className="chainItem" key={chain.sequence}>
    <div className="chainMarker"><span>S{chain.sequence}</span></div>
    <div><span className="sectionLabel">RESONANCE CHAIN {chain.sequence}</span><h2>{chain.name}</h2><FormattedText text={chain.description} /></div>
  </article>)}</div>;
}

function FormattedText({ text }: { text: string }) {
  return <div className="formattedText">{text.split("\n").filter(Boolean).map((line, index) => <p key={index}>{line}</p>)}</div>;
}

function DetailEmpty({ text }: { text: string }) {
  return <div className="emptyState detailEmpty"><div className="emptyGlyph">◇</div><h2>{text}</h2><p>Execute uma nova sincronização para buscar os detalhes atuais.</p></div>;
}

function initials(name: string) {
  return name.split(/\s+/).slice(0, 2).map((part) => part[0]).join("").toUpperCase();
}
