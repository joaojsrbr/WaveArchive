import { useEffect, useState, type FormEvent } from 'react';
import { evaluateBuild, saveBuildConfig } from './lib/backend';
import type { BuildEvaluation } from './types';

export function BuildTheorycraftModal({
  buildId,
  onClose,
  onError,
}: {
  buildId: number;
  onClose: () => void;
  onError: (message: string) => void;
}) {
  const [evaluation, setEvaluation] = useState<BuildEvaluation>();
  const [saving, setSaving] = useState(false);
  useEffect(() => {
    void evaluateBuild(buildId)
      .then(setEvaluation)
      .catch((cause) => onError(messageFrom(cause)));
  }, [buildId]);
  async function submit(event: FormEvent) {
    event.preventDefault();
    if (!evaluation) return;
    setSaving(true);
    try {
      setEvaluation(await saveBuildConfig(evaluation.config));
      onError('');
    } catch (cause) {
      onError(messageFrom(cause));
    } finally {
      setSaving(false);
    }
  }
  function update(key: keyof BuildEvaluation['config'], value: number) {
    if (evaluation)
      setEvaluation({ ...evaluation, config: { ...evaluation.config, [key]: value } });
  }
  function updateString(key: keyof BuildEvaluation['config'], value: string) {
    if (evaluation)
      setEvaluation({ ...evaluation, config: { ...evaluation.config, [key]: value } });
  }
  return (
    <div className="modalBackdrop" onMouseDown={onClose}>
      <form
        className="theoryModal"
        onSubmit={submit}
        onMouseDown={(event) => event.stopPropagation()}
      >
        <div className="modalHeader">
          <div>
            <span className="sectionLabel">BUILDS → ENGINE DE DANO</span>
            <h2>{evaluation?.build.name ?? 'Carregando…'}</h2>
          </div>
          <button type="button" onClick={onClose}>
            ×
          </button>
        </div>
        {!evaluation ? (
          <div className="skeleton theorySkeleton" />
        ) : (
          <div className="theoryLayout">
            <div>
              <p className="theoryNotice">
                A arma e os atributos escritos nos Echoes são interpretados automaticamente. Informe
                abaixo somente a base do personagem e o cenário.
              </p>
              <div className="buildFormGrid">
                <label className="buildField">
                  Escala
                  <select
                    value={evaluation.config.scalingType}
                    onChange={(event) =>
                      setEvaluation({
                        ...evaluation,
                        config: {
                          ...evaluation.config,
                          scalingType: event.target.value as 'ATK' | 'HP' | 'DEF',
                        },
                      })
                    }
                  >
                    <option>ATK</option>
                    <option>HP</option>
                    <option>DEF</option>
                  </select>
                </label>
                <label className="buildField">
                  Motion Value (%)
                  <input
                    type="number"
                    step="any"
                    value={evaluation.config.motionValue * 100}
                    onChange={(event) => update('motionValue', Number(event.target.value) / 100)}
                  />
                </label>
                <label className="buildField">
                  ATK-base do personagem
                  <input
                    type="number"
                    value={evaluation.config.baseAtk}
                    onChange={(event) => update('baseAtk', Number(event.target.value))}
                  />
                </label>
                <label className="buildField">
                  HP-base do personagem
                  <input
                    type="number"
                    value={evaluation.config.baseHp}
                    onChange={(event) => update('baseHp', Number(event.target.value))}
                  />
                </label>
                <label className="buildField">
                  DEF-base do personagem
                  <input
                    type="number"
                    value={evaluation.config.baseDef}
                    onChange={(event) => update('baseDef', Number(event.target.value))}
                  />
                </label>
                <label className="buildField">
                  Dano fixo
                  <input
                    type="number"
                    value={evaluation.config.flatDamage}
                    onChange={(event) => update('flatDamage', Number(event.target.value))}
                  />
                </label>
                <label className="buildField">
                  Nível do inimigo
                  <input
                    type="number"
                    min={1}
                    max={120}
                    value={evaluation.config.enemyLevel}
                    onChange={(event) => update('enemyLevel', Number(event.target.value))}
                  />
                </label>
                <label className="buildField">
                  Resistência (%)
                  <input
                    type="number"
                    step="any"
                    value={evaluation.config.enemyResistance * 100}
                    onChange={(event) =>
                      update('enemyResistance', Number(event.target.value) / 100)
                    }
                  />
                </label>
                <label className="buildField">
                  DEF Ignore (%)
                  <input
                    type="number"
                    step="any"
                    value={evaluation.config.defenseIgnore * 100}
                    onChange={(event) => update('defenseIgnore', Number(event.target.value) / 100)}
                  />
                </label>
                <label className="buildField">
                  Bônus adicionais (%)
                  <input
                    value={jsonPercents(evaluation.config.extraDamageBonusesJson)}
                    onChange={(event) =>
                      updateString('extraDamageBonusesJson', percentsJSON(event.target.value))
                    }
                    placeholder="20, 15"
                  />
                </label>
              </div>
              {evaluation.stats.unparsedStats.length > 0 && (
                <div className="unparsedStats">
                  <strong>Revisar atributos não reconhecidos</strong>
                  {evaluation.stats.unparsedStats.map((stat, index) => (
                    <code key={index}>{stat}</code>
                  ))}
                </div>
              )}
            </div>
            <aside>
              <span className="sectionLabel">ATRIBUTOS FINAIS</span>
              <div className="theoryStats">
                <Stat label="ATK" value={evaluation.stats.totalAtk} />
                <Stat label="HP" value={evaluation.stats.totalHp} />
                <Stat label="DEF" value={evaluation.stats.totalDef} />
                <Stat label="CRIT Rate" value={evaluation.stats.critRate * 100} suffix="%" />
                <Stat label="CRIT DMG" value={evaluation.stats.critDamage * 100} suffix="%" />
                <Stat label="Energy Regen" value={evaluation.stats.energyRegen * 100} suffix="%" />
              </div>
              <div className="theoryDamage">
                <small>DANO ESPERADO</small>
                <strong>{format(evaluation.damage.expectedDamage)}</strong>
                <span>
                  Crítico {format(evaluation.damage.criticalDamage)} · sem crítico{' '}
                  {format(evaluation.damage.nonCriticalDamage)}
                </span>
              </div>
              <div className="parsedSources">
                <strong>Fontes interpretadas</strong>
                <span>Arma: {format(evaluation.stats.weaponAtk)} ATK-base</span>
                {Object.entries(evaluation.stats.damageBonuses).map(([name, value]) => (
                  <span key={name}>
                    {name}: +{(value * 100).toFixed(1)}%
                  </span>
                ))}
              </div>
            </aside>
          </div>
        )}
        <div className="modalActions">
          <button type="button" onClick={onClose}>
            Fechar
          </button>
          <button className="primaryButton" disabled={!evaluation || saving}>
            {saving ? 'Calculando…' : 'Salvar e recalcular'}
          </button>
        </div>
      </form>
    </div>
  );
}
function Stat({ label, value, suffix = '' }: { label: string; value: number; suffix?: string }) {
  return (
    <div>
      <span>{label}</span>
      <strong>
        {format(value)}
        {suffix}
      </strong>
    </div>
  );
}
function format(value: number) {
  return new Intl.NumberFormat('pt-BR', { maximumFractionDigits: 1 }).format(value);
}
function jsonPercents(value: string) {
  try {
    return (JSON.parse(value) as number[]).map((item) => item * 100).join(', ');
  } catch {
    return '';
  }
}
function percentsJSON(value: string) {
  return JSON.stringify(
    value
      .split(/[,;+]/)
      .map((item) => Number(item.trim()) / 100)
      .filter(Number.isFinite)
  );
}
function messageFrom(cause: unknown) {
  return cause instanceof Error ? cause.message : String(cause);
}
