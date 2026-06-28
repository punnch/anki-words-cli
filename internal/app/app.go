// Package app wires storage, terminal input, OpenAI, AnkiConnect, and
// validation into the local TUI workflow.
package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/punnch/ankiwords/internal/anki"
	"github.com/punnch/ankiwords/internal/config"
	"github.com/punnch/ankiwords/internal/logger"
	"github.com/punnch/ankiwords/internal/model"
	"github.com/punnch/ankiwords/internal/openai"
	"github.com/punnch/ankiwords/internal/repository"
	"github.com/punnch/ankiwords/internal/validation"
	"go.uber.org/zap"
)

const localUserID = "1"

var errInputInterrupted = errors.New("input interrupted")

// App owns the runtime dependencies and local terminal workflow.
type App struct {
	cfg    config.Config
	store  repository.Repository
	anki   *anki.Client
	openai *openai.Client
	logger *logger.Logger
	in     *bufio.Reader
	out    io.Writer
	done   bool
}

// New constructs the local application.
func New(
	cfg config.Config,
	st repository.Repository,
	logger *logger.Logger,
) (*App, error) {
	return &App{
		cfg:    cfg,
		store:  st,
		anki:   anki.NewClient(cfg.AnkiConnectURL),
		openai: openai.NewClient(cfg.OpenAIAPIKey, cfg.OpenAIBaseURL, cfg.OpenAIModel),
		logger: logger,
		in:     bufio.NewReader(os.Stdin),
		out:    os.Stdout,
	}, nil
}

// Run starts the local terminal UI.
func (a *App) Run(ctx context.Context) error {
	a.logger.Info("tui starting")

	for ctx.Err() == nil && !a.done {
		prefs, err := a.ensureUser(ctx)
		if err != nil {
			a.logger.Error("user load error", zap.Error(err))
			return err
		}

		a.clear()
		a.printf("Anki Words\n")
		a.printf("==========\n\n")
		a.printf("Deck:  %s\n", prefs.ActiveDeck)
		a.printf("Model: %s\n\n", prefs.PreferredModel)
		a.printf("1. Create cards\n")
		a.printf("2. List decks\n")
		a.printf("3. Change deck\n")
		a.printf("4. List models\n")
		a.printf("5. Change model\n")
		a.printf("6. Show generation preset\n")
		a.printf("7. Edit generation preset\n")
		a.printf("8. Quit\n\n")

		choice, err := a.prompt("Choose: ")
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, errInputInterrupted) {
				return nil
			}

			return err
		}

		if ctx.Err() != nil {
			return nil
		}

		switch strings.TrimSpace(choice) {
		case "1":
			a.createCards(ctx)
		case "2":
			a.listDecks(ctx)
		case "3":
			a.changeDeck(ctx)
		case "4":
			a.listModels(ctx)
		case "5":
			a.changeModel(ctx)
		case "6":
			a.showFormat(ctx)
		case "7":
			a.editFormat(ctx)
		case "8", "q", "quit", "exit":
			return nil
		default:
			a.message("Unknown menu option.")
		}
	}

	return ctx.Err()
}

func (a *App) createCards(ctx context.Context) {
	wordsText, err := a.prompt("Words, comma-separated: ")
	if err != nil {
		if errors.Is(err, errInputInterrupted) {
			return
		}

		a.message("Could not read words.")

		return
	}

	words := parseWordList(wordsText)
	a.logger.Debug("parsed words", zap.Strings("words", words))
	if len(words) == 0 {
		a.message("Enter at least one word.")
		return
	}

	result := a.createCardsFromWords(ctx, words)
	a.message(result)
}

func (a *App) listDecks(ctx context.Context) {
	decks, err := a.anki.DeckNames(ctx)
	if err != nil {
		a.logger.Error("anki deckNames error", zap.Error(err))
		a.message("Could not read decks from AnkiConnect.")
		return
	}

	a.message(strings.Join(decks, "\n"))
}

func (a *App) changeDeck(ctx context.Context) {
	deck, err := a.prompt("Deck name: ")
	if err != nil {
		if errors.Is(err, errInputInterrupted) {
			return
		}

		a.message("Could not read deck name.")

		return
	}

	result := a.setDeck(ctx, strings.TrimSpace(deck))
	a.message(result)
}

