package flashtext2

import (
	"github.com/rivo/uniseg"
	"strings"
)

type Node struct {
	endOfWord bool
	cleanWord string // cleanWord is used only if `endOfWord` is true, otherwise its a zero value
	children  map[string]*Node
}

type KeywordProcessor struct {
	trie *Node
	len  uint
}

func (kp *KeywordProcessor) Len() uint {
	return kp.len
}

func (kp *KeywordProcessor) AddKeyword(word string) {
	kp.AddKeywordWithCleanWord(word, word)
}

func (kp *KeywordProcessor) AddKeywordWithCleanWord(word, cleanWord string) {
	node := kp.trie
	text := word
	state := -1

	for len(text) != 0 {
		word, text, state = uniseg.FirstWordInString(text, state)
		nodeMaybeNil, exists := node.children[word]
		if !exists {
			child := &Node{
				endOfWord: false,
				cleanWord: "",
				children:  make(map[string]*Node),
			}
			node.children[word] = child
			node = child
		} else {
			node = nodeMaybeNil
		}
	}

	if !node.endOfWord {
		kp.len++
		node.endOfWord = true
	}
	node.cleanWord = cleanWord
}

func (kp *KeywordProcessor) FirstKeyword(text string, state int) (isMatch bool, longestMatch, remainingText string, newState int) {
	startingText := text // save the text at the beginning so we can rollback if necessary
	startingState := state

	node := kp.trie
	// we need to be able to distinguish from a match that is an empty string to a ...
	longestMatch = ""
	isMatch = false
	var textAfterKeyword string
	var stateAfterKeyword int
	var word string
	var childExists bool
	for len(text) != 0 {
		word, text, state = uniseg.FirstWordInString(text, state)
		node, childExists = node.children[word]
		if childExists {
			if node.endOfWord {
				isMatch = true
				longestMatch = node.cleanWord
				textAfterKeyword = text
				stateAfterKeyword = state
			}
		} else {
			if isMatch {
				return isMatch, longestMatch, textAfterKeyword, stateAfterKeyword
			} else {
				// re-start the search from the second word of the `startingText`
				_, text, state = uniseg.FirstWordInString(startingText, startingState)
				startingText = text // save the text at the beginning so we can rollback if necessary
				startingState = state
				node = kp.trie
				isMatch = false
			}
		}
	}
	return isMatch, longestMatch, text, state
}

func (kp *KeywordProcessor) ExtractKeywordsAsSlice(text string) []string {
	var keywordsFound []string
	for isMatch, longestMatch, text, state := kp.FirstKeyword(text, -1); isMatch; {
		keywordsFound = append(keywordsFound, longestMatch)
		isMatch, longestMatch, text, state = kp.FirstKeyword(text, state)
	}
	return keywordsFound
}

func (kp *KeywordProcessor) ReplaceKeywords(text string) string {
	var builder strings.Builder
	builder.Grow(len(text))

	state := -1
	startingText := text // save the text at the beginning so we can rollback if necessary
	startingState := -1
	textBeforeMatch := text
	lenTextAtStartTraversal := 0 // dummy value
	node := kp.trie
	isMatch := false // we need to be able to distinguish from a match that is an empty string to a ...

	var longestMatch string
	var word string
	var childExists bool
	for len(text) != 0 {
		prevText := text
		word, text, state = uniseg.FirstWordInString(text, state)
		node, childExists = node.children[word]
		if childExists {
			if lenTextAtStartTraversal == 0 {
				lenTextAtStartTraversal = len(prevText)
			}
			if node.endOfWord {
				isMatch = true
				longestMatch = node.cleanWord
			}
		} else {
			if isMatch {
				builder.WriteString(textBeforeMatch[:len(textBeforeMatch)-lenTextAtStartTraversal])
				builder.WriteString(longestMatch)
				textBeforeMatch = prevText
			} else {
				// reset the state and re-start the search from the second word of the `startingText`
				_, text, state = uniseg.FirstWordInString(startingText, startingState)
			}
			startingText = text
			startingState = state
			node = kp.trie
			isMatch = false
			lenTextAtStartTraversal = 0
		}
	}
	// TODO: if nothing was replaced in the input-string we can just return a pointer to the original string
	if isMatch {
		builder.WriteString(textBeforeMatch[:len(textBeforeMatch)-lenTextAtStartTraversal])
		builder.WriteString(longestMatch)
	} else {
		builder.WriteString(textBeforeMatch)
	}
	return builder.String()
}

func (kp *KeywordProcessor) ContainsWord(word string) bool {
	node := kp.trie
	text := word
	state := -1
	var exists bool
	for len(text) != 0 {
		word, text, state = uniseg.FirstWordInString(text, state)
		node, exists = node.children[word]
		if !exists {
			return false
		}
	}
	return node.endOfWord
}

func NewKeywordProcessor() KeywordProcessor {
	return KeywordProcessor{trie: &Node{endOfWord: false, children: make(map[string]*Node)}, len: 0}
}

// TODO: rename all instances of `Word`, `CleanWord`, and `KeyWord` to something more intuitive
