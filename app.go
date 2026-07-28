package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	assetcache "wavearchive/internal/assets"
	"wavearchive/internal/database"
	"wavearchive/internal/domain"
	"wavearchive/internal/repository"
	guidesource "wavearchive/internal/sources/guide"
	"wavearchive/internal/sources/nanoka"
	"wavearchive/internal/usecase"
)

type App struct {
	ctx         context.Context
	logger      *slog.Logger
	db          *database.Database
	assets      *assetcache.Cache
	catalog     *usecase.CharacterCatalog
	weapons     *usecase.WeaponCatalog
	echoes      *usecase.EchoCatalog
	builds      *usecase.BuildManager
	teams       *usecase.TeamManager
	damage      *usecase.DamageCalculator
	ai          *usecase.AIAnalyzer
	evaluator   *usecase.BuildEvaluator
	theorycraft *usecase.TheorycraftManager
	assistant   *usecase.AssistantService
	workspace   *usecase.WorkspaceManager
	guides      *usecase.GuideManager
	initErr     error
	dataDir     string
	dbPath      string

	syncMu     sync.Mutex
	syncCancel context.CancelFunc
	syncing    bool
}

func NewApp(logger *slog.Logger) *App {
	return &App{logger: logger}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.initErr = a.initResources()
	if a.initErr != nil {
		a.logger.Error("failed to initialise application", "error", a.initErr)
	}
}

func (a *App) initResources() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	dataDir := filepath.Join(configDir, "WaveArchive")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}

	dbPath, err := database.ResolveApplicationPath(filepath.Join(dataDir, "wavearchive.db"))
	if err != nil {
		return err
	}
	db, err := database.Open(dbPath)
	if err != nil {
		return err
	}

	repo := repository.NewCharacterSQLite(db.SQL())
	weaponRepo := repository.NewWeaponSQLite(db.SQL())
	echoRepo := repository.NewEchoSQLite(db.SQL())
	buildRepo := repository.NewBuildSQLite(db.SQL())
	teamRepo := repository.NewTeamSQLite(db.SQL())
	theoryRepo := repository.NewTheorycraftSQLite(db.SQL())
	aiHistoryRepo := repository.NewAIHistorySQLite(db.SQL())
	workspaceRepo := repository.NewWorkspaceSQLite(db.SQL())
	guideRepo := repository.NewGuideSQLite(db.SQL())
	client := nanoka.NewClient(nil)
	a.assets = assetcache.NewCache(filepath.Join(dataDir, "assets"), nil)
	a.db = db
	a.dataDir = dataDir
	a.dbPath = dbPath
	a.catalog = usecase.NewCharacterCatalog(repo, client, a.assets, a.logger)
	a.weapons = usecase.NewWeaponCatalog(weaponRepo, client, a.assets, a.logger)
	a.echoes = usecase.NewEchoCatalog(echoRepo, client, a.assets, a.logger)
	a.builds = usecase.NewBuildManager(buildRepo)
	a.teams = usecase.NewTeamManager(teamRepo)
	a.damage = usecase.NewDamageCalculator()
	a.ai = usecase.NewAIAnalyzer(nil)
	a.evaluator = usecase.NewBuildEvaluator(buildRepo, weaponRepo, theoryRepo, teamRepo, a.damage)
	a.theorycraft = usecase.NewTheorycraftManager(theoryRepo, a.evaluator)
	a.assistant = usecase.NewAssistantService(a.ai, aiHistoryRepo, a.evaluator, repo, guideRepo)
	a.workspace = usecase.NewWorkspaceManager(workspaceRepo)
	a.guides = usecase.NewGuideManager(guideRepo, guidesource.NewClient(nil))
	return nil
}

func (a *App) assetServerHandler() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if a.assets == nil {
			http.NotFound(writer, request)
			return
		}
		a.assets.Handler().ServeHTTP(writer, request)
	})
}

func (a *App) Close() {
	if a.db != nil {
		_ = a.db.Close()
	}
}

func (a *App) ListCharacters(filter domain.CharacterFilter) ([]domain.Character, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.catalog.List(a.context(), filter)
}

