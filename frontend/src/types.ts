export type Character = {
  id: number;
  name: string;
  nickname: string;
  rarity: number;
  elementCode: number;
  element: string;
  weaponTypeCode: number;
  weaponType: string;
  iconPath: string;
  backgroundPath: string;
  gender?: string;
  owned: boolean;
  level: number;
  sequence: number;
  favorite: boolean;
  gameVersion: string;
  apiOrder: number;
};

export type CharacterAccountUpdate = {
  characterId: number;
  owned: boolean;
  level: number;
  sequence: number;
  favorite: boolean;
};

export type CharacterFilter = {
  query: string;
  element: number;
  rarity: number;
  weaponType?: number;
  gender?: string;
  account?: 'all' | 'owned' | 'missing';
  minLevel?: number;
  maxLevel?: number;
  minSequence?: number;
  maxSequence?: number;
  ownedOnly: boolean;
  favorites: boolean;
  sort: 'name' | 'api' | 'rarity' | 'element' | 'id';
};

export type CatalogStatus = {
  count: number;
  version: string;
  lastSyncAt?: string;
};

export type SyncResult = {
  version: string;
  count: number;
  syncedAt: string;
};

export type Skill = {
  nodeId: string;
  type: string;
  name: string;
  description: string;
  iconPath: string;
  levelsJson: string;
  sortOrder: number;
};

export type ResonanceChain = {
  sequence: number;
  name: string;
  description: string;
  iconPath: string;
};

export type Material = {
  id: number;
  name: string;
  rarity: number;
  type: number;
  description: string;
  iconPath: string;
  sources: string[];
  gameVersion: string;
};
export type CharacterContentSearchResult = {
  kind: 'skill' | 'material';
  entityId: string;
  characterId: number;
  characterName: string;
  title: string;
  subtitle: string;
  iconPath: string;
};
export type MaterialCost = { material: Material; quantity: number };
export type AscensionStage = { stage: number; unlockLevel: number; costs: MaterialCost[] };
export type SkillValueRow = { name: string; values: string[] };
export type SkillLevelCost = { level: number; costs: MaterialCost[] };
export type SkillProgression = {
  nodeId: string;
  nodeType: number;
  type: string;
  name: string;
  iconPath: string;
  maxLevel: number;
  unlockCosts: MaterialCost[];
  levelCosts: SkillLevelCost[];
  values: SkillValueRow[];
};
export type CharacterStat = {
  ascension: number;
  level: number;
  hp: number;
  atk: number;
  def: number;
};
export type CharacterProgression = {
  ascensions: AscensionStage[];
  skills: SkillProgression[];
  levelExp: number[];
  stats: CharacterStat[];
};
export type CharacterTag = {
  id: number;
  name: string;
  description: string;
  iconPath: string;
  color: string;
};
export type LoreEntry = { title: string; content: string; iconPath: string };
export type ForteAction = { name: string; description: string; inputs: string[]; images: string[] };
export type ForteGuide = {
  iconPath: string;
  descriptions: string[];
  features: string[];
  actions: ForteAction[];
};
export type WeaknessStats = {
  buildUp: number;
  buildUpMax: number;
  totalBonus: number;
  breakRatio: number;
  mastery: number;
};
export type SkillBranch = { id: number; name: string; description: string; iconPath: string };
export type SkillTreeNodeInfo = {
  nodeId: string;
  nodeType: number;
  coordinate: number;
  parentNodes: number[];
  branchIds: number[];
  unlockCondition: number;
};
export type CharacterExtras = {
  tags: CharacterTag[];
  stories: LoreEntry[];
  goods: LoreEntry[];
  forte: ForteGuide;
  weakness: WeaknessStats;
  skillBranches: SkillBranch[];
  skillTree: SkillTreeNodeInfo[];
};
export type ProgressionPlanRequest = {
  characterId: number;
  currentLevel: number;
  targetLevel: number;
  currentSkills: Record<string, number>;
  targetSkills: Record<string, number>;
  includeUnlocks: boolean;
};
export type ProgressionPlan = {
  characterId: number;
  ascensions: MaterialCost[];
  skills: MaterialCost[];
  total: MaterialCost[];
};

