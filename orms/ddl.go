package orms

import (
	"fmt"
	"strings"
)

// ddlDoc is a CREATE TABLE statement split into its top-level item list and the
// text after the outermost ')'. The suffix is preserved verbatim because it
// carries driver-specific clauses (SQLite WITHOUT ROWID/STRICT, ClickHouse
// ENGINE/ORDER BY/SETTINGS).
type ddlDoc struct {
	items  []string
	suffix string
}

// parseCreateTable locates the outermost parenthesised body of a CREATE TABLE
// statement and splits it into top-level items. Quotes (' " ` [ ]) and doubled
// escapes are respected, so commas inside types, defaults or comments do not
// split an item.
func parseCreateTable(ddl string) (ddlDoc, error) {
	open := strings.IndexByte(ddl, '(')
	if open < 0 {
		return ddlDoc{}, fmt.Errorf("orms: CREATE TABLE statement has no column list")
	}

	depth := 1
	var quote byte
	closeAt := -1
scan:
	for i := open + 1; i < len(ddl); i++ {
		c := ddl[i]
		if quote != 0 {
			if c == quote {
				if i+1 < len(ddl) && ddl[i+1] == quote {
					i++
					continue
				}
				quote = 0
			}
			continue
		}
		switch c {
		case '\'', '"', '`':
			quote = c
		case '[':
			quote = ']'
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				closeAt = i
				break scan
			}
		}
	}
	if closeAt < 0 {
		return ddlDoc{}, fmt.Errorf("orms: unbalanced parentheses in CREATE TABLE statement")
	}
	return ddlDoc{
		items:  splitItems(ddl[open+1 : closeAt]),
		suffix: ddl[closeAt+1:],
	}, nil
}

// splitItems splits a CREATE TABLE body on top-level commas.
func splitItems(body string) []string {
	var items []string
	depth := 0
	var quote byte
	start := 0
	for i := 0; i < len(body); i++ {
		c := body[i]
		if quote != 0 {
			if c == quote {
				if i+1 < len(body) && body[i+1] == quote {
					i++
					continue
				}
				quote = 0
			}
			continue
		}
		switch c {
		case '\'', '"', '`':
			quote = c
		case '[':
			quote = ']'
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				items = appendItem(items, body[start:i])
				start = i + 1
			}
		}
	}
	return appendItem(items, body[start:])
}

func appendItem(items []string, item string) []string {
	if item = strings.TrimSpace(item); item != "" {
		items = append(items, item)
	}
	return items
}

// itemIdent extracts the leading identifier of a CREATE TABLE item. It accepts
// a quoted identifier (", `, [] with doubled escapes) or a bare
// [A-Za-z0-9_$]+ name and reports false for anything else (CONSTRAINT, PRIMARY
// KEY, KEY, ...).
func itemIdent(item string) (string, bool) {
	item = strings.TrimSpace(item)
	if item == "" {
		return "", false
	}
	switch item[0] {
	case '"', '`':
		return unquoteIdent(item, item[0])
	case '[':
		return unquoteIdent(item, ']')
	default:
		i := 0
		for i < len(item) && isBareIdentByte(item[i]) {
			i++
		}
		if i == 0 {
			return "", false
		}
		return item[:i], true
	}
}

func unquoteIdent(item string, closeCh byte) (string, bool) {
	var b strings.Builder
	for i := 1; i < len(item); i++ {
		c := item[i]
		if c == closeCh {
			if i+1 < len(item) && item[i+1] == closeCh {
				b.WriteByte(closeCh)
				i++
				continue
			}
			return b.String(), true
		}
		b.WriteByte(c)
	}
	return "", false
}

func isBareIdentByte(c byte) bool {
	return c == '_' || c == '$' ||
		(c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9')
}

// reorderItems rebuilds the item list: columns in the requested order first,
// then every non-column item (constraints, indexes) in its original relative
// order. Column items that the caller left out of order are dropped, which is
// how hard mode removes database-only columns during a rebuild.
//
// A bare identifier can collide with a keyword (MySQL's KEY index begins with
// the identifier "key"); a quoted item is preferred as the column definition
// and the bare one is kept as a non-column item.
func reorderItems(doc ddlDoc, columns, order []string) ddlDoc {
	isColumn := make(map[string]bool, len(columns))
	for _, name := range columns {
		isColumn[name] = true
	}

	columnItems := make(map[string]string, len(columns))
	var others []string
	for _, item := range doc.items {
		name, ok := itemIdent(item)
		if !ok || !isColumn[name] {
			others = append(others, item)
			continue
		}
		if _, seen := columnItems[name]; seen && !isQuotedItem(item) {
			others = append(others, item)
			continue
		}
		columnItems[name] = item
	}

	out := make([]string, 0, len(doc.items))
	used := make(map[string]bool, len(order))
	for _, name := range order {
		item, ok := columnItems[name]
		if !ok {
			continue
		}
		out = append(out, item)
		used[name] = true
	}
	out = append(out, others...)
	return ddlDoc{items: out, suffix: doc.suffix}
}

// isQuotedItem reports whether a CREATE TABLE item starts with a quoted
// identifier.
func isQuotedItem(item string) bool {
	item = strings.TrimSpace(item)
	if item == "" {
		return false
	}
	switch item[0] {
	case '"', '`', '[':
		return true
	}
	return false
}