func (a *App) CatalogStatus() (domain.CatalogStatus, error) {
	if err := a.ready(); err != nil {
		return domain.CatalogStatus{}, err
	}
	return a.catalog.Status(a.context())
}

func (a *App) GetCharacter(id int64) (domain.CharacterProfile, error) {
	if err := a.ready(); err != nil {
		return domain.CharacterProfile{}, err
	}
	return a.catalog.Get(a.context(), id)
}

func (a *App) CalculateCharacterProgression(request domain.ProgressionPlanRequest) (domain.ProgressionPlan, error) {
	if err := a.ready(); err != nil {
		return domain.ProgressionPlan{}, err
	}
	return a.catalog.CalculateProgression(a.context(), request)
}

func (a *App) UpdateCharacterAccount(update domain.CharacterAccountUpdate) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.catalog.UpdateAccount(a.context(), update)
}

func (a *App) ListWeapons(filter domain.WeaponFilter) ([]domain.Weapon, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.weapons.List(a.context(), filter)
}

func (a *App) GetWeapon(id int64) (domain.Weapon, error) {
	if err := a.ready(); err != nil {
		return domain.Weapon{}, err
	}
	return a.weapons.Get(a.context(), id)
}

func (a *App) UpdateWeaponAccount(update domain.WeaponAccountUpdate) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.weapons.UpdateAccount(a.context(), update)
}

func (a *App) ListEchoes(filter domain.EchoFilter) ([]domain.Echo, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.echoes.List(a.context(), filter)
}

func (a *App) GetEcho(id int64) (domain.Echo, error) {
	if err := a.ready(); err != nil {
		return domain.Echo{}, err
	}
	return a.echoes.Get(a.context(), id)
}

func (a *App) ListSonatas() ([]domain.Sonata, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.echoes.ListSonatas(a.context())
}

func (a *App) ListOwnedEchoes() ([]domain.OwnedEcho, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.echoes.ListOwned(a.context())
}

func (a *App) SaveOwnedEcho(item domain.OwnedEcho) (domain.OwnedEcho, error) {
	if err := a.ready(); err != nil {
		return domain.OwnedEcho{}, err
	}
	return a.echoes.SaveOwned(a.context(), item)
}

func (a *App) DeleteOwnedEcho(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.echoes.DeleteOwned(a.context(), id)
}

func (a *App) ListBuilds() ([]domain.Build, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.builds.List(a.context())
}

func (a *App) SaveBuild(build domain.Build) (domain.Build, error) {
	if err := a.ready(); err != nil {
		return domain.Build{}, err
	}
	if build.GameVersion == "" {
		if status, err := a.catalog.Status(a.context()); err == nil {
			build.GameVersion = status.Version
		}
	}
	return a.builds.Save(a.context(), build)
}

func (a *App) DuplicateBuild(id int64) (domain.Build, error) {
	if err := a.ready(); err != nil {
		return domain.Build{}, err
	}
	return a.builds.Duplicate(a.context(), id)
}

func (a *App) DeleteBuild(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.builds.Delete(a.context(), id)
}

func (a *App) RestoreBuild(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.builds.Restore(a.context(), id)
}

func (a *App) ListBuildVersions(id int64) ([]domain.BuildVersion, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.builds.History(a.context(), id)
}

func (a *App) GetSettings() (domain.AppSettings, error) {
	if err := a.ready(); err != nil {
		return domain.AppSettings{}, err
	}
	return a.workspace.Settings(a.context())
}
func (a *App) SaveSettings(settings domain.AppSettings) (domain.AppSettings, error) {
	if err := a.ready(); err != nil {
		return domain.AppSettings{}, err
	}
	return a.workspace.SaveSettings(a.context(), settings)
}
func (a *App) GetAccountSummary() (domain.AccountSummary, error) {
	if err := a.ready(); err != nil {
		return domain.AccountSummary{}, err
	}
	return a.workspace.Account(a.context())
}
func (a *App) SaveAccountSummary(account domain.AccountSummary) (domain.AccountSummary, error) {
	if err := a.ready(); err != nil {
		return domain.AccountSummary{}, err
	}
	return a.workspace.SaveAccount(a.context(), account)
}
func (a *App) ListEnemies() ([]domain.Enemy, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.workspace.Enemies(a.context())
}
func (a *App) ListFormulaVersions() ([]domain.FormulaVersion, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.workspace.Formulas(a.context())
}
func (a *App) DashboardSummary() (domain.DashboardSummary, error) {
	if err := a.ready(); err != nil {
		return domain.DashboardSummary{}, err
	}
	summary, err := a.workspace.Dashboard(a.context())
	if err != nil {
		return summary, err
	}
	if builds, listErr := a.builds.List(a.context()); listErr == nil {
		if len(builds) > 3 {
			builds = builds[:3]
		}
		summary.RecentBuilds = builds
	}
	return summary, nil
}