export type Weapon = {
  id: number;
  name: string;
  rarity: number;
  typeCode: number;
  type: string;
  description: string;
  effectName: string;
  effect: string;
  iconPath: string;
  paramsJson: string;
  baseAtk: number;
  subStat: string;
  gameVersion: string;
  owned: boolean;
  level: number;
  rank: number;
  favorite: boolean;
};

export type WeaponFilter = {
  query: string;
  type: number;
  rarity: number;
  subStat?: string;
  account?: 'all' | 'owned' | 'missing';
  minAtk?: number;
  maxAtk?: number;
  minLevel?: number;
  maxLevel?: number;
  minRank?: number;
  maxRank?: number;
  ownedOnly: boolean;
  favorites: boolean;
  sort: 'name' | 'rarity' | 'type' | 'atk' | 'id';
};

export type WeaponAccountUpdate = {
  weaponId: number;
  owned: boolean;
  level: number;
  rank: number;
  favorite: boolean;
};

export type Build = {
  id: number;
  name: string;
  characterId: number;
  characterName: string;
  characterIcon: string;
  characterLevel: number;
  sequence: number;
  weaponId?: number;
  weaponName: string;
  weaponIcon: string;
  weaponLevel: number;
  weaponRank: number;
  normalAttackLevel: number;
  resonanceSkillLevel: number;
  forteLevel: number;
  liberationLevel: number;
  introLevel: number;
  echoes: OwnedEcho[];
  targetEnemyId?: number;
  rotationId?: number;
  conditions: string;
  notes: string;
  favorite: boolean;
  locked: boolean;
  gameVersion: string;
  createdAt: string;
  updatedAt: string;
};

export type Echo = {
  id: number;
  name: string;
  code: string;
  type: string;
  class: string;
  cost: number;
  place: string;
  iconPath: string;
  skill: string;
  raritiesJson: string;
  sonataIdsJson: string;
  gameVersion: string;
  ownedCount: number;
  favorite: boolean;
};

export type Sonata = {
  id: number;
  name: string;
  iconPath: string;
  twoPiece: string;
  fivePiece: string;
  gameVersion: string;
};

export type EchoFilter = {
  query: string;
  cost: number;
  sonataId: number;
  class?: string;
  type?: string;
  place?: string;
  rarity?: number;
  minOwned?: number;
  ownedOnly: boolean;
  favorites?: boolean;
  sort: 'name' | 'cost' | 'id';
};

export type OwnedEcho = {
  id: number;
  echoId: number;
  echoName: string;
  iconPath: string;
  cost: number;
  mainStat: string;
  substatsJson: string;
  level: number;
  sonataId?: number;
  sonataName: string;
  characterId?: number;
  characterName: string;
  locked: boolean;
  favorite: boolean;
  note: string;
};

export type TeamMember = {
  slot: number;
  characterId: number;
  characterName: string;
  characterIcon: string;
  buildId?: number;
  buildName: string;
  role: string;
  customRole: string;
};

export type Team = {
  id: number;
  name: string;
  members: TeamMember[];
  notes: string;
  favorite: boolean;
  locked: boolean;
  gameVersion: string;
  createdAt: string;
  updatedAt: string;
};

export type DamageInput = {
  scalingStat: number;
  motionValue: number;
  flatDamage: number;
  flatBonusDamage: number;
  characterLevel: number;
  enemyLevel: number;
  enemyResistance: number;
  resistancePenetration: number;
  defenseIgnore: number;
  damageReduction: number;
  additionalDamageReduction: number;
  elementReduction: number;
  additionalElementReduction: number;
  damageBonuses: number[];
  amplifications: number[];
  specialBonuses: number[];
  critRate: number;
  critDamage: number;
};

export type Insight = {
  severity: 'info' | 'tip' | 'warning';
  title: string;
  message: string;
};

export type DamageResult = {
  formulaVersion: string;
  formulaConfidence: string;
  baseDamage: number;
  effectiveResistance: number;
  resistanceMultiplier: number;
  defenseMultiplier: number;
  damageReductionMultiplier: number;
  elementReductionMultiplier: number;
  damageBonusMultiplier: number;
  amplificationMultiplier: number;
  specialDamageMultiplier: number;
  nonCriticalDamage: number;
  criticalDamage: number;
  expectedDamage: number;
  insights: Insight[];
};

