package analysis

type MarkupKind string

const (
	MarkupPlainText MarkupKind = "plaintext"
	MarkupMarkdown  MarkupKind = "markdown"
)

type Position struct {
	Line      uint
	Character uint
}

type Range struct {
	Start Position
	End   Position
}

type MarkupContent struct {
	Kind  MarkupKind
	Value string
}

type Hover struct {
	Contents MarkupContent
	Range    *Range
}

type Location struct {
	URI   string
	Range Range
}

type CompletionItem struct {
	Label         string
	Detail        string
	Documentation MarkupContent
}
