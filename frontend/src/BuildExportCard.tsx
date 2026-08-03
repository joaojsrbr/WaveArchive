import { useEffect, useMemo, useRef, useState, type RefObject } from 'react';
import {
  Activity,
  Crosshair,
  Download,
  Flame,
  Heart,
  Shield,
  Sparkles,
  Swords,
  Waves,
  X,
  Zap,
} from 'lucide-react';
import { toPng } from 'html-to-image';
import { evaluateBuild } from './lib/backend';
import { buildSonataProgress } from './lib/sonata';
import type { Build, BuildEvaluation, Character, OwnedEcho, Sonata } from './types';

type Props = {
  build: Build;
  character?: Character;
  sonatas: Sonata[];
  onClose: () => void;
  onError: (message: string) => void;
};

const skills = [
  ['Ataque Normal', 'normalAttackLevel', Swords],
  ['Habilidade', 'resonanceSkillLevel', Zap],
  ['Forte', 'forteLevel', Activity],
  ['Liberação', 'liberationLevel', Sparkles],
  ['Intro', 'introLevel', Crosshair],
] as const;

export function BuildExportCard({ build, character, sonatas, onClose, onError }: Props) {
  const cardRef = useRef<HTMLDivElement>(null);
  const [evaluation, setEvaluation] = useState<BuildEvaluation>();
  const [loading, setLoading] = useState(true);
  const [exporting, setExporting] = useState(false);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    void evaluateBuild(build.id)
      .then((result) => {
        if (!cancelled) setEvaluation(result);
      })
      .catch((cause) => {
        if (!cancelled) onError(messageFrom(cause));
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [build.id, onError]);

  const sonataProgress = useMemo(
    () => buildSonataProgress(build.echoes, sonatas),
    [build.echoes, sonatas]
  );

  async function download() {
    if (!cardRef.current || exporting) return;
    setExporting(true);
    try {
      const scale = 2048 / cardRef.current.offsetWidth;
      const dataUrl = await toPng(cardRef.current, {
        cacheBust: true,
        pixelRatio: scale,
        width: cardRef.current.offsetWidth,
        height: cardRef.current.offsetHeight,
        backgroundColor: '#071019',
      });
      const link = document.createElement('a');
      link.download = `${safeFilename(build.name || build.characterName || 'build')}-wavearchive.png`;
      link.href = dataUrl;
      link.click();
      onError('');
    } catch (cause) {
      onError(`Não foi possível exportar o card: ${messageFrom(cause)}`);
    } finally {
      setExporting(false);
    }
  }

  return (
    <div
      className="buildExportOverlay"
      role="dialog"
      aria-modal="true"
      aria-labelledby="build-export-title"
    >
      <div className="buildExportModal">
        <header className="buildExportToolbar">
          <div>
            <span className="sectionLabel">EXPORTAR BUILD</span>
            <h2 id="build-export-title">Prévia do card</h2>
            <p>PNG 2048 × 960 · somente dados salvos</p>
          </div>
          <div>
            <button onClick={onClose}>
              <X size={15} /> Fechar
            </button>
            <button
              className="saveTeamButton"
              disabled={loading || exporting}
              onClick={() => void download()}
            >
              <Download size={15} />
              {exporting ? 'Exportando...' : loading ? 'Carregando...' : 'Baixar PNG'}
            </button>
          </div>
        </header>
        <div className="buildExportViewport">
          <ExportCardCanvas
            build={build}
            character={character}
            sonataProgress={sonataProgress}
            evaluation={evaluation}
            cardRef={cardRef}
          />
        </div>
      </div>
    </div>
  );
}

function ExportCardCanvas({
  build,
  character,
  sonataProgress,
  evaluation,
  cardRef,
}: {
  build: Build;
  character?: Character;
  sonataProgress: ReturnType<typeof buildSonataProgress>;
  evaluation?: BuildEvaluation;
  cardRef: RefObject<HTMLDivElement | null>;
}) {
  const stats = evaluation?.stats;
  const statItems = [
    ['HP', stats ? formatNumber(stats.totalHp) : '—', Heart],
    ['ATK', stats ? formatNumber(stats.totalAtk) : '—', Swords],
    ['DEF', stats ? formatNumber(stats.totalDef) : '—', Shield],
    ['Taxa CRIT', stats ? formatPercent(stats.critRate) : '—', Crosshair],
    ['Dano CRIT', stats ? formatPercent(stats.critDamage) : '—', Sparkles],
    ['Recarga', stats ? formatPercent(stats.energyRegen) : '—', Activity],
  ] as const;
  const backdrop = character?.backgroundPath || character?.iconPath || build.characterIcon;

  return (
    <div className="exportBuildCard" ref={cardRef}>
      {isAssetPath(backdrop) && (
        <img className="exportCharacterBackdrop" src={backdrop} alt="" crossOrigin="anonymous" />
      )}
      <div className="exportCardScrim" aria-hidden="true" />

      <aside className="exportIdentityRail">
        <span className="exportBrand">
          <Waves size={22} /> WAVEARCHIVE
        </span>
        <h1>{build.characterName}</h1>
        <p>
          <strong>Nv. {build.characterLevel}/90</strong>
          <strong>S{build.sequence}</strong>
        </p>
        {character && (
          <div className="exportIdentityTraits">
            <span>
              <Flame size={20} /> {character.element}
            </span>
            <span>
              <Swords size={20} /> {character.weaponType}
            </span>
          </div>
        )}
        <Sparkles className="exportRailMark" size={44} />
      </aside>

      <aside className="exportBuildMeta">
        <i />
        <span>DADOS DA BUILD</span>
        <i />
        {build.gameVersion && <small>VERSÃO {build.gameVersion}</small>}
      </aside>

      <section className="exportStats">
        {statItems.map(([label, value, Icon]) => (
          <article key={label}>
            <span>
              <Icon size={18} /> {label}
            </span>
            <strong>{value}</strong>
          </article>
        ))}
      </section>

      <section className="exportWeapon">
        {isAssetPath(build.weaponIcon) && (
          <img src={build.weaponIcon} alt="" crossOrigin="anonymous" />
        )}
        <strong>{build.weaponName || 'Arma não definida'}</strong>
        {build.weaponName && (
          <small>
            Nv. {build.weaponLevel} · R{build.weaponRank}
          </small>
        )}
      </section>

      <section className="exportSonatas">
        {sonataProgress.length ? (
          sonataProgress.slice(0, 1).map(({ sonata, count, tiers }) => (
            <article key={sonata.id}>
              {isAssetPath(sonata.iconPath) ? (
                <img src={sonata.iconPath} alt="" crossOrigin="anonymous" />
              ) : (
                <Waves size={24} />
              )}
              <div>
                <strong>
                  {sonata.name} ({count})
                </strong>
                {tiers.map((tier) => (
                  <small key={tier.pieces}>
                    {tier.pieces} peças · {tier.description}
                  </small>
                ))}
              </div>
            </article>
          ))
        ) : (
          <span className="exportUnavailable">Sonata não definida</span>
        )}
      </section>

      <section className="exportSkillLine" aria-label="Níveis de habilidade">
        <div className="exportSkillHeading">
          <i />
          <span>HABILIDADES</span>
          <i />
        </div>
        {skills.map(([label, key, Icon]) => {
          const level = build[key];
          return (
            <article key={key}>
              <span className="exportSkillIcon">
                <Icon size={19} />
              </span>
              <small>{label}</small>
              <strong>{level > 0 ? `Nv. ${level}` : '—'}</strong>
            </article>
          );
        })}
        <aside>
          <span>Ecos</span>
          <strong>{build.echoes.reduce((sum, echo) => sum + echo.cost, 0)}/12</strong>
        </aside>
      </section>

      <span className="exportEchoHeading">ECOS EQUIPADOS</span>
      <section className="exportEchoGallery">
        {Array.from({ length: 4 }, (_, index) => (
          <ExportEcho
            echo={build.echoes[index + 1]}
            index={index + 1}
            featured={false}
            key={index}
          />
        ))}
        <ExportEcho echo={build.echoes[0]} index={0} featured />
      </section>
    </div>
  );
}

function ExportEcho({
  echo,
  index,
  featured,
}: {
  echo?: OwnedEcho;
  index: number;
  featured: boolean;
}) {
  if (!echo) {
    return (
      <article className={`exportEcho ${featured ? 'featured' : ''} empty`}>
        <span>Slot {index + 1}</span>
        <strong>Echo não equipado</strong>
      </article>
    );
  }
  const substats = parseSubstats(echo.substatsJson);
  return (
    <article className={`exportEcho ${featured ? 'featured' : ''}`}>
      <div className="exportEchoArt">
        {isAssetPath(echo.iconPath) && <img src={echo.iconPath} alt="" crossOrigin="anonymous" />}
      </div>
      <header>
        <span>
          CUSTO {echo.cost} · +{echo.level}
        </span>
        <strong>{echo.echoName}</strong>
        <small>{echo.mainStat || 'Atributo principal não definido'}</small>
      </header>
      <div className="exportEchoSubstats">
        {Array.from({ length: 5 }, (_, substatIndex) => (
          <span className={!substats[substatIndex] ? 'missing' : ''} key={substatIndex}>
            <i>{String(substatIndex + 1).padStart(2, '0')}</i>
            {substats[substatIndex] || '—'}
          </span>
        ))}
      </div>
    </article>
  );
}

function parseSubstats(value: string): string[] {
  try {
    const parsed = JSON.parse(value);
    return Array.isArray(parsed) ? parsed.map(String).filter(Boolean).slice(0, 5) : [];
  } catch {
    return [];
  }
}

function formatNumber(value: number) {
  return new Intl.NumberFormat('pt-BR', { maximumFractionDigits: 0 }).format(value);
}

function formatPercent(value: number) {
  return new Intl.NumberFormat('pt-BR', { style: 'percent', maximumFractionDigits: 1 }).format(
    value
  );
}

function isAssetPath(path?: string) {
  return Boolean(
    path && (path.startsWith('/cache/') || path.startsWith('data:') || path.startsWith('blob:'))
  );
}

function safeFilename(value: string) {
  return value
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .replace(/[^a-zA-Z0-9_-]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .toLowerCase();
}

function messageFrom(cause: unknown) {
  return cause instanceof Error ? cause.message : String(cause);
}