export type AIAnalysisRequest = {
  provider: string;
  endpoint: string;
  model: string;
  apiKey: string;
  mode: string;
  context: string;
  dataJson: string;
};

export type AIAnalysisResult = {
  text: string;
  provider: string;
  model: string;
};

export type BuildConfig = {
  buildId: number;
  scalingType: 'ATK' | 'HP' | 'DEF';
  baseAtk: number;
  baseHp: number;
  baseDef: number;
  motionValue: number;
  flatDamage: number;
  enemyLevel: number;
  enemyResistance: number;
  defenseIgnore: number;
  damageReduction: number;
  elementReduction: number;
  extraDamageBonusesJson: string;
};

export type BuildStats = {
  baseAtk: number;
  weaponAtk: number;
  atkPercent: number;
  flatAtk: number;
  totalAtk: number;
  baseHp: number;
  hpPercent: number;
  flatHp: number;
  totalHp: number;
  baseDef: number;
  defPercent: number;
  flatDef: number;
  totalDef: number;
  critRate: number;
  critDamage: number;
  energyRegen: number;
  damageBonuses: Record<string, number>;
  unparsedStats: string[];
  scalingStat: number;
};

export type BuildEvaluation = {
  build: Build;
  config: BuildConfig;
  stats: BuildStats;
  damage: DamageResult;
};

export type Buff = {
  id: number;
  teamId: number;
  sourceSlot: number;
  targetSlot: number;
  name: string;
  group: string;
  value: number;
  scope: string;
  condition: string;
  active: boolean;
  duration: number;
  triggerAction: string;
};

export type RotationAction = {
  id: number;
  order: number;
  slot: number;
  actionType: string;
  name: string;
  motionValue: number;
  castTime: number;
  energy: number;
  concerto: number;
  cooldown: number;
  energyCost: number;
  notes: string;
};

export type Rotation = {
  id: number;
  teamId: number;
  name: string;
  duration: number;
  notes: string;
  actions: RotationAction[];
};

export type RotationResult = {
  rotation: Rotation;
  actions: {
    action: RotationAction;
    damage: number;
    startTime: number;
    endTime: number;
    activeBuffs: string[];
    expiredBuffs: string[];
  }[];
  totalDamage: number;
  duration: number;
  dps: number;
  energyBySlot: Record<number, number>;
  concertoBySlot: Record<number, number>;
  fieldTimeBySlot: Record<number, number>;
  warnings: string[];
  errors: string[];
};

export type TeamTheorycraft = {
  team: Team;
  buffs: Buff[];
  rotations: Rotation[];
  warnings: string[];
};

export type AIMessage = {
  id: number;
  conversationId: number;
  role: 'user' | 'assistant';
  content: string;
  createdAt: string;
};

export type AIConversation = {
  id: number;
  title: string;
  contextType: string;
  contextId?: number;
  provider: string;
  model: string;
  createdAt: string;
  updatedAt: string;
  messages: AIMessage[];
  sources: KnowledgeSource[];
};

export type AssistantRequest = {
  conversationId: number;
  contextType: string;
  contextId?: number;
  question: string;
  endpoint: string;
  model: string;
  provider: string;
  apiKey: string;
  mode: string;
};
export type AIProviderStatus = {
  provider: string;
  online: boolean;
  models: string[];
  message: string;
};
export type CharacterGuide = {
  id: string;
  characterId: number;
  name: string;
  source: string;
  likeCount: number;
  language: string;
  teams: number[][];
  dataJson: string;
  syncedAt: string;
};
export type BuildExportIcons = {
  elementIconPath: string;
  weaponTypeIconPath: string;
};
export type KnowledgeSource = {
  entityType: string;
  entityId: string;
  title: string;
  snippet: string;
};