func (a *App) ListTeams() ([]domain.Team, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.teams.List(a.context())
}

func (a *App) SaveTeam(team domain.Team) (domain.Team, error) {
	if err := a.ready(); err != nil {
		return domain.Team{}, err
	}
	if team.GameVersion == "" {
		if status, err := a.catalog.Status(a.context()); err == nil {
			team.GameVersion = status.Version
		}
	}
	return a.teams.Save(a.context(), team)
}

func (a *App) DuplicateTeam(id int64) (domain.Team, error) {
	if err := a.ready(); err != nil {
		return domain.Team{}, err
	}
	return a.teams.Duplicate(a.context(), id)
}

func (a *App) DeleteTeam(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.teams.Delete(a.context(), id)
}

func (a *App) RestoreTeam(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.teams.Restore(a.context(), id)
}

func (a *App) CalculateDamage(input domain.DamageInput) (domain.DamageResult, error) {
	if err := a.ready(); err != nil {
		return domain.DamageResult{}, err
	}
	return a.damage.Calculate(input)
}

func (a *App) AnalyzeWithAI(request domain.AIAnalysisRequest) (domain.AIAnalysisResult, error) {
	if err := a.ready(); err != nil {
		return domain.AIAnalysisResult{}, err
	}
	return a.ai.Analyze(a.context(), request)
}
func (a *App) TestAIProvider(request domain.AIAnalysisRequest) (domain.AIProviderStatus, error) {
	if err := a.ready(); err != nil {
		return domain.AIProviderStatus{}, err
	}
	return a.ai.Status(a.context(), request)
}

func (a *App) EvaluateBuild(id int64) (domain.BuildEvaluation, error) {
	if err := a.ready(); err != nil {
		return domain.BuildEvaluation{}, err
	}
	return a.evaluator.Evaluate(a.context(), id, nil)
}

func (a *App) SaveBuildConfig(config domain.BuildConfig) (domain.BuildEvaluation, error) {
	if err := a.ready(); err != nil {
		return domain.BuildEvaluation{}, err
	}
	return a.evaluator.SaveConfig(a.context(), config)
}

func (a *App) GetTeamTheorycraft(teamID int64) (domain.TeamTheorycraft, error) {
	if err := a.ready(); err != nil {
		return domain.TeamTheorycraft{}, err
	}
	return a.theorycraft.Team(a.context(), teamID)
}

func (a *App) SaveBuff(buff domain.Buff) (domain.Buff, error) {
	if err := a.ready(); err != nil {
		return domain.Buff{}, err
	}
	return a.theorycraft.SaveBuff(a.context(), buff)
}

func (a *App) DeleteBuff(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.theorycraft.DeleteBuff(a.context(), id)
}

func (a *App) ListAIConversations() ([]domain.AIConversation, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.assistant.List(a.context())
}

func (a *App) AssistantChat(request domain.AssistantRequest) (domain.AIConversation, error) {
	if err := a.ready(); err != nil {
		return domain.AIConversation{}, err
	}
	return a.assistant.Chat(a.context(), request)
}

func (a *App) AssistantChatStream(request domain.AssistantRequest) (domain.AIConversation, error) {
	if err := a.ready(); err != nil {
		return domain.AIConversation{}, err
	}
	return a.assistant.ChatStream(a.context(), request, func(chunk string) {
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "ai:chunk", chunk)
		}
	})
}

