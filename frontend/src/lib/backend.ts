import type { AIAnalysisRequest, AIAnalysisResult, AIConversation, AssistantRequest, BackendAPI, Buff, Build, BuildConfig, BuildEvaluation, CatalogStatus, Character, CharacterAccountUpdate, CharacterFilter, CharacterProfile, DamageInput, DamageResult, Echo, EchoFilter, OwnedEcho, ProgressionPlan, ProgressionPlanRequest, Sonata, SyncResult, Team, TeamTheorycraft, Weapon, WeaponAccountUpdate, WeaponFilter } from "../types";
import { getPreviewCharacter, listPreviewCharacters, listPreviewTeams, previewCharacters, savePreviewTeam } from "./previewData";

const emptyStatus: CatalogStatus = { count: 0, version: "" };

function api(): BackendAPI | undefined {
  return window.go?.main?.App;
}

export async function listCharacters(filter: CharacterFilter): Promise<Character[]> {
  const backend = api();
  return backend ? backend.ListCharacters(filter) : import.meta.env.DEV ? listPreviewCharacters(filter) : [];
}

export async function catalogStatus(): Promise<CatalogStatus> {
  const backend = api();
  return backend ? backend.CatalogStatus() : import.meta.env.DEV ? { count: previewCharacters.length, version: "3.6.1" } : emptyStatus;
}

export async function getCharacter(id: number): Promise<CharacterProfile> {
  const backend = api();
  if (!backend && import.meta.env.DEV) return getPreviewCharacter(id);
  if (!backend) {
    throw new Error("Os detalhes estão disponíveis no aplicativo desktop.");
  }
  return backend.GetCharacter(id);
}

export async function calculateCharacterProgression(request: ProgressionPlanRequest): Promise<ProgressionPlan> {
  const backend = api();
  if (!backend) throw new Error("O cálculo de materiais está disponível no aplicativo desktop.");
  return backend.CalculateCharacterProgression(request);
}

export async function syncCharacters(): Promise<SyncResult> {
  const backend = api();
  if (!backend) {
    throw new Error("A sincronização está disponível no aplicativo desktop.");
  }
  return backend.SyncCharacters();
}

export async function cancelSync(): Promise<boolean> {
  return (await api()?.CancelSync()) ?? false;
}

export async function restoreLatestSnapshot(): Promise<string> {
  const backend = api();
  if (!backend) {
    throw new Error("A restauração está disponível no aplicativo desktop.");
  }
  return backend.RestoreLatestSnapshot();
}

export async function updateCharacterAccount(update: CharacterAccountUpdate): Promise<void> {
  const backend = api();
  if (!backend) {
    throw new Error("A conta local está disponível no aplicativo desktop.");
  }
  await backend.UpdateCharacterAccount(update);
}

export async function listWeapons(filter: WeaponFilter): Promise<Weapon[]> {
  return (await api()?.ListWeapons(filter)) ?? [];
}

export async function getWeapon(id: number): Promise<Weapon> {
  const backend = api();
  if (!backend) throw new Error("O catálogo de armas está disponível no aplicativo desktop.");
  return backend.GetWeapon(id);
}

export async function updateWeaponAccount(update: WeaponAccountUpdate): Promise<void> {
  const backend = api();
  if (!backend) throw new Error("A conta local está disponível no aplicativo desktop.");
  await backend.UpdateWeaponAccount(update);
}

export async function listEchoes(filter: EchoFilter): Promise<Echo[]> {
  return (await api()?.ListEchoes(filter)) ?? [];
}

export async function getEcho(id: number): Promise<Echo> {
  const backend = api();
  if (!backend) throw new Error("O catálogo de Echoes está disponível no aplicativo desktop.");
  return backend.GetEcho(id);
}

export async function listSonatas(): Promise<Sonata[]> {
  return (await api()?.ListSonatas()) ?? [];
}

export async function listOwnedEchoes(): Promise<OwnedEcho[]> {
  return (await api()?.ListOwnedEchoes()) ?? [];
}

export async function saveOwnedEcho(echo: OwnedEcho): Promise<OwnedEcho> {
  const backend = api();
  if (!backend) throw new Error("O inventário de Echoes está disponível no aplicativo desktop.");
  return backend.SaveOwnedEcho(echo);
}

export async function deleteOwnedEcho(id: number): Promise<void> {
  const backend = api();
  if (!backend) throw new Error("O inventário de Echoes está disponível no aplicativo desktop.");
  await backend.DeleteOwnedEcho(id);
}

export async function listBuilds(): Promise<Build[]> {
  return (await api()?.ListBuilds()) ?? [];
}

export async function saveBuild(build: Build): Promise<Build> {
  const backend = api();
  if (!backend) throw new Error("Builds estão disponíveis no aplicativo desktop.");
  return backend.SaveBuild(build);
}

export async function duplicateBuild(id: number): Promise<Build> {
  const backend = api();
  if (!backend) throw new Error("Builds estão disponíveis no aplicativo desktop.");
  return backend.DuplicateBuild(id);
}

export async function deleteBuild(id: number): Promise<void> {
  const backend = api();
  if (!backend) throw new Error("Builds estão disponíveis no aplicativo desktop.");
  await backend.DeleteBuild(id);
}

export async function restoreBuild(id: number): Promise<void> {
  const backend = api();
  if (!backend) throw new Error("Builds estão disponíveis no aplicativo desktop.");
  await backend.RestoreBuild(id);
}

export async function listTeams(): Promise<Team[]> {
  const backend = api();
  return backend ? backend.ListTeams() : import.meta.env.DEV ? listPreviewTeams() : [];
}

