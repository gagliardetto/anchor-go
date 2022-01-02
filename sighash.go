package main

import (
	"strings"
	"unicode"

	. "github.com/gagliardetto/utilz"
)

func ToSnakeForSighash(s string) string {
	return ToRustSnakeCase(s)
}

type reader struct {
	runes []rune
	index int
}

func newReader(s string) *reader {
	return &reader{
		runes: SplitStringByRune(s),
		index: -1,
	}
}

func (r reader) This() (int, rune) {
	return r.index, r.runes[r.index]
}

func (r reader) HasNext() bool {
	return r.index < len(r.runes)-1
}

func (r reader) Peek() (int, rune) {
	if r.HasNext() {
		return r.index + 1, r.runes[r.index+1]
	}
	return -1, rune(0)
}

func (r *reader) Move() bool {
	if r.HasNext() {
		r.index++
		return true
	}
	return false
}

// #[cfg(feature = "unicode")]
// fn get_iterator(s: &str) -> unicode_segmentation::UnicodeWords {
//     use unicode_segmentation::UnicodeSegmentation;
//     s.unicode_words()
// }
func splitByUnicode(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		// TODO: see https://unicode.org/reports/tr29/#Word_Boundaries
		return !(unicode.IsLetter(r) || unicode.IsDigit(r)) || unicode.Is(unicode.Extender, r)
	})
	return parts
}

// #[cfg(not(feature = "unicode"))]
func splitIntoWords(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	})
	return parts
}

type WordMode int

const (
	/// There have been no lowercase or uppercase characters in the current
	/// word.
	Boundary WordMode = iota
	/// The previous cased character in the current word is lowercase.
	Lowercase
	/// The previous cased character in the current word is uppercase.
	Uppercase
)

// ToRustSnakeCase converts the given string to a snake_case string.
// Ported from https://github.com/withoutboats/heck/blob/c501fc95db91ce20eaef248a511caec7142208b4/src/lib.rs#L75
func ToRustSnakeCase(s string) string {

	builder := new(strings.Builder)

	first_word := true
	words := splitIntoWords(s)
	for _, word := range words {
		char_indices := newReader(word)
		init := 0
		mode := Boundary

		for char_indices.Move() {
			i, c := char_indices.This()

			// Skip underscore characters
			if c == '_' {
				if init == i {
					init += 1
				}
				continue
			}

			if next_i, next := char_indices.Peek(); next_i != -1 {

				// The mode including the current character, assuming the
				// current character does not result in a word boundary.
				next_mode := func() WordMode {
					if unicode.IsLower(c) {
						return Lowercase
					} else if unicode.IsUpper(c) {
						return Uppercase
					} else {
						return mode
					}
				}()

				// Word boundary after if next is underscore or current is
				// not uppercase and next is uppercase
				if next == '_' || (next_mode == Lowercase && unicode.IsUpper(next)) {
					if !first_word {
						// TODO:
						// boundary(f)?;
						builder.WriteString(strings.ToLower("_"))
					}
					{ // TODO:
						// with_word(&word[init..next_i], f)?;
						builder.WriteString(strings.ToLower(word[init:next_i]))
					}

					first_word = false
					init = next_i
					mode = Boundary

					// Otherwise if current and previous are uppercase and next
					// is lowercase, word boundary before
				} else if mode == Uppercase && unicode.IsUpper(c) && unicode.IsLower(next) {
					if !first_word {
						// TODO:
						// boundary(f)?;
						builder.WriteString(strings.ToLower("_"))
					} else {
						first_word = false
					}
					{
						// TODO:
						// with_word(&word[init..i], f)?;
						builder.WriteString(strings.ToLower(word[init:i]))
					}
					init = i
					mode = Boundary

					// Otherwise no word boundary, just update the mode
				} else {
					mode = next_mode
				}

			} else {
				// Collect trailing characters as a word
				if !first_word {
					// TODO:
					// boundary(f)?;
					builder.WriteString(strings.ToLower("_"))
				} else {
					first_word = false
				}
				{
					// TODO:
					// with_word(&word[init..], f)?;
					builder.WriteString(strings.ToLower(word[init:]))
				}
				break
			}
		}
	}

	return builder.String()
}