func (a *App) DeleteAIConversation(id int64) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.assistant.Delete(a.context(), id)
}

func (a *App) ListCharacterGuides(characterID int64) ([]domain.CharacterGuide, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.guides.List(a.context(), characterID)
}

func (a *App) ListAllCharacterGuides() ([]domain.CharacterGuide, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.guides.ListAll(a.context())
}
func (a *App) SyncCharacterGuides(characterID int64, language string) ([]domain.CharacterGuide, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.guides.Sync(a.context(), characterID, language)
}
func (a *App) SearchLocalKnowledge(query string, limit int) ([]domain.KnowledgeSource, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	return a.guides.Search(a.context(), query, limit)
}

func (a *App) SyncCharacters() (domain.SyncResult, error) {
	if err := a.ready(); err != nil {
		return domain.SyncResult{}, err
	}
	a.syncMu.Lock()
	if a.syncing {
		a.syncMu.Unlock()
		return domain.SyncResult{}, errors.New("a synchronization is already running")
	}
	ctx, cancel := context.WithCancel(a.context())
	a.syncing = true
	a.syncCancel = cancel
	a.syncMu.Unlock()
	defer func() {
		cancel()
		a.syncMu.Lock()
		a.syncing = false
		a.syncCancel = nil
		a.syncMu.Unlock()
	}()

	if err := a.createSnapshot(ctx); err != nil {
		a.logger.Warn("could not create pre-sync snapshot", "error", err)
	}
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "catalog:sync", map[string]any{"stage": "detecting", "progress": 0})
	}
	result, err := a.catalog.Sync(ctx, func(stage string, progress int) {
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "catalog:sync", map[string]any{"stage": stage, "progress": progress * 55 / 100})
		}
	})
	if err != nil {
		if errors.Is(err, context.Canceled) && a.ctx != nil {
			runtime.EventsEmit(a.ctx, "catalog:sync", map[string]any{"stage": "cancelled", "progress": 0})
		}
		a.logger.Error("character sync failed", "error", err)
		return domain.SyncResult{}, err
	}
	if _, err := a.weapons.Sync(ctx, result.Version, func(stage string, progress int) {
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "catalog:sync", map[string]any{"stage": stage, "progress": 55 + progress*23/100})
		}
	}); err != nil {
		a.logger.Error("weapon sync failed", "error", err)
		return domain.SyncResult{}, err
	}
	if _, err := a.echoes.Sync(ctx, result.Version, func(stage string, progress int) {
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "catalog:sync", map[string]any{"stage": stage, "progress": 78 + progress*22/100})
		}
	}); err != nil {
		a.logger.Error("echo sync failed", "error", err)
		return domain.SyncResult{}, err
	}
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "catalog:sync", map[string]any{"stage": "done", "progress": 100})
	}
	return result, nil
}

func (a *App) CancelSync() bool {
	a.syncMu.Lock()
	defer a.syncMu.Unlock()
	if !a.syncing || a.syncCancel == nil {
		return false
	}
	a.syncCancel()
	return true
}

func (a *App) RestoreLatestSnapshot() (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	a.syncMu.Lock()
	defer a.syncMu.Unlock()
	if a.syncing {
		return "", errors.New("cancel the current synchronization before restoring")
	}
	snapshots, err := a.snapshotFiles()
	if err != nil {
		return "", err
	}
	if len(snapshots) == 0 {
		return "", errors.New("no snapshot is available")
	}
	latest := snapshots[len(snapshots)-1]
	return a.restoreDatabase(latest)
}

func (a *App) CreateManualBackup() (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	name := "manual-" + time.Now().UTC().Format("20060102T150405.000000000Z") + ".db"
	path := filepath.Join(a.dataDir, "backups", name)
	if err := a.db.Backup(path); err != nil {
		return "", err
	}
	return name, nil
}
func (a *App) ListBackups() ([]string, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	dir := filepath.Join(a.dataDir, "backups")
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".db") {
			names = append(names, entry.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(names)))
	return names, nil
}
func (a *App) RestoreBackup(name string) (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	a.syncMu.Lock()
	defer a.syncMu.Unlock()
	if a.syncing {
		return "", errors.New("cancel synchronization before restoring a backup")
	}
	if filepath.Base(name) != name || !strings.HasSuffix(name, ".db") {
		return "", errors.New("invalid backup name")
	}
	path := filepath.Join(a.dataDir, "backups", name)
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	return a.restoreDatabase(path)
}

