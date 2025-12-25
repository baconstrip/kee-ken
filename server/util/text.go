package util

import "unicode"

// Check if the string contains no punctuation and only common scripts
func IsValidName(str string) bool {
	for _, r := range str {
		if unicode.IsSpace(r) {
			continue
		}

		// Check for punctuation marks
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			return false
		}

		// Check if the character belongs to a valid script (Latin, Greek, Cyrillic, or CJK)
		if !(unicode.IsLetter(r) || runeIsCJK(r) || runeIsEmoji(r) || unicode.IsDigit(r)) {
			return false
		}
	}
	return true
}

// Check if the rune is a CJK character (Chinese, Japanese, Korean)
func runeIsCJK(r rune) bool {
	// Unicode range for CJK characters (Chinese, Japanese, Korean)
	return (r >= 0x4E00 && r <= 0x9FFF) || // CJK Ideographs
		(r >= 0x3040 && r <= 0x309F) || // Hiragana (Japanese)
		(r >= 0x30A0 && r <= 0x30FF) || // Katakana (Japanese)
		(r >= 0xAC00 && r <= 0xD7AF) // Hangul (Korean)
}

// Check if the rune is an emoji character
func runeIsEmoji(r rune) bool {
	// Unicode ranges for emojis
	return (r >= 0x1F600 && r <= 0x1F64F) || // Emoticons
		(r >= 0x1F300 && r <= 0x1F5FF) || // Symbols and pictographs
		(r >= 0x1F680 && r <= 0x1F6FF) || // Transport and map symbols
		(r >= 0x1F700 && r <= 0x1F77F) || // Alchemical symbols
		(r >= 0x2600 && r <= 0x26FF) || // Miscellaneous symbols
		(r >= 0x2700 && r <= 0x27BF) || // Dingbats
		(r >= 0x2B50 && r <= 0x2B50) || // Star emoji
		(r >= 0x1F900 && r <= 0x1F9FF) // Supplemental symbols and pictographs
}
