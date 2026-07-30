package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"wavearchive/internal/domain"
	convenesource "wavearchive/internal/sources/convene"
)

const maxConveneLogTail = 16 << 20

var conveneURLPattern = regexp.MustCompile(`(?i)https://aki-gm-resources(?:-oversea)?\.aki-game\.(?:net|com)/aki/gacha/index\.html#/record[^\s"'<>)]*`)

type ConveneHistory struct {
	repository domain.ConveneHistoryRepository
	client     *convenesource.Client
}

func NewConveneHistory(
	repository domain.ConveneHistoryRepository,
	client *convenesource.Client,
) *ConveneHistory {
	return &ConveneHistory{repository: repository, client: client}
}

func (h *ConveneHistory) Overview(ctx context.Context) (domain.ConveneOverview, error) {
	profile, err := h.repository.GetConveneProfile(ctx)
	if err != nil {
		return domain.ConveneOverview{}, err
	}
	pulls, err := h.repository.ListConvenePulls(ctx)
	if err != nil {
		return domain.ConveneOverview{}, err
	}
	pools := fallbackConvenePoolDefinitions()
	if profile != nil {
		stored, listErr := h.repository.ListConvenePoolDefinitions(ctx, profile.ID)
		if listErr != nil {
			return domain.ConveneOverview{}, listErr
		}
		if len(stored) > 0 {
			pools = stored
		}
	}
	return buildConveneOverview(profile, pulls, pools), nil
}

func (h *ConveneHistory) Delete(ctx context.Context) error {
	return h.repository.DeleteConveneHistory(ctx)
}

func (h *ConveneHistory) ImportURL(
	ctx context.Context,
	rawURL string,
	source string,
) (domain.ConveneImportResult, error) {
	credentials, err := convenesource.ParseHistoryURL(rawURL)
	if err != nil {
		return domain.ConveneImportResult{}, err
	}
	pools, catalogErr := h.client.FetchPoolCatalog(ctx, credentials)
	if catalogErr != nil {
		pools = convenesource.Pools
	}
	pulls, poolsUpdated, err := h.client.FetchAll(ctx, credentials, pools)
	if err != nil {
		return domain.ConveneImportResult{}, err
	}
	if len(pulls) == 0 {
		return domain.ConveneImportResult{}, errors.New("nenhum giro foi encontrado; abra o histórico no jogo e gere uma URL nova")
	}
	profile, imported, duplicates, err := h.repository.SaveImportedPulls(ctx, domain.ConveneImportPayload{
		PlayerID:     credentials.PlayerID,
		ServerID:     credentials.ServerID,
		Region:       credentials.Region,
		LanguageCode: credentials.LanguageCode,
		Pulls:        pulls,
		Pools:        poolDefinitions(pools),
	})
	if err != nil {
		return domain.ConveneImportResult{}, err
	}
	overview, err := h.Overview(ctx)
	if err != nil {
		return domain.ConveneImportResult{}, err
	}
	return domain.ConveneImportResult{
		Imported:       imported,
		Duplicates:     duplicates,
		PoolsUpdated:   poolsUpdated,
		Profile:        profile,
		Overview:       overview,
		Source:         source,
		HistoryPartial: true,
	}, nil
}

func (h *ConveneHistory) ImportLog(
	ctx context.Context,
	path string,
) (domain.ConveneImportResult, error) {
	rawURL, err := extractConveneURL(path)
	if err != nil {
		return domain.ConveneImportResult{}, err
	}
	return h.ImportURL(ctx, rawURL, "client_log")
}

func (h *ConveneHistory) ImportFromGame(ctx context.Context) (domain.ConveneImportResult, error) {
	logs := existingConveneLogs(conveneLogCandidates())
	for _, log := range logs {
		rawURL, err := extractConveneURL(log.path)
		if err != nil {
			continue
		}
		return h.ImportURL(ctx, rawURL, "game_log")
	}
	if len(logs) > 0 {
		return domain.ConveneImportResult{}, errors.New("os arquivos de log foram encontrados, mas não contêm uma URL de Convene; abra Convene → Histórico no jogo e tente novamente")
	}
	return domain.ConveneImportResult{}, errors.New("Client.log não encontrado ou sem URL recente; abra o Histórico de Convocação no jogo e tente novamente")
}

