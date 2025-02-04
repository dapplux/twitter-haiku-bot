package ai

type TextProcessor interface {
	GenerateSummary(text string) (string, error)
	GenerateHaiku(summary string) (string, error)
}