func (a *App) listModels(ctx context.Context) {
	models, err := a.anki.ModelNames(ctx)
	if err != nil {
		a.logger.Error("anki modelNames error", zap.Error(err))
		a.message("Could not read models from AnkiConnect.")
		return
	}

	a.message(strings.Join(models, "\n"))
}

func (a *App) changeModel(ctx context.Context) {
	modelName, err := a.prompt("Model name: ")
	if err != nil {
		if errors.Is(err, errInputInterrupted) {
			return
		}

		a.message("Could not read model name.")

		return
	}

	result := a.setModel(ctx, strings.TrimSpace(modelName))
	a.message(result)
}

func (a *App) showFormat(ctx context.Context) {
	prefs, err := a.ensureUser(ctx)
	if err != nil {
		a.logger.Error("user load error", zap.Error(err))
		a.message("Could not load settings.")
		return
	}

	a.message(prefs.GenerationTemplate)
}

func (a *App) editFormat(ctx context.Context) {
	a.printf("\nEnter generation preset. Finish with a single '.' on its own line.\n\n")

	var lines []string
	for {
		line, err := a.prompt("> ")
		if err != nil {
			if errors.Is(err, errInputInterrupted) {
				return
			}

			a.message("Could not read generation preset.")

			return
		}

		if strings.TrimSpace(line) == "." {
			break
		}

		lines = append(lines, line)
	}

	template := strings.TrimSpace(strings.Join(lines, "\n"))
	if template == "" {
		a.message("Generation preset was not changed.")
		return
	}

	result := a.setFormat(ctx, template)
	a.message(result)
}

func (a *App) setDeck(ctx context.Context, deck string) string {
	prefs, err := a.ensureUser(ctx)
	if err != nil {
		a.logger.Error("user load error", zap.Error(err))
		return "Could not load settings."
	}

	if deck == "" {
		return "Deck name is required."
	}

	decks, err := a.anki.DeckNames(ctx)
	if err != nil {
		a.logger.Error("anki deckNames error", zap.Error(err))
		return "Could not read decks from AnkiConnect."
	}

	if !containsString(decks, deck) {
		return "Deck does not exist. Use List decks to see available decks."
	}

	prefs.ActiveDeck = deck
	if err := a.store.UpsertUser(prefs); err != nil {
		a.logger.Error("store update error", zap.Error(err))
		return "Could not save deck selection."
	}

	return "Current deck changed to:\n" + deck
}

func (a *App) setModel(ctx context.Context, modelName string) string {
	prefs, err := a.ensureUser(ctx)
	if err != nil {
		a.logger.Error("user load error", zap.Error(err))
		return "Could not load settings."
	}

	if modelName == "" {
		return "Model name is required."
	}

	models, err := a.anki.ModelNames(ctx)
	if err != nil {
		a.logger.Error("anki modelNames error", zap.Error(err))
		return "Could not read models from AnkiConnect."
	}

	if !containsString(models, modelName) {
		return "Model does not exist. Use List models to see available models."
	}

	modelFields, err := a.anki.ModelFieldNames(ctx, modelName)
	if err != nil {
		a.logger.Error("anki modelFieldNames error", zap.Error(err))
		return "Could not read Anki model."
	}

	if len(modelFields) == 0 {
		return "Anki model has no fields."
	}

	prefs.PreferredModel = modelName
	if err := a.store.UpsertUser(prefs); err != nil {
		a.logger.Error("store update error", zap.Error(err))
		return "Could not save model selection."
	}

	return "Current model changed to:\n" + modelName
}

func (a *App) setFormat(ctx context.Context, template string) string {
	prefs, err := a.ensureUser(ctx)
	if err != nil {
		a.logger.Error("user load error", zap.Error(err))
		return "Could not load settings."
	}

	if strings.TrimSpace(template) == "" {
		return "Generation preset is required."
	}

	prefs.GenerationTemplate = template
	if err := a.store.UpsertUser(prefs); err != nil {
		a.logger.Error("store update error", zap.Error(err))
		return "Could not save generation preset."
	}

	return "Generation preset updated."
}