func (a *App) ExportArchive() (string, error) {
	if err := a.ready(); err != nil {
		return "", err
	}
	ctx := a.context()
	bundle := domain.ArchiveBundle{SchemaVersion: 1, ExportedAt: time.Now().UTC().Format(time.RFC3339)}
	var err error
	if bundle.Builds, err = a.builds.List(ctx); err != nil {
		return "", err
	}
	if bundle.Teams, err = a.teams.List(ctx); err != nil {
		return "", err
	}
	for _, team := range bundle.Teams {
		theory, theoryErr := a.theorycraft.Team(ctx, team.ID)
		if theoryErr == nil {
			bundle.Theorycraft = append(bundle.Theorycraft, theory)
		}
	}
	bundle.Settings, _ = a.workspace.Settings(ctx)
	bundle.Goals, _ = a.workspace.Goals(ctx)
	bundle.Convenes, _ = a.workspace.Convenes(ctx)
	raw, err := json.MarshalIndent(bundle, "", "  ")
	return string(raw), err
}

func (a *App) ImportArchive(payload string) (domain.ArchiveReport, error) {
	report := domain.ArchiveReport{Warnings: []string{}}
	if err := a.ready(); err != nil {
		return report, err
	}
	if len(payload) == 0 || len(payload) > 16<<20 {
		return report, errors.New("archive must be between 1 byte and 16 MiB")
	}
	var bundle domain.ArchiveBundle
	if err := json.Unmarshal([]byte(payload), &bundle); err != nil {
		return report, fmt.Errorf("invalid archive JSON: %w", err)
	}
	if bundle.SchemaVersion != 1 {
		return report, fmt.Errorf("unsupported archive schema %d", bundle.SchemaVersion)
	}
	if len(bundle.Builds) > 1000 || len(bundle.Teams) > 500 || len(bundle.Goals) > 2000 || len(bundle.Convenes) > 10000 {
		return report, errors.New("archive exceeds safe item limits")
	}
	safety := filepath.Join(a.dataDir, "snapshots", "pre-import-"+time.Now().UTC().Format("20060102T150405.000000000Z")+".db")
	if err := a.db.Backup(safety); err != nil {
		return report, fmt.Errorf("create pre-import backup: %w", err)
	}
	ctx := a.context()
	buildMap := map[int64]int64{}
	teamMap := map[int64]int64{}
	for _, build := range bundle.Builds {
		old := build.ID
		build.ID = 0
		if len(build.Echoes) > 0 {
			build.Echoes = nil
			report.Warnings = append(report.Warnings, "Echoes pessoais não foram vinculados; use backup completo para preservar IDs de inventário.")
		}
		saved, err := a.builds.Save(ctx, build)
		if err != nil {
			return report, fmt.Errorf("import build %q: %w", build.Name, err)
		}
		buildMap[old] = saved.ID
		report.Builds++
	}
	for _, team := range bundle.Teams {
		old := team.ID
		team.ID = 0
		for index := range team.Members {
			if team.Members[index].BuildID != nil {
				if mapped, ok := buildMap[*team.Members[index].BuildID]; ok {
					team.Members[index].BuildID = &mapped
				} else {
					team.Members[index].BuildID = nil
				}
			}
		}
		saved, err := a.teams.Save(ctx, team)
		if err != nil {
			return report, fmt.Errorf("import team %q: %w", team.Name, err)
		}
		teamMap[old] = saved.ID
		report.Teams++
	}
	for _, theory := range bundle.Theorycraft {
		newTeam, ok := teamMap[theory.Team.ID]
		if !ok {
			continue
		}
		for _, buff := range theory.Buffs {
			buff.ID = 0
			buff.TeamID = newTeam
			if _, err := a.theorycraft.SaveBuff(ctx, buff); err != nil {
				return report, err
			}
			report.Buffs++
		}
		for _, rotation := range theory.Rotations {
			rotation.ID = 0
			rotation.TeamID = newTeam
			for index := range rotation.Actions {
				rotation.Actions[index].ID = 0
			}
			if _, err := a.theorycraft.SaveRotation(ctx, rotation); err != nil {
				return report, err
			}
			report.Rotations++
		}
	}
	for _, goal := range bundle.Goals {
		goal.ID = 0
		if _, err := a.workspace.SaveGoal(ctx, goal); err != nil {
			return report, err
		}
		report.Goals++
	}
	for _, record := range bundle.Convenes {
		record.ID = 0
		if _, err := a.workspace.SaveConvene(ctx, record); err != nil {
			return report, err
		}
		report.Convenes++
	}
	if bundle.Settings.Density != "" {
		_, _ = a.workspace.SaveSettings(ctx, bundle.Settings)
	}
	return report, nil
}

