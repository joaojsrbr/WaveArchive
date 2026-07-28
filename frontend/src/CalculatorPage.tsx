import { useEffect, useState } from "react";
import { calculateDamage, evaluateBuild, listBuilds, listEnemies, listFormulaVersions } from "./lib/backend";
import type { Build, BuildEvaluation, DamageInput, DamageResult, Enemy, FormulaVersion } from "./types";

const initialInput: DamageInput = {
  scalingStat: 2000, motionValue: 2, flatDamage: 0, flatBonusDamage: 0,
  characterLevel: 90, enemyLevel: 90, enemyResistance: .1, resistancePenetration: 0,
  defenseIgnore: 0, damageReduction: 0, additionalDamageReduction: 0,
  elementReduction: 0, additionalElementReduction: 0,
  damageBonuses: [.6], amplifications: [], specialBonuses: [],
  critRate: .7, critDamage: 2.4
};

export function CalculatorPage({ onError }: { onError: (message: string) => void }) {
  const [input, setInput] = useState<DamageInput>(readInput);
  const [result, setResult] = useState<DamageResult>();
  const [advanced, setAdvanced] = useState(false);
  const [builds, setBuilds] = useState<Build[]>([]);
  const [buildEvaluation, setBuildEvaluation] = useState<BuildEvaluation>();
  const [enemies,setEnemies]=useState<Enemy[]>([]);
  const [formulas,setFormulas]=useState<FormulaVersion[]>([]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void calculateDamage(input).then((next) => {
        setResult(next);
        localStorage.setItem("wavearchive:damage-input", JSON.stringify(input));
        onError("");
      }).catch((cause) => onError(messageFrom(cause)));
    }, 100);
    return () => window.clearTimeout(timer);
  }, [input]);
  useEffect(() => { void Promise.all([listBuilds(),listEnemies(),listFormulaVersions()]).then(([b,e,f])=>{setBuilds(b);setEnemies(e);setFormulas(f)}).catch((cause) => onError(messageFrom(cause))); }, []);
  async function selectBuild(id: number) {
    if (!id) { setBuildEvaluation(undefined); return; }
    try { setBuildEvaluation(await evaluateBuild(id)); onError(""); }
    catch (cause) { onError(messageFrom(cause)); }
  }
  function loadBuildScenario() {
    if (!buildEvaluation) return;
    let extras: number[] = []; try { extras = JSON.parse(buildEvaluation.config.extraDamageBonusesJson); } catch { /* validated by backend */ }
    setInput({
      ...input, scalingStat: buildEvaluation.stats.scalingStat,
      motionValue: buildEvaluation.config.motionValue, flatDamage: buildEvaluation.config.flatDamage,
      characterLevel: buildEvaluation.build.characterLevel, enemyLevel: buildEvaluation.config.enemyLevel,
      enemyResistance: buildEvaluation.config.enemyResistance, defenseIgnore: buildEvaluation.config.defenseIgnore,
      damageReduction: buildEvaluation.config.damageReduction, elementReduction: buildEvaluation.config.elementReduction,
      damageBonuses: [...Object.values(buildEvaluation.stats.damageBonuses), ...extras],
      critRate: buildEvaluation.stats.critRate, critDamage: buildEvaluation.stats.critDamage
    });
  }

  function numberField(key: keyof DamageInput, value: number) {
    setInput({ ...input, [key]: Number.isFinite(value) ? value : 0 });
  }
  function percentField(key: keyof DamageInput, value: number) {
    numberField(key, value / 100);
  }
  function groupField(key: "damageBonuses" | "amplifications" | "specialBonuses", value: string) {
    setInput({ ...input, [key]: value.split(/[,;+]/).map((part) => Number(part.trim()) / 100).filter(Number.isFinite) });
  }

  return <div className="calculatorPage">
    <div className="pageIntro"><div><span className="eyebrow">ENGINE MATEMÁTICA DETERMINÍSTICA</span><h1>CALCULADORA</h1><p>Todos os grupos seguem a equação oficial; a análise inteligente apenas explica o resultado.</p></div><button className="secondaryButton" onClick={() => setInput(initialInput)}>Restaurar exemplo</button></div>
    <div className="scenarioBar"><label>Inimigo<select defaultValue={0} onChange={e=>{const enemy=enemies.find(item=>item.id===Number(e.target.value));if(enemy)setInput({...input,enemyLevel:enemy.level,enemyResistance:enemy.resistance,damageReduction:enemy.damageReduction,elementReduction:enemy.elementReduction})}}><option value={0}>Cenário manual</option>{enemies.map(enemy=><option key={enemy.id} value={enemy.id}>{enemy.name}</option>)}</select></label><div>{formulas.map(formula=><span className={formula.active?"active":""} key={formula.id}>{formula.name} · {formula.confidence.replace("_"," ")}</span>)}</div></div>
    {builds.length > 0 && <div className="buildCalculatorLink"><label>Usar build salva<select defaultValue={0} onChange={(event) => void selectBuild(Number(event.target.value))}><option value={0}>Selecionar build…</option>{builds.map((build) => <option key={build.id} value={build.id}>{build.name} · {build.characterName}</option>)}</select></label>{buildEvaluation && <><div><span>ATK {formatDamage(buildEvaluation.stats.totalAtk)}</span><span>CR {Math.round(buildEvaluation.stats.critRate * 100)}%</span><span>CD {Math.round(buildEvaluation.stats.critDamage * 100)}%</span><span>Automático {formatDamage(buildEvaluation.damage.expectedDamage)}</span></div><button className="primaryButton" onClick={loadBuildScenario}>Carregar na calculadora</button></>}</div>}
    <div className="calculatorLayout">
      <div className="calculatorForm">
        <section><header><span>01</span><div><h2>Dano-base</h2><p>Escala × Motion Value + dano fixo</p></div></header><div className="calcFields">
          <CalcField label="Atributo de escala" value={input.scalingStat} onChange={(value) => numberField("scalingStat", value)} suffix="ATK/HP/DEF" />
          <CalcField label="Motion Value" value={input.motionValue * 100} onChange={(value) => percentField("motionValue", value)} suffix="%" />
          <CalcField label="Dano fixo" value={input.flatDamage} onChange={(value) => numberField("flatDamage", value)} />
          <CalcField label="Bônus fixo" value={input.flatBonusDamage} onChange={(value) => numberField("flatBonusDamage", value)} />
        </div></section>
        <section><header><span>02</span><div><h2>Alvo e defesa</h2><p>Resistência, nível e redução de dano</p></div></header><div className="calcFields">
          <CalcField label="Nível do personagem" value={input.characterLevel} onChange={(value) => numberField("characterLevel", value)} />
          <CalcField label="Nível do inimigo" value={input.enemyLevel} onChange={(value) => numberField("enemyLevel", value)} />
          <CalcField label="Resistência do inimigo" value={input.enemyResistance * 100} onChange={(value) => percentField("enemyResistance", value)} suffix="%" />
          <CalcField label="Penetração de resistência" value={input.resistancePenetration * 100} onChange={(value) => percentField("resistancePenetration", value)} suffix="%" />
          <CalcField label="DEF Ignore" value={input.defenseIgnore * 100} onChange={(value) => percentField("defenseIgnore", value)} suffix="%" />
          <CalcField label="Redução de dano" value={input.damageReduction * 100} onChange={(value) => percentField("damageReduction", value)} suffix="%" />
        </div></section>
        <section><header><span>03</span><div><h2>Multiplicadores do jogador</h2><p>Valores do mesmo campo são somados; grupos diferentes multiplicam</p></div></header><div className="calcFields groupFields">
          <label>Bônus de dano aditivos<small>Separe fontes com vírgula</small><input value={percentGroup(input.damageBonuses)} onChange={(event) => groupField("damageBonuses", event.target.value)} placeholder="40, 20, 15" /><i>%</i></label>
          <label>Amplify / Deepen<small>Grupo multiplicativo</small><input value={percentGroup(input.amplifications)} onChange={(event) => groupField("amplifications", event.target.value)} placeholder="15, 25" /><i>%</i></label>
          <label>Bônus especiais<small>Grupo multiplicativo raro</small><input value={percentGroup(input.specialBonuses)} onChange={(event) => groupField("specialBonuses", event.target.value)} placeholder="10" /><i>%</i></label>
        </div></section>
        <section><header><span>04</span><div><h2>Crítico</h2><p>Dano esperado = dano final × [1 + CR × (CD − 1)]</p></div></header><div className="calcFields">
          <CalcField label="Taxa crítica" value={input.critRate * 100} onChange={(value) => percentField("critRate", value)} suffix="%" />
          <CalcField label="Dano crítico total" value={input.critDamage * 100} onChange={(value) => percentField("critDamage", value)} suffix="%" />
        </div></section>
        <button className="advancedToggle" onClick={() => setAdvanced(!advanced)}>{advanced ? "Ocultar" : "Mostrar"} reduções elementais avançadas</button>
        {advanced && <section><div className="calcFields">
          <CalcField label="Redução adicional de dano" value={input.additionalDamageReduction * 100} onChange={(value) => percentField("additionalDamageReduction", value)} suffix="%" />
          <CalcField label="Redução elemental-base" value={input.elementReduction * 100} onChange={(value) => percentField("elementReduction", value)} suffix="%" />
          <CalcField label="Redução elemental adicional" value={input.additionalElementReduction * 100} onChange={(value) => percentField("additionalElementReduction", value)} suffix="%" />
        </div></section>}
      </div>
      <aside className="damageResults">
        <span className="sectionLabel">RESULTADO EM TEMPO REAL · {result?.formulaVersion ?? "—"} · {result?.formulaConfidence ?? ""}</span>
        <div className="damagePrimary"><small>DANO ESPERADO</small><strong>{formatDamage(result?.expectedDamage)}</strong><em>Média considerando crítico</em></div>
        <div className="damageVariants"><div><span>Sem crítico</span><strong>{formatDamage(result?.nonCriticalDamage)}</strong></div><div><span>Crítico</span><strong>{formatDamage(result?.criticalDamage)}</strong></div><div><span>Dano-base</span><strong>{formatDamage(result?.baseDamage)}</strong></div></div>
        <button className="breakdownToggle" onClick={() => setAdvanced(true)}>Ver decomposição</button>
        {result && <div className="multiplierList">
          <Multiplier label="Atributo × MV + fixos" value={result.baseDamage} raw />
          <Multiplier label="Resistência" value={result.resistanceMultiplier} />
          <Multiplier label="Defesa" value={result.defenseMultiplier} />
          <Multiplier label="Redução de dano" value={result.damageReductionMultiplier} />
          <Multiplier label="Redução elemental" value={result.elementReductionMultiplier} />
          <Multiplier label="Bônus de dano" value={result.damageBonusMultiplier} />
          <Multiplier label="Amplificação" value={result.amplificationMultiplier} />
          <Multiplier label="Bônus especial" value={result.specialDamageMultiplier} />
        </div>}
        <div className="aiAnalysis"><header><span>✦</span><div><small>ANÁLISE INTELIGENTE LOCAL</small><strong>Leitura do resultado</strong></div></header>{result?.insights.map((insight, index) => <article className={insight.severity} key={index}><b>{insight.title}</b><p>{insight.message}</p></article>)}</div>
      </aside>
    </div>
  </div>;
}

function CalcField({ label, value, onChange, suffix }: { label: string; value: number; onChange: (value: number) => void; suffix?: string }) {
  return <label>{label}<input type="number" step="any" value={rounded(value)} onChange={(event) => onChange(Number(event.target.value))} />{suffix && <i>{suffix}</i>}</label>;
}
function Multiplier({ label, value, raw=false }: { label: string; value: number;raw?:boolean }) { return <div><span>{label}</span><strong>{raw?formatDamage(value):`× ${value.toFixed(4)}`}</strong></div>; }
function percentGroup(values: number[]) { return values.map((value) => rounded(value * 100)).join(", "); }
function rounded(value: number) { return Math.round(value * 10000) / 10000; }
function formatDamage(value?: number) { return value === undefined ? "—" : new Intl.NumberFormat("pt-BR", { maximumFractionDigits: 0 }).format(value); }
function readInput(): DamageInput { try { return { ...initialInput, ...JSON.parse(localStorage.getItem("wavearchive:damage-input") ?? "") }; } catch { return initialInput; } }
function messageFrom(cause: unknown) { return cause instanceof Error ? cause.message : String(cause); }
