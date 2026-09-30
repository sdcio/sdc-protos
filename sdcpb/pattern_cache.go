package sdcpb

import (
	"fmt"
	"regexp"
	"sync"

	logf "github.com/sdcio/logger"
)

// compiledPatternEntry is the outcome of compiling a pattern text. The error
// is cached as well, so that an invalid pattern is neither recompiled nor
// logged again on every call.
type compiledPatternEntry struct {
	re  *regexp.Regexp
	err error
}

// patternCache maps pattern text to its compiled pattern.
//
// It is intentionally unbounded and never evicted: consumers walk whole
// configuration trees cyclically, so a capped cache would thrash. The key is
// pure pattern text, hence a stale entry can only cost memory, never
// correctness. Growth is limited to the number of distinct pattern texts seen
// across all loaded schemas.
var patternCache sync.Map // map[string]*compiledPatternEntry

// compiledPattern returns the compiled pattern for the given pattern text,
// compiling and caching it on first use.
func compiledPattern(patternText string) (*regexp.Regexp, error) {
	if v, ok := patternCache.Load(patternText); ok {
		e := v.(*compiledPatternEntry)
		return e.re, e.err
	}

	// The set of metacharacters is not the same between XML schema and perl/python/go REs.
	// XML: .\?*+{}()[] (https://www.w3.org/TR/xmlschema-2/#dt-metac)
	// Go:  \.+*?()|[]{}^$ (go/libexec/src/regexp/regexp.go:714)
	// we need therefore to escape some values
	// TODO check about '^'
	re, err := regexp.Compile(xmlRegexConvert(patternText))
	entry := &compiledPatternEntry{re: re}
	if err != nil {
		entry.err = fmt.Errorf("unable to compile regex: %w", err)
	}

	actual, loaded := patternCache.LoadOrStore(patternText, entry)
	if !loaded && err != nil {
		// only the goroutine that stored the entry logs, so once per pattern
		logf.DefaultLogger.Error(err, "unable to compile regex", "pattern", patternText)
	}
	e := actual.(*compiledPatternEntry)
	return e.re, e.err
}

// resetPatternCache empties the pattern cache. Used by tests only.
func resetPatternCache() {
	patternCache.Range(func(k, _ any) bool {
		patternCache.Delete(k)
		return true
	})
}