func (a *App) createCardsFromWords(ctx context.Context, words []string) string {
	prefs, err := a.ensureUser(ctx)
	if err != nil {
		a.logger.Error("user load error", zap.Error(err))
		return "Could not load settings."
	}

	modelFields, err := a.anki.ModelFieldNames(ctx, prefs.PreferredModel)
	if err != nil {
		a.logger.Error("anki modelFieldNames error", zap.Error(err))
		return "Could not read Anki model."
	}

	if len(modelFields) == 0 {
		return "Anki model has no fields."
	}

	formatter := model.SentenceFormatter{Color: a.cfg.SentenceHighlightColor}
	a.logger.Debug(
		"openai request",
		zap.Strings("words", words),
		zap.String("active_deck", prefs.ActiveDeck),
		zap.String("model", prefs.PreferredModel),
		zap.String("openai_model", a.cfg.OpenAIModel),
	)

	cards, err := a.openai.GenerateCards(ctx, prefs.GenerationTemplate, words, prefs.PreferredModel, modelFields, formatter)
	if err != nil {
		a.logger.Error("openai error", zap.Error(err))
		return "Card generation failed. Please try again."
	}

	a.logger.Debug("openai response", zap.Any("cards", cards))

	if err := validation.ValidateGeneratedCards(cards, words, modelFields, formatter); err != nil {
		a.logger.Error("validation failure", zap.Error(err))
		return "Card generation failed. Please try again."
	}

	created := 0
	for _, card := range cards {
		note := anki.Note{
			DeckName:  prefs.ActiveDeck,
			ModelName: prefs.PreferredModel,
			Fields:    card.Fields,
			Tags:      []string{"ai-generated"},
		}

		a.logger.Debug("ankiconnect request", zap.Any("note", note))

		noteID, err := a.anki.AddNote(ctx, note)
		if err != nil {
			a.logger.Error("ankiconnect error", zap.Error(err))
			return "Could not create cards in Anki."
		}

		a.logger.Debug("ankiconnect response", zap.Int64("note_id", noteID))
		created++
	}

	return fmt.Sprintf("Created %d cards in:\n%s", created, prefs.ActiveDeck)
}

// ensureUser loads local preferences or creates initial values from Anki on first run.
func (a *App) ensureUser(ctx context.Context) (model.UserPrefs, error) {
	if prefs, ok, err := a.store.GetUser(localUserID); err != nil {
		return model.UserPrefs{}, err
	} else if ok {
		return prefs, nil
	}

	decks, err := a.anki.DeckNames(ctx)
	if err != nil {
		return model.UserPrefs{}, fmt.Errorf("load decks from AnkiConnect: %w", err)
	}
	if len(decks) == 0 {
		return model.UserPrefs{}, fmt.Errorf("no Anki decks available")
	}

	models, err := a.anki.ModelNames(ctx)
	if err != nil {
		return model.UserPrefs{}, fmt.Errorf("load models from AnkiConnect: %w", err)
	}
	if len(models) == 0 {
		return model.UserPrefs{}, fmt.Errorf("no Anki note models available")
	}

	prefs := model.UserPrefs{
		UserID:             localUserID,
		ActiveDeck:         decks[0],
		GenerationTemplate: a.cfg.DefaultGenerationPrompt,
		PreferredModel:     models[0],
	}

	if err := a.store.UpsertUser(prefs); err != nil {
		return model.UserPrefs{}, err
	}

	return prefs, nil
}

func (a *App) prompt(label string) (string, error) {
	a.printf("%s", label)

	text, err := a.in.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	if strings.Contains(text, "\x03") {
		a.done = true
		return "", errInputInterrupted
	}

	return strings.TrimRight(text, "\r\n"), err
}

func (a *App) message(text string) {
	if a.done {
		return
	}

	a.printf("\n%s\n\n", text)
	_, _ = a.prompt("Press Enter to continue...")
}

func (a *App) clear() {
	a.printf("\033[2J\033[H")
}

func (a *App) printf(format string, args ...any) {
	fmt.Fprintf(a.out, format, args...)
}

// parseWordList splits a comma-separated command argument into target words.
func parseWordList(args string) []string {
	raw := strings.Split(args, ",")
	out := make([]string, 0, len(raw))

	for _, part := range raw {
		word := strings.TrimSpace(part)
		if word != "" {
			out = append(out, word)
		}
	}

	return out
}

// containsString reports whether values contains target.
func containsString(values []string, target string) bool {
	return slices.Contains(values, target)
}