export type AppSettings = {
  density: 'compact' | 'comfortable' | 'spacious';
  sidebarCollapsed: boolean;
  dataSource: 'nanoka' | 'arikatsu';
  dataChannel: 'live' | 'latest' | string;
  dataVersion: string;
  aiProvider: 'ollama' | 'lmstudio' | 'gemini';
  aiEndpoint: string;
  aiModel: string;
  aiMode: 'strict' | 'assisted' | 'general';
  reduceMotion: boolean;
};
export type DataSourceOption = {
  id: string;
  provider: 'nanoka' | 'arikatsu';
  channel: string;
  version: string;
  label: string;
  description: string;
  syncReady: boolean;
  preRelease: boolean;
};
export type AccountSummary = {
  id: number;
  name: string;
  notes: string;
  astrite: number;
  radiantTides: number;
  ownedCharacters: number;
  ownedWeapons: number;
  ownedEchoes: number;
};
export type PlannerGoal = {
  id: number;
  title: string;
  goalType: string;
  targetName: string;
  requiredAmount: number;
  ownedAmount: number;
  shellCredits: number;
  priority: number;
  dueDate: string;
  completed: boolean;
  notes: string;
  createdAt: string;
  updatedAt: string;
};
export type ConveneRecord = {
  id: number;
  banner: string;
  bannerType: string;
  itemName: string;
  rarity: number;
  pullNumber: number;
  guaranteed: boolean;
  obtainedAt: string;
  notes: string;
};
export type ConveneProfile = {
  id: number;
  playerId: string;
  serverId: string;
  region: string;
  languageCode: string;
  lastImportedAt: string;
  historyPartial: boolean;
};
export type ConvenePull = {
  id: number;
  profileId: number;
  poolType: number;
  poolName: string;
  resourceId: string;
  resourceType: string;
  itemName: string;
  rarity: number;
  quantity: number;
  obtainedAt: string;
  sourceIndex: number;
  iconPath: string;
};
export type ConvenePoolSummary = {
  poolType: number;
  name: string;
  shortName: string;
  kind: string;
  total: number;
  count5: number;
  count4: number;
  count3: number;
  currentPity: number;
  hardPity: number;
  currentPity4: number;
  averagePity5: number;
  guaranteeState: 'guaranteed' | 'not_guaranteed' | 'not_applicable' | 'unknown';
  historyPartial: boolean;
  recentFiveStar: ConvenePull[];
};
export type ConveneOverview = {
  profile?: ConveneProfile;
  pools: ConvenePoolSummary[];
  pulls: ConvenePull[];
  total: number;
  count5: number;
  count4: number;
  count3: number;
  lastImportedAt: string;
};
export type ConveneImportResult = {
  imported: number;
  duplicates: number;
  poolsUpdated: number;
  profile: ConveneProfile;
  overview: ConveneOverview;
  source: string;
  historyPartial: boolean;
};
export type Enemy = {
  id: number;
  name: string;
  level: number;
  resistance: number;
  damageReduction: number;
  elementReduction: number;
  notes: string;
};
export type FormulaVersion = {
  id: number;
  name: string;
  gameVersion: string;
  defenseConstant: number;
  levelFactor: number;
  confidence: string;
  references: string;
  roundingPolicy: string;
  active: boolean;
};
export type BuildVersion = { id: number; buildId: number; snapshot: string; createdAt: string };
export type DashboardSummary = {
  characters: number;
  weapons: number;
  echoes: number;
  builds: number;
  teams: number;
  recentBuilds: Build[];
};
export type ArchiveReport = {
  builds: number;
  teams: number;
  buffs: number;
  rotations: number;
  goals: number;
  convenes: number;
  warnings: string[];
};
export type Diagnostics = {
  dataDirectory: string;
  databasePath: string;
  databaseBytes: number;
  migrations: number;
  gameVersion: string;
  catalogCount: number;
  goVersion: string;
};

export type CharacterProfile = {
  character: Character;
  description: string;
  birthday: string;
  gender: string;
  region: string;
  faction: string;
  talentName: string;
  talentDescription: string;
  signatureWeapon?: Weapon;
  skills: Skill[];
  chains: ResonanceChain[];
  progression: CharacterProgression;
  extras: CharacterExtras;
};

