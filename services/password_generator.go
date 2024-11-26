package services

import (
	"crypto/rand"
	"math/big"
	"strings"
)

type PasswordGeneratorService interface {
	GeneratePassword(options *PasswordOptions) (string, error)
	GeneratePassphrase(options *PassphraseOptions) (string, error)
	ValidateGeneratedPassword(password string, options *PasswordOptions) bool
}

type PasswordOptions struct {
	Length           int
	IncludeUpper    bool
	IncludeLower    bool
	IncludeNumbers  bool
	IncludeSpecial  bool
	MinSpecial      int
	MinNumbers      int
	MinUpper        int
	ExcludeSimilar  bool
	ExcludeAmbiguous bool
}

type PassphraseOptions struct {
	WordCount    int
	Separator    string
	Capitalize   bool
	IncludeNumber bool
}

type passwordGeneratorService struct {
	wordList []string
}

const (
	upperChars    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowerChars    = "abcdefghijklmnopqrstuvwxyz"
	numberChars   = "0123456789"
	specialChars  = "!@#$%^&*"
	similarChars  = "iIlL1oO0"
	ambiguousChars = "{}[]()/\\'\"`~,;:.<>"
)

func NewPasswordGeneratorService() PasswordGeneratorService {
	return &passwordGeneratorService{
		wordList: loadWordList(),
	}
}

func (s *passwordGeneratorService) GeneratePassword(options *PasswordOptions) (string, error) {
	if options.Length < 1 {
		return "", fmt.Errorf("password length must be at least 1")
	}

	var chars string
	var requiredChars []rune

	if options.IncludeUpper {
		chars += upperChars
		for i := 0; i < options.MinUpper; i++ {
			c, err := s.randomChar(upperChars)
			if err != nil {
				return "", err
			}
			requiredChars = append(requiredChars, c)
		}
	}

	if options.IncludeLower {
		chars += lowerChars
	}

	if options.IncludeNumbers {
		chars += numberChars
		for i := 0; i < options.MinNumbers; i++ {
			c, err := s.randomChar(numberChars)
			if err != nil {
				return "", err
			}
			requiredChars = append(requiredChars, c)
		}
	}

	if options.IncludeSpecial {
		chars += specialChars
		for i := 0; i < options.MinSpecial; i++ {
			c, err := s.randomChar(specialChars)
			if err != nil {
				return "", err
			}
			requiredChars = append(requiredChars, c)
		}
	}

	if options.ExcludeSimilar {
		chars = removeChars(chars, similarChars)
	}

	if options.ExcludeAmbiguous {
		chars = removeChars(chars, ambiguousChars)
	}

	if len(chars) == 0 {
		return "", fmt.Errorf("no characters available with current options")
	}

	password := make([]rune, options.Length)

	// Place required chars at random positions
	positions := make([]int, options.Length)
	for i := range positions {
		positions[i] = i
	}
	s.shuffle(positions)

	for i, c := range requiredChars {
		if i >= len(positions) {
			break
		}
		password[positions[i]] = c
	}

	// Fill remaining positions with random chars
	for i := len(requiredChars); i < options.Length; i++ {
		c, err := s.randomChar(chars)
		if err != nil {
			return "", err
		}
		password[positions[i]] = c
	}

	return string(password), nil
}

func (s *passwordGeneratorService) GeneratePassphrase(options *PassphraseOptions) (string, error) {
	if options.WordCount < 1 {
		return "", fmt.Errorf("word count must be at least 1")
	}

	words := make([]string, options.WordCount)
	for i := 0; i < options.WordCount; i++ {
		idx, err := s.randomInt(len(s.wordList))
		if err != nil {
			return "", err
		}
		word := s.wordList[idx]
		if options.Capitalize {
			word = strings.Title(word)
		}
		words[i] = word
	}

	passphrase := strings.Join(words, options.Separator)

	if options.IncludeNumber {
		num, err := s.randomInt(100)
		if err != nil {
			return "", err
		}
		passphrase += options.Separator + fmt.Sprintf("%02d", num)
	}

	return passphrase, nil
}

func (s *passwordGeneratorService) ValidateGeneratedPassword(password string, options *PasswordOptions) bool {
	if len(password) != options.Length {
		return false
	}

	var upperCount, numberCount, specialCount int
	for _, c := range password {
		switch {
		case strings.ContainsRune(upperChars, c):
			upperCount++
		case strings.ContainsRune(numberChars, c):
			numberCount++
		case strings.ContainsRune(specialChars, c):
			specialCount++
		}
	}


	return upperCount >= options.MinUpper &&
		numberCount >= options.MinNumbers &&
		specialCount >= options.MinSpecial
}

func (s *passwordGeneratorService) randomChar(chars string) (rune, error) {
	idx, err := s.randomInt(len(chars))
	if err != nil {
		return 0, err
	}
	return rune(chars[idx]), nil
}

func (s *passwordGeneratorService) randomInt(max int) (int, error) {
	if max <= 0 {
		return 0, nil
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}

func (s *passwordGeneratorService) shuffle(slice []int) {
	for i := len(slice) - 1; i > 0; i-- {
		j, _ := s.randomInt(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

func removeChars(str, chars string) string {
	for _, c := range chars {
		str = strings.ReplaceAll(str, string(c), "")
	}
	return str
}

func loadWordList() []string {
	// This would typically load from a file or embedded resource
	return []string{
		"correct", "horse", "battery", "staple",
		"apple", "banana", "cherry", "dragon",
		"elephant", "falcon", "giraffe", "hammer",
		// Add more words as needed
	}
}