func (a *App) Diagnostics() (domain.Diagnostics, error) {
	if err := a.ready(); err != nil {
		return domain.Diagnostics{}, err
	}
	var d domain.Diagnostics
	d.DataDirectory = a.dataDir
	d.DatabasePath = a.dbPath
	if info, err := os.Stat(a.dbPath); err == nil {
		d.DatabaseBytes = info.Size()
	}
	_ = a.db.SQL().QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&d.Migrations)
	if status, err := a.catalog.Status(a.context()); err == nil {
		d.GameVersion = status.Version
		d.CatalogCount = status.Count
	}
	if build, ok := debug.ReadBuildInfo(); ok {
		d.GoVersion = build.GoVersion
	}
	return d, nil
}

func (a *App) restoreDatabase(snapshot string) (string, error) {
	safety := filepath.Join(a.dataDir, "snapshots", "before-restore-"+time.Now().UTC().Format("20060102T150405.000000000Z")+".db")
	if err := a.db.Backup(safety); err != nil {
		return "", fmt.Errorf("create restore safety snapshot: %w", err)
	}
	_ = a.db.Checkpoint()
	if err := a.db.Close(); err != nil {
		return "", err
	}
	a.db = nil
	a.catalog = nil
	a.weapons = nil
	a.echoes = nil
	a.builds = nil
	a.teams = nil
	a.damage = nil
	a.ai = nil
	a.evaluator = nil
	a.theorycraft = nil
	a.assistant = nil
	a.workspace = nil
	a.guides = nil
	if err := database.RestoreFile(snapshot, a.dbPath); err != nil {
		a.initErr = a.initResources()
		return "", err
	}
	a.initErr = a.initResources()
	if a.initErr != nil {
		return "", a.initErr
	}
	return filepath.Base(snapshot), nil
}

func (a *App) createSnapshot(ctx context.Context) error {
	status, err := a.catalog.Status(ctx)
	if err != nil || status.Count == 0 {
		return err
	}
	snapshotDir := filepath.Join(a.dataDir, "snapshots")
	name := "pre-sync-" + time.Now().UTC().Format("20060102T150405.000000000Z") + ".db"
	if err := a.db.Backup(filepath.Join(snapshotDir, name)); err != nil {
		return err
	}
	files, err := a.snapshotFiles()
	if err != nil {
		return err
	}
	for len(files) > 5 {
		if err := os.Remove(files[0]); err != nil {
			return err
		}
		files = files[1:]
	}
	return nil
}

func (a *App) snapshotFiles() ([]string, error) {
	dir := filepath.Join(a.dataDir, "snapshots")
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".db") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	sort.Strings(files)
	return files, nil
}

func (a *App) ready() error {
	if a.initErr != nil {
		return fmt.Errorf("application database is unavailable: %w", a.initErr)
	}
	if a.catalog == nil || a.weapons == nil || a.echoes == nil || a.builds == nil || a.teams == nil ||
		a.damage == nil || a.ai == nil || a.evaluator == nil || a.theorycraft == nil || a.assistant == nil || a.workspace == nil || a.guides == nil {
		return errors.New("application is still starting")
	}
	return nil
}

func (a *App) context() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}
