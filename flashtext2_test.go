package flashtext2

import (
	"testing"
)

func assert(b bool) {
	if !b {
		panic("Assertion error")
	}
}

func sliceEq[T comparable](x []T, y []T) bool {
	if len(x) != len(y) {
		return false
	}
	n := len(x)
	for i := 0; i < n; i++ {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

func TestKeywordProcessor_Len(t *testing.T) {
	kp := NewKeywordProcessor()
	assert(kp.len == 0)

	kp.AddKeyword("hello")
	assert(kp.len == 1)

	// it shouldnt increase the count if the word is already present
	kp.AddKeyword("hello")
	assert(kp.len == 1)

	kp.AddKeywordWithCleanWord("hello", "new clean word")
	assert(kp.len == 1)

	kp.AddKeywordWithCleanWord("hello", "overwriting the clean word again ...")
	assert(kp.len == 1)

	kp.AddKeyword("HELLO")
	assert(kp.len == 2)

	kp.AddKeyword("hello, hey!")
	assert(kp.len == 3)

	kp.AddKeyword("hello, ")
	assert(kp.len == 4)
}

func TestKeywordProcessor_AddKeyword(t *testing.T) {
	kp := NewKeywordProcessor()
	assert(kp.len == 0)

	kp.AddKeyword("py")
	out := kp.ExtractKeywordsAsSlice("py")
	assert(sliceEq(out, []string{"py"}))

	kp.AddKeywordWithCleanWord("py", "Python")
	out = kp.ExtractKeywordsAsSlice("py")
	assert(sliceEq(out, []string{"Python"}))
}

func TestKeywordProcessor_AddKeywordWithCleanWord(t *testing.T) {
}

func TestKeywordProcessor_ContainsWord(t *testing.T) {
	kp := NewKeywordProcessor()

	assert(!kp.ContainsWord("hello"))

	kp.AddKeyword("hello")
	assert(kp.ContainsWord("hello"))

	kp.AddKeyword("hello world!")
	assert(kp.ContainsWord("hello world!"))

	assert(!kp.ContainsWord("hello world"))

	kp.AddKeyword("hello world!")
	assert(kp.ContainsWord("hello world!"))
}

func TestKeywordProcessor_FirstKeyword(t *testing.T) {
	kp := NewKeywordProcessor()
	kp.AddKeyword("hello")
	kp.AddKeyword("hello hello")
	kp.AddKeyword(" world")

	isMatch, longestMatch, remainingText, state := kp.FirstKeyword("hello hello world", -1)
	assert(isMatch)
	assert(longestMatch == "hello hello")
	assert(remainingText == " world")

	isMatch, longestMatch, remainingText, state = kp.FirstKeyword(remainingText, state)
	assert(isMatch)
	assert(longestMatch == " world")
	assert(remainingText == "")

	isMatch, _, remainingText, _ = kp.FirstKeyword(remainingText, state)
	assert(!isMatch)
	assert(remainingText == "")

	isMatch, _, remainingText, _ = kp.FirstKeyword(remainingText, state)
	assert(!isMatch)
	assert(remainingText == "")
}

func TestKeywordProcessor_ExtractKeywordsAsSlice(t *testing.T) {
	var out []string
	kp := NewKeywordProcessor()

	kp.AddKeyword("hello")

	out = kp.ExtractKeywordsAsSlice("")
	assert(len(out) == 0)

	out = kp.ExtractKeywordsAsSlice(" ")
	assert(len(out) == 0)

	out = kp.ExtractKeywordsAsSlice("hell")
	assert(len(out) == 0)

	out = kp.ExtractKeywordsAsSlice("hello")
	assert(sliceEq(out, []string{"hello"}))

	out = kp.ExtractKeywordsAsSlice(" hello")
	assert(sliceEq(out, []string{"hello"}))

	out = kp.ExtractKeywordsAsSlice("hello ")
	assert(sliceEq(out, []string{"hello"}))

	out = kp.ExtractKeywordsAsSlice(" hello ")
	assert(sliceEq(out, []string{"hello"}))

	out = kp.ExtractKeywordsAsSlice("hellohello")
	assert(sliceEq(out, []string{}))

	out = kp.ExtractKeywordsAsSlice("hello hello")
	assert(sliceEq(out, []string{"hello", "hello"}))

	kp.AddKeyword("Hello")

	out = kp.ExtractKeywordsAsSlice("helloHello")
	assert(sliceEq(out, []string{}))

	out = kp.ExtractKeywordsAsSlice("HelloHello")
	assert(sliceEq(out, []string{}))

	kp.AddKeyword("hello world")

	out = kp.ExtractKeywordsAsSlice("hello world")
	assert(sliceEq(out, []string{"hello world"}))

	out = kp.ExtractKeywordsAsSlice("hello worldhello world")
	assert(sliceEq(out, []string{"hello"}))

	out = kp.ExtractKeywordsAsSlice("hello world hello world")
	assert(sliceEq(out, []string{"hello world", "hello world"}))

	kp.AddKeyword("hey there")

	out = kp.ExtractKeywordsAsSlice("hey hey there")
	assert(sliceEq(out, []string{"hey there"}))

	// -----
	kp = NewKeywordProcessor()
	kp.AddKeyword("hello ")
	kp.AddKeyword("world")
	out = kp.ExtractKeywordsAsSlice("hello world")
	assert(sliceEq(out, []string{"hello ", "world"}))

	kp = NewKeywordProcessor()
	kp.AddKeyword("hello")
	kp.AddKeyword(" world")
	out = kp.ExtractKeywordsAsSlice("hello world")
	assert(sliceEq(out, []string{"hello", " world"}))

	// -----

	kp = NewKeywordProcessor()
	kp.AddKeyword("a b c d x")
	kp.AddKeyword("d e f")

	out = kp.ExtractKeywordsAsSlice("a b c d e f")
	assert(sliceEq(out, []string{"d e f"}))

	kp.AddKeyword(" d e f")

	out = kp.ExtractKeywordsAsSlice("a b c d e f")
	assert(sliceEq(out, []string{" d e f"}))
}

func TestKeywordProcessor_ReplaceKeywords(t *testing.T) {
	var replaceTestCases = [...]struct {
		words []string
		in    string
		out   string
	}{
		{[]string{}, "", ""},
		{[]string{"hello", "hello"}, "", ""},
		{[]string{"hello", "hello"}, "hey", "hey"},
		{[]string{"hello", "hello"}, "hello", "hello"},
		{[]string{"hello", "hello"}, "hello world", "hello world"},
		{[]string{"hello", "hey"}, "hello world", "hey world"},
		{[]string{"hello", "hey", "hey", "hello"}, "hey jack, hello sarah", "hello jack, hey sarah"},
		{[]string{"hello", "", "hey", ""}, "hey jack, hello sarah", " jack,  sarah"},
		{[]string{"a b c", "abc"}, "a b a b c a b", "a b abc a b"},
	}

	for idx, test := range replaceTestCases {
		assert(len(test.words)%2 == 0)
		kp := NewKeywordProcessor()
		for i := 0; i < len(test.words); i += 2 {
			kp.AddKeywordWithCleanWord(test.words[i], test.words[i+1])
		}

		if kp.ReplaceKeywords(test.in) != test.out {
			t.Fatalf("\nidx=(%v) \ninput=(%s) \nexpected-output=(%s) \nactual-output=(%s)", idx, test.in, test.out, kp.ReplaceKeywords(test.in))
		}
	}
}
