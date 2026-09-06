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

type CompletionItemKind uint8

const (
	Function CompletionItemKind = 3
	Variable CompletionItemKind = 6
	Keyword  CompletionItemKind = 14
)

func (cik CompletionItemKind) String() string {
	switch cik {
	case Function:
		return "Function"
	case Variable:
		return "Variable"
	case Keyword:
		return "Keyword"
	default:
		return "Unknown kind"
	}
}

type CompletionItem struct {
	Label         string
	Detail        string
	Kind          CompletionItemKind
	Documentation MarkupContent
}