func extractConveneURL(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("abrir Client.log: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	offset := info.Size() - maxConveneLogTail
	if offset < 0 {
		offset = 0
	}
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return "", err
	}
	raw, err := io.ReadAll(io.LimitReader(file, maxConveneLogTail))
	if err != nil {
		return "", fmt.Errorf("ler Client.log: %w", err)
	}
	if strings.EqualFold(filepath.Base(path), "Client.log") {
		if rawURL := lastConveneURL(decodeConveneClientLog(raw)); rawURL != "" {
			return rawURL, nil
		}
	}
	if rawURL := lastConveneURL(raw); rawURL != "" {
		return rawURL, nil
	}
	return "", errors.New("URL do histórico não encontrada no arquivo de log")
}

func lastConveneURL(content []byte) string {
	matches := conveneURLPattern.FindAllString(string(content), -1)
	if len(matches) == 0 {
		return ""
	}
	return matches[len(matches)-1]
}

func decodeConveneClientLog(content []byte) []byte {
	decoded := make([]byte, len(content))
	for index, value := range content {
		mask := byte(0xEF)
		if value&1 != 0 {
			mask = 0xA5
		}
		decoded[index] = value ^ mask
	}
	return decoded
}

type conveneLogFile struct {
	path       string
	modifiedAt int64
}

func existingConveneLogs(candidates []string) []conveneLogFile {
	logs := make([]conveneLogFile, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		key := strings.ToLower(filepath.Clean(candidate))
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() {
			continue
		}
		logs = append(logs, conveneLogFile{path: candidate, modifiedAt: info.ModTime().UnixNano()})
	}
	sort.SliceStable(logs, func(left, right int) bool {
		return logs[left].modifiedAt > logs[right].modifiedAt
	})
	return logs
}

