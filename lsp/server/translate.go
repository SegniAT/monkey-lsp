package server

import (
	"github.com/SegniAT/monkey-lsp/analysis"
	"github.com/SegniAT/monkey-lsp/lsp/protocol"
)

func toProtocolHover(h *analysis.Hover) *protocol.Hover {
	if h == nil {
		return nil
	}

	var r *protocol.Range
	if h.Range != nil {
		r = &protocol.Range{
			Start: toProtocolPosition(h.Range.Start),
			End:   toProtocolPosition(h.Range.End),
		}
	}

	return &protocol.Hover{
		Contents: toProtocolMarkupContent(h.Contents),
		Range:    r,
	}
}

func toProtocolLocation(l *analysis.Location) *protocol.Location {
	if l == nil {
		return nil
	}

	return &protocol.Location{
		URI: l.URI,
		Range: protocol.Range{
			Start: toProtocolPosition(l.Range.Start),
			End:   toProtocolPosition(l.Range.End),
		},
	}
}

func toProtocolCompletionItems(items []analysis.CompletionItem) []protocol.CompletionItem {
	result := make([]protocol.CompletionItem, 0, len(items))
	for _, item := range items {
		result = append(result, protocol.CompletionItem{
			Label:         item.Label,
			Detail:        item.Detail,
			Documentation: toProtocolMarkupContent(item.Documentation),
		})
	}
	return result
}

func toProtocolPosition(p analysis.Position) protocol.Position {
	return protocol.Position{
		Line:      p.Line - 1,
		Character: p.Character - 1,
	}
}

func toProtocolMarkupContent(c analysis.MarkupContent) protocol.MarkupContent {
	return protocol.MarkupContent{
		Kind:  protocol.MarkupKind(c.Kind),
		Value: c.Value,
	}
}