export async function saveTeam(team: Team): Promise<Team> {
  const backend = api();
  if (!backend && import.meta.env.DEV) return savePreviewTeam(team);
  if (!backend) throw new Error("Equipes estão disponíveis no aplicativo desktop.");
  return backend.SaveTeam(team);
}

export async function duplicateTeam(id: number): Promise<Team> {
  const backend = api();
  if (!backend) throw new Error("Equipes estão disponíveis no aplicativo desktop.");
  return backend.DuplicateTeam(id);
}

export async function deleteTeam(id: number): Promise<void> {
  const backend = api();
  if (!backend) throw new Error("Equipes estão disponíveis no aplicativo desktop.");
  await backend.DeleteTeam(id);
}

export async function restoreTeam(id: number): Promise<void> {
  const backend = api();
  if (!backend) throw new Error("Equipes estão disponíveis no aplicativo desktop.");
  await backend.RestoreTeam(id);
}

export async function calculateDamage(input: DamageInput): Promise<DamageResult> {
  const backend = api();
  if (!backend) throw new Error("A calculadora está disponível no aplicativo desktop.");
  return backend.CalculateDamage(input);
}

export async function analyzeWithAI(request: AIAnalysisRequest): Promise<AIAnalysisResult> {
  const backend = api();
  if (!backend) throw new Error("A análise por IA está disponível no aplicativo desktop.");
  return backend.AnalyzeWithAI(request);
}

export async function evaluateBuild(id: number): Promise<BuildEvaluation> {
  const backend = api(); if (!backend) throw new Error("Theorycraft disponível no desktop.");
  return backend.EvaluateBuild(id);
}
export async function saveBuildConfig(config: BuildConfig): Promise<BuildEvaluation> {
  const backend = api(); if (!backend) throw new Error("Theorycraft disponível no desktop.");
  return backend.SaveBuildConfig(config);
}
export async function getTeamTheorycraft(teamId: number): Promise<TeamTheorycraft> {
  const backend = api(); if (!backend) throw new Error("Theorycraft disponível no desktop.");
  return backend.GetTeamTheorycraft(teamId);
}
export async function saveBuff(buff: Buff): Promise<Buff> {
  const backend = api(); if (!backend) throw new Error("Buffs disponíveis no desktop.");
  return backend.SaveBuff(buff);
}
export async function deleteBuff(id: number): Promise<void> { await api()?.DeleteBuff(id); }
export async function listAIConversations(): Promise<AIConversation[]> { return (await api()?.ListAIConversations()) ?? []; }
export async function assistantChat(request: AssistantRequest): Promise<AIConversation> {
  const backend = api(); if (!backend) throw new Error("Assistente disponível no desktop.");
  return backend.AssistantChat(request);
}
export async function assistantChatStream(request:AssistantRequest):Promise<AIConversation>{const b=api();if(!b)throw new Error("Streaming disponível no desktop.");return b.AssistantChatStream(request);}
export async function testAIProvider(request:AIAnalysisRequest){const b=api();if(!b)throw new Error("Teste disponível no desktop.");return b.TestAIProvider(request);}
export async function listCharacterGuides(id:number){return (await api()?.ListCharacterGuides(id))??[];}
export async function listAllCharacterGuides(){return (await api()?.ListAllCharacterGuides())??[];}
export async function syncCharacterGuides(id:number,language="en"){const b=api();if(!b)throw new Error("Guias disponíveis no desktop.");return b.SyncCharacterGuides(id,language);}
export async function searchLocalKnowledge(query:string,limit=8){return (await api()?.SearchLocalKnowledge(query,limit))??[];}
export async function deleteAIConversation(id: number): Promise<void> { await api()?.DeleteAIConversation(id); }
export async function listBuildVersions(id:number){return (await api()?.ListBuildVersions(id))??[];}
export async function getSettings(){const b=api();if(!b)throw new Error("Configurações disponíveis no desktop.");return b.GetSettings();}
export async function saveSettings(value:import("../types").AppSettings){const b=api();if(!b)throw new Error("Configurações disponíveis no desktop.");return b.SaveSettings(value);}
export async function listDataSourceOptions(){const b=api();if(!b)throw new Error("Fontes disponíveis no desktop.");return b.ListDataSourceOptions();}
export async function getAccountSummary(){const b=api();if(!b)throw new Error("Conta disponível no desktop.");return b.GetAccountSummary();}
export async function saveAccountSummary(value:import("../types").AccountSummary){const b=api();if(!b)throw new Error("Conta disponível no desktop.");return b.SaveAccountSummary(value);}
export async function listEnemies(){return (await api()?.ListEnemies())??[];}
export async function listFormulaVersions(){return (await api()?.ListFormulaVersions())??[];}
export async function dashboardSummary(){const b=api();if(!b)throw new Error("Dashboard disponível no desktop.");return b.DashboardSummary();}
export async function exportArchive(){const b=api();if(!b)throw new Error("Exportação disponível no desktop.");return b.ExportArchive();}
export async function importArchive(payload:string){const b=api();if(!b)throw new Error("Importação disponível no desktop.");return b.ImportArchive(payload);}
export async function createManualBackup(){const b=api();if(!b)throw new Error("Backup disponível no desktop.");return b.CreateManualBackup();}
export async function listBackups(){return (await api()?.ListBackups())??[];}
export async function restoreBackup(name:string){const b=api();if(!b)throw new Error("Restauração disponível no desktop.");return b.RestoreBackup(name);}
export async function diagnostics(){const b=api();if(!b)throw new Error("Diagnóstico disponível no desktop.");return b.Diagnostics();}