func conveneLogCandidates() []string {
	candidates := make([]string, 0, 80)
	for letter := 'C'; letter <= 'Z'; letter++ {
		root := string(letter) + `:\`
		bases := []string{
			filepath.Join(root, "Wuthering Waves"),
			filepath.Join(root, "Wuthering Waves", "Wuthering Waves Game"),
			filepath.Join(root, "Wuthering Waves Game"),
			filepath.Join(root, "Games", "Wuthering Waves"),
			filepath.Join(root, "Games", "Wuthering Waves", "Wuthering Waves Game"),
			filepath.Join(root, "Games", "Steam", "steamapps", "common", "Wuthering Waves"),
			filepath.Join(root, "Games", "Steam", "steamapps", "common", "Wuthering Waves", "Wuthering Waves Game"),
			filepath.Join(root, "Games", "Wuthering Waves Game"),
			filepath.Join(root, "SteamLibrary", "steamapps", "common", "Wuthering Waves"),
			filepath.Join(root, "Steam", "steamapps", "common", "Wuthering Waves"),
			filepath.Join(root, "Program Files (x86)", "Steam", "steamapps", "common", "Wuthering Waves"),
			filepath.Join(root, "Program Files", "Steam", "steamapps", "common", "Wuthering Waves"),
			filepath.Join(root, "Program Files", "Wuthering Waves"),
			filepath.Join(root, "Program Files (x86)", "Wuthering Waves"),
			filepath.Join(root, "XboxGames", "Wuthering Waves", "Content"),
		}
		epic, _ := filepath.Glob(filepath.Join(root, "Program Files", "Epic Games", "WutheringWaves*"))
		epic32, _ := filepath.Glob(filepath.Join(root, "Program Files (x86)", "Epic Games", "WutheringWaves*"))
		bases = append(bases, epic...)
		bases = append(bases, epic32...)
		for _, base := range bases {
			candidates = append(candidates,
				filepath.Join(base, "Client", "Saved", "Logs", "Client.log"),
				filepath.Join(base, "Wuthering Waves Game", "Client", "Saved", "Logs", "Client.log"),
				filepath.Join(base, "Client", "Binaries", "Win64", "ThirdParty", "KrPcSdk_Global", "KRSDKRes", "KRSDKWebView", "debug.log"),
				filepath.Join(base, "Wuthering Waves Game", "Client", "Binaries", "Win64", "ThirdParty", "KrPcSdk_Global", "KRSDKRes", "KRSDKWebView", "debug.log"),
			)
		}
	}
	return candidates
}

func buildConveneOverview(
	profile *domain.ConveneProfile,
	pulls []domain.ConvenePull,
	pools []domain.ConvenePoolDefinition,
) domain.ConveneOverview {
	if pulls == nil {
		pulls = make([]domain.ConvenePull, 0)
	}
	if pools == nil {
		pools = make([]domain.ConvenePoolDefinition, 0)
	}
	overview := domain.ConveneOverview{
		Profile: profile,
		Pools:   make([]domain.ConvenePoolSummary, 0, len(pools)),
		Pulls:   pulls,
	}
	if profile != nil {
		overview.LastImportedAt = profile.LastImportedAt
	}
	byPool := map[int][]domain.ConvenePull{}
	poolByType := make(map[int]domain.ConvenePoolDefinition, len(pools))
	for _, pool := range pools {
		poolByType[pool.PoolType] = pool
	}
	for index := range overview.Pulls {
		pool, ok := poolByType[overview.Pulls[index].PoolType]
		if ok {
			overview.Pulls[index].PoolName = pool.Name
		}
		pull := overview.Pulls[index]
		byPool[pull.PoolType] = append(byPool[pull.PoolType], pull)
		overview.Total++
		switch pull.Rarity {
		case 5:
			overview.Count5++
		case 4:
			overview.Count4++
		case 3:
			overview.Count3++
		}
	}
	for _, pool := range pools {
		overview.Pools = append(overview.Pools, summarizeConvenePool(pool, byPool[pool.PoolType]))
	}
	return overview
}

func summarizeConvenePool(
	pool domain.ConvenePoolDefinition,
	pulls []domain.ConvenePull,
) domain.ConvenePoolSummary {
	summary := domain.ConvenePoolSummary{
		PoolType:       pool.PoolType,
		Name:           pool.Name,
		ShortName:      pool.ShortName,
		Kind:           pool.Kind,
		HardPity:       pool.HardPity,
		GuaranteeState: "not_applicable",
		RecentFiveStar: make([]domain.ConvenePull, 0, 5),
	}
	for _, pull := range pulls {
		switch pull.Rarity {
		case 5:
			summary.Count5++
			if len(summary.RecentFiveStar) < 5 {
				summary.RecentFiveStar = append(summary.RecentFiveStar, pull)
			}
		case 4:
			summary.Count4++
		case 3:
			summary.Count3++
		}
	}
	summary.Total = len(pulls)

	chronological := append([]domain.ConvenePull(nil), pulls...)
	sort.SliceStable(chronological, func(i, j int) bool {
		if chronological[i].ObtainedAt == chronological[j].ObtainedAt {
			return chronological[i].SourceIndex > chronological[j].SourceIndex
		}
		return chronological[i].ObtainedAt < chronological[j].ObtainedAt
	})
	pity5, pity4 := 0, 0
	seenFive := false
	intervals := make([]int, 0)
	var latestFive *domain.ConvenePull
	for index := range chronological {
		pity5++
		pity4++
		if chronological[index].Rarity == 5 {
			if seenFive {
				intervals = append(intervals, pity5)
			}
			seenFive = true
			pity5 = 0
			pity4 = 0
			item := chronological[index]
			latestFive = &item
		} else if chronological[index].Rarity == 4 {
			pity4 = 0
		}
	}
	summary.CurrentPity = pity5
	summary.CurrentPity4 = pity4
	summary.HistoryPartial = !seenFive
	if len(intervals) > 0 {
		total := 0
		for _, value := range intervals {
			total += value
		}
		summary.AveragePity5 = float64(total) / float64(len(intervals))
	}
	if pool.PoolType == 1 {
		summary.GuaranteeState = "unknown"
		if latestFive != nil {
			if permanentFiveStars[strings.ToLower(latestFive.ItemName)] {
				summary.GuaranteeState = "guaranteed"
			} else {
				summary.GuaranteeState = "not_guaranteed"
			}
		}
	}
	return summary
}

func fallbackConvenePoolDefinitions() []domain.ConvenePoolDefinition {
	return poolDefinitions(convenesource.Pools)
}

func poolDefinitions(pools []convenesource.Pool) []domain.ConvenePoolDefinition {
	result := make([]domain.ConvenePoolDefinition, 0, len(pools))
	for _, pool := range pools {
		result = append(result, pool.Domain())
	}
	return result
}

var permanentFiveStars = map[string]bool{
	"calcharo": true,
	"encore":   true,
	"jianxin":  true,
	"lingyang": true,
	"verina":   true,
}
