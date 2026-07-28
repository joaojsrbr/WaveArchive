package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"wavearchive/internal/domain"
)

type AssistantService struct {
	analyzer   *AIAnalyzer
	history    domain.AIHistoryRepository
	evaluator  *BuildEvaluator
	characters domain.CharacterRepository
	guides     domain.GuideRepository
}

func NewAssistantService(analyzer *AIAnalyzer, history domain.AIHistoryRepository, evaluator *BuildEvaluator, characters domain.CharacterRepository, guides domain.GuideRepository) *AssistantService {
	return &AssistantService{analyzer: analyzer, history: history, evaluator: evaluator, characters: characters, guides: guides}
}

func (s *AssistantService) List(ctx context.Context) ([]domain.AIConversation, error) {
	return s.history.List(ctx)
}

func (s *AssistantService) Delete(ctx context.Context, id int64) error {
	return s.history.Delete(ctx, id)
}

func (s *AssistantService) Chat(ctx context.Context, request domain.AssistantRequest) (domain.AIConversation, error) {
	return s.chat(ctx, request, nil)
}

func (s *AssistantService) ChatStream(ctx context.Context, request domain.AssistantRequest, emit func(string)) (domain.AIConversation, error) {
	return s.chat(ctx, request, emit)
}

func (s *AssistantService) chat(ctx context.Context, request domain.AssistantRequest, emit func(string)) (domain.AIConversation, error) {
	request.Question = strings.TrimSpace(request.Question)
	if request.Question == "" {
		return domain.AIConversation{}, errors.New("question is required")
	}
	var conversation domain.AIConversation
	var err error
	if request.ConversationID > 0 {
		conversation, err = s.history.Get(ctx, request.ConversationID)
		if err != nil {
			return conversation, err
		}
	} else {
		title := request.Question
		if len([]rune(title)) > 60 {
			title = string([]rune(title)[:60]) + "…"
		}
		conversation, err = s.history.Create(ctx, domain.AIConversation{
			Title: title, ContextType: request.ContextType, ContextID: request.ContextID,
			Provider: request.Provider, Model: request.Model,
		})
		if err != nil {
			return conversation, err
		}
	}
	if _, err := s.history.AddMessage(ctx, domain.AIMessage{
		ConversationID: conversation.ID, Role: "user", Content: request.Question,
	}); err != nil {
		return conversation, err
	}
	contextData, err := s.contextData(ctx, conversation.ContextType, conversation.ContextID)
	if err != nil {
		return conversation, err
	}
	conversation, err = s.history.Get(ctx, conversation.ID)
	if err != nil {
		return conversation, err
	}
	sources := []domain.KnowledgeSource{}
	if s.guides != nil {
		sources, _ = s.guides.Search(ctx, request.Question, 8)
	}
	payload, _ := json.Marshal(map[string]any{
		"context": contextData,
		"history": conversation.Messages,
		"sources": sources,
	})
	analysisRequest := domain.AIAnalysisRequest{
		Provider: request.Provider, Endpoint: request.Endpoint, Model: request.Model,
		APIKey: request.APIKey, Mode: request.Mode,
		Context:  "Responda à última pergunta do usuário usando o contexto local e o histórico.",
		DataJSON: string(payload),
	}
	var analysis domain.AIAnalysisResult
	if emit != nil {
		analysis, err = s.analyzer.Stream(ctx, analysisRequest, emit)
	} else {
		analysis, err = s.analyzer.Analyze(ctx, analysisRequest)
	}
	if err != nil {
		return conversation, err
	}
	if _, err := s.history.AddMessage(ctx, domain.AIMessage{
		ConversationID: conversation.ID, Role: "assistant", Content: analysis.Text,
	}); err != nil {
		return conversation, err
	}
	final, err := s.history.Get(ctx, conversation.ID)
	if err == nil {
		final.Sources = sources
	}
	return final, err
}

func (s *AssistantService) contextData(ctx context.Context, contextType string, contextID *int64) (any, error) {
	switch strings.ToLower(contextType) {
	case "build":
		if contextID == nil {
			return nil, errors.New("build context requires an id")
		}
		return s.evaluator.Evaluate(ctx, *contextID, nil)
	case "team":
		if contextID == nil {
			return nil, errors.New("team context requires an id")
		}
		return s.evaluator.TeamTheorycraft(ctx, *contextID)
	case "rotation":
		if contextID == nil {
			return nil, errors.New("rotation context requires an id")
		}
		return s.evaluator.EvaluateRotation(ctx, *contextID)
	case "character":
		if s.characters == nil {
			return nil, errors.New("character context is unavailable")
		}
		if contextID == nil {
			return nil, errors.New("character context requires an id")
		}
		profile, err := s.characters.GetProfile(ctx, *contextID)
		if err != nil {
			return nil, err
		}
		guides := []domain.CharacterGuide{}
		if s.guides != nil {
			guides, _ = s.guides.List(ctx, *contextID)
		}
		return map[string]any{"profile": profile, "officialGuides": guides}, nil
	default:
		return map[string]any{
			"scope": "general",
			"note":  "Nenhum contexto específico foi selecionado. Não invente dados da conta.",
		}, nil
	}
}