export type BackendAPI = {
  ListCharacters(filter: CharacterFilter): Promise<Character[]>;
  SearchCharacterContent(query: string, limit: number): Promise<CharacterContentSearchResult[]>;
  GetCharacter(id: number): Promise<CharacterProfile>;
  CalculateCharacterProgression(request: ProgressionPlanRequest): Promise<ProgressionPlan>;
  CatalogStatus(): Promise<CatalogStatus>;
  SyncCharacters(): Promise<SyncResult>;
  CancelSync(): Promise<boolean>;
  RestoreLatestSnapshot(): Promise<string>;
  UpdateCharacterAccount(update: CharacterAccountUpdate): Promise<void>;
  ListWeapons(filter: WeaponFilter): Promise<Weapon[]>;
  GetWeapon(id: number): Promise<Weapon>;
  UpdateWeaponAccount(update: WeaponAccountUpdate): Promise<void>;
  ListEchoes(filter: EchoFilter): Promise<Echo[]>;
  GetEcho(id: number): Promise<Echo>;
  ListSonatas(): Promise<Sonata[]>;
  ListOwnedEchoes(): Promise<OwnedEcho[]>;
  SaveOwnedEcho(echo: OwnedEcho): Promise<OwnedEcho>;
  DeleteOwnedEcho(id: number): Promise<void>;
  ListBuilds(): Promise<Build[]>;
  SaveBuild(build: Build): Promise<Build>;
  DuplicateBuild(id: number): Promise<Build>;
  DeleteBuild(id: number): Promise<void>;
  RestoreBuild(id: number): Promise<void>;
  ListTeams(): Promise<Team[]>;
  SaveTeam(team: Team): Promise<Team>;
  DuplicateTeam(id: number): Promise<Team>;
  DeleteTeam(id: number): Promise<void>;
  RestoreTeam(id: number): Promise<void>;
  CalculateDamage(input: DamageInput): Promise<DamageResult>;
  AnalyzeWithAI(request: AIAnalysisRequest): Promise<AIAnalysisResult>;
  EvaluateBuild(id: number): Promise<BuildEvaluation>;
  SaveBuildConfig(config: BuildConfig): Promise<BuildEvaluation>;
  GetTeamTheorycraft(teamId: number): Promise<TeamTheorycraft>;
  SaveBuff(buff: Buff): Promise<Buff>;
  DeleteBuff(id: number): Promise<void>;
  ListAIConversations(): Promise<AIConversation[]>;
  AssistantChat(request: AssistantRequest): Promise<AIConversation>;
  AssistantChatStream(request: AssistantRequest): Promise<AIConversation>;
  DeleteAIConversation(id: number): Promise<void>;
  TestAIProvider(request: AIAnalysisRequest): Promise<AIProviderStatus>;
  ListCharacterGuides(characterId: number): Promise<CharacterGuide[]>;
  ListAllCharacterGuides(): Promise<CharacterGuide[]>;
  SyncCharacterGuides(characterId: number, language: string): Promise<CharacterGuide[]>;
  GetBuildExportIcons(characterId: number): Promise<BuildExportIcons>;
  SearchLocalKnowledge(query: string, limit: number): Promise<KnowledgeSource[]>;
  ListBuildVersions(id: number): Promise<BuildVersion[]>;
  GetSettings(): Promise<AppSettings>;
  SaveSettings(settings: AppSettings): Promise<AppSettings>;
  ListDataSourceOptions(): Promise<DataSourceOption[]>;
  GetAccountSummary(): Promise<AccountSummary>;
  SaveAccountSummary(account: AccountSummary): Promise<AccountSummary>;
  ListEnemies(): Promise<Enemy[]>;
  ListFormulaVersions(): Promise<FormulaVersion[]>;
  DashboardSummary(): Promise<DashboardSummary>;
  GetConveneOverview(): Promise<ConveneOverview>;
  DeleteConveneHistory(): Promise<void>;
  ImportConveneURL(url: string): Promise<ConveneImportResult>;
  ImportConveneFromGame(): Promise<ConveneImportResult>;
  ImportConveneFromLogFile(): Promise<ConveneImportResult>;
  ExportArchive(): Promise<string>;
  ImportArchive(payload: string): Promise<ArchiveReport>;
  CreateManualBackup(): Promise<string>;
  ListBackups(): Promise<string[]>;
  RestoreBackup(name: string): Promise<string>;
  Diagnostics(): Promise<Diagnostics>;
};
